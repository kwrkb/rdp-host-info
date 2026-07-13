package diag

import "strconv"

type PortListeningCheck struct {
	// ReadPort は設定されたRDPポートを返す。fromRegistry が false の場合は
	// レジストリから読めなかったため既定値(3389)を仮定していることを示す。
	ReadPort func() (port uint32, fromRegistry bool)
	// IsListening は指定ポートが TCP で待ち受け状態か判定する。
	IsListening func(port uint32) (bool, error)
}

func (c PortListeningCheck) Name() string     { return "port_listening" }
func (c PortListeningCheck) NeedsAdmin() bool { return false }

func (c PortListeningCheck) Run() Result {
	port, fromRegistry := c.ReadPort()

	listening, err := c.IsListening(port)
	if err != nil {
		return Result{
			Status:  StatusUnknown,
			Message: "TCP " + strconv.Itoa(int(port)) + " listening state could not be determined",
			Hint:    "コマンドプロンプトで netstat -an を実行し、該当ポートの状態を確認してください。",
		}
	}

	portNote := ""
	if !fromRegistry {
		portNote = " (assumed default)"
	}

	if !listening {
		return Result{
			Status:  StatusNG,
			Message: "TCP " + strconv.Itoa(int(port)) + portNote + " is not listening",
			Hint:    "Remote Desktop Services が起動しているか、リモートデスクトップが有効になっているか確認してください。",
		}
	}

	return Result{
		Status:  StatusOK,
		Message: "TCP " + strconv.Itoa(int(port)) + portNote + " is listening",
	}
}
