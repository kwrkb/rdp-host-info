package hostinfo

import "strings"

type AccountType int

const (
	AccountUnknown AccountType = iota
	AccountLocal
	AccountDomain
	AccountAzureAD
	AccountMicrosoft
)

// UserCandidate は接続元で入力するユーザー名の候補 1 件。
type UserCandidate struct {
	Value string // 例: `OMEN16\yugo`, `MicrosoftAccount\user@example.com`
	Label string // 例: "local account", "Microsoft account"
}

// UserLogin はアカウント種別判定の結果。Candidates は確度順。
type UserLogin struct {
	AccountType AccountType
	Candidates  []UserCandidate
	Notes       []string
}

// AccountData は winsys が返す生データ。判定ロジックは持たない。
type AccountData struct {
	UserSID      string // SID 文字列（"S-1-5-21-..." / "S-1-12-1-..."）
	Domain       string // LookupAccountSid のドメイン部
	User         string // 同アカウント名部
	UPN          string // GetUserNameEx(NameUserPrincipal)。取得失敗は空
	JoinKnown    bool   // NetGetJoinInformation が成功したか
	DomainJoined bool   // NetSetupDomainName だったか
	JoinedDomain string
	MSAChecked   bool     // IdentityCRL レジストリを読めたか（キー不在も「読めた・0件」）
	MSASubKeys   []string // UserExtendedProperties のサブキー名（生のまま）
}

// azureADSIDPrefix は Entra ID (Azure AD) アカウントの SID プレフィックス。
const azureADSIDPrefix = "S-1-12-1-"

// Classify はアカウント種別を判定し、RDP 接続時に入力すべきユーザー名の
// 候補を確度順で生成する。断定できない場合は複数候補 + 注意書きを返す。
func Classify(pcName string, a AccountData) UserLogin {
	if pcName == "" {
		pcName = a.Domain
	}
	if a.User == "" {
		// provider 失敗。候補を捏造せず unknown 扱いにする。
		return UserLogin{AccountType: AccountUnknown}
	}

	// 1. Entra ID (Azure AD): SID プレフィックスで確定
	if strings.HasPrefix(a.UserSID, azureADSIDPrefix) {
		if a.UPN != "" {
			return UserLogin{
				AccountType: AccountAzureAD,
				Candidates:  []UserCandidate{{Value: `AzureAD\` + a.UPN, Label: "Microsoft Entra ID account"}},
			}
		}
		return UserLogin{
			AccountType: AccountAzureAD,
			Candidates:  []UserCandidate{{Value: `AzureAD\` + a.User, Label: "Microsoft Entra ID account"}},
			Notes:       []string{"正確な UPN はこの PC で whoami /upn を実行して確認してください。"},
		}
	}

	// 2. ドメイン参加: 参加済みかつ SID のドメイン部が PC 名と異なる
	if a.DomainJoined && !strings.EqualFold(a.Domain, pcName) {
		return UserLogin{
			AccountType: AccountDomain,
			Candidates:  []UserCandidate{{Value: a.Domain + `\` + a.User, Label: "domain account"}},
		}
	}

	// 3. Microsoft アカウント: 非公開レジストリのサブキー（メールアドレス）で判定
	if a.MSAChecked {
		emails := msaEmails(a.MSASubKeys)
		if len(emails) > 0 {
			candidates := make([]UserCandidate, 0, len(emails)+1)
			for _, email := range emails {
				candidates = append(candidates, UserCandidate{Value: `MicrosoftAccount\` + email, Label: "Microsoft account"})
			}
			candidates = append(candidates, UserCandidate{Value: pcName + `\` + a.User, Label: "local account"})
			return UserLogin{
				AccountType: AccountMicrosoft,
				Candidates:  candidates,
				Notes:       []string{"Microsoft アカウントでは PIN ではなくアカウントのパスワードでサインインしてください。"},
			}
		}

		// 4. ローカル: MSA の痕跡なし
		return UserLogin{
			AccountType: AccountLocal,
			Candidates:  []UserCandidate{{Value: pcName + `\` + a.User, Label: "local account"}},
		}
	}

	// 5. MSA レジストリを読めなかった: ローカルか MSA か断定できない
	return UserLogin{
		AccountType: AccountUnknown,
		Candidates:  []UserCandidate{{Value: pcName + `\` + a.User, Label: "local account"}},
		Notes:       []string{`Microsoft アカウントの場合は MicrosoftAccount\メールアドレス 形式も試してください。`},
	}
}

// msaEmails はサブキー名からメールアドレスらしいもの（"@" を含む）だけを残す。
func msaEmails(subKeys []string) []string {
	var emails []string
	for _, k := range subKeys {
		if strings.Contains(k, "@") {
			emails = append(emails, k)
		}
	}
	return emails
}
