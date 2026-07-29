package main

import (
	"errors"
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
	// flag.CommandLine（ExitOnError）は使わない。-help やパースエラーが
	// Parse 内で即 exit してしまい、-lang の検証がその後段では素通りになる
	// （Codex 指摘: -lang xx -help / -version -lang xx が exit 0 になる不具合）。
	// ContinueOnError にして、-lang 検証を help/version/エラー処理より先に行う。
	fs := flag.NewFlagSet("rdp-host-info", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {} // ErrHelp 時の自動 usage 呼び出しを抑止し、出力タイミングを自前で制御する
	langFlag := fs.String("lang", string(msg.English), `output language: "en" or "ja"`)
	showVersion := fs.Bool("version", false, "print version and exit")
	parseErr := fs.Parse(os.Args[1:])

	lang := msg.Lang(*langFlag)
	if !msg.ValidLang(lang) {
		fmt.Fprintf(os.Stderr, "invalid -lang value: %q (expected \"en\" or \"ja\")\n", *langFlag)
		os.Exit(2)
	}

	if parseErr != nil {
		if errors.Is(parseErr, flag.ErrHelp) {
			fmt.Print(msg.Format(lang, msgid.UsageText))
			return
		}
		fmt.Fprint(os.Stderr, msg.Format(lang, msgid.UsageText))
		os.Exit(2)
	}

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
