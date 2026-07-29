package diag

import "github.com/kwrkb/rdp-host-info/internal/msgid"

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
			MsgID:   msgid.PortUnknown,
			MsgArgs: []any{port},
			HintID:  msgid.PortUnknownHint,
		}
	}

	if !listening {
		id := msgid.PortNotListening
		if !fromRegistry {
			id = msgid.PortNotListeningAssumed
		}
		return Result{
			Status:  StatusNG,
			MsgID:   id,
			MsgArgs: []any{port},
			HintID:  msgid.PortNotListeningHint,
		}
	}

	id := msgid.PortListening
	if !fromRegistry {
		id = msgid.PortListeningAssumed
	}
	return Result{
		Status:  StatusOK,
		MsgID:   id,
		MsgArgs: []any{port},
	}
}
