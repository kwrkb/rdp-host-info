package render

import (
	"fmt"
	"strings"

	"github.com/kwrkb/rdp-host-info/internal/diag"
	"github.com/kwrkb/rdp-host-info/internal/hostinfo"
)

func orUnknown(s string) string {
	if s == "" {
		return "unknown"
	}
	return s
}

func HostInfoText(info hostinfo.HostInfo) string {
	var b strings.Builder

	fmt.Fprintln(&b, "Remote Desktop Host Information")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "PC Name:")
	fmt.Fprintf(&b, "  %s\n", orUnknown(info.PCName))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Windows:")
	fmt.Fprintf(&b, "  %s\n", orUnknown(info.Edition.DisplayName))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Connection Address:")
	localIP := "unknown"
	if len(info.LocalIPv4) > 0 {
		localIP = info.LocalIPv4[0]
	}
	fmt.Fprintf(&b, "  Local IP:      %s\n", localIP)
	fmt.Fprintf(&b, "  Tailscale IP:  %s\n", orUnknown(info.TailscaleIP))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Username:")
	writeLogin(&b, info.Login)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Recommended:")
	fmt.Fprintf(&b, "  %s\n", orUnknown(info.Recommended))

	return b.String()
}

// writeLogin はユーザー名候補をアカウント種別ラベル付きで揃えて出力する。
// 複数候補は Value を最長幅で左揃えし、注意書きは "* " 付きで続ける。
func writeLogin(b *strings.Builder, login hostinfo.UserLogin) {
	if len(login.Candidates) == 0 {
		fmt.Fprintln(b, "  unknown")
		return
	}

	width := 0
	for _, c := range login.Candidates {
		width = max(width, len(c.Value))
	}
	for _, c := range login.Candidates {
		fmt.Fprintf(b, "  %-*s    (%s)\n", width, c.Value, c.Label)
	}
	for _, note := range login.Notes {
		fmt.Fprintf(b, "  * %s\n", note)
	}
}

func statusLabel(s diag.Status) string {
	switch s {
	case diag.StatusOK:
		return "[OK]"
	case diag.StatusNG:
		return "[NG]"
	case diag.StatusWarn:
		return "[WARN]"
	default:
		return "[??]"
	}
}

func StatusText(results []diag.Result) string {
	var b strings.Builder

	fmt.Fprintln(&b, "Remote Desktop Status")
	fmt.Fprintln(&b)
	for _, r := range results {
		admin := ""
		if r.NeedsAdmin {
			admin = " (admin required)"
		}
		fmt.Fprintf(&b, "%s %s%s\n", statusLabel(r.Status), r.Message, admin)
		if r.Hint != "" {
			for line := range strings.SplitSeq(r.Hint, "\n") {
				fmt.Fprintf(&b, "  %s\n", line)
			}
		}
	}

	return b.String()
}
