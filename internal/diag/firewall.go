package diag

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
			Status:  StatusUnknown,
			Message: "Windows Firewall state could not be determined",
			Hint:    "wf.msc（セキュリティが強化された Windows ファイアウォール）で「リモート デスクトップ」の受信ルールを確認してください。",
		}
	}

	profiles := profileNames(active)

	if enabled {
		note := ""
		if viaFallback {
			note = ", port rule"
		}
		return Result{
			Status:  StatusOK,
			Message: "Windows Firewall allows Remote Desktop (" + profiles + " profile, active" + note + ")",
		}
	}

	if active == FwProfilePublic {
		return Result{
			Status:  StatusNG,
			Message: "Network is set to Public and Remote Desktop is not allowed",
			Hint:    "ファイアウォールのリモートデスクトップ許可が、現在のネットワークに適用されていません。設定 > ネットワークとインターネット でネットワークを「プライベート」に変更してください。",
		}
	}

	return Result{
		Status:  StatusNG,
		Message: "Windows Firewall blocks Remote Desktop (" + profiles + " profile, active)",
		Hint:    "wf.msc で「リモート デスクトップ」の受信ルールを現在のプロファイルで有効にしてください。サードパーティ製セキュリティソフト使用時は、そちらの設定も確認してください。",
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
