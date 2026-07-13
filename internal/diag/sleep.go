package diag

import "strconv"

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
			Status:  StatusUnknown,
			Message: "Sleep settings could not be determined",
			Hint:    "設定 > システム > 電源 でスリープ設定を確認してください。",
		}
	}

	if ac == 0 && (!hasDC || dc == 0) {
		return Result{
			Status:  StatusOK,
			Message: "PC does not sleep automatically",
		}
	}

	// Modern Standby (S0) 機ではタイムアウト値の意味が従来スリープと異なる
	// ため、Hint は断定を避ける。
	hint := "スリープ中はリモートデスクトップ接続を受け付けられない場合があります。常時接続したい場合は 設定 > システム > 電源 でスリープを「なし」にすることを検討してください。"

	var msg string
	switch {
	case ac > 0 && hasDC && dc > 0:
		msg = "PC sleeps after " + formatDuration(ac) + " (plugged in) / " + formatDuration(dc) + " (on battery)"
	case ac > 0:
		msg = "PC sleeps after " + formatDuration(ac)
	default:
		msg = "PC sleeps after " + formatDuration(dc) + " (on battery)"
	}

	return Result{
		Status:  StatusWarn,
		Message: msg,
		Hint:    hint,
	}
}

// formatDuration は秒数を "15 minutes" / "90 seconds" の形式にする。
func formatDuration(seconds uint32) string {
	if seconds%60 == 0 {
		return strconv.Itoa(int(seconds/60)) + " minutes"
	}
	return strconv.Itoa(int(seconds)) + " seconds"
}
