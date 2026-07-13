package winsys

import "golang.org/x/sys/windows"

// ComputerName は物理 DNS ホスト名を返す。
func ComputerName() (string, error) {
	return windows.ComputerName()
}

// CurrentUserName は現在のプロセストークンのユーザー名を DOMAIN\User 形式で返す。
func CurrentUserName() (string, error) {
	token := windows.GetCurrentProcessToken()

	tokenUser, err := token.GetTokenUser()
	if err != nil {
		return "", err
	}

	account, domain, _, err := tokenUser.User.Sid.LookupAccount("")
	if err != nil {
		return "", err
	}

	return domain + `\` + account, nil
}
