# Implementation Notes

作業中の判断・選択・妥協の記録。

## Phase 3 — ファイアウォール / グループ所属 / スリープ

- **`IsRuleGroupCurrentlyEnabled` を `GetProperty` で呼ぶ**: COM の IDL 上はパラメータ付き propget プロパティであり、go-ole の `CallMethod` では DISP_E_MEMBERNOTFOUND (0x80020003) になる。実機デバッグで判明し `oleutil.GetProperty(policy, "IsRuleGroupCurrentlyEnabled", group)` に修正。この失敗はフォールバック（ルール列挙）で隠れて動いてしまうため、実機出力の `port rule` 表記から発見した。
- **グループ所属は SID の存在のみで判定（SE_GROUP_ENABLED を見ない）**: UAC 非昇格トークンでは Administrators が SE_GROUP_USE_FOR_DENY_ONLY で載り ENABLED が立たない。代替案の `CheckTokenMembership` は deny-only SID を「非所属」と返すため不採用（VISION の趣旨どおり）。RDP ログオンは新規トークンを生成するので deny-only でも所属の証拠と判断。実機（非昇格）で `[OK] (Administrators)` を確認。
- **winsys → diag の import を許容**: `diag.TokenGroup` を winsys が参照する。既存の winsys → hostinfo（`hostinfo.Edition`）と同方向で、禁止されているのは diag/hostinfo → winsys のみ。TokenGroup を第三のパッケージに切り出す案は過剰と判断。
- **ファイアウォールのフォールバック時、ポート範囲指定ルール（"3000-4000"）は一致扱いしない**: 範囲パースまでやると誤検知・複雑化のリスクが上回る。範囲ルールで開けている環境は稀で、その場合は NG 側に倒れて Hint で手動確認を促せる。
- **DC（バッテリー）タイムアウトの読取失敗は err にしない**: デスクトップ機には DC 値が無いのが正常のため `hasDC=false` で続行。AC 失敗のみ Unknown に落とす。
- **gofmt は新規ファイルのみ適用**: 既存の未整形ファイル（rdpenabled.go 等）が gofmt -l に出るが、Phase 3 の差分を最小に保つため触れない。

## Phase 7 — 公開準備とリリース (v0.1.0)

- **golden テスト FAIL の根本原因は改行コード、テスト側の緩和では直さない**: `core.autocrlf=true` の Windows 機（ローカル・GitHub Actions windows-latest 共通）でチェックアウトすると `.golden`/`.go` が CRLF 化される一方、`render.HostInfoText` 等は LF を生成するため不一致で FAIL していた。`readGolden` 側で `\r` を除去する対処ではなく `.gitattributes`（`eol=lf`）でチェックアウト結果自体を固定した（CLAUDE.md の「一時的な回避策は禁止」に従う）。
- **`.gitattributes` 追加後は `git checkout` だけでは再チェックアウトされない**: 既存ファイルは内容不変とみなされ、attribute 変更だけでは working tree が書き換わらない（削除→checkout で強制書き換えが必要）。この作業中に一括削除＋再checkoutを行い、ハーネスの破壊的操作チェックに一度ブロックされた（未コミットの `ci.yml` 編集が巻き戻された）。以後は「まず `git status` で作業ツリーがクリーンかを確認してから一括操作する」を徹底し、失われた編集は再適用して復旧した。
- **golangci-lint のバージョン選定は go.mod のターゲット Go に強く依存する**: バイナリ配布版（`golangci-lint-action` の既定）は latest でも古い Go でビルドされており、`go.mod` の `go 1.26.5`（非常に新しい）を解釈できず `can't load config` で失敗する。`install-mode: goinstall` でも `version: latest` は v1 系（`.golangci.yml` の v2 config schema と非互換）を解決してしまい、かつ v2 系はモジュールパスが `github.com/golangci/golangci-lint/v2/...` に変わっているため `install-mode: goinstall` 自体が非対応（`invalid version: unknown revision cmd/golangci-lint/v2.12.2`）。最終的にアクションを使わず `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2` を CI ステップで直接実行する方式に落ち着いた。
- **CI は windows-latest 固定が必須**: `internal/winsys/*_windows.go` が `//go:build windows` のため `GOOS=linux go build ./...` はビルド不能（`undefined: winsys.ReadEdition` 等）。matrix に linux を足さない。
- **リリースビルド（GoReleaser）は ubuntu-latest からのクロスコンパイルで十分**: 依存（`golang.org/x/sys`, `go-ole`）は cgo 不要の純 Go のため、Windows amd64/arm64 バイナリを Linux ランナーから生成できる。CI（実行・lint 検証）と Release（成果物ビルド）で runner を使い分けた。
