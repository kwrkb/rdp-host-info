# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## コマンド

- ビルド: `go build ./...`
- vet: `go vet ./...`
- 全テスト: `go test ./...`
- 単一パッケージ: `go test ./internal/diag/...`
- 単一テスト: `go test ./internal/diag/ -run TestRDPEnabledCheck`
- lint: `golangci-lint run ./...`（設定は `.golangci.yml`）
- 実行: `go run .`

module は `github.com/kwrkb/rdp-host-info`（Go 1.26）。依存は `golang.org/x/sys` と `github.com/go-ole/go-ole`（ファイアウォール COM 用）の 2 つのみ。

## アーキテクチャ

### DI seam（最重要）

`internal/diag` と `internal/hostinfo` は `internal/winsys` を import しない。各 Check / Collector は抽象化した関数型（provider）を注入され、`main.go`（+ `checks.go`）が `winsys` の実装を配線する。これにより実 OS 無しでテスト可能になっている。新しい診断項目を追加するときもこの構造を踏襲する。

禁止方向は「diag / hostinfo → winsys」のみ。逆方向（winsys → hostinfo / diag）は共有型を返すために許容されている（例: `winsys.ReadEdition` → `hostinfo.Edition`、`winsys.ListTokenGroups` → `diag.TokenGroup`、`winsys.CurrentAccount` → `hostinfo.AccountData`）。

### Check インターフェース

`internal/diag/check.go` が中核。`Check` は `Name() / NeedsAdmin() / Run() Result` を持つ。`diag.RunAll` が各 `Run()` を実行し、panic を recover して `StatusUnknown` に落とす。`Status` のゼロ値は `StatusUnknown`（取得失敗時のデフォルト）。

### アカウント種別判定（hostinfo）

RDP 接続用ユーザー名の候補生成は `internal/hostinfo/account.go` の `Classify` に集約されている。winsys（`CurrentAccount`）は生データ（SID 文字列 / UPN / ドメイン参加状態 / MSA レジストリのサブキー名）のみ返し、判定・フィルタは hostinfo 側で行う。判定は確度順（AzureAD SID → ドメイン参加 → MSA レジストリ → ローカル）で、順序を崩すと誤判定する（AzureAD 機はドメイン参加も true になりうる）。断定できない場合は複数候補 + Notes を返す。

### winsys の隔離

Windows 依存コードは `internal/winsys/*_windows.go` に集約し、ビルドタグで隔離する。ロジックを持たせず「取得と型変換のみ」に留め、テスト対象を最小化する（実装の妥当性検証は smoke test で行う）。例外は `netinfo.go` — `net` パッケージのみで実装されており OS 非依存のためユニットテスト対象になっている。

### render とデータフロー

`internal/render` の `HostInfoText` / `StatusText` が VISION.md 準拠のテキストを生成する。golden test（`testdata/*.golden`）で出力の回帰を検知する。将来 JSON/GUI 出力を追加する場合もこの層に追加する。

全体のデータフローは `winsys`（OS からの取得）→ `hostinfo` / `diag`（OS 非依存のロジック・判定）→ `render`（整形）→ `main`（配線・exit code）。exit code は NG が1つでもあれば 1。

## VISION 由来の必須規約

正典は `VISION.md`。実装判断に迷ったら VISION.md を優先する。特に以下は違反しやすいので注意する。

- **ロケール依存のコマンド出力を文字列パースしない**。一次情報源はレジストリ / Windows API（`golang.org/x/sys/windows`）/ 既知 SID / GUID。外部コマンドや PowerShell 呼び出しは他に手段がない場合の最終手段。
- **取得失敗は成功とも失敗とも偽らず `StatusUnknown`** として扱う。断定できない情報（ユーザー名形式など）は断定せず候補を複数提示する。
- **最小権限で API を開く**。サービス状態確認は `golang.org/x/sys/windows/svc/mgr.Connect()`（`SC_MANAGER_ALL_ACCESS` を要求し管理者権限が必要になる）を使わず、raw SCM を `SC_MANAGER_CONNECT` + `SERVICE_QUERY_STATUS` で開く（`internal/winsys/scm_windows.go` 参照）。管理者権限が必要な項目は `NeedsAdmin()` で明示する。
- 本ツールは診断・情報表示専用であり、Windows の設定を勝手に変更しない。

## テスト方針

- 各 Check は fake provider を注入したテーブル駆動テストで、値0/値1/欠落/アクセス拒否 → OK/NG/Unknown の全分岐を検証する（`internal/diag/checks_test.go` 参照）。
- `render` は golden file 比較（`internal/render/testdata/`）。
- `winsys` の Windows 実装はロジックを持たないため、`//go:build windows` の smoke test のみで十分とする。

## 個人情報の取り扱い

- テスト・golden ファイル・ドキュメントの例示値には、実機 `go run .` の出力（実PC名・実ユーザー名・実IP・Tailscale IP 等）を使わない。VISION.md の架空例（`OMEN16` / `yugo` / `192.168.1.20` / `100.80.10.5`）のような fictional な値のみを使用する。
- git commit 前には `git status` / 差分を確認し、実機出力やログが混入していないか確認する。

## 参照

- `VISION.md` — プロジェクトの正典仕様
- `PLAN.md` — 全体計画・フェーズ分割・進捗
- `LESSONS.md` — 過去に踏んだ落とし穴（COM の propget、deny-only SID 等）と、実装中の判断・妥協の記録。同種の作業前に確認する
