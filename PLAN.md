# PLAN.md — rdp-host-info 実装計画

VISION.md を正典仕様とする。実装中に矛盾があれば VISION.md を優先し、本ファイルを更新する。

## ゴール

VISION.md の MVP を Go CLI として実装する。

1. **接続情報の表示** — PC名 / Windowsエディション / ローカルIPv4 / Tailscale IP / RDP接続用ユーザー名（アカウント種別に応じた形式）/ 推奨接続先
2. **接続受け入れ状態の確認** — エディション対応 / RDP有効 / TermService稼働 / ファイアウォール（アクティブプロファイル）/ ポート待受 / グループ所属 / スリープ設定
3. **人間向け出力** — `[OK]/[NG]/[WARN]/[??]` + 問題項目への短い説明（Hint）

## 技術方針（要約）

- **ロケール依存のコマンド出力を文字列パースしない**。一次情報源はレジストリ / Windows API / 既知SID / GUID
- 取得失敗は成功とも失敗とも偽らず **Unknown** として扱う
- 管理者権限なしで動作することを基本とする（必要な項目は明示）
- 依存は 2 つのみ: `golang.org/x/sys`、`github.com/go-ole/go-ole`（firewall COM 用）。cobra / testify 等は不採用
- 断定できない情報（ユーザー名形式など）は候補として複数提示する

## ディレクトリ構成

```
rdp-host-info/
├── go.mod                  // module github.com/kwrkb/rdp-host-info
├── main.go                 // 配線(DI)と exit code のみ。ロジックを置かない
└── internal/
    ├── diag/               // OS非依存: Check interface, Status, Result, Runner + 各チェック
    ├── hostinfo/           // OS非依存: HostInfo モデル、アカウント種別→ユーザー名候補生成
    ├── winsys/             // Windows依存コードを全て隔離（_windows.go）
    └── render/             // テキスト出力（golden test 対象）
```

**依存方向**: `diag` / `hostinfo` は `winsys` を import しない。`main.go` が winsys の実装を関数フィールド/小さな interface として注入する。将来の JSON 出力 / GUI は `render` の追加のみで対応。

### コア抽象

```go
type Status int // StatusOK / StatusNG / StatusWarn / StatusUnknown

type Result struct {
    Status  Status
    Message string // 一行の状態説明
    Hint    string // NG/WARN 時の対処ヒント（空なら省略）
}

type Check interface {
    Name() string     // 安定した識別子（将来の JSON key）
    NeedsAdmin() bool // true なら出力に明示
    Run(ctx context.Context) Result
}
```

取得エラーは必ず `StatusUnknown` に落とし、Hint に手動確認方法を書く。panic 禁止。

## 主要データソース

| 項目 | 取得手段 |
|---|---|
| PC名 | `windows.GetComputerNameEx(ComputerNamePhysicalDnsHostname)` |
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

3 は非公開レジストリ依存のため、読めない/曖昧なら AccountType=Unknown とし候補を複数列挙して断定しない。

## フェーズ分割

- [x] **Phase 0 — Scaffold**
  - [x] `go mod init github.com/kwrkb/rdp-host-info`
  - [x] `diag`（Check/Status/Result/Runner）、`render`（テキスト整形）
  - [x] ダミーチェック 1 個で `go run .` が end-to-end で動く
  - [x] golden test の器を作る
- [x] **Phase 1 — 接続情報**
  - [x] PCName / Edition / LocalIPv4 / TailscaleIP / 現在ユーザー名
  - [x] netinfo（CGNAT 判定含む）は純 Go なのでこの時点でユニットテスト
- [x] **Phase 2 — 簡単なチェック群**（レジストリ/単純 API、低リスク）
  - [x] エディション対応チェック
  - [x] RDP 有効（fDenyTSConnections）
  - [x] ポート待受（PortNumber + TCP テーブル）
  - [x] TermService 稼働
- [x] **Phase 3 — 難しいチェック群**
  - [x] ファイアウォール（COM、失敗時 Unknown フォールバックを最初から実装）
  - [x] グループ所属（トークン）
  - [x] スリープ（電源 API）
- [x] **Phase 4 — アカウント種別 + ユーザー名形式候補**
  - [x] 種別判定ロジック
  - [x] 候補生成（最も曖昧な領域。テストを厚くする）
- [x] **Phase 5 — 仕上げ**
  - [x] Recommended 判定（Tailscale IP 優先）
  - [x] Hint 文言を VISION の例文に揃える
  - [x] `--version`、NeedsAdmin ラベル表示、README
- [x] **Phase 6 — 品質**
  - [x] golden test 網羅
  - [x] `go vet` / `golangci-lint`
  - [x] 実機マトリクス検証（読み取り専用項目。状態変更を伴う項目は下記「検証手順」6 に残置）

各フェーズ末で `go build ./... && go vet ./... && go test ./...` を通す。

## テスト戦略

- **単体（OS非依存）**: 各 Check に fake の取得関数を注入し、テーブル駆動で「値0 / 値1 / 値欠落 / アクセス拒否 → OK/NG/Unknown」を全分岐検証
- **Golden output test**: fake providers 一式で `render` の全出力を `testdata/*.golden` と比較
- **winsys**: ロジックを持たせない（変換のみ）。`//go:build windows` の smoke test を少数（error なし・値の形だけ検証）
- **手動マトリクス**: 本機（Win11 Pro）で Settings / コントロールパネルと突合

