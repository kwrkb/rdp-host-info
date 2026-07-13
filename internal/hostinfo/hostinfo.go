package hostinfo

type Edition struct {
	ID              string
	DisplayName     string
	SupportsRDPHost bool
}

type HostInfo struct {
	PCName      string
	Login       UserLogin
	TailscaleIP string
	Recommended string
	Edition     Edition
	LocalIPv4   []string
}

// Providers はホスト情報取得手段の注入点。各関数はエラー時に空値/ゼロ値を返してよく、
// Collect はそれを "unknown" として扱い、他フィールドの取得は継続する。
type Providers struct {
	ComputerName func() (string, error)
	Edition      func() (Edition, error)
	LocalIPv4    func() ([]string, error)
	TailscaleIP  func() (string, error)
	Account      func() (AccountData, error)
}

func Collect(p Providers) HostInfo {
	info := HostInfo{}

	if p.ComputerName != nil {
		if v, err := p.ComputerName(); err == nil {
			info.PCName = v
		}
	}
	if p.Edition != nil {
		if v, err := p.Edition(); err == nil {
			info.Edition = v
		}
	}
	if p.LocalIPv4 != nil {
		if v, err := p.LocalIPv4(); err == nil {
			info.LocalIPv4 = v
		}
	}
	if p.TailscaleIP != nil {
		if v, err := p.TailscaleIP(); err == nil {
			info.TailscaleIP = v
		}
	}
	if p.Account != nil {
		if v, err := p.Account(); err == nil {
			info.Login = Classify(info.PCName, v)
		}
	}

	info.Recommended = recommend(info)
	return info
}

func recommend(info HostInfo) string {
	if info.TailscaleIP != "" {
		return info.TailscaleIP
	}
	if len(info.LocalIPv4) > 0 {
		return info.LocalIPv4[0]
	}
	return ""
}
