package render

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/kwrkb/rdp-host-info/internal/diag"
	"github.com/kwrkb/rdp-host-info/internal/hostinfo"
)

func readGolden(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read golden file %s: %v", name, err)
	}
	return string(data)
}

func baseHostInfo(login hostinfo.UserLogin) hostinfo.HostInfo {
	return hostinfo.HostInfo{
		PCName:      "OMEN16",
		Login:       login,
		TailscaleIP: "100.80.10.5",
		Recommended: "100.80.10.5",
		Edition: hostinfo.Edition{
			ID:              "Professional",
			DisplayName:     "Windows 11 Pro",
			SupportsRDPHost: true,
		},
		LocalIPv4: []string{"192.168.1.20"},
	}
}

func TestHostInfoText(t *testing.T) {
	tests := []struct {
		name   string
		golden string
		login  hostinfo.UserLogin
	}{
		{
			name:   "local account",
			golden: "hostinfo.golden",
			login: hostinfo.UserLogin{
				AccountType: hostinfo.AccountLocal,
				Candidates:  []hostinfo.UserCandidate{{Value: `OMEN16\yugo`, Label: "local account"}},
			},
		},
		{
			name:   "microsoft account with two candidates and note",
			golden: "hostinfo_msa.golden",
			login: hostinfo.UserLogin{
				AccountType: hostinfo.AccountMicrosoft,
				Candidates: []hostinfo.UserCandidate{
					{Value: `MicrosoftAccount\user@example.com`, Label: "Microsoft account"},
					{Value: `OMEN16\yugo`, Label: "local account"},
				},
				Notes: []string{"Microsoft アカウントでは PIN ではなくアカウントのパスワードでサインインしてください。"},
			},
		},
		{
			name:   "azure ad account",
			golden: "hostinfo_azuread.golden",
			login: hostinfo.UserLogin{
				AccountType: hostinfo.AccountAzureAD,
				Candidates:  []hostinfo.UserCandidate{{Value: `AzureAD\yugo@example.com`, Label: "Microsoft Entra ID account"}},
			},
		},
		{
			name:   "domain account",
			golden: "hostinfo_domain.golden",
			login: hostinfo.UserLogin{
				AccountType: hostinfo.AccountDomain,
				Candidates:  []hostinfo.UserCandidate{{Value: `CORP\yugo`, Label: "domain account"}},
			},
		},
		{
			name:   "unknown account type with note",
			golden: "hostinfo_account_unknown.golden",
			login: hostinfo.UserLogin{
				AccountType: hostinfo.AccountUnknown,
				Candidates:  []hostinfo.UserCandidate{{Value: `OMEN16\yugo`, Label: "local account"}},
				Notes:       []string{`Microsoft アカウントの場合は MicrosoftAccount\メールアドレス 形式も試してください。`},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HostInfoText(baseHostInfo(tt.login))
			want := readGolden(t, tt.golden)
			if got != want {
				t.Errorf("HostInfoText mismatch:\ngot:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

func TestHostInfoText_UnknownFields(t *testing.T) {
	got := HostInfoText(hostinfo.HostInfo{})
	want := readGolden(t, "hostinfo_unknown.golden")
	if got != want {
		t.Errorf("HostInfoText mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// TestStatusText_RealChecks は実際の Check 実装（fake provider 注入）の
// Message/Hint 文言を end-to-end で golden と比較する。
func TestStatusText_RealChecks(t *testing.T) {
	edition := func(display string, supports bool) func() (hostinfo.Edition, error) {
		return func() (hostinfo.Edition, error) {
			return hostinfo.Edition{DisplayName: display, SupportsRDPHost: supports}, nil
		}
	}
	port3389 := func() (uint32, bool) { return 3389, true }

	tests := []struct {
		name   string
		golden string
		checks []diag.Check
	}{
		{
			name:   "all ok with sleep warning (VISION full set)",
			golden: "status_all_ok.golden",
			checks: []diag.Check{
				diag.EditionSupportCheck{ReadEdition: edition("Windows 11 Pro", true)},
				diag.RDPEnabledCheck{ReadDWORD: func(string, string) (uint32, error) { return 0, nil }},
				diag.ServiceRunningCheck{
					ServiceName: "TermService", DisplayName: "Remote Desktop Services",
					QueryRunning: func(string) (bool, error) { return true, nil },
				},
				diag.FirewallCheck{
					ReadPort: port3389,
					Query:    func(uint32) (uint32, bool, bool, error) { return diag.FwProfilePrivate, true, false, nil },
				},
				diag.PortListeningCheck{
					ReadPort:    port3389,
					IsListening: func(uint32) (bool, error) { return true, nil },
				},
				diag.GroupMembershipCheck{
					ListTokenGroups: func() ([]diag.TokenGroup, error) {
						return []diag.TokenGroup{{SID: diag.SIDAdministrators}}, nil
					},
				},
				diag.SleepCheck{ReadTimeouts: func() (uint32, uint32, bool, error) { return 900, 0, false, nil }},
			},
		},
		{
			name:   "ng with hints (home edition, rdp disabled)",
			golden: "status_ng.golden",
			checks: []diag.Check{
				diag.EditionSupportCheck{ReadEdition: edition("Windows 11 Home", false)},
				diag.RDPEnabledCheck{ReadDWORD: func(string, string) (uint32, error) { return 1, nil }},
			},
		},
		{
			name:   "all unknown with manual-check hints",
			golden: "status_all_unknown.golden",
			checks: []diag.Check{
				diag.EditionSupportCheck{
					ReadEdition: func() (hostinfo.Edition, error) { return hostinfo.Edition{}, errors.New("boom") },
				},
				diag.RDPEnabledCheck{ReadDWORD: func(string, string) (uint32, error) { return 0, errors.New("boom") }},
				diag.FirewallCheck{
					ReadPort: port3389,
					Query:    func(uint32) (uint32, bool, bool, error) { return 0, false, false, errors.New("boom") },
				},
				diag.GroupMembershipCheck{
					ListTokenGroups: func() ([]diag.TokenGroup, error) { return nil, errors.New("boom") },
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StatusText(diag.RunAll(tt.checks))
			want := readGolden(t, tt.golden)
			if got != want {
				t.Errorf("StatusText mismatch:\ngot:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

func TestStatusText(t *testing.T) {
	results := []diag.Result{
		{Name: "edition", Status: diag.StatusOK, Message: "Windows supports Remote Desktop hosting"},
		{Name: "rdp_enabled", Status: diag.StatusNG, Message: "Remote Desktop is disabled", Hint: "設定からリモートデスクトップを有効にしてください。"},
		{Name: "group", Status: diag.StatusUnknown, Message: "group membership could not be determined", NeedsAdmin: true},
		{Name: "sleep", Status: diag.StatusWarn, Message: "PC sleeps after 15 minutes", Hint: "スリープ中はリモートデスクトップ接続を受け付けられない場合があります。"},
	}

	got := StatusText(results)
	want := readGolden(t, "status.golden")
	if got != want {
		t.Errorf("StatusText mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}
