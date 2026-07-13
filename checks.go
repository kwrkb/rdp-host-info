package main

import (
	"github.com/kwrkb/rdp-host-info/internal/diag"
	"github.com/kwrkb/rdp-host-info/internal/winsys"
)

func buildChecks() []diag.Check {
	return []diag.Check{
		diag.EditionSupportCheck{
			ReadEdition: winsys.ReadEdition,
		},
		diag.RDPEnabledCheck{
			ReadDWORD: winsys.RegReadLocalMachineDWORD,
		},
		diag.ServiceRunningCheck{
			ServiceName:  "TermService",
			DisplayName:  "Remote Desktop Services",
			QueryRunning: winsys.IsServiceRunning,
		},
		diag.PortListeningCheck{
			ReadPort:    winsys.ReadRDPPort,
			IsListening: winsys.IsPortListeningTCP4,
		},
	}
}
