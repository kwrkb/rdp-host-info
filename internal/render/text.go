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
	fmt.Fprintln(&b, "Windows:")
	fmt.Fprintf(&b, "  %s\n", orUnknown(info.Edition.DisplayName))
	fmt.Fprintln(&b, "Connection Address:")
	localIP := "unknown"
	if len(info.LocalIPv4) > 0 {
		localIP = info.LocalIPv4[0]
	}
	fmt.Fprintf(&b, "  Local IP:      %s\n", localIP)
	fmt.Fprintf(&b, "  Tailscale IP:  %s\n", orUnknown(info.TailscaleIP))
	fmt.Fprintln(&b, "Username:")
	fmt.Fprintf(&b, "  %s\n", orUnknown(info.UserName))
	fmt.Fprintln(&b, "Recommended:")
	fmt.Fprintf(&b, "  %s\n", orUnknown(info.Recommended))

	return b.String()
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
			for _, line := range strings.Split(r.Hint, "\n") {
				fmt.Fprintf(&b, "  %s\n", line)
			}
		}
	}

	return b.String()
}
