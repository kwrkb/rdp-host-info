package diag

import "github.com/kwrkb/rdp-host-info/internal/msgid"

type ServiceRunningCheck struct {
	ServiceName  string
	DisplayName  string
	QueryRunning func(serviceName string) (bool, error)
}

func (c ServiceRunningCheck) Name() string     { return "service_" + c.ServiceName }
func (c ServiceRunningCheck) NeedsAdmin() bool { return false }

func (c ServiceRunningCheck) Run() Result {
	running, err := c.QueryRunning(c.ServiceName)
	if err != nil {
		return Result{
			Status:  StatusUnknown,
			MsgID:   msgid.ServiceUnknown,
			MsgArgs: []any{c.DisplayName},
			HintID:  msgid.ServiceUnknownHint,
		}
	}

	if !running {
		return Result{
			Status:   StatusNG,
			MsgID:    msgid.ServiceNotRunning,
			MsgArgs:  []any{c.DisplayName},
			HintID:   msgid.ServiceNotRunningHint,
			HintArgs: []any{c.DisplayName},
		}
	}

	return Result{
		Status:  StatusOK,
		MsgID:   msgid.ServiceRunning,
		MsgArgs: []any{c.DisplayName},
	}
}
