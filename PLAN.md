# PLAN.md — rdp-host-info 実装計画

仕様の正典は VISION.md、内部構造の索引は NOTES.md、判断の記録は LESSONS.md。
本ファイルはフェーズ境界と現在地だけを持つ。実装中に矛盾があれば VISION.md を優先する。

## 現在地

**v0.2.0 リリース済み（2026-08-16）。Phase 0〜9 完了、未着手のフェーズなし。**

残作業は、状態変更を伴う実機検証（NOTES.md「検証手順」6）が手動残置になっていること。
各フェーズの実装経緯と却下した代替案は LESSONS.md、それ以前の細かい進捗は
`git log -p PLAN.md` に残っている。

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
  - [x] 実機マトリクス検証（読み取り専用項目。状態変更を伴う項目は手動残置）
- [x] **Phase 7 — 公開準備とリリース (v0.1.0)**
  - [x] `.gitattributes`（`eol=lf`）、`LICENSE`（MIT）
  - [x] CI（`.github/workflows/ci.yml`, windows-latest 固定）: build/vet/test/golangci-lint
  - [x] `.goreleaser.yaml` + リリースワークフロー（タグ `v*` → windows amd64/arm64
    zip + checksums + GitHub Release）
  - [x] README に License / バイナリ入手方法 / CI バッジ、リポジトリ description・topics
  - [x] 個人情報の混入なしを確認して `public` 化、`v0.1.0` タグ push →
    配布 zip を実機で実行して検証
- [x] **Phase 8 — 出力の i18n（既定 en / `-lang ja`）**
  - [x] `internal/msgid` + `internal/msg` カタログを導入し、`diag` /
    `hostinfo` から文言を排除（ID と値のみを返す）
  - [x] golden test を en/ja 両言語でフルカバー、カタログ完全性テスト
  - [x] README を英語主・`README.ja.md` 従に再編
- [x] **Phase 9 — Scoop 配布 (v0.2.0)**
  - [x] `.goreleaser.yaml` に `scoops:`（`directory: bucket`）を追加し、
    `release.yml` で 1Password から `SCOOP_GITHUB_TOKEN` を供給
  - [x] README（英/日）に `scoop install rdp-host-info` の導線を追加
  - [x] `v0.2.0` タグ push → `kwrkb/scoop-bucket` に manifest が発行されることを確認

各フェーズ末で `go build ./... && go vet ./... && go test ./...` を通す。
