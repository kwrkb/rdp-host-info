package diag

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
			Message: c.DisplayName + " status could not be determined",
			Hint:    "サービスの状態を確認できませんでした。services.msc で確認してください。",
		}
	}

	if !running {
		return Result{
			Status:  StatusNG,
			Message: c.DisplayName + " is not running",
			Hint:    "services.msc から " + c.DisplayName + " を開始してください。",
		}
	}

	return Result{
		Status:  StatusOK,
		Message: c.DisplayName + " is running",
	}
}
