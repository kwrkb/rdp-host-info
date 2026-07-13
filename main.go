package main

import (
	"fmt"
	"os"

	"github.com/kwrkb/rdp-host-info/internal/diag"
	"github.com/kwrkb/rdp-host-info/internal/hostinfo"
	"github.com/kwrkb/rdp-host-info/internal/render"
	"github.com/kwrkb/rdp-host-info/internal/winsys"
)

func main() {
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

func exitCode(results []diag.Result) int {
	for _, r := range results {
		if r.Status == diag.StatusNG {
			return 1
		}
	}
	return 0
}
