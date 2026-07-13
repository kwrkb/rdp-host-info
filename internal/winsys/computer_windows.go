package winsys

import "golang.org/x/sys/windows"

// ComputerName は物理 DNS ホスト名を返す。
func ComputerName() (string, error) {
	return windows.ComputerName()
}
