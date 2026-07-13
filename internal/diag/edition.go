package diag

import "github.com/kwrkb/rdp-host-info/internal/hostinfo"

type EditionSupportCheck struct {
	ReadEdition func() (hostinfo.Edition, error)
}

func (c EditionSupportCheck) Name() string     { return "edition_support" }
func (c EditionSupportCheck) NeedsAdmin() bool { return false }

func (c EditionSupportCheck) Run() Result {
	edition, err := c.ReadEdition()
	if err != nil {
		return Result{
			Status:  StatusUnknown,
			Message: "Windows edition could not be determined",
			Hint:    "設定 > システム > バージョン情報 でエディションを確認してください。",
		}
	}

	if !edition.SupportsRDPHost {
		return Result{
			Status:  StatusNG,
			Message: edition.DisplayName + " does not support Remote Desktop hosting",
			Hint:    "Windows Home エディションはリモートデスクトップのホストになれません。Pro/Enterprise/Education への変更が必要です。",
		}
	}

	return Result{
		Status:  StatusOK,
		Message: "Windows supports Remote Desktop hosting (" + edition.DisplayName + ")",
	}
}
