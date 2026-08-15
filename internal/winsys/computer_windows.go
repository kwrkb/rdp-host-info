package winsys

import "golang.org/x/sys/windows"

// ComputerName は NetBIOS コンピューター名を返す（windows.ComputerName の
// 中身は GetComputerName で、最大 15 文字。DNS ホスト名ではない）。
//
// GetComputerNameEx(ComputerNamePhysicalDnsHostname) に差し替えてはいけない。
// 戻り値は hostinfo.Classify にも渡っており、ローカルアカウント候補
// `PC名\ユーザー名` の PC 名部と、ドメイン参加判定での LookupAccountSid の
// ドメイン部との比較に使われる。どちらも Windows のローカルアカウント認証局名
// ＝ NetBIOS 名が正しく、DNS ホスト名にすると 15 文字超のホスト名を持つ
// ドメイン参加機でローカルアカウントをドメインアカウントと誤判定する。
//
// 既知の制約: そのため 15 文字超のホスト名の機では、表示される PC 名が
// 切り詰められた NetBIOS 名になる。
func ComputerName() (string, error) {
	return windows.ComputerName()
}
