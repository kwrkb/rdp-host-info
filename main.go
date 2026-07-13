package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/kwrkb/rdp-host-info/internal/diag"
	"github.com/kwrkb/rdp-host-info/internal/hostinfo"
	"github.com/kwrkb/rdp-host-info/internal/render"
	"github.com/kwrkb/rdp-host-info/internal/winsys"
)

// version はリリースビルド時に -ldflags "-X main.version=v0.1.0" で上書きする。
var version = "dev"

const usageText = `rdp-host-info - show RDP connection info and check host readiness (Windows)

Usage: rdp-host-info [options]

Run on the PC that accepts Remote Desktop connections (the host).
Prints the connection info a client needs (PC name, IP addresses,
username format by account type) and diagnoses whether the host can
accept connections, as [OK]/[NG]/[WARN]/[??] lines with fix hints.

Diagnosis only: never changes any Windows settings.
Admin rights are not required; checks that need them are labeled
"(admin required)". Exits 1 if any check is [NG], otherwise 0.

Options:
  -help       print this help
  -version    print version and exit
`

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprint(flag.CommandLine.Output(), usageText)
	}
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("rdp-host-info " + resolveVersion())
		return
	}
	info := hostinfo.Collect(hostinfo.Providers{
		ComputerName: winsys.ComputerName,
		Edition:      winsys.ReadEdition,
		LocalIPv4:    winsys.LocalIPv4,
		TailscaleIP:  winsys.TailscaleIP,
		Account:      winsys.CurrentAccount,
	})

	checks := buildChecks()
	results := diag.RunAll(checks)

	fmt.Print(render.HostInfoText(info))
	fmt.Println()
	fmt.Print(render.StatusText(results))

	os.Exit(exitCode(results))
}

// resolveVersion は ldflags 未指定時（go install 等）にビルド情報から
// モジュールバージョンを補う。
func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

func exitCode(results []diag.Result) int {
	for _, r := range results {
		if r.Status == diag.StatusNG {
			return 1
		}
	}
	return 0
}