## 検証手順

1. `go build ./...` / `go vet ./...` / `go test ./...` / `golangci-lint run ./...`
2. 本機で `go run .` → VISION の「想定利用フロー」の出力例と見比べる
3. 非昇格ターミナルで実行し、admin なしで Unknown にならないこと（なる項目は NeedsAdmin ラベルが出ること）を確認
4. `whoami /user`（SID）・`whoami /upn` と Username 候補を突合
5. `-version` / `--help` の出力確認
6. 状態を変えて再実行（手動、ツール自身は設定を変更しない）:
   - ネットワークを一時的に「パブリック」へ → firewall が `[NG]` + Hint → 戻す
   - RDP を一時的に無効化 → `[NG]` + exit code 1 → 再有効化
   - Tailscale 停止 → Recommended がローカル IPv4 に切り替わる
   - 昇格ターミナルでも実行し、非昇格と表示差がないこと

## リスク

- **firewall COM が最大の複雑性**。間接文字列 `@FirewallAPI.dll,-28752` が見つからない場合は「ルール列挙で LocalPort==設定ポート && TCP」にフォールバック。サードパーティ AV 環境では実態と乖離しうる → メッセージに限定を明記
- **MSA 判定は非公開レジストリ依存**。壊れたら Unknown + 複数候補提示に退避（VISION が許容）
- **Administrators 判定**: TokenGroups 列挙方式でも「所属しているが LSA ポリシーで拒否」は検出不能 → 文言を「グループ所属」に限定
- **Modern Standby (S0)** 機では STANDBY_TIMEOUT の意味が従来スリープと異なる → WARN 文言は断定を避ける
- **GetExtendedTcpTable の構造体自前定義**はバグりやすい → テストを厚めに
- go-ole の IDispatch は型ミスマッチが実行時エラー → COM 部分は必ず recover/エラー → Unknown 経路を通す

## 進捗メモ

- **Phase 6 完了 = MVP 完成**
  - status golden 3 件追加（all_ok / ng / all_unknown）。実 Check（fake provider 注入）→ `RunAll` → `StatusText` の end-to-end で文言の回帰を検知する形にした
  - golangci-lint v2 導入（winget）。`.golangci.yml` は govet/staticcheck/errcheck/unused/ineffassign + gofmt。defer での後始末（handle Close/Free）は errcheck 除外
  - 既存の gofmt 未整形 3 ファイル（rdpenabled/edition_windows/tcptable_windows）を整形
  - 実機マトリクスの読み取り専用項目を検証済み（SID/UPN 突合、非昇格実行、-version/--help）。ネットワーク切替・RDP 無効化・Tailscale 停止・昇格実行は手動残置（検証手順 6）
- **Phase 5 完了**
  - Recommended（Tailscale 優先）と NeedsAdmin ラベルは Phase 1〜3 で実装済みだったためチェックのみ更新
  - rdp_enabled の NG Hint を VISION の 2 行例文に揃えた。firewall の NG Message は英語 Message / 日本語 Hint の規約を優先し VISION の日本語例文（「ネットワークが「パブリック」...」）には揃えない（Hint 側はほぼ一致済み）
  - `-version`: stdlib flag + `debug.ReadBuildInfo()` フォールバック（cobra 不採用の方針どおり）
- **Phase 4 完了**（`hostinfo.Classify` + `winsys.CurrentAccount`、`HostInfo.UserName` → `Login UserLogin` に置換）
  - winsys は生データ（SID/UPN/Join/MSA サブキー）のみ返し、判定・"@" フィルタは hostinfo 側（OS 非依存でテスト可能）
  - MSA レジストリ: キー不在は「MSA 痕跡なし = ローカル確定材料」（MSAChecked=true・0件）、読み取り失敗のみ Unknown 退避
  - UPN / NetGetJoinInformation / MSA の取得失敗は best-effort で握り、SID 取得失敗のみ error
  - HostInfoText にセクション間空行を追加（VISION 例準拠）。golden 全更新 + 種別別 4 件追加
  - 実機（MSA サインイン機）で複数メール候補 + ローカル候補 + 注意書きの表示を確認
- **Phase 3 完了**（firewall / group_membership / sleep の 3 Check 追加、go-ole v1.3.0 導入）
  - `IsRuleGroupCurrentlyEnabled` は IDL 上パラメータ付き **propget** のため go-ole では `CallMethod` ではなく `GetProperty` で呼ぶ（`CallMethod` は DISP_E_MEMBERNOTFOUND 0x80020003 になり、フォールバックのルール列挙経路に落ちる）
  - グループ所属は **SE_GROUP_ENABLED を問わず「SID が TokenGroups に存在すれば所属」** と判定。UAC 非昇格トークンでは Administrators が SE_GROUP_USE_FOR_DENY_ONLY で載るが、RDP ログオンは新規トークンを生成するため所属の証拠として OK 扱い（実機の非昇格実行で検証済み）
  - winsys → diag の import（`diag.TokenGroup`）は既存の winsys → hostinfo と同方向で許容。禁止は diag/hostinfo → winsys のみ
