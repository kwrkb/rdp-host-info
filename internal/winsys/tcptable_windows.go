package winsys

import (
	"encoding/binary"
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modiphlpapi           = windows.NewLazySystemDLL("iphlpapi.dll")
	procGetExtendedTCPTable = modiphlpapi.NewProc("GetExtendedTcpTable")
)

const (
	afInet                   = 2
	tcpTableOwnerPidListener = 3
)

// tcpTableRow はネイティブ構造体 MIB_TCPROW_OWNER_PID と同じレイアウト
// (すべて uint32 のフィールド。LocalPort/RemotePort はネットワークバイトオーダーの
// 上位16bitに格納される)。
type tcpTableRow struct {
	State      uint32
	LocalAddr  uint32
	LocalPort  uint32
	RemoteAddr uint32
	RemotePort uint32
	OwningPid  uint32
}

// IsPortListeningTCP4 は指定した TCP ポート(IPv4)が LISTEN 状態か調べる。
func IsPortListeningTCP4(port uint32) (bool, error) {
	rows, err := getExtendedTCPTable()
	if err != nil {
		return false, err
	}
	for _, row := range rows {
		rowPort := binary.BigEndian.Uint16([]byte{byte(row.LocalPort), byte(row.LocalPort >> 8)})
		if uint32(rowPort) == port {
			return true, nil
		}
	}
	return false, nil
}

func getExtendedTCPTable() ([]tcpTableRow, error) {
	var size uint32
	buf := make([]byte, 8)

	for range 5 {
		ret, _, _ := procGetExtendedTCPTable.Call(
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)),
			0, // bOrder
			uintptr(afInet),
			uintptr(tcpTableOwnerPidListener),
			0,
		)

		switch ret {
		case 0: // NO_ERROR
			return parseTCPTable(buf), nil
		case 122: // ERROR_INSUFFICIENT_BUFFER
			buf = make([]byte, size)
			continue
		default:
			return nil, errors.New("GetExtendedTcpTable failed")
		}
	}

	return nil, errors.New("GetExtendedTcpTable: buffer size did not converge")
}

func parseTCPTable(buf []byte) []tcpTableRow {
	if len(buf) < 4 {
		return nil
	}
	numEntries := binary.LittleEndian.Uint32(buf[0:4])

	const rowSize = 24 // 6 * uint32
	rows := make([]tcpTableRow, 0, numEntries)

	offset := 4
	for range numEntries {
		if offset+rowSize > len(buf) {
			break
		}
		row := tcpTableRow{
			State:      binary.LittleEndian.Uint32(buf[offset : offset+4]),
			LocalAddr:  binary.LittleEndian.Uint32(buf[offset+4 : offset+8]),
			LocalPort:  binary.LittleEndian.Uint32(buf[offset+8 : offset+12]),
			RemoteAddr: binary.LittleEndian.Uint32(buf[offset+12 : offset+16]),
			RemotePort: binary.LittleEndian.Uint32(buf[offset+16 : offset+20]),
			OwningPid:  binary.LittleEndian.Uint32(buf[offset+20 : offset+24]),
		}
		rows = append(rows, row)
		offset += rowSize
	}

	return rows
}
