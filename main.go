package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/kwrkb/rdp-host-info/internal/diag"
	"github.com/kwrkb/rdp-host-info/internal/hostinfo"
	"github.com/kwrkb/rdp-host-info/internal/msg"
	"github.com/kwrkb/rdp-host-info/internal/msgid"
	"github.com/kwrkb/rdp-host-info/internal/render"
	"github.com/kwrkb/rdp-host-info/internal/winsys"
)

// version はリリースビルド時に -ldflags "-X main.version=v0.1.0" で上書きする。
var version = "dev"

func main() {
	langFlag := flag.String("lang", string(msg.English), `output language: "en" or "ja"`)
	flag.Usage = func() {
		lang := msg.Lang(*langFlag)
		if !msg.ValidLang(lang) {
			lang = msg.English
		}
		_, _ = fmt.Fprint(flag.CommandLine.Output(), msg.Format(lang, msgid.UsageText))
	}
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("rdp-host-info " + resolveVersion())
		return
	}

	lang := msg.Lang(*langFlag)
	if !msg.ValidLang(lang) {
		fmt.Fprintf(os.Stderr, "invalid -lang value: %q (expected \"en\" or \"ja\")\n", *langFlag)
		os.Exit(2)
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

	fmt.Print(render.HostInfoText(lang, info))
	fmt.Println()
	fmt.Print(render.StatusText(lang, results))

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
