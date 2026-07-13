package diag

import (
	"errors"
	"strings"
	"testing"

	"github.com/kwrkb/rdp-host-info/internal/hostinfo"
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
		name    string
		query   func(uint32) (uint32, bool, bool, error)
		status  Status
		wantMsg string // 空でなければ Message に含まれること
	}{
		{
			name:   "query error is unknown",
			query:  query(0, false, false, errors.New("COM failure")),
			status: StatusUnknown,
		},
		{
			name:    "enabled on private",
			query:   query(FwProfilePrivate, true, false, nil),
			status:  StatusOK,
			wantMsg: "(Private profile, active)",
		},
		{
			name:    "enabled on domain and private",
			query:   query(FwProfileDomain|FwProfilePrivate, true, false, nil),
			status:  StatusOK,
			wantMsg: "(Domain+Private profile, active)",
		},
		{
			name:    "enabled via fallback rule enumeration",
			query:   query(FwProfilePrivate, true, true, nil),
			status:  StatusOK,
			wantMsg: "port rule",
		},
		{
			name:    "disabled on public only network",
			query:   query(FwProfilePublic, false, false, nil),
			status:  StatusNG,
			wantMsg: "Public",
		},
		{
			name:    "disabled on private",
			query:   query(FwProfilePrivate, false, false, nil),
			status:  StatusNG,
			wantMsg: "blocks Remote Desktop",
		},
		{
			name:    "no active profile does not crash",
			query:   query(0, true, false, nil),
			status:  StatusOK,
			wantMsg: "(none profile, active)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := FirewallCheck{ReadPort: defaultPort, Query: tt.query}
			got := c.Run()
			if got.Status != tt.status {
				t.Errorf("Status = %v, want %v", got.Status, tt.status)
			}
			if tt.wantMsg != "" && !strings.Contains(got.Message, tt.wantMsg) {
				t.Errorf("Message = %q, want substring %q", got.Message, tt.wantMsg)
			}
		})
	}
}

func TestGroupMembershipCheck(t *testing.T) {
	const seGroupEnabled = 0x00000004
	const seGroupUseForDenyOnly = 0x00000010

	tests := []struct {
		name    string
		list    func() ([]TokenGroup, error)
		status  Status
		wantMsg string
	}{
		{
			name:   "list error is unknown",
			list:   func() ([]TokenGroup, error) { return nil, errors.New("access denied") },
			status: StatusUnknown,
		},
		{
			name: "administrators enabled",
			list: func() ([]TokenGroup, error) {
				return []TokenGroup{{SID: SIDAdministrators, Attributes: seGroupEnabled}}, nil
			},
			status:  StatusOK,
			wantMsg: "(Administrators)",
		},
		{
			name: "administrators deny-only counts as member",
			list: func() ([]TokenGroup, error) {
				return []TokenGroup{{SID: SIDAdministrators, Attributes: seGroupUseForDenyOnly}}, nil
			},
			status:  StatusOK,
			wantMsg: "(Administrators)",
		},
		{
			name: "remote desktop users only",
			list: func() ([]TokenGroup, error) {
				return []TokenGroup{{SID: SIDRemoteDesktopUsers, Attributes: seGroupEnabled}}, nil
			},
			status:  StatusOK,
			wantMsg: "(Remote Desktop Users)",
		},
		{
			name: "both groups",
			list: func() ([]TokenGroup, error) {
				return []TokenGroup{
					{SID: SIDAdministrators, Attributes: seGroupEnabled},
					{SID: SIDRemoteDesktopUsers, Attributes: seGroupEnabled},
				}, nil
			},
			status:  StatusOK,
			wantMsg: "(Administrators, Remote Desktop Users)",
		},
		{
			name: "unrelated groups only",
			list: func() ([]TokenGroup, error) {
				return []TokenGroup{{SID: "S-1-5-32-545", Attributes: seGroupEnabled}}, nil
			},
			status: StatusNG,
		},
		{
			name:   "empty group list",
			list:   func() ([]TokenGroup, error) { return nil, nil },
			status: StatusNG,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := GroupMembershipCheck{ListTokenGroups: tt.list}
			got := c.Run()
			if got.Status != tt.status {
				t.Errorf("Status = %v, want %v", got.Status, tt.status)
			}
			if tt.wantMsg != "" && !strings.Contains(got.Message, tt.wantMsg) {
				t.Errorf("Message = %q, want substring %q", got.Message, tt.wantMsg)
			}
		})
	}
}

func TestSleepCheck(t *testing.T) {
	read := func(ac, dc uint32, hasDC bool, err error) func() (uint32, uint32, bool, error) {
		return func() (uint32, uint32, bool, error) { return ac, dc, hasDC, err }
	}

	tests := []struct {
		name    string
		read    func() (uint32, uint32, bool, error)
		status  Status
		wantMsg string
	}{
		{
			name:   "read error is unknown",
			read:   read(0, 0, false, errors.New("boom")),
			status: StatusUnknown,
		},
		{
			name:   "never sleeps without battery",
			read:   read(0, 0, false, nil),
			status: StatusOK,
		},
		{
			name:   "never sleeps with battery",
			read:   read(0, 0, true, nil),
			status: StatusOK,
		},
		{
			name:    "sleeps on AC without battery",
			read:    read(900, 0, false, nil),
			status:  StatusWarn,
			wantMsg: "15 minutes",
		},
		{
			name:    "sleeps on AC and battery",
			read:    read(900, 600, true, nil),
			status:  StatusWarn,
			wantMsg: "15 minutes (plugged in) / 10 minutes (on battery)",
		},
		{
			name:    "sleeps on battery only",
			read:    read(0, 600, true, nil),
			status:  StatusWarn,
			wantMsg: "10 minutes (on battery)",
		},
		{
			name:    "non-whole-minute timeout shown in seconds",
			read:    read(90, 0, false, nil),
			status:  StatusWarn,
			wantMsg: "90 seconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := SleepCheck{ReadTimeouts: tt.read}
			got := c.Run()
			if got.Status != tt.status {
				t.Errorf("Status = %v, want %v", got.Status, tt.status)
			}
			if tt.wantMsg != "" && !strings.Contains(got.Message, tt.wantMsg) {
				t.Errorf("Message = %q, want substring %q", got.Message, tt.wantMsg)
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
