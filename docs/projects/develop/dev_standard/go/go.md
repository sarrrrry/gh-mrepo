# Go コーディング規約

本ドキュメントは、Goプロジェクトにおける一般的なコーディング規約を定める。
公式の [Effective Go](https://go.dev/doc/effective_go) および [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) を基盤とし、プロジェクト固有のルールを追加する。

---

## 1. フォーマットとリンター

### 1.1 フォーマッタ

- **gofumpt** を使用する(`gofmt` の厳格版)
- **goimports** でimportの整理を行う
- CI/エディタ保存時に自動フォーマットを適用する

### 1.2 リンター

- **golangci-lint** を使用する
- 最低限有効にするリンター: `errcheck`, `govet`, `staticcheck`
- プロジェクトルートの `.golangci.yml` に設定を集約する

---

## 2. 命名規則

### 2.1 基本方針

| 対象 | 規則 | 例 |
|------|------|----|
| パッケージ | 小文字、単数形、短く | `config`, `domain`, `executor` |
| エクスポートされる型/関数 | UpperCamelCase | `NewProfile`, `ConfigLoader` |
| 非エクスポートの型/関数 | lowerCamelCase | `findProfile`, `extractUserFlag` |
| 定数 | UpperCamelCase(エクスポート時) | `MaxRetries`, `DefaultTimeout` |
| インターフェース | 動作を表す名前(er接尾辞) | `Reader`, `ConfigLoader` |
| ファイル名 | snake_case | `loader_test.go`, `run.go` |

### 2.2 パッケージ名

- パッケージ名とディレクトリ名を一致させる
- `common`, `util`, `utils`, `helper`, `helpers`, `misc` は禁止
- パッケージ名をエクスポートされる識別子のプレフィックスとして繰り返さない

```go
// Good
config.NewLoader()

// Bad - パッケージ名を繰り返している
config.NewConfigLoader()
```

### 2.3 インターフェース

- メソッドが1つの場合は `er` 接尾辞を使う: `Reader`, `Writer`, `Closer`
- 利用側(consumer)のパッケージでインターフェースを定義する(Accept interfaces, return structs)
- 不要に大きなインターフェースを作らない

```go
// Good - 利用側で必要最小限のインターフェースを定義
type ConfigLoader interface {
    Load() ([]domain.Profile, error)
}

// Bad - 実装側で全メソッドを含む巨大なインターフェースを定義
type ConfigService interface {
    Load() ([]domain.Profile, error)
    Save(profile domain.Profile) error
    Delete(name string) error
    Validate() error
}
```

### 2.4 略語

- 一般に認知された略語は全て大文字にする: `URL`, `HTTP`, `ID`, `JSON`, `API`
- ローカル変数では短い名前を許容する(`i`, `n`, `err`, `ctx`, `cfg`)

---

## 3. プロジェクト構成

### 3.1 ディレクトリレイアウト

```
project-root/
  main.go              # エントリーポイント
  internal/            # 外部パッケージから参照不可
    domain/            # エンティティ、値オブジェクト
    app/               # アプリケーションロジック(ユースケース)
    config/            # 設定の読み込み・初期化
    executor/          # 外部コマンド実行
    selector/          # UI操作(プロンプト等)
  docs/                # ドキュメント
```

### 3.2 `internal` パッケージ

- 外部に公開しないコードは全て `internal/` 配下に置く
- Goコンパイラが `internal` パッケージへの外部アクセスを禁止するため、意図しない依存を防げる

---

## 4. import

### 4.1 グループ分け

importは以下の3グループに分け、空行で区切る:

```go
import (
    // 1. 標準ライブラリ
    "fmt"
    "os"

    // 2. サードパーティ
    "github.com/BurntSushi/toml"

    // 3. プロジェクト内パッケージ
    "github.com/sarrrrry/gh-mrepo/internal/domain"
)
```

- goimportsの `local-prefixes` 設定で自動的にグループ分けされる
- ドット(`.`)インポート、ブランク(`_`)インポートは必要最小限に留める

---

## 5. エラーハンドリング

### 5.1 基本ルール

- エラーは常に呼び出し元に返すか、適切に処理する。無視しない
- `_` でのエラー破棄は明確な理由がある場合のみ

```go
// Good
if err != nil {
    return fmt.Errorf("loading config: %w", err)
}

// Bad - エラーを無視
result, _ := doSomething()
```

### 5.2 エラーのラップ

- `fmt.Errorf` と `%w` でエラーをラップし、文脈情報を付加する
- ラップメッセージは小文字で始め、末尾にコロンを付ける

```go
return fmt.Errorf("loading config from %s: %w", path, err)
```

### 5.3 センチネルエラーとカスタムエラー型

- パッケージ間で判別が必要なエラーは `var ErrXxx = errors.New(...)` で定義する
- 追加情報が必要な場合はカスタムエラー型を定義する

```go
var ErrEmptyName = errors.New("profile name must not be empty")

type ExitError struct {
    Code int
}

func (e *ExitError) Error() string {
    return fmt.Sprintf("exit status %d", e.Code)
}
```

### 5.4 エラー判定

- `errors.Is()` / `errors.As()` を使用する。型アサーションでの直接比較は避ける

```go
// Good
if errors.Is(err, ErrNotFound) { ... }

var exitErr *ExitError
if errors.As(err, &exitErr) { ... }
```

---

## 6. 関数とメソッド

### 6.1 関数設計

- 関数は1つの責務に集中させる
- 引数は5つ以下を目安とする。超える場合は構造体にまとめる
- 戻り値は `(結果, error)` のパターンに従う

### 6.2 コンストラクタ

- `NewXxx` の命名規約に従う
- パッケージ内に型が1つしかない場合は `New` でも可

```go
func NewLoader(path string) *Loader {
    return &Loader{path: path}
}
```

### 6.3 メソッドレシーバ

- 状態を変更しない場合は値レシーバ
- 状態を変更する場合、または構造体が大きい場合はポインタレシーバ
- 1つの型に対してレシーバの種類を混在させない

```go
// 値レシーバ - 状態を変更しない
func (p Profile) String() string { ... }

// ポインタレシーバ - 状態を変更する
func (a *App) Run(user string, args []string) error { ... }
```

---

## 7. 構造体と型

### 7.1 構造体の初期化

- フィールド名を明示した初期化を使う(順序依存を避ける)

```go
// Good
p := Profile{
    Name:        "work",
    GHConfigDir: "/path/to/config",
    Root:        "/path/to/root",
}

// Bad - 順序依存
p := Profile{"work", "/path/to/config", "/path/to/root"}
```

### 7.2 ゼロ値の活用

- 型のゼロ値が有効な状態になるよう設計する
- `sync.Mutex`, `bytes.Buffer` などはゼロ値でそのまま使える

---

## 8. 並行処理

### 8.1 goroutine

- goroutineのライフサイクルを常に管理する(リークを防ぐ)
- `sync.WaitGroup` または `errgroup.Group` で終了を待機する
- `context.Context` でキャンセルを伝搬する

### 8.2 チャネル

- チャネルの方向を型で制限する(`chan<-`, `<-chan`)
- 送信側がチャネルをクローズする

### 8.3 データ競合

- 共有状態へのアクセスは `sync.Mutex` またはチャネルで保護する
- `-race` フラグでテストを実行し、データ競合を検出する

```bash
go test -race ./...
```

---

## 9. テスト

### 9.1 ファイル配置

- テストファイルはテスト対象と同じパッケージに配置する
- ファイル名は `xxx_test.go`

### 9.2 テスト関数

- `Test` プレフィックスで始め、テスト対象と条件を明記する

```go
func TestNewProfile_EmptyNameReturnsError(t *testing.T) { ... }
func TestFindByDirectory_MatchesPrefix(t *testing.T) { ... }
```

### 9.3 テーブル駆動テスト

- 複数のケースがある場合はテーブル駆動テストを使用する

```go
func TestNewProfile(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr error
    }{
        {name: "valid", input: "work", wantErr: nil},
        {name: "empty name", input: "", wantErr: ErrEmptyName},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := NewProfile(tt.input, "/path", "")
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("got %v, want %v", err, tt.wantErr)
            }
        })
    }
}
```

### 9.4 テストヘルパー

- テストヘルパー関数には `t.Helper()` を呼び出す
- テスト用のフェイク/スタブはテストファイル内に定義する

### 9.5 テストにおける外部依存

- インターフェースを利用してモックに差し替える
- ファイルシステム操作は `t.TempDir()` を使う
- 外部コマンド実行はインターフェース経由で抽象化する

---

## 10. コメントとドキュメント

### 10.1 Godoc規約

- エクスポートされる型・関数・メソッドにはコメントを付ける
- コメントは対象の名前で始める

```go
// Profile はGitHubアカウントの設定プロファイルを表す値オブジェクト。
type Profile struct { ... }

// NewProfile は新しいProfileを生成する。名前が空の場合はエラーを返す。
func NewProfile(name, ghConfigDir, root string) (Profile, error) { ... }
```

### 10.2 コメントの方針

- 「何を」ではなく「なぜ」を書く
- 自明なコードにはコメントを付けない
- TODOコメントには担当者や課題番号を含める: `// TODO(#123): ...`

---

## 11. その他

### 11.1 `context.Context`

- 関数の第1引数として渡す
- 構造体のフィールドに保持しない

```go
func (a *App) Run(ctx context.Context, args []string) error { ... }
```

### 11.2 `defer`

- リソースの解放(ファイルクローズ、ロック解除等)は `defer` を使う
- `defer` はエラーチェックの直後に書く

```go
f, err := os.Open(path)
if err != nil {
    return err
}
defer f.Close()
```

### 11.3 型変換

- 明示的な型変換を行う。暗黙的な変換に頼らない
- `strconv` パッケージを使う(`fmt.Sprintf` での数値変換は避ける)

### 11.4 不要な`else`の回避

- early returnを活用し、ネストを浅く保つ

```go
// Good
if err != nil {
    return err
}
// 正常系の処理

// Bad
if err != nil {
    return err
} else {
    // 正常系の処理
}
```

---

## 参考資料

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
