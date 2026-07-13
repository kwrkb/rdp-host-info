package winsys

const rdpTcpKey = `SYSTEM\CurrentControlSet\Control\Terminal Server\WinStations\RDP-Tcp`

const defaultRDPPort = 3389

// ReadRDPPort は RDP-Tcp の PortNumber を読む。読めない場合は既定値 3389 を返し、
// fromRegistry を false にする。
func ReadRDPPort() (port uint32, fromRegistry bool) {
	v, err := RegReadLocalMachineDWORD(rdpTcpKey, "PortNumber")
	if err != nil {
		return defaultRDPPort, false
	}
	return v, true
}
