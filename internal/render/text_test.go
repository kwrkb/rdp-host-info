package render

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/kwrkb/rdp-host-info/internal/diag"
	"github.com/kwrkb/rdp-host-info/internal/hostinfo"
	"github.com/kwrkb/rdp-host-info/internal/msg"
	"github.com/kwrkb/rdp-host-info/internal/msgid"
)

var update = flag.Bool("update", false, "update golden files")

func goldenPath(lang msg.Lang, name string) string {
	if lang == msg.Japanese {
		ext := filepath.Ext(name)
		name = name[:len(name)-len(ext)] + ".ja" + ext
	}
	return filepath.Join("testdata", name)
}

func checkGolden(t *testing.T, lang msg.Lang, name, got string) {
	t.Helper()
	path := goldenPath(lang, name)
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden file %s: %v", path, err)
		}
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden file %s: %v", path, err)
	}
	want := string(data)
	if got != want {
		t.Errorf("mismatch for %s:\ngot:\n%s\nwant:\n%s", path, got, want)
	}
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

var langs = []msg.Lang{msg.English, msg.Japanese}

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
				Candidates:  []hostinfo.UserCandidate{{Value: `OMEN16\yugo`, Label: msgid.LabelLocalAccount}},
			},
		},
		{
			name:   "microsoft account with two candidates and note",
			golden: "hostinfo_msa.golden",
			login: hostinfo.UserLogin{
				AccountType: hostinfo.AccountMicrosoft,
				Candidates: []hostinfo.UserCandidate{
					{Value: `MicrosoftAccount\user@example.com`, Label: msgid.LabelMicrosoftAccount},
					{Value: `OMEN16\yugo`, Label: msgid.LabelLocalAccount},
				},
				Notes: []msgid.ID{msgid.NoteMSAPassword},
			},
		},
		{
			name:   "azure ad account",
			golden: "hostinfo_azuread.golden",
			login: hostinfo.UserLogin{
				AccountType: hostinfo.AccountAzureAD,
				Candidates:  []hostinfo.UserCandidate{{Value: `AzureAD\yugo@example.com`, Label: msgid.LabelAzureADAccount}},
			},
		},
		{
			name:   "domain account",
			golden: "hostinfo_domain.golden",
			login: hostinfo.UserLogin{
				AccountType: hostinfo.AccountDomain,
				Candidates:  []hostinfo.UserCandidate{{Value: `CORP\yugo`, Label: msgid.LabelDomainAccount}},
			},
		},
		{
			name:   "unknown account type with note",
			golden: "hostinfo_account_unknown.golden",
			login: hostinfo.UserLogin{
				AccountType: hostinfo.AccountUnknown,
				Candidates:  []hostinfo.UserCandidate{{Value: `OMEN16\yugo`, Label: msgid.LabelLocalAccount}},
				Notes:       []msgid.ID{msgid.NoteMaybeMSA},
			},
		},
	}

	for _, tt := range tests {
		for _, lang := range langs {
			t.Run(tt.name+"_"+string(lang), func(t *testing.T) {
				got := HostInfoText(lang, baseHostInfo(tt.login))
				checkGolden(t, lang, tt.golden, got)
			})
		}
	}
}

func TestHostInfoText_UnknownFields(t *testing.T) {
	for _, lang := range langs {
		t.Run(string(lang), func(t *testing.T) {
			got := HostInfoText(lang, hostinfo.HostInfo{})
			checkGolden(t, lang, "hostinfo_unknown.golden", got)
		})
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
		for _, lang := range langs {
			t.Run(tt.name+"_"+string(lang), func(t *testing.T) {
				got := StatusText(lang, diag.RunAll(tt.checks))
				checkGolden(t, lang, tt.golden, got)
			})
		}
	}
}

func TestStatusText(t *testing.T) {
	results := []diag.Result{
		{Name: "edition", Status: diag.StatusOK, MsgID: msgid.EditionSupported, MsgArgs: []any{"Windows 11 Pro"}},
		{Name: "rdp_enabled", Status: diag.StatusNG, MsgID: msgid.RDPDisabled, HintID: msgid.RDPDisabledHint},
		{Name: "group", Status: diag.StatusUnknown, MsgID: msgid.GroupUnknown, NeedsAdmin: true},
		{Name: "sleep", Status: diag.StatusWarn, MsgID: msgid.SleepWarnAC, MsgArgs: []any{uint32(900)}, HintID: msgid.SleepWarnHint},
	}

	for _, lang := range langs {
		t.Run(string(lang), func(t *testing.T) {
			got := StatusText(lang, results)
			checkGolden(t, lang, "status.golden", got)
		})
	}
}
