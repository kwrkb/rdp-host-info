// Package diag は RDP 接続受け入れ状態の診断項目（Check）を提供する。
// OS 非依存であり、Windows からの取得は関数型 provider として注入される
// （internal/winsys を import しない）。文言は持たず、internal/msgid の
// ID と値だけを返す。文言解決は internal/msg・整形は internal/render が担う。
package diag

import "github.com/kwrkb/rdp-host-info/internal/msgid"

type Status int

const (
	StatusUnknown Status = iota
	StatusOK
	StatusNG
	StatusWarn
)

type Result struct {
	Name       string
	NeedsAdmin bool
	Status     Status
	MsgID      msgid.ID
	MsgArgs    []any
	HintID     msgid.ID
	HintArgs   []any
}

type Check interface {
	Name() string
	NeedsAdmin() bool
	Run() Result
}

func RunAll(checks []Check) []Result {
	results := make([]Result, len(checks))
	for i, c := range checks {
		results[i] = runOne(c)
	}
	return results
}

func runOne(c Check) (result Result) {
	defer func() {
		if r := recover(); r != nil {
			result = Result{
				Status: StatusUnknown,
				MsgID:  msgid.InternalError,
				HintID: msgid.InternalErrorHint,
			}
		}
		result.Name = c.Name()
		result.NeedsAdmin = c.NeedsAdmin()
	}()
	return c.Run()
}
