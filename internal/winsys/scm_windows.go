package winsys

import "golang.org/x/sys/windows"

// IsServiceRunning はサービスの稼働状態を最小権限（SC_MANAGER_CONNECT +
// SERVICE_QUERY_STATUS）で確認する。svc/mgr.Connect は SC_MANAGER_ALL_ACCESS を
// 要求し管理者権限が必要になるため使わない。
func IsServiceRunning(serviceName string) (bool, error) {
	scm, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return false, err
	}
	defer windows.CloseServiceHandle(scm)

	nameUTF16, err := windows.UTF16PtrFromString(serviceName)
	if err != nil {
		return false, err
	}

	svc, err := windows.OpenService(scm, nameUTF16, windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return false, err
	}
	defer windows.CloseServiceHandle(svc)

	var status windows.SERVICE_STATUS
	if err := windows.QueryServiceStatus(svc, &status); err != nil {
		return false, err
	}

	return status.CurrentState == windows.SERVICE_RUNNING, nil
}
