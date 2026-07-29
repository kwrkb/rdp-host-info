package diag

import "github.com/kwrkb/rdp-host-info/internal/msgid"

// アクティブプロファイルのビットマスク。NET_FW_PROFILE_TYPE2 と同値。
const (
	FwProfileDomain  uint32 = 1
	FwProfilePrivate uint32 = 2
	FwProfilePublic  uint32 = 4
)

type FirewallCheck struct {
	// ReadPort は設定されたRDPポートを返す（フォールバックのルール列挙用）。
	ReadPort func() (port uint32, fromRegistry bool)
	// Query はファイアウォールの事実のみを返し、判定は持たない。
	// activeProfiles: CurrentProfileTypes のビットマスク。
	// ruleEnabled: RDPルール群（またはフォールバックで見つけたポート一致ルール）が
	// 現在のプロファイルで有効か。
	// viaFallback: ルールグループの間接文字列が使えず、ルール列挙で判定した。
	Query func(port uint32) (activeProfiles uint32, ruleEnabled bool, viaFallback bool, err error)
}

func (c FirewallCheck) Name() string     { return "firewall" }
func (c FirewallCheck) NeedsAdmin() bool { return false }

func (c FirewallCheck) Run() Result {
	port, _ := c.ReadPort()

	active, enabled, viaFallback, err := c.Query(port)
	if err != nil {
		return Result{
			Status: StatusUnknown,
			MsgID:  msgid.FirewallUnknown,
			HintID: msgid.FirewallUnknownHint,
		}
	}

	profiles := profileNames(active)

	if enabled {
		id := msgid.FirewallOK
		if viaFallback {
			id = msgid.FirewallOKFallback
		}
		return Result{
			Status:  StatusOK,
			MsgID:   id,
			MsgArgs: []any{profiles},
		}
	}

	if active == FwProfilePublic {
		return Result{
			Status: StatusNG,
			MsgID:  msgid.FirewallPublicBlocked,
			HintID: msgid.FirewallPublicBlockedHint,
		}
	}

	return Result{
		Status:  StatusNG,
		MsgID:   msgid.FirewallBlocked,
		MsgArgs: []any{profiles},
		HintID:  msgid.FirewallBlockedHint,
	}
}

// profileNames はプロファイルのビットマスクを "Private" / "Domain+Private" の形式にする。
func profileNames(mask uint32) string {
	names := ""
	add := func(s string) {
		if names != "" {
			names += "+"
		}
		names += s
	}
	if mask&FwProfileDomain != 0 {
		add("Domain")
	}
	if mask&FwProfilePrivate != 0 {
		add("Private")
	}
	if mask&FwProfilePublic != 0 {
		add("Public")
	}
	if names == "" {
		return "none"
	}
	return names
}
