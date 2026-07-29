package diag

import (
	"errors"
	"reflect"
	"testing"

	"github.com/kwrkb/rdp-host-info/internal/hostinfo"
	"github.com/kwrkb/rdp-host-info/internal/msgid"
)

func TestEditionSupportCheck(t *testing.T) {
	tests := []struct {
		name   string
		read   func() (hostinfo.Edition, error)
		status Status
	}{
		{
			name: "pro supports host",
			read: func() (hostinfo.Edition, error) {
				return hostinfo.Edition{DisplayName: "Windows 11 Pro", SupportsRDPHost: true}, nil
			},
			status: StatusOK,
		},
		{
			name: "home does not support host",
			read: func() (hostinfo.Edition, error) {
				return hostinfo.Edition{DisplayName: "Windows 11 Home", SupportsRDPHost: false}, nil
			},
			status: StatusNG,
		},
		{
			name:   "read error is unknown",
			read:   func() (hostinfo.Edition, error) { return hostinfo.Edition{}, errors.New("boom") },
			status: StatusUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := EditionSupportCheck{ReadEdition: tt.read}
			got := c.Run().Status
			if got != tt.status {
				t.Errorf("Status = %v, want %v", got, tt.status)
			}
		})
	}
}

func TestRDPEnabledCheck(t *testing.T) {
	tests := []struct {
		name   string
		read   func(string, string) (uint32, error)
		status Status
	}{
		{
			name:   "zero means enabled",
			read:   func(string, string) (uint32, error) { return 0, nil },
			status: StatusOK,
		},
		{
			name:   "nonzero means disabled",
			read:   func(string, string) (uint32, error) { return 1, nil },
			status: StatusNG,
		},
		{
			name:   "missing value is unknown",
			read:   func(string, string) (uint32, error) { return 0, errors.New("not found") },
			status: StatusUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := RDPEnabledCheck{ReadDWORD: tt.read}
			got := c.Run().Status
			if got != tt.status {
				t.Errorf("Status = %v, want %v", got, tt.status)
			}
		})
	}
}

func TestServiceRunningCheck(t *testing.T) {
	tests := []struct {
		name   string
		query  func(string) (bool, error)
		status Status
	}{
		{
			name:   "running",
			query:  func(string) (bool, error) { return true, nil },
			status: StatusOK,
		},
		{
			name:   "stopped",
			query:  func(string) (bool, error) { return false, nil },
			status: StatusNG,
		},
		{
			name:   "query error is unknown",
			query:  func(string) (bool, error) { return false, errors.New("access denied") },
			status: StatusUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := ServiceRunningCheck{ServiceName: "TermService", DisplayName: "Remote Desktop Services", QueryRunning: tt.query}
			got := c.Run().Status
			if got != tt.status {
				t.Errorf("Status = %v, want %v", got, tt.status)
			}
		})
	}
}

func TestPortListeningCheck(t *testing.T) {
	tests := []struct {
		name        string
		readPort    func() (uint32, bool)
		isListening func(uint32) (bool, error)
		status      Status
	}{
		{
			name:        "listening",
			readPort:    func() (uint32, bool) { return 3389, true },
			isListening: func(uint32) (bool, error) { return true, nil },
			status:      StatusOK,
		},
		{
			name:        "not listening",
			readPort:    func() (uint32, bool) { return 3389, true },
			isListening: func(uint32) (bool, error) { return false, nil },
			status:      StatusNG,
		},
		{
			name:        "check error is unknown",
			readPort:    func() (uint32, bool) { return 3389, false },
			isListening: func(uint32) (bool, error) { return false, errors.New("access denied") },
			status:      StatusUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := PortListeningCheck{ReadPort: tt.readPort, IsListening: tt.isListening}
			got := c.Run().Status
			if got != tt.status {
				t.Errorf("Status = %v, want %v", got, tt.status)
			}
		})
	}
}

