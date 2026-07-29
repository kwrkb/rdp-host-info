package diag

import "github.com/kwrkb/rdp-host-info/internal/msgid"

type RDPEnabledCheck struct {
	// ReadDWORD は HKLM\SYSTEM\CurrentControlSet\Control\Terminal Server の
	// fDenyTSConnections を読む関数を注入する。
	ReadDWORD func(subKey, name string) (uint32, error)
}

const (
	terminalServerKey = `SYSTEM\CurrentControlSet\Control\Terminal Server`
	denyTSConnections = "fDenyTSConnections"
)

func (c RDPEnabledCheck) Name() string     { return "rdp_enabled" }
func (c RDPEnabledCheck) NeedsAdmin() bool { return false }

func (c RDPEnabledCheck) Run() Result {
	v, err := c.ReadDWORD(terminalServerKey, denyTSConnections)
	if err != nil {
		return Result{
			Status: StatusUnknown,
			MsgID:  msgid.RDPEnabledUnknown,
			HintID: msgid.RDPEnabledUnknownHint,
		}
	}

	if v != 0 {
		return Result{
			Status: StatusNG,
			MsgID:  msgid.RDPDisabled,
			HintID: msgid.RDPDisabledHint,
		}
	}

	return Result{
		Status: StatusOK,
		MsgID:  msgid.RDPEnabled,
	}
}
