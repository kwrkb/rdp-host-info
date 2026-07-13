# Implementation Notes

作業中の判断・選択・妥協の記録。

## Phase 3 — ファイアウォール / グループ所属 / スリープ

- **`IsRuleGroupCurrentlyEnabled` を `GetProperty` で呼ぶ**: COM の IDL 上はパラメータ付き propget プロパティであり、go-ole の `CallMethod` では DISP_E_MEMBERNOTFOUND (0x80020003) になる。実機デバッグで判明し `oleutil.GetProperty(policy, "IsRuleGroupCurrentlyEnabled", group)` に修正。この失敗はフォールバック（ルール列挙）で隠れて動いてしまうため、実機出力の `port rule` 表記から発見した。
- **グループ所属は SID の存在のみで判定（SE_GROUP_ENABLED を見ない）**: UAC 非昇格トークンでは Administrators が SE_GROUP_USE_FOR_DENY_ONLY で載り ENABLED が立たない。代替案の `CheckTokenMembership` は deny-only SID を「非所属」と返すため不採用（VISION の趣旨どおり）。RDP ログオンは新規トークンを生成するので deny-only でも所属の証拠と判断。実機（非昇格）で `[OK] (Administrators)` を確認。
- **winsys → diag の import を許容**: `diag.TokenGroup` を winsys が参照する。既存の winsys → hostinfo（`hostinfo.Edition`）と同方向で、禁止されているのは diag/hostinfo → winsys のみ。TokenGroup を第三のパッケージに切り出す案は過剰と判断。
- **ファイアウォールのフォールバック時、ポート範囲指定ルール（"3000-4000"）は一致扱いしない**: 範囲パースまでやると誤検知・複雑化のリスクが上回る。範囲ルールで開けている環境は稀で、その場合は NG 側に倒れて Hint で手動確認を促せる。
- **DC（バッテリー）タイムアウトの読取失敗は err にしない**: デスクトップ機には DC 値が無いのが正常のため `hasDC=false` で続行。AC 失敗のみ Unknown に落とす。
- **gofmt は新規ファイルのみ適用**: 既存の未整形ファイル（rdpenabled.go 等）が gofmt -l に出るが、Phase 3 の差分を最小に保つため触れない。
