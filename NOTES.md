# NOTES.md — 内部構造メモ

コードを読めば分かるが、読むのが高コストな内部構造の索引。

**正典性はゼロ。** 上書き自由・鮮度非保証で、この文書とコードが食い違ったらコードが正しい。
仕様の正典は VISION.md、フェーズと現在地は PLAN.md、判断の記録は LESSONS.md。

## パッケージ構成と依存方向

```
rdp-host-info/
├── go.mod                  // module github.com/kwrkb/rdp-host-info
├── main.go                 // 配線(DI)と exit code のみ。ロジックを置かない
├── checks.go               // Check 一覧の組み立て（winsys の実装を注入）
└── internal/
    ├── diag/               // OS非依存: Check interface, Status, Result, Runner + 各チェック
    ├── hostinfo/           // OS非依存: HostInfo モデル、アカウント種別→ユーザー名候補生成
    ├── winsys/             // Windows依存コードを全て隔離（_windows.go）
    ├── msgid/              // 文言 ID の定義（msgid.All に全 ID を登録）
    ├── msg/                // ID → 言語別文言のカタログ（msg.Format）
    └── render/             // テキスト出力（golden test 対象）
```

**依存方向**: `diag` / `hostinfo` は `winsys` を import しない。`main.go`（+ `checks.go`）が
winsys の実装を関数型 provider として注入する。文言も同様に `diag` / `hostinfo` は `msg` を
import せず、`msgid.ID` と値だけを返す。

禁止方向は「diag / hostinfo → winsys」と「diag / hostinfo → msg」のみ。逆方向
（winsys → hostinfo / diag）は共有型を返すために許容されている。

**データフロー**: `winsys`（OS からの取得）→ `hostinfo` / `diag`（OS 非依存のロジック・判定）
→ `msg`（言語別の文言解決）→ `render`（整形）→ `main`（配線・exit code）。
exit code は NG が1つでもあれば 1。

`Check` / `Result` / `Status` の型定義はここに複製しない。`internal/diag/check.go` を見ること。
（かつて PLAN.md がこの型を転記しており、i18n 対応で `Message string` → `MsgID`/`MsgArgs` に
変わった後も古い定義が残っていた。転記はしない。）

## 主要データソース

| 項目 | 取得手段 |
|---|---|
| PC名 | `windows.ComputerName()`（`internal/winsys/computer_windows.go`）。中身は `GetComputerName` で **NetBIOS 名**（最大 15 文字）を返す。`GetComputerNameEx(ComputerNamePhysicalDnsHostname)` ではないため、15 文字超のホスト名では実際の DNS 名と食い違う |
| エディション | レジストリ `HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion` の `EditionID`（Home 判定: Core 系）。`ProductName` は Win11 でも "Windows 10" を返すため build ≥ 22000 で表示補正 |
| RDP有効 | `HKLM\SYSTEM\CurrentControlSet\Control\Terminal Server` の `fDenyTSConnections`（0=有効） |
| RDPポート番号 | `...\Terminal Server\WinStations\RDP-Tcp` の `PortNumber`。読めなければ 3389 と仮定し「既定値と仮定」と明記 |
| TermService稼働 | raw SCM API を**最小権限**で: `OpenSCManager(SC_MANAGER_CONNECT)` → `OpenService(SERVICE_QUERY_STATUS)` → `QueryServiceStatusEx`。`svc/mgr.Connect()` は ALL_ACCESS 要求のため使わない |
| ポート待受 | iphlpapi `GetExtendedTcpTable`（自前バインド）。IPv4/IPv6 両方走査、ERROR_INSUFFICIENT_BUFFER 再試行必須 |
| ファイアウォール | `INetFwPolicy2`（go-ole IDispatch 経由）: `CurrentProfileTypes` でアクティブプロファイル取得 + `IsRuleGroupCurrentlyEnabled("@FirewallAPI.dll,-28752")`（ロケール非依存の間接文字列）。Public のみアクティブ+無効なら「ネットワークがパブリック」NG |
| グループ所属 | `GetTokenInformation(TokenGroups)` を well-known SID `S-1-5-32-555`(Remote Desktop Users) / `S-1-5-32-544`(Administrators) と比較。`CheckTokenMembership` は UAC 非昇格時に不正確なため使わない |
| スリープ | powrprof `PowerGetActiveScheme` → `PowerReadACValueIndex` / `PowerReadDCValueIndex`（GUID_SLEEP_SUBGROUP / GUID_STANDBY_TIMEOUT）。0=OK、>0=WARN。取得失敗は黙って Unknown（best-effort） |
| ローカルIPv4 | `net` パッケージで列挙。デフォルトルート側優先は `net.Dial("udp", "8.8.8.8:80")` の LocalAddr で決定（送信なし）。ループバック/リンクローカル/CGNAT 除外 |
| Tailscale IP | CGNAT 100.64.0.0/10 のインターフェース走査（Tailscale CLI 非依存） |
| 現在ユーザー名 | トークン `TokenUser` SID → `LookupAccountSid` |

