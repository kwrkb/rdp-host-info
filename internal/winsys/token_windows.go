package winsys

import (
	"golang.org/x/sys/windows"

	"github.com/kwrkb/rdp-host-info/internal/diag"
)

// ListTokenGroups は現在のプロセストークンに含まれるグループ SID と属性を列挙する。
// 所属判定のロジックは持たず、diag.GroupMembershipCheck 側で行う。
func ListTokenGroups() ([]diag.TokenGroup, error) {
	token := windows.GetCurrentProcessToken() // 擬似ハンドルのため Close 不要

	tg, err := token.GetTokenGroups()
	if err != nil {
		return nil, err
	}

	groups := make([]diag.TokenGroup, 0, tg.GroupCount)
	for _, g := range tg.AllGroups() {
		groups = append(groups, diag.TokenGroup{
			SID:        g.Sid.String(),
			Attributes: g.Attributes,
		})
	}
	return groups, nil
}
