# rdp-host-info

![CI](https://github.com/kwrkb/rdp-host-info/actions/workflows/ci.yml/badge.svg)

English | [日本語](README.ja.md)

A Windows CLI tool that runs on the Remote Desktop (RDP) host PC and shows, in one shot, both the information you need to connect and whether the host is currently ready to accept a connection.

When a connection fails, the cause is usually a host-side setting: RDP disabled, network profile set to public, wrong username format, or sleep. This tool diagnoses those and displays the results for humans. **It never changes any settings** — diagnosis and display only.

```
Remote Desktop Host Information

PC Name:
  OMEN16

Windows:
  Windows 11 Pro

Connection Address:
  Local IP:      192.168.1.20
  Tailscale IP:  100.80.10.5

Username:
  OMEN16\yugo    (local account)

Recommended:
  100.80.10.5

Remote Desktop Status

[OK] Windows supports Remote Desktop hosting (Windows 11 Pro)
[OK] Remote Desktop is enabled
[OK] Remote Desktop Services is running
[OK] Windows Firewall allows Remote Desktop (Private profile, active)
[OK] TCP 3389 is listening
[OK] User is a member of a group allowed to connect (Administrators)
  This only checks group membership. If the "Deny log on through Remote Desktop Services" policy denies this user, membership alone won't let them connect. Check secpol.msc > Local Policies > User Rights Assignment.
[WARN] PC sleeps after 15 minutes
  Remote Desktop connections may be refused while the PC is asleep. For an always-reachable host, consider setting sleep to "Never" under Settings > System > Power.
```

## Installation

Windows 10/11 only.

The easiest way is to grab a prebuilt binary from [Releases](https://github.com/kwrkb/rdp-host-info/releases):

1. Download the zip matching your environment from the latest release (e.g. `rdp-host-info_<version>_windows_amd64.zip`)
2. Extract it and run `rdp-host-info.exe`

Or via `go install`:

```powershell
go install github.com/kwrkb/rdp-host-info@latest
```

Or from the repository:

```powershell
git clone https://github.com/kwrkb/rdp-host-info.git
cd rdp-host-info
go build .
```

## Usage

Just run it on the PC that will accept the connection (the host):

```powershell
rdp-host-info
```

- No administrator privileges required (items that need admin are marked `(admin required)`)
- Exit code: `1` if any check is `NG`, otherwise `0`
- `rdp-host-info -version` prints the version
- `rdp-host-info -lang ja` prints the same output in Japanese (default is English)

## Reading the output

| Label | Meaning |
|---|---|
| `[OK]` | No problem |
| `[NG]` | A problem blocks connections; the line below shows how to fix it |
| `[WARN]` | Connections work but need attention (e.g. sleep settings) |
| `[??]` | Could not be determined (not treated as a failure); the line below shows how to check manually |

### Username format

The username you type into the RDP client differs by account type. The `Username:` field shows the detected result:

| Account type | Format |
|---|---|
| Local account | `PCNAME\username` |
| Microsoft account | `MicrosoftAccount\email` |
| Microsoft Entra ID (Azure AD) | `AzureAD\UPN` |
| Domain | `DOMAIN\username` |

When the type can't be determined with confidence, multiple candidates are shown instead of guessing. Microsoft accounts need the account password, not the PIN.

## Limitations

- Never changes settings; any remediation shown must be applied manually
- Microsoft account detection is a best-effort read of an undocumented registry key (falls back to multiple candidates if unreadable)
- Sleep detection is based on the legacy sleep timeout value; machines using Modern Standby (S0) may behave differently than reported
- Firewall checks cover only Windows Firewall; third-party security software needs separate verification

## Development

```powershell
go build ./...
go vet ./...
go test ./...
golangci-lint run ./...
```

See `VISION.md` (canonical spec) and `PLAN.md` for detailed design.

## License

MIT License. See [LICENSE](LICENSE).
