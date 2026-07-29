// Package msg resolves msgid.ID values to language-specific text. It is the
// only package that knows what output language was selected; diag and
// hostinfo return IDs and raw values without ever choosing wording.
package msg

import (
	"fmt"
	"strconv"

	"github.com/kwrkb/rdp-host-info/internal/msgid"
)

type Lang string

const (
	English  Lang = "en"
	Japanese Lang = "ja"
)

func ValidLang(l Lang) bool {
	return l == English || l == Japanese
}

type entry struct {
	en string
	ja string
}

var catalog = map[msgid.ID]entry{
	msgid.InternalError:     {"internal error", "内部エラー"},
	msgid.InternalErrorHint: {"This item could not be checked. Please verify manually.", "この項目は確認できませんでした。手動で確認してください。"},

	msgid.EditionUnknown:         {"Windows edition could not be determined", "Windows のエディションを確認できませんでした"},
	msgid.EditionUnknownHint:     {"Check Settings > System > About for your edition.", "設定 > システム > バージョン情報 でエディションを確認してください。"},
	msgid.EditionUnsupported:     {"%s does not support Remote Desktop hosting", "%s はリモートデスクトップのホストとして使用できません"},
	msgid.EditionUnsupportedHint: {"Windows Home cannot host Remote Desktop. Upgrade to Pro/Enterprise/Education.", "Windows Home エディションはリモートデスクトップのホストになれません。Pro/Enterprise/Education への変更が必要です。"},
	msgid.EditionSupported:       {"Windows supports Remote Desktop hosting (%s)", "Windows はリモートデスクトップのホストに対応しています（%s）"},

	msgid.RDPEnabledUnknown:     {"Remote Desktop enabled state could not be determined", "リモートデスクトップの有効状態を確認できませんでした"},
	msgid.RDPEnabledUnknownHint: {"Check Settings > System > Remote Desktop.", "設定 > システム > リモートデスクトップ で状態を確認してください。"},
	msgid.RDPDisabled:           {"Remote Desktop is disabled", "リモートデスクトップが無効です"},
	msgid.RDPDisabledHint:       {"Enable Remote Desktop from Windows Settings.\nSettings > System > Remote Desktop", "Windows の設定からリモートデスクトップを有効にしてください。\n設定 > システム > リモートデスクトップ"},
	msgid.RDPEnabled:            {"Remote Desktop is enabled", "リモートデスクトップが有効です"},

	msgid.ServiceUnknown:        {"%s status could not be determined", "%s の状態を確認できませんでした"},
	msgid.ServiceUnknownHint:    {"Service status could not be checked. Please verify with services.msc.", "サービスの状態を確認できませんでした。services.msc で確認してください。"},
	msgid.ServiceNotRunning:     {"%s is not running", "%s が実行されていません"},
	msgid.ServiceNotRunningHint: {"Start %s from services.msc.", "services.msc から %s を開始してください。"},
	msgid.ServiceRunning:        {"%s is running", "%s が実行されています"},

	msgid.FirewallUnknown:           {"Windows Firewall state could not be determined", "Windows ファイアウォールの状態を確認できませんでした"},
	msgid.FirewallUnknownHint:       {"Check the Remote Desktop inbound rule in wf.msc (Windows Firewall with Advanced Security).", "wf.msc（セキュリティが強化された Windows ファイアウォール）で「リモート デスクトップ」の受信ルールを確認してください。"},
	msgid.FirewallOK:                {"Windows Firewall allows Remote Desktop (%s profile, active)", "Windows ファイアウォールがリモートデスクトップを許可しています（%s プロファイル、有効）"},
	msgid.FirewallOKFallback:        {"Windows Firewall allows Remote Desktop (%s profile, active, port rule)", "Windows ファイアウォールがリモートデスクトップを許可しています（%s プロファイル、有効、ポートルール）"},
	msgid.FirewallPublicBlocked:     {"Network is set to Public and Remote Desktop is not allowed", "ネットワークが「パブリック」に設定されており、リモートデスクトップが許可されていません"},
	msgid.FirewallPublicBlockedHint: {"The firewall's Remote Desktop allowance does not apply to the current network. Change the network to Private under Settings > Network & internet.", "ファイアウォールのリモートデスクトップ許可が、現在のネットワークに適用されていません。設定 > ネットワークとインターネット でネットワークを「プライベート」に変更してください。"},
	msgid.FirewallBlocked:           {"Windows Firewall blocks Remote Desktop (%s profile, active)", "Windows ファイアウォールがリモートデスクトップをブロックしています（%s プロファイル、有効）"},
	msgid.FirewallBlockedHint:       {"Enable the Remote Desktop inbound rule for the current profile in wf.msc. If you use third-party security software, check its settings too.", "wf.msc で「リモート デスクトップ」の受信ルールを現在のプロファイルで有効にしてください。サードパーティ製セキュリティソフト使用時は、そちらの設定も確認してください。"},

	msgid.PortUnknown:             {"TCP %d listening state could not be determined", "TCP %d の待ち受け状態を確認できませんでした"},
	msgid.PortUnknownHint:         {"Run netstat -an in Command Prompt and check the port's state.", "コマンドプロンプトで netstat -an を実行し、該当ポートの状態を確認してください。"},
	msgid.PortNotListening:        {"TCP %d is not listening", "TCP %d が待ち受けていません"},
	msgid.PortNotListeningAssumed: {"TCP %d (assumed default) is not listening", "TCP %d（既定値と仮定）が待ち受けていません"},
	msgid.PortNotListeningHint:    {"Check whether Remote Desktop Services is running and Remote Desktop is enabled.", "Remote Desktop Services が起動しているか、リモートデスクトップが有効になっているか確認してください。"},
	msgid.PortListening:           {"TCP %d is listening", "TCP %d が待ち受けています"},
	msgid.PortListeningAssumed:    {"TCP %d (assumed default) is listening", "TCP %d（既定値と仮定）が待ち受けています"},

	msgid.GroupUnknown:       {"Group membership could not be determined", "グループ所属を確認できませんでした"},
	msgid.GroupUnknownHint:   {"Check in lusrmgr.msc or Settings > System > Remote Desktop whether the user is a member of Remote Desktop Users or Administrators.", "lusrmgr.msc または 設定 > システム > リモートデスクトップ で、ユーザーが Remote Desktop Users か Administrators に含まれるか確認してください。"},
	msgid.GroupNotMember:     {"User is not in Remote Desktop Users or Administrators", "ユーザーが Remote Desktop Users にも Administrators にも含まれていません"},
	msgid.GroupNotMemberHint: {"Add the connecting user from Settings > System > Remote Desktop > Remote Desktop users.", "設定 > システム > リモートデスクトップ > リモートデスクトップユーザー から、接続に使うユーザーを追加してください。"},
	msgid.GroupOKAdmin:       {"User is a member of a group allowed to connect (Administrators)", "ユーザーは接続を許可されたグループに所属しています（Administrators）"},
	msgid.GroupOKRDU:         {"User is a member of a group allowed to connect (Remote Desktop Users)", "ユーザーは接続を許可されたグループに所属しています（Remote Desktop Users）"},
	msgid.GroupOKBoth:        {"User is a member of a group allowed to connect (Administrators, Remote Desktop Users)", "ユーザーは接続を許可されたグループに所属しています（Administrators, Remote Desktop Users）"},
	msgid.GroupOKHint:        {`This only checks group membership. If the "Deny log on through Remote Desktop Services" policy denies this user, membership alone won't let them connect. Check secpol.msc > Local Policies > User Rights Assignment.`, "これはグループ所属のみの確認です。「リモート デスクトップ サービスを使ったログオンを拒否する」ポリシーで拒否されている場合は、所属していても接続できません。secpol.msc > ローカル ポリシー > ユーザー権利の割り当て で確認してください。"},

	msgid.SleepUnknown:     {"Sleep settings could not be determined", "スリープ設定を確認できませんでした"},
	msgid.SleepUnknownHint: {"Check sleep settings under Settings > System > Power.", "設定 > システム > 電源 でスリープ設定を確認してください。"},
	msgid.SleepNever:       {"PC does not sleep automatically", "PC は自動的にスリープしません"},
	msgid.SleepWarnBoth:    {"PC sleeps after %s (plugged in) / %s (on battery)", "PC は %s（電源接続時）/ %s（バッテリー時）でスリープします"},
	msgid.SleepWarnAC:      {"PC sleeps after %s", "PC は %s でスリープします"},
	msgid.SleepWarnDC:      {"PC sleeps after %s (on battery)", "PC は %s（バッテリー時）でスリープします"},
	msgid.SleepWarnHint:    {`Remote Desktop connections may be refused while the PC is asleep. For an always-reachable host, consider setting sleep to "Never" under Settings > System > Power.`, "スリープ中はリモートデスクトップ接続を受け付けられない場合があります。常時接続したい場合は 設定 > システム > 電源 でスリープを「なし」にすることを検討してください。"},

	msgid.NoteAzureADUPN:        {"Run whoami /upn on this PC to confirm the exact UPN.", "正確な UPN はこの PC で whoami /upn を実行して確認してください。"},
	msgid.NoteMSAPassword:       {"For a Microsoft account, sign in with the account password, not the PIN.", "Microsoft アカウントでは PIN ではなくアカウントのパスワードでサインインしてください。"},
	msgid.NoteMaybeMSA:          {`If this is a Microsoft account, also try the MicrosoftAccount\email format.`, `Microsoft アカウントの場合は MicrosoftAccount\メールアドレス 形式も試してください。`},
	msgid.LabelLocalAccount:     {"local account", "ローカルアカウント"},
	msgid.LabelDomainAccount:    {"domain account", "ドメインアカウント"},
	msgid.LabelAzureADAccount:   {"Microsoft Entra ID account", "Microsoft Entra ID アカウント"},
	msgid.LabelMicrosoftAccount: {"Microsoft account", "Microsoft アカウント"},

	msgid.HeaderHostInfo:      {"Remote Desktop Host Information", "リモートデスクトップ ホスト情報"},
	msgid.HeaderPCName:        {"PC Name:", "PC 名:"},
	msgid.HeaderWindows:       {"Windows:", "Windows:"},
	msgid.HeaderConnAddr:      {"Connection Address:", "接続先アドレス:"},
	msgid.LabelLocalIP:        {"Local IP:", "ローカル IP:"},
	msgid.LabelTailscaleIP:    {"Tailscale IP:", "Tailscale IP:"},
	msgid.HeaderUsername:      {"Username:", "ユーザー名:"},
	msgid.HeaderRecommended:   {"Recommended:", "推奨:"},
	msgid.HeaderStatus:        {"Remote Desktop Status", "リモートデスクトップ状態"},
	msgid.SuffixAdminRequired: {" (admin required)", "（要管理者権限）"},
	msgid.Unknown:             {"unknown", "不明"},

	msgid.UsageText: {en: usageTextEn, ja: usageTextJa},
}

