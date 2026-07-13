# LESSONS.md

開発で得た教訓の記録。同じミスを繰り返さないためのルール集。

## Phase 3 — ファイアウォール / グループ所属 / スリープ (2026-07-13)

### COM のパラメータ付きプロパティは CallMethod では呼べない
- `INetFwPolicy2.IsRuleGroupCurrentlyEnabled` を go-ole の `CallMethod` で呼ぶと DISP_E_MEMBERNOTFOUND (0x80020003) になった。IDL 上はメソッドではなくパラメータ付き propget プロパティだった
- **ルール**: COM メンバーを IDispatch 経由で呼ぶ前に IDL の宣言（method / propget / propput）を確認し、propget は `oleutil.GetProperty(disp, name, args...)` で呼ぶ

### フォールバック経路は本経路の失敗を隠す — 出力に経路マーカーを入れて検証する
- 上記の CallMethod 失敗はフォールバック（ルール列挙）が動いたため、テスト・ビルドは全て通り一見正常だった。実機出力に付けた `port rule` マーカーで初めてフォールバック経路に落ちていることに気づいた
- **ルール**: フォールバックを実装したら、どちらの経路で結果を得たかを出力・ログで区別できるようにし、実機検証では本経路で動いていることまで確認する

### UAC 非昇格トークンでは Administrators は deny-only で載る
- 非昇格プロセスの TokenGroups では Administrators SID が SE_GROUP_USE_FOR_DENY_ONLY 属性（ENABLED なし）で列挙される。`CheckTokenMembership` は deny-only を「非所属」と返すため、所属確認の用途には不適
- **ルール**: 「グループに所属しているか」を判定するときは TokenGroups を自前列挙し、SID の存在で判定する（SE_GROUP_ENABLED や CheckTokenMembership に頼らない）。判定の意味（実効権限ではなく所属）を文言にも反映する

### go mod tidy はコードが import する前の依存を消す
- `go get go-ole` 直後に `go mod tidy` を実行したら、まだどのファイルも import していなかったため依存が削除された
- **ルール**: 新しい依存は「import を書いたコードを作成した後」に `go get` する。先に追加する場合は tidy を実装後まで遅らせる

### 実機依存の値の欠落は err と正常系を区別して設計する
- デスクトップ機にはバッテリー(DC)の電源設定値が存在しないのが正常。DC 読取失敗を err にすると、デスクトップで常に Unknown になってしまう
- **ルール**: provider の戻り値を設計する時点で「環境によって存在しない値」を洗い出し、`hasX bool` などで欠落を正常系として表現する（err は本当に取得に失敗した場合のみ）
