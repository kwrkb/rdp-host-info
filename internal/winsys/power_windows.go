package winsys

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modpowrprof               = windows.NewLazySystemDLL("powrprof.dll")
	procPowerGetActiveScheme  = modpowrprof.NewProc("PowerGetActiveScheme")
	procPowerReadACValueIndex = modpowrprof.NewProc("PowerReadACValueIndex")
	procPowerReadDCValueIndex = modpowrprof.NewProc("PowerReadDCValueIndex")
)

// GUID_SLEEP_SUBGROUP {238C9FA8-0AAD-41ED-83F4-97BE242C8F20}
var guidSleepSubgroup = windows.GUID{
	Data1: 0x238C9FA8, Data2: 0x0AAD, Data3: 0x41ED,
	Data4: [8]byte{0x83, 0xF4, 0x97, 0xBE, 0x24, 0x2C, 0x8F, 0x20},
}

// GUID_STANDBY_TIMEOUT {29F6C1DB-86DA-48C5-9FDB-F2B67B1F44DA}
var guidStandbyTimeout = windows.GUID{
	Data1: 0x29F6C1DB, Data2: 0x86DA, Data3: 0x48C5,
	Data4: [8]byte{0x9F, 0xDB, 0xF2, 0xB6, 0x7B, 0x1F, 0x44, 0xDA},
}

// ReadSleepTimeouts はアクティブな電源プランのスリープタイムアウト（秒、0=なし）を
// AC(電源接続時)/DC(バッテリー時)それぞれ返す。DC 値が読めない場合（デスクトップ機）は
// hasDC=false で続行し、AC 値が読めない場合のみ err を返す。
func ReadSleepTimeouts() (acSeconds, dcSeconds uint32, hasDC bool, err error) {
	var scheme *windows.GUID
	ret, _, _ := procPowerGetActiveScheme.Call(0, uintptr(unsafe.Pointer(&scheme)))
	if ret != 0 || scheme == nil {
		return 0, 0, false, errors.New("PowerGetActiveScheme failed")
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(scheme)))

	ret, _, _ = procPowerReadACValueIndex.Call(
		0,
		uintptr(unsafe.Pointer(scheme)),
		uintptr(unsafe.Pointer(&guidSleepSubgroup)),
		uintptr(unsafe.Pointer(&guidStandbyTimeout)),
		uintptr(unsafe.Pointer(&acSeconds)),
	)
	if ret != 0 {
		return 0, 0, false, errors.New("PowerReadACValueIndex failed")
	}

	ret, _, _ = procPowerReadDCValueIndex.Call(
		0,
		uintptr(unsafe.Pointer(scheme)),
		uintptr(unsafe.Pointer(&guidSleepSubgroup)),
		uintptr(unsafe.Pointer(&guidStandbyTimeout)),
		uintptr(unsafe.Pointer(&dcSeconds)),
	)
	if ret != 0 {
		return acSeconds, 0, false, nil
	}

	return acSeconds, dcSeconds, true, nil
}
