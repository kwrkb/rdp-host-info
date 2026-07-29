package diag

import "github.com/kwrkb/rdp-host-info/internal/msgid"

type SleepCheck struct {
	// ReadTimeouts はスリープまでの秒数を返す（0 = スリープしない）。
	// hasDC はバッテリー駆動時(DC)の値が読めたか（デスクトップ機では false）。
	// AC 値が読めない場合は err を返す。
	ReadTimeouts func() (acSeconds, dcSeconds uint32, hasDC bool, err error)
}

func (c SleepCheck) Name() string     { return "sleep" }
func (c SleepCheck) NeedsAdmin() bool { return false }

func (c SleepCheck) Run() Result {
	ac, dc, hasDC, err := c.ReadTimeouts()
	if err != nil {
		return Result{
			Status: StatusUnknown,
			MsgID:  msgid.SleepUnknown,
			HintID: msgid.SleepUnknownHint,
		}
	}

	if ac == 0 && (!hasDC || dc == 0) {
		return Result{
			Status: StatusOK,
			MsgID:  msgid.SleepNever,
		}
	}

	// Modern Standby (S0) 機ではタイムアウト値の意味が従来スリープと異なる
	// ため、Hint（SleepWarnHint）は断定を避ける。
	switch {
	case ac > 0 && hasDC && dc > 0:
		return Result{Status: StatusWarn, MsgID: msgid.SleepWarnBoth, MsgArgs: []any{ac, dc}, HintID: msgid.SleepWarnHint}
	case ac > 0:
		return Result{Status: StatusWarn, MsgID: msgid.SleepWarnAC, MsgArgs: []any{ac}, HintID: msgid.SleepWarnHint}
	default:
		return Result{Status: StatusWarn, MsgID: msgid.SleepWarnDC, MsgArgs: []any{dc}, HintID: msgid.SleepWarnHint}
	}
}
