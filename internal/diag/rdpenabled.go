package diag

type RDPEnabledCheck struct {
	// ReadDWORD は HKLM\SYSTEM\CurrentControlSet\Control\Terminal Server の
	// fDenyTSConnections を読む関数を注入する。
	ReadDWORD func(subKey, name string) (uint32, error)
}

const (
	terminalServerKey  = `SYSTEM\CurrentControlSet\Control\Terminal Server`
	denyTSConnections  = "fDenyTSConnections"
)

func (c RDPEnabledCheck) Name() string     { return "rdp_enabled" }
func (c RDPEnabledCheck) NeedsAdmin() bool { return false }

func (c RDPEnabledCheck) Run() Result {
	v, err := c.ReadDWORD(terminalServerKey, denyTSConnections)
	if err != nil {
		return Result{
			Status:  StatusUnknown,
			Message: "Remote Desktop enabled state could not be determined",
			Hint:    "設定 > システム > リモートデスクトップ で状態を確認してください。",
		}
	}

	if v != 0 {
		return Result{
			Status:  StatusNG,
			Message: "Remote Desktop is disabled",
			Hint:    "Windowsの設定からリモートデスクトップを有効にしてください。\n設定 > システム > リモートデスクトップ",
		}
	}

	return Result{
		Status:  StatusOK,
		Message: "Remote Desktop is enabled",
	}
}
