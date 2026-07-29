package diag

import "github.com/kwrkb/rdp-host-info/internal/msgid"

// RDP接続を許可する well-known SID。
const (
	SIDRemoteDesktopUsers = "S-1-5-32-555"
	SIDAdministrators     = "S-1-5-32-544"
)

// TokenGroup はトークンに含まれるグループ SID とその属性。
// UAC 非昇格トークンでは Administrators が SE_GROUP_USE_FOR_DENY_ONLY
// （ENABLED なし）で載るが、SID の存在自体が所属の証拠であり、RDP ログオンは
// 新規トークンを生成するため、属性を問わず「存在すれば所属」と判定する。
type TokenGroup struct {
	SID        string
	Attributes uint32
}

type GroupMembershipCheck struct {
	ListTokenGroups func() ([]TokenGroup, error)
}

func (c GroupMembershipCheck) Name() string     { return "group_membership" }
func (c GroupMembershipCheck) NeedsAdmin() bool { return false }

func (c GroupMembershipCheck) Run() Result {
	groups, err := c.ListTokenGroups()
	if err != nil {
		return Result{
			Status: StatusUnknown,
			MsgID:  msgid.GroupUnknown,
			HintID: msgid.GroupUnknownHint,
		}
	}

	isAdmin := false
	isRDU := false
	for _, g := range groups {
		switch g.SID {
		case SIDAdministrators:
			isAdmin = true
		case SIDRemoteDesktopUsers:
			isRDU = true
		}
	}

	if !isAdmin && !isRDU {
		return Result{
			Status: StatusNG,
			MsgID:  msgid.GroupNotMember,
			HintID: msgid.GroupNotMemberHint,
		}
	}

	// グループ所属のみの確認であり、「リモート デスクトップ サービスを使った
	// ログオンを拒否する」ポリシーで拒否されている場合は所属していても
	// 接続できない（このチェックでは検出できない）ため、断定を避ける（GroupOKHint）。
	id := msgid.GroupOKRDU
	switch {
	case isAdmin && isRDU:
		id = msgid.GroupOKBoth
	case isAdmin:
		id = msgid.GroupOKAdmin
	}

	return Result{
		Status: StatusOK,
		MsgID:  id,
		HintID: msgid.GroupOKHint,
	}
}
