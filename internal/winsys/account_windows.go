package winsys

import (
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	"github.com/kwrkb/rdp-host-info/internal/hostinfo"
)

// Microsoft アカウントのメールアドレスがサブキー名として現れる非公開レジストリキー。
const msaUserExtendedProperties = `SOFTWARE\Microsoft\IdentityCRL\UserExtendedProperties`

// CurrentAccount はアカウント種別判定に必要な生データを収集する。
// 判定ロジックは持たず hostinfo.Classify 側で行う。
// SID（LookupAccount）の取得失敗のみを error とし、UPN / ドメイン参加情報 /
// MSA レジストリの失敗はゼロ値（JoinKnown=false / MSAChecked=false 等)の
// まま続行する（best-effort）。
func CurrentAccount() (hostinfo.AccountData, error) {
	sid, domain, user, err := currentUserAccount()
	if err != nil {
		return hostinfo.AccountData{}, err
	}

	data := hostinfo.AccountData{
		UserSID: sid,
		Domain:  domain,
		User:    user,
	}

	if upn, err := userPrincipalName(); err == nil {
		data.UPN = upn
	}

	if domainJoined, name, err := joinInformation(); err == nil {
		data.JoinKnown = true
		data.DomainJoined = domainJoined
		data.JoinedDomain = name
	}

	switch subKeys, err := regListCurrentUserSubKeys(msaUserExtendedProperties); err {
	case nil:
		data.MSAChecked = true
		data.MSASubKeys = subKeys
	case registry.ErrNotExist:
		// キー不在は「MSA サインインの痕跡なし」という確定情報。
		data.MSAChecked = true
	}

	return data, nil
}

// currentUserAccount は現在のプロセストークンのユーザー SID 文字列と
// ドメイン / アカウント名を返す。
func currentUserAccount() (sid, domain, user string, err error) {
	token := windows.GetCurrentProcessToken()

	tokenUser, err := token.GetTokenUser()
	if err != nil {
		return "", "", "", err
	}

	account, domain, _, err := tokenUser.User.Sid.LookupAccount("")
	if err != nil {
		return "", "", "", err
	}

	return tokenUser.User.Sid.String(), domain, account, nil
}

// userPrincipalName は GetUserNameEx(NameUserPrincipal) で UPN を返す。
// 非ドメイン・非 Entra ID 環境では ERROR_NONE_MAPPED で失敗する（正常系）。
func userPrincipalName() (string, error) {
	size := uint32(256)
	for range 2 {
		buf := make([]uint16, size)
		err := windows.GetUserNameEx(windows.NameUserPrincipal, &buf[0], &size)
		if err == nil {
			return windows.UTF16ToString(buf[:size]), nil
		}
		if err != windows.ERROR_MORE_DATA {
			return "", err
		}
		// size に必要長が入っているので再確保して 1 回だけ再試行する。
	}
	return "", windows.ERROR_MORE_DATA
}

// joinInformation は NetGetJoinInformation でドメイン参加状態を返す。
// name はドメイン参加時は参加ドメイン名、それ以外はワークグループ名。
func joinInformation() (domainJoined bool, name string, err error) {
	var buf *uint16
	var joinType uint32
	if err := windows.NetGetJoinInformation(nil, &buf, &joinType); err != nil {
		return false, "", err
	}
	defer windows.NetApiBufferFree((*byte)(unsafe.Pointer(buf)))

	return joinType == windows.NetSetupDomainName, windows.UTF16PtrToString(buf), nil
}