### アカウント種別判定（dsregcmd / WMI 不使用、確度順）

1. **AzureAD**: ユーザー SID が `S-1-12-1-` で始まる → `AzureAD\UPN`（UPN は `GetUserNameEx(NameUserPrincipal)`）
2. **ドメイン参加**: `NetGetJoinInformation` が `NetSetupDomainName` かつ SID ドメイン部 ≠ コンピュータ名 → `DOMAIN\user`
3. **Microsoftアカウント**: `HKCU\SOFTWARE\Microsoft\IdentityCRL\UserExtendedProperties` のサブキー名（メールアドレス）で判定 → `MicrosoftAccount\email` と `PC名\ユーザー名` の**両候補**を提示し「PIN ではなくパスワードが必要」の注意書き
4. **ローカル**: 上記いずれでもない → `PC名\ユーザー名`

順序を崩すと誤判定する（AzureAD 機はドメイン参加も true になりうる）。3 は非公開レジストリ
依存のため、読めない/曖昧なら AccountType=Unknown とし候補を複数列挙して断定しない。

## テスト戦略

- **単体（OS非依存）**: 各 Check に fake の取得関数を注入し、テーブル駆動で「値0 / 値1 / 値欠落 / アクセス拒否 → OK/NG/Unknown」を全分岐検証
- **Golden output test**: fake providers 一式で `render` の全出力を `testdata/*.golden` と比較。en/ja 両言語分。`go test ./internal/render/... -update` で再生成
- **カタログ完全性**: `internal/msg/catalog_test.go` が `msgid.All` の全 ID を en/ja 両方で検証（`msgid.ID` は文字列型でコンパイラが取りこぼしを検出できないため）
- **winsys**: ロジックを持たせない（変換のみ）。`//go:build windows` の smoke test を少数（error なし・値の形だけ検証）
- **手動マトリクス**: 本機（Win11 Pro）で Settings / コントロールパネルと突合

## 検証手順

1. `go build ./...` / `go vet ./...` / `go test ./...` / `golangci-lint run ./...`
2. 本機で `go run .` → VISION の「想定利用フロー」の出力例と見比べる
3. 非昇格ターミナルで実行し、admin なしで Unknown にならないこと（なる項目は NeedsAdmin ラベルが出ること）を確認
4. `whoami /user`（SID）・`whoami /upn` と Username 候補を突合
5. `-version` / `--help` / `-lang ja` の出力確認
6. **手動残置**。状態を変えて再実行（ツール自身は設定を変更しない）:
   - ネットワークを一時的に「パブリック」へ → firewall が `[NG]` + Hint → 戻す
   - RDP を一時的に無効化 → `[NG]` + exit code 1 → 再有効化
   - Tailscale 停止 → Recommended がローカル IPv4 に切り替わる
   - 昇格ターミナルでも実行し、非昇格と表示差がないこと

## 既知の制約

いずれも実装済みの挙動。詳細な経緯は LESSONS.md。

- **ファイアウォール判定が最も複雑**。間接文字列 `@FirewallAPI.dll,-28752` が見つからない場合は「ルール列挙で LocalPort==設定ポート && TCP && Action==Allow」にフォールバックする。ポート範囲指定ルール（`3000-4000`）は一致扱いしない。サードパーティ AV 環境では実態と乖離しうるため、メッセージに限定を明記している
- **グループ所属は「所属」までしか見ていない**。TokenGroups 列挙方式では「所属しているが LSA ポリシー（deny-logon 権利）で拒否」を検出できない。文言を「グループのメンバーである」に限定し、Hint で `secpol.msc` での手動確認を案内している
- **MSA 判定は非公開レジストリ依存**。壊れたら Unknown + 複数候補提示に退避する（VISION が許容）
- **Modern Standby (S0)** 機では STANDBY_TIMEOUT の意味が従来スリープと異なるため、WARN 文言は断定を避けている
- **`GetExtendedTcpTable` の構造体は自前定義**でバグりやすいため、テストを厚めにしている
- **go-ole の IDispatch は型ミスマッチが実行時エラー**になる。COM 部分は必ず recover / エラー → Unknown 経路を通す
- **PC 名は NetBIOS 名（最大 15 文字）**。`hostinfo.Classify` がローカルアカウント候補の `PC名\ユーザー名` とドメイン参加判定に同じ値を使っており、そこは NetBIOS 名でなければならないため、DNS ホスト名には差し替えられない。15 文字超のホスト名の機では表示名が切り詰められる（`internal/winsys/computer_windows.go` にも同じ注意書きがある）
