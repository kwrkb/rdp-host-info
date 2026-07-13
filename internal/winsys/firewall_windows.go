package winsys

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

const (
	// INetFwPolicy2 のルールグループをロケール非依存で指す間接文字列
	// （「リモート デスクトップ」グループ）。
	rdpRuleGroup = "@FirewallAPI.dll,-28752"

	hresultSFalse      = 1
	hresultChangedMode = 0x80010106 // RPC_E_CHANGED_MODE
	fwProfileTypesAll  = 0x7FFFFFFF // NET_FW_PROFILE2_ALL
	fwIPProtocolTCP    = 6
	fwRuleDirectionIn  = 1
	fwActionAllow      = 1 // NET_FW_ACTION_ALLOW（0 は NET_FW_ACTION_BLOCK）
)

// QueryRDPFirewall はファイアウォールの事実のみを返す（判定は diag 側）。
// activeProfiles: CurrentProfileTypes のビットマスク。
// ruleEnabled: RDPルールグループ（またはフォールバックで見つけた port 宛の
// 受信許可ルール）が現在のプロファイルで有効か。
// viaFallback: 間接文字列でのグループ判定が失敗し、ルール列挙で判定した。
// go-ole の IDispatch は型ミスマッチが実行時 panic になりうるため、
// 関数全体を recover で保護し error に変換する。
func QueryRDPFirewall(port uint32) (activeProfiles uint32, ruleEnabled bool, viaFallback bool, err error) {
	// STA（COINIT_APARTMENTTHREADED）は初期化・呼び出し・未初期化が同一 OS
	// スレッドで行われることを要求する。goroutine は LockOSThread しない限り
	// 別スレッドへ再スケジュールされうるため、関数全体をスレッド固定する。
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("firewall COM: %v", r)
		}
	}()

	needUninit := true
	if initErr := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); initErr != nil {
		oleErr, ok := initErr.(*ole.OleError)
		switch {
		case ok && oleErr.Code() == hresultSFalse:
			// 既に初期化済み。参照カウントは増えているため CoUninitialize は必要。
		case ok && oleErr.Code() == hresultChangedMode:
			// 別モードで初期化済み。そのまま利用でき、対応する CoUninitialize は不要。
			needUninit = false
		default:
			return 0, false, false, initErr
		}
	}
	if needUninit {
		defer ole.CoUninitialize()
	}

	unknown, err := oleutil.CreateObject("HNetCfg.FwPolicy2")
	if err != nil {
		return 0, false, false, err
	}
	defer unknown.Release()

	policy, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return 0, false, false, err
	}
	defer policy.Release()

	profilesVar, err := oleutil.GetProperty(policy, "CurrentProfileTypes")
	if err != nil {
		return 0, false, false, err
	}
	activeProfiles = uint32(profilesVar.Val)
	profilesVar.Clear()

	// IsRuleGroupCurrentlyEnabled は IDL 上パラメータ付き propget のため、
	// CallMethod ではなく GetProperty で呼ぶ（CallMethod は DISP_E_MEMBERNOTFOUND）。
	groupVar, groupErr := oleutil.GetProperty(policy, "IsRuleGroupCurrentlyEnabled", rdpRuleGroup)
	if groupErr == nil {
		enabled, ok := groupVar.Value().(bool)
		groupVar.Clear()
		if !ok {
			return 0, false, false, errors.New("firewall COM: IsRuleGroupCurrentlyEnabled returned non-bool")
		}
		return activeProfiles, enabled, false, nil
	}

	// 間接文字列が解決できない環境向けフォールバック: ルール列挙で
	// 「現在のプロファイルで有効な、port 宛 TCP 受信許可(Allow)ルール」を探す。
	// ブロックルールは対象外（Action を確認する）。
	enabled, err := hasEnabledInboundTCPRule(policy, port, activeProfiles)
	if err != nil {
		return 0, false, false, err
	}
	return activeProfiles, enabled, true, nil
}

var errRuleFound = errors.New("rule found")

func hasEnabledInboundTCPRule(policy *ole.IDispatch, port, activeProfiles uint32) (bool, error) {
	rulesVar, err := oleutil.GetProperty(policy, "Rules")
	if err != nil {
		return false, err
	}
	defer rulesVar.Clear()

	rules := rulesVar.ToIDispatch()
	if rules == nil {
		return false, errors.New("firewall COM: Rules is not IDispatch")
	}

	err = oleutil.ForEach(rules, func(v *ole.VARIANT) error {
		defer v.Clear()
		rule := v.ToIDispatch()
		if rule == nil {
			return nil
		}
		if ruleMatchesInboundTCP(rule, port, activeProfiles) {
			return errRuleFound
		}
		return nil
	})
	if errors.Is(err, errRuleFound) {
		return true, nil
	}
	return false, err
}

func ruleMatchesInboundTCP(rule *ole.IDispatch, port, activeProfiles uint32) bool {
	boolProp := func(name string) bool {
		v, err := oleutil.GetProperty(rule, name)
		if err != nil {
			return false
		}
		defer v.Clear()
		b, _ := v.Value().(bool)
		return b
	}
	intProp := func(name string) (int64, bool) {
		v, err := oleutil.GetProperty(rule, name)
		if err != nil {
			return 0, false
		}
		defer v.Clear()
		return v.Val, true
	}
	strProp := func(name string) string {
		v, err := oleutil.GetProperty(rule, name)
		if err != nil {
			return ""
		}
		defer v.Clear()
		return v.ToString()
	}

	if !boolProp("Enabled") {
		return false
	}
	// Action を見ないと、ポート宛の有効な受信ブロックルールを
	// 誤って「許可されている」と判定してしまう。
	if action, ok := intProp("Action"); !ok || action != fwActionAllow {
		return false
	}
	if proto, ok := intProp("Protocol"); !ok || proto != fwIPProtocolTCP {
		return false
	}
	if dir, ok := intProp("Direction"); !ok || dir != fwRuleDirectionIn {
		return false
	}
	profiles, ok := intProp("Profiles")
	if !ok {
		return false
	}
	if uint32(profiles) != fwProfileTypesAll && uint32(profiles)&activeProfiles == 0 {
		return false
	}
	return localPortsMatch(strProp("LocalPorts"), port)
}

// localPortsMatch は INetFwRule.LocalPorts（"3389" / "80,3389" / "*" 等）が
// port を含むか調べる。範囲指定（"3000-4000"）は誤検知を避けるため一致扱いしない。
func localPortsMatch(localPorts string, port uint32) bool {
	if localPorts == "*" {
		return true
	}
	want := strconv.Itoa(int(port))
	for p := range strings.SplitSeq(localPorts, ",") {
		if strings.TrimSpace(p) == want {
			return true
		}
	}
	return false
}
