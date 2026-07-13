# rdp-host-info

リモートデスクトップ（RDP）のホスト側 PC で実行し、「接続に必要な情報」と「接続を受け入れられる状態か」を一度に表示する Windows 用 CLI ツール。

接続がうまくいかないとき、原因はたいてい ホスト側の設定（RDP 無効 / ネットワークがパブリック / ユーザー名の形式違い / スリープ）にある。このツールはそれらを診断して人間向けに表示する。**設定の変更は一切行わない**（診断・表示専用）。

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
[OK] User is allowed to connect (Administrators)
[WARN] PC sleeps after 15 minutes
  スリープ中はリモートデスクトップ接続を受け付けられない場合があります。常時接続したい場合は 設定 > システム > 電源 でスリープを「なし」にすることを検討してください。
```

## インストール

Windows 10/11 専用。

```powershell
go install github.com/kwrkb/rdp-host-info@latest
```

またはリポジトリから:

```powershell
git clone https://github.com/kwrkb/rdp-host-info.git
cd rdp-host-info
go build .
```

## 使い方

接続を受ける側（ホスト）の PC で実行するだけ:

```powershell
rdp-host-info
```

- 管理者権限は不要（管理者権限が必要な項目には `(admin required)` が付く）
- exit code: NG が 1 つでもあれば `1`、それ以外は `0`
- `rdp-host-info -version` でバージョン表示

## 出力の読み方

| ラベル | 意味 |
|---|---|
| `[OK]` | 問題なし |
| `[NG]` | 接続を妨げる問題あり。下の行に対処方法を表示 |
| `[WARN]` | 接続できるが注意が必要（スリープ設定など） |
| `[??]` | 確認できなかった（失敗と断定しない）。手動確認方法を表示 |

### ユーザー名の形式

RDP クライアントに入力するユーザー名はアカウント種別で異なる。`Username:` 欄に判定結果を表示する:

| アカウント種別 | 入力形式 |
|---|---|
| ローカルアカウント | `PC名\ユーザー名` |
| Microsoft アカウント | `MicrosoftAccount\メールアドレス` |
| Microsoft Entra ID (Azure AD) | `AzureAD\UPN` |
| ドメイン | `ドメイン名\ユーザー名` |

判定できない場合は候補を複数表示する（断定しない）。Microsoft アカウントでは PIN ではなくアカウントのパスワードが必要。

## 制限事項

- 設定は変更しない。表示された対処方法は手動で実施する
- Microsoft アカウントの判定は非公開レジストリに依存する best-effort（読めない場合は候補を複数提示）
- スリープ判定は従来型スリープのタイムアウト値に基づく。Modern Standby (S0) 搭載機では実際の挙動と異なる場合がある
- ファイアウォール判定は Windows ファイアウォールのみ対象。サードパーティ製セキュリティソフトは別途確認が必要

## 開発

```powershell
go build ./...
go vet ./...
go test ./...
golangci-lint run ./...
```

詳細な設計は `VISION.md`（正典仕様）と `PLAN.md` を参照。
