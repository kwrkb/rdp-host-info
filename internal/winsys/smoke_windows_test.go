//go:build windows

package winsys

import (
	"strings"
	"testing"
)

// winsys はロジックを持たない（取得と型変換のみ）ため、実 OS 上で
// 「エラーなく妥当な形の値が返る」ことだけを検証する。

func TestListTokenGroups_Smoke(t *testing.T) {
	groups, err := ListTokenGroups()
	if err != nil {
		t.Fatalf("ListTokenGroups() error = %v", err)
	}
	if len(groups) == 0 {
		t.Fatal("ListTokenGroups() returned no groups; every token has at least Everyone")
	}
	for _, g := range groups {
		if !strings.HasPrefix(g.SID, "S-1-") {
			t.Errorf("SID %q does not look like a SID", g.SID)
		}
	}
}

func TestReadSleepTimeouts_Smoke(t *testing.T) {
	ac, dc, hasDC, err := ReadSleepTimeouts()
	if err != nil {
		t.Fatalf("ReadSleepTimeouts() error = %v", err)
	}
	// 秒値は 0（なし）または現実的な範囲（1年未満）であること。
	const yearSeconds = 365 * 24 * 3600
	if ac > yearSeconds {
		t.Errorf("acSeconds = %d, implausibly large", ac)
	}
	if hasDC && dc > yearSeconds {
		t.Errorf("dcSeconds = %d, implausibly large", dc)
	}
}

func TestQueryRDPFirewall_Smoke(t *testing.T) {
	active, _, _, err := QueryRDPFirewall(3389)
	if err != nil {
		t.Fatalf("QueryRDPFirewall() error = %v", err)
	}
	// CurrentProfileTypes は Domain|Private|Public(=7) か ALL(0x7FFFFFFF) の範囲。
	if active != 0x7FFFFFFF && active > 7 {
		t.Errorf("activeProfiles = %#x, outside NET_FW_PROFILE_TYPE2 mask", active)
	}
}
