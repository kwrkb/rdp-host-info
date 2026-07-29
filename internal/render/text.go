// Package render は hostinfo / diag の結果を VISION.md 準拠の
// 人間向けテキストに整形する。出力の回帰は golden test で検知する。
// 文言解決（言語ごとの表現）は internal/msg に委譲し、この層はレイアウトのみ持つ。
package render

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/kwrkb/rdp-host-info/internal/diag"
	"github.com/kwrkb/rdp-host-info/internal/hostinfo"
	"github.com/kwrkb/rdp-host-info/internal/msg"
	"github.com/kwrkb/rdp-host-info/internal/msgid"
)

func orUnknown(lang msg.Lang, s string) string {
	if s == "" {
		return msg.Format(lang, msgid.Unknown)
	}
	return s
}

func HostInfoText(lang msg.Lang, info hostinfo.HostInfo) string {
	var b strings.Builder

	fmt.Fprintln(&b, msg.Format(lang, msgid.HeaderHostInfo))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, msg.Format(lang, msgid.HeaderPCName))
	fmt.Fprintf(&b, "  %s\n", orUnknown(lang, info.PCName))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, msg.Format(lang, msgid.HeaderWindows))
	fmt.Fprintf(&b, "  %s\n", orUnknown(lang, info.Edition.DisplayName))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, msg.Format(lang, msgid.HeaderConnAddr))
	localIP := msg.Format(lang, msgid.Unknown)
	if len(info.LocalIPv4) > 0 {
		localIP = info.LocalIPv4[0]
	}
	localLabel := msg.Format(lang, msgid.LabelLocalIP)
	tsLabel := msg.Format(lang, msgid.LabelTailscaleIP)
	addrWidth := max(utf8.RuneCountInString(localLabel), utf8.RuneCountInString(tsLabel))
	fmt.Fprintf(&b, "  %-*s  %s\n", addrWidth, localLabel, localIP)
	fmt.Fprintf(&b, "  %-*s  %s\n", addrWidth, tsLabel, orUnknown(lang, info.TailscaleIP))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, msg.Format(lang, msgid.HeaderUsername))
	writeLogin(&b, lang, info.Login)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, msg.Format(lang, msgid.HeaderRecommended))
	fmt.Fprintf(&b, "  %s\n", orUnknown(lang, info.Recommended))

	return b.String()
}

// writeLogin はユーザー名候補をアカウント種別ラベル付きで揃えて出力する。
// 複数候補は Value を最長幅で左揃えし、注意書きは "* " 付きで続ける。
func writeLogin(b *strings.Builder, lang msg.Lang, login hostinfo.UserLogin) {
	if len(login.Candidates) == 0 {
		fmt.Fprintln(b, "  "+msg.Format(lang, msgid.Unknown))
		return
	}

	width := 0
	for _, c := range login.Candidates {
		width = max(width, len(c.Value))
	}
	for _, c := range login.Candidates {
		fmt.Fprintf(b, "  %-*s    (%s)\n", width, c.Value, msg.Format(lang, c.Label))
	}
	for _, note := range login.Notes {
		fmt.Fprintf(b, "  * %s\n", msg.Format(lang, note))
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

func StatusText(lang msg.Lang, results []diag.Result) string {
	var b strings.Builder

	fmt.Fprintln(&b, msg.Format(lang, msgid.HeaderStatus))
	fmt.Fprintln(&b)
	for _, r := range results {
		admin := ""
		if r.NeedsAdmin {
			admin = msg.Format(lang, msgid.SuffixAdminRequired)
		}
		message := msg.Format(lang, r.MsgID, r.MsgArgs...)
		fmt.Fprintf(&b, "%s %s%s\n", statusLabel(r.Status), message, admin)
		if r.HintID != "" {
			hint := msg.Format(lang, r.HintID, r.HintArgs...)
			for line := range strings.SplitSeq(hint, "\n") {
				fmt.Fprintf(&b, "  %s\n", line)
			}
		}
	}

	return b.String()
}
