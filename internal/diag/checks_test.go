package diag

import (
	"errors"
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