func TestFirewallCheck(t *testing.T) {
	defaultPort := func() (uint32, bool) { return 3389, true }
	query := func(active uint32, enabled, viaFallback bool, err error) func(uint32) (uint32, bool, bool, error) {
		return func(uint32) (uint32, bool, bool, error) { return active, enabled, viaFallback, err }
	}

	tests := []struct {
		name      string
		query     func(uint32) (uint32, bool, bool, error)
		status    Status
		wantMsgID msgid.ID
		wantArgs  []any // nil に対しては検証しない
	}{
		{
			name:      "query error is unknown",
			query:     query(0, false, false, errors.New("COM failure")),
			status:    StatusUnknown,
			wantMsgID: msgid.FirewallUnknown,
		},
		{
			name:      "enabled on private",
			query:     query(FwProfilePrivate, true, false, nil),
			status:    StatusOK,
			wantMsgID: msgid.FirewallOK,
			wantArgs:  []any{"Private"},
		},
		{
			name:      "enabled on domain and private",
			query:     query(FwProfileDomain|FwProfilePrivate, true, false, nil),
			status:    StatusOK,
			wantMsgID: msgid.FirewallOK,
			wantArgs:  []any{"Domain+Private"},
		},
		{
			name:      "enabled via fallback rule enumeration",
			query:     query(FwProfilePrivate, true, true, nil),
			status:    StatusOK,
			wantMsgID: msgid.FirewallOKFallback,
			wantArgs:  []any{"Private"},
		},
		{
			name:      "disabled on public only network",
			query:     query(FwProfilePublic, false, false, nil),
			status:    StatusNG,
			wantMsgID: msgid.FirewallPublicBlocked,
		},
		{
			name:      "disabled on private",
			query:     query(FwProfilePrivate, false, false, nil),
			status:    StatusNG,
			wantMsgID: msgid.FirewallBlocked,
			wantArgs:  []any{"Private"},
		},
		{
			name:      "no active profile does not crash",
			query:     query(0, true, false, nil),
			status:    StatusOK,
			wantMsgID: msgid.FirewallOK,
			wantArgs:  []any{"none"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := FirewallCheck{ReadPort: defaultPort, Query: tt.query}
			got := c.Run()
			if got.Status != tt.status {
				t.Errorf("Status = %v, want %v", got.Status, tt.status)
			}
			if got.MsgID != tt.wantMsgID {
				t.Errorf("MsgID = %v, want %v", got.MsgID, tt.wantMsgID)
			}
			if tt.wantArgs != nil && !reflect.DeepEqual(got.MsgArgs, tt.wantArgs) {
				t.Errorf("MsgArgs = %v, want %v", got.MsgArgs, tt.wantArgs)
			}
		})
	}
}

func TestGroupMembershipCheck(t *testing.T) {
	const seGroupEnabled = 0x00000004
	const seGroupUseForDenyOnly = 0x00000010

	tests := []struct {
		name      string
		list      func() ([]TokenGroup, error)
		status    Status
		wantMsgID msgid.ID
	}{
		{
			name:      "list error is unknown",
			list:      func() ([]TokenGroup, error) { return nil, errors.New("access denied") },
			status:    StatusUnknown,
			wantMsgID: msgid.GroupUnknown,
		},
		{
			name: "administrators enabled",
			list: func() ([]TokenGroup, error) {
				return []TokenGroup{{SID: SIDAdministrators, Attributes: seGroupEnabled}}, nil
			},
			status:    StatusOK,
			wantMsgID: msgid.GroupOKAdmin,
		},
		{
			name: "administrators deny-only counts as member",
			list: func() ([]TokenGroup, error) {
				return []TokenGroup{{SID: SIDAdministrators, Attributes: seGroupUseForDenyOnly}}, nil
			},
			status:    StatusOK,
			wantMsgID: msgid.GroupOKAdmin,
		},
		{
			name: "remote desktop users only",
			list: func() ([]TokenGroup, error) {
				return []TokenGroup{{SID: SIDRemoteDesktopUsers, Attributes: seGroupEnabled}}, nil
			},
			status:    StatusOK,
			wantMsgID: msgid.GroupOKRDU,
		},
		{
			name: "both groups",
			list: func() ([]TokenGroup, error) {
				return []TokenGroup{
					{SID: SIDAdministrators, Attributes: seGroupEnabled},
					{SID: SIDRemoteDesktopUsers, Attributes: seGroupEnabled},
				}, nil
			},
			status:    StatusOK,
			wantMsgID: msgid.GroupOKBoth,
		},
		{
			name: "unrelated groups only",
			list: func() ([]TokenGroup, error) {
				return []TokenGroup{{SID: "S-1-5-32-545", Attributes: seGroupEnabled}}, nil
			},
			status:    StatusNG,
			wantMsgID: msgid.GroupNotMember,
		},
		{
			name:      "empty group list",
			list:      func() ([]TokenGroup, error) { return nil, nil },
			status:    StatusNG,
			wantMsgID: msgid.GroupNotMember,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := GroupMembershipCheck{ListTokenGroups: tt.list}
			got := c.Run()
			if got.Status != tt.status {
				t.Errorf("Status = %v, want %v", got.Status, tt.status)
			}
			if got.MsgID != tt.wantMsgID {
				t.Errorf("MsgID = %v, want %v", got.MsgID, tt.wantMsgID)
			}
		})
	}
}

