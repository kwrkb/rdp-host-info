# LESSONS.md

開発で得た教訓の記録。同じミスを繰り返さないためのルール集。

## Phase 4〜6 — アカウント種別 / 仕上げ / 品質 (2026-07-13)

### フィールドの置換は互換フィールドを残さずコンパイルエラーで消費者を洗い出す
- `HostInfo.UserName string` を `Login UserLogin` に差し替える際、旧フィールドを残す案もあったが、削除したことでコンパイルエラーが全消費者（render / main / テスト）を漏れなく指し示した。互換のための二重管理は「断定しない候補提示」という新仕様とも矛盾する
- **ルール**: 同一リポジトリ内で完結する型変更は、旧フィールドを残さず置換し、コンパイルエラーを消費者の検出器として使う

### 「キー不在」と「読み取り失敗」は別の情報として扱う
- MSA 判定レジストリ（IdentityCRL\UserExtendedProperties）で、キー不在（ErrNotExist）は「MSA サインインの痕跡なし＝ローカル確定の材料」であり、アクセス失敗（判定不能→Unknown 退避）とは意味が異なる。同一視すると常に Unknown になり判定精度が落ちる
- **ルール**: 存在確認を兼ねる読み取りでは、NotExist とその他のエラーを分岐させ、それぞれが何を意味するかをコメントで明示する

### 判定の確度順は「後の条件も真になりうる」前提で早期 return にする
- アカウント種別判定で AzureAD 参加機はドメイン参加判定も true になりうる。SID プレフィックス（確度最高）→ ドメイン → MSA → ローカルの順で早期 return しないと誤判定する
- **ルール**: 複数の判定条件が排他でない分類ロジックは、確度順に並べて早期 return し、順序の理由をコメントに書く

### 文言の回帰は「実装を通した golden」で検知する
- render の golden を手組み Result で書くと、Check 実装側の Message/Hint 変更を検知できない。fake provider を注入した実 Check → `RunAll` → `StatusText` を通す golden にしたことで、文言変更が必ず golden 差分に現れるようになった
- **ルール**: 出力文言をテストしたいときは、文言を持つ層（Check 実装）を実際に通して golden を生成する。手組みデータの golden はレイアウト検証に限定する

### lint の除外は行単位の `_ =` ではなく設定ファイルに理由付きで集約する
- errcheck が defer の後始末（handle Close/Free）を 8 箇所指摘。各行に `_ =` を撒くとノイズになるため、`.golangci.yml` の `exclude-functions` に関数単位で列挙し「失敗しても対処のしようがない」理由をコメントした
- **ルール**: 系統的に無視してよい lint 指摘は、設定ファイルに理由コメント付きで集約する。行単位の抑制は単発の例外に限る

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
