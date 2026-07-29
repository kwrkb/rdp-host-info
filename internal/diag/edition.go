package diag

import (
	"github.com/kwrkb/rdp-host-info/internal/hostinfo"
	"github.com/kwrkb/rdp-host-info/internal/msgid"
)

type EditionSupportCheck struct {
	ReadEdition func() (hostinfo.Edition, error)
}

func (c EditionSupportCheck) Name() string     { return "edition_support" }
func (c EditionSupportCheck) NeedsAdmin() bool { return false }

func (c EditionSupportCheck) Run() Result {
	edition, err := c.ReadEdition()
	if err != nil {
		return Result{
			Status: StatusUnknown,
			MsgID:  msgid.EditionUnknown,
			HintID: msgid.EditionUnknownHint,
		}
	}

	if !edition.SupportsRDPHost {
		return Result{
			Status:  StatusNG,
			MsgID:   msgid.EditionUnsupported,
			MsgArgs: []any{edition.DisplayName},
			HintID:  msgid.EditionUnsupportedHint,
		}
	}

	return Result{
		Status:  StatusOK,
		MsgID:   msgid.EditionSupported,
		MsgArgs: []any{edition.DisplayName},
	}
}