const usageTextEn = `rdp-host-info - show RDP connection info and check host readiness (Windows)

Usage: rdp-host-info [options]

Run on the PC that accepts Remote Desktop connections (the host).
Prints the connection info a client needs (PC name, IP addresses,
username format by account type) and diagnoses whether the host can
accept connections, as [OK]/[NG]/[WARN]/[??] lines with fix hints.

Diagnosis only: never changes any Windows settings.
Admin rights are not required; checks that need them are labeled
"(admin required)". Exits 1 if any check is [NG], otherwise 0.

Options:
  -help       print this help
  -lang       output language: "en" or "ja" (default "en")
  -version    print version and exit
`

const usageTextJa = `rdp-host-info - RDP接続情報の表示とホストの受け入れ状態の診断（Windows専用）

使い方: rdp-host-info [options]

リモートデスクトップ接続を受け入れる側（ホスト）の PC で実行してください。
接続元が必要とする情報（PC名、IPアドレス、アカウント種別ごとのユーザー名形式）を表示し、
ホストが接続を受け入れられる状態かを [OK]/[NG]/[WARN]/[??] の行と対処方法で診断します。

診断・表示専用: Windows の設定は一切変更しません。
管理者権限は不要です。管理者権限が必要な項目には "(admin required)" が付きます。
[NG] が1つでもあれば終了コード 1、それ以外は 0 です。

オプション:
  -help       このヘルプを表示
  -lang       出力言語: "en" または "ja"（既定 "en"）
  -version    バージョンを表示して終了
`

