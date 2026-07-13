package winsys

import "golang.org/x/sys/windows/registry"

// RegReadLocalMachineDWORD は HKLM 配下の DWORD 値を読む。
func RegReadLocalMachineDWORD(subKey, name string) (uint32, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, subKey, registry.QUERY_VALUE)
	if err != nil {
		return 0, err
	}
	defer k.Close()

	v, _, err := k.GetIntegerValue(name)
	if err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// regListCurrentUserSubKeys は HKCU 配下キーのサブキー名を列挙する。
// キーが存在しない場合は registry.ErrNotExist を返す。
func regListCurrentUserSubKeys(subKey string) ([]string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, subKey, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return nil, err
	}
	defer k.Close()

	return k.ReadSubKeyNames(-1)
}

// RegReadLocalMachineString は HKLM 配下の文字列値を読む。
func RegReadLocalMachineString(subKey, name string) (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, subKey, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	v, _, err := k.GetStringValue(name)
	if err != nil {
		return "", err
	}
	return v, nil
}
