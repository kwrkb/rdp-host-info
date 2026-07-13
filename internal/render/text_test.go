package render

import (
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

func TestHostInfoText(t *testing.T) {
	info := hostinfo.HostInfo{
		PCName:      "OMEN16",
		UserName:    "OMEN16\\yugo",
		TailscaleIP: "100.80.10.5",
		Recommended: "100.80.10.5",
		Edition: hostinfo.Edition{
			ID:              "Professional",
			DisplayName:     "Windows 11 Pro",
			SupportsRDPHost: true,
		},
		LocalIPv4: []string{"192.168.1.20"},
	}

	got := HostInfoText(info)
	want := readGolden(t, "hostinfo.golden")
	if got != want {
		t.Errorf("HostInfoText mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestHostInfoText_UnknownFields(t *testing.T) {
	got := HostInfoText(hostinfo.HostInfo{})
	want := readGolden(t, "hostinfo_unknown.golden")
	if got != want {
		t.Errorf("HostInfoText mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestStatusText(t *testing.T) {
	results := []diag.Result{
		{Name: "edition", Status: diag.StatusOK, Message: "Windows supports Remote Desktop hosting"},
		{Name: "rdp_enabled", Status: diag.StatusNG, Message: "Remote Desktop is disabled", Hint: "設定からリモートデスクトップを有効にしてください。"},
		{Name: "group", Status: diag.StatusUnknown, Message: "group membership could not be determined", NeedsAdmin: true},
	}

	got := StatusText(results)
	want := readGolden(t, "status.golden")
	if got != want {
		t.Errorf("StatusText mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}