// Format resolves id to lang's text and formats it with args. Sleep IDs get
// their numeric seconds arguments turned into localized durations first,
// since duration wording (pluralization, unit placement) isn't a plain
// %s substitution.
func Format(lang Lang, id msgid.ID, args ...any) string {
	switch id {
	case msgid.SleepWarnBoth:
		args = []any{duration(lang, args[0]), duration(lang, args[1])}
	case msgid.SleepWarnAC, msgid.SleepWarnDC:
		args = []any{duration(lang, args[0])}
	}

	e, ok := catalog[id]
	if !ok {
		return string(id)
	}
	tmpl := e.en
	if lang == Japanese {
		tmpl = e.ja
	}
	if len(args) == 0 {
		return tmpl
	}
	return fmt.Sprintf(tmpl, args...)
}

// duration formats a seconds value (as passed via Result.MsgArgs, typically
// uint32) into localized prose, e.g. "15 minutes" / "15分" or
// "90 seconds" / "90秒".
func duration(lang Lang, seconds any) string {
	s := toUint32(seconds)
	if lang == Japanese {
		if s%60 == 0 {
			return strconv.Itoa(int(s/60)) + "分"
		}
		return strconv.Itoa(int(s)) + "秒"
	}
	if s%60 == 0 {
		m := s / 60
		if m == 1 {
			return "1 minute"
		}
		return strconv.Itoa(int(m)) + " minutes"
	}
	return strconv.Itoa(int(s)) + " seconds"
}

func toUint32(v any) uint32 {
	switch n := v.(type) {
	case uint32:
		return n
	case int:
		return uint32(n)
	default:
		return 0
	}
}
