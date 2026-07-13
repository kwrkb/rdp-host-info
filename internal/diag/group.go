package diag

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
			Status:  StatusUnknown,
			Message: "Group membership could not be determined",
			Hint:    "lusrmgr.msc または 設定 > システム > リモートデスクトップ で、ユーザーが Remote Desktop Users か Administrators に含まれるか確認してください。",
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
			Status:  StatusNG,
			Message: "User is not in Remote Desktop Users or Administrators",
			Hint:    "設定 > システム > リモートデスクトップ > リモートデスクトップユーザー から、接続に使うユーザーを追加してください。",
		}
	}

	names := ""
	if isAdmin {
		names = "Administrators"
	}
	if isRDU {
		if names != "" {
			names += ", "
		}
		names += "Remote Desktop Users"
	}

	return Result{
		Status:  StatusOK,
		Message: "User is allowed to connect (" + names + ")",
	}
}