func TestSleepCheck(t *testing.T) {
	read := func(ac, dc uint32, hasDC bool, err error) func() (uint32, uint32, bool, error) {
		return func() (uint32, uint32, bool, error) { return ac, dc, hasDC, err }
	}

	tests := []struct {
		name      string
		read      func() (uint32, uint32, bool, error)
		status    Status
		wantMsgID msgid.ID
		wantArgs  []any
	}{
		{
			name:      "read error is unknown",
			read:      read(0, 0, false, errors.New("boom")),
			status:    StatusUnknown,
			wantMsgID: msgid.SleepUnknown,
		},
		{
			name:      "never sleeps without battery",
			read:      read(0, 0, false, nil),
			status:    StatusOK,
			wantMsgID: msgid.SleepNever,
		},
		{
			name:      "never sleeps with battery",
			read:      read(0, 0, true, nil),
			status:    StatusOK,
			wantMsgID: msgid.SleepNever,
		},
		{
			name:      "sleeps on AC without battery",
			read:      read(900, 0, false, nil),
			status:    StatusWarn,
			wantMsgID: msgid.SleepWarnAC,
			wantArgs:  []any{uint32(900)},
		},
		{
			name:      "sleeps on AC and battery",
			read:      read(900, 600, true, nil),
			status:    StatusWarn,
			wantMsgID: msgid.SleepWarnBoth,
			wantArgs:  []any{uint32(900), uint32(600)},
		},
		{
			name:      "sleeps on battery only",
			read:      read(0, 600, true, nil),
			status:    StatusWarn,
			wantMsgID: msgid.SleepWarnDC,
			wantArgs:  []any{uint32(600)},
		},
		{
			name:      "non-whole-minute timeout shown in seconds",
			read:      read(90, 0, false, nil),
			status:    StatusWarn,
			wantMsgID: msgid.SleepWarnAC,
			wantArgs:  []any{uint32(90)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := SleepCheck{ReadTimeouts: tt.read}
			got := c.Run()
			if got.Status != tt.status {
				t.Errorf("Status = %v, want %v", got.Status, tt.status)
			}
			if got.MsgID != tt.wantMsgID {
				t.Errorf("MsgID = %v, want %v", got.MsgID, tt.wantMsgID)
			}
			if tt.wantArgs != nil && !reflect.DeepEqual(got.MsgArgs, tt.wantArgs) {
				t.Errorf("MsgArgs = %v, want %v", got.MsgArgs, tt.wantArgs)
			}
		})
	}
}

func TestRunAll_RecoversFromPanic(t *testing.T) {
	results := RunAll([]Check{panickyCheck{}})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusUnknown {
		t.Errorf("Status = %v, want StatusUnknown", results[0].Status)
	}
	if results[0].Name != "panicky" {
		t.Errorf("Name = %q, want panicky", results[0].Name)
	}
}

type panickyCheck struct{}

func (panickyCheck) Name() string     { return "panicky" }
func (panickyCheck) NeedsAdmin() bool { return false }
func (panickyCheck) Run() Result      { panic("boom") }
