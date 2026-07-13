// Package winsys は Windows からの情報取得を隔離する層。ロジックを持たず
// 「取得と型変換のみ」に留める（判定は diag / hostinfo 側）。Windows 依存
// コードは *_windows.go に置く。本ファイルのみ net パッケージだけで実装
// されており OS 非依存。
package winsys

import (
	"errors"
	"net"
	"net/netip"
)

var tailscaleCGNAT = netip.MustParsePrefix("100.64.0.0/10")

func isUsableIPv4(ip net.IP) bool {
	v4 := ip.To4()
	if v4 == nil {
		return false
	}
	if v4.IsLoopback() || v4.IsLinkLocalUnicast() || v4.IsUnspecified() {
		return false
	}
	return true
}

// LocalIPv4 は利用可能なローカル IPv4 アドレスを列挙する。
// デフォルトルート側のアドレスが分かればそれを先頭にする。
func LocalIPv4() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, a := range addrs {
		ipNet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		if !isUsableIPv4(ipNet.IP) {
			continue
		}
		ip := ipNet.IP.To4()
		if tailscaleCGNAT.Contains(netip.AddrFrom4([4]byte(ip))) {
			continue
		}
		ips = append(ips, ip.String())
	}
	if len(ips) == 0 {
		return nil, errors.New("no usable local IPv4 address found")
	}

	if preferred, err := defaultRouteIPv4(); err == nil {
		for i, ip := range ips {
			if ip == preferred {
				ips[0], ips[i] = ips[i], ips[0]
				break
			}
		}
	}

	return ips, nil
}

// defaultRouteIPv4 はパケットを送信せずに、既定ルートで使われるローカルアドレスを求める。
func defaultRouteIPv4() (string, error) {
	conn, err := net.Dial("udp4", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return "", errors.New("unexpected local addr type")
	}
	return localAddr.IP.String(), nil
}

// TailscaleIP は CGNAT レンジ(100.64.0.0/10)にあるインターフェースアドレスを探す。
func TailscaleIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, a := range addrs {
		ipNet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		ip := ipNet.IP.To4()
		if ip == nil {
			continue
		}
		if tailscaleCGNAT.Contains(netip.AddrFrom4([4]byte(ip))) {
			return ip.String(), nil
		}
	}
	return "", errors.New("no tailscale IP found")
}
