package winsys

import (
	"fmt"
	"strconv"

	"github.com/kwrkb/rdp-host-info/internal/hostinfo"
)

const currentVersionKey = `SOFTWARE\Microsoft\Windows NT\CurrentVersion`

// homeEditionIDs は RDP ホストとして動作できない Home 系 EditionID。
var homeEditionIDs = map[string]bool{
	"Core":                 true,
	"CoreN":                true,
	"CoreSingleLanguage":   true,
	"CoreCountrySpecific":  true,
	"CoreARM":              true,
	"Home":                 true,
	"HomeN":                true,
	"HomeSingleLanguage":   true,
	"HomeCountrySpecific":  true,
	"HomeBasic":            true,
	"HomeBasicN":           true,
	"HomePremium":          true,
	"HomePremiumN":         true,
}

// ReadEdition はレジストリから Windows エディション情報を読む。
func ReadEdition() (hostinfo.Edition, error) {
	editionID, err := RegReadLocalMachineString(currentVersionKey, "EditionID")
	if err != nil {
		return hostinfo.Edition{}, err
	}

	productName, err := RegReadLocalMachineString(currentVersionKey, "ProductName")
	if err != nil {
		productName = editionID
	}

	displayName := productName
	if buildStr, err := RegReadLocalMachineString(currentVersionKey, "CurrentBuildNumber"); err == nil {
		if build, err := strconv.Atoi(buildStr); err == nil && build >= 22000 {
			displayName = win11DisplayName(productName)
		}
	}

	return hostinfo.Edition{
		ID:              editionID,
		DisplayName:     displayName,
		SupportsRDPHost: !homeEditionIDs[editionID],
	}, nil
}

// win11DisplayName は "Windows 10 Pro" のような表示名を "Windows 11 Pro" に補正する。
// ProductName レジストリ値は Windows 11 でも更新されないことがあるための対応。
func win11DisplayName(productName string) string {
	const prefix = "Windows 10 "
	if len(productName) > len(prefix) && productName[:len(prefix)] == prefix {
		return "Windows 11 " + productName[len(prefix):]
	}
	if productName == "Windows 10" {
		return "Windows 11"
	}
	return fmt.Sprintf("%s (Windows 11)", productName)
}
