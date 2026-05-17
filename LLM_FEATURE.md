# LLM ワイン情報取得機能

このドキュメントでは、Cellar App バックエンドに追加された LLM ベースのワイン情報取得機能について説明します。

## 概要

LLM（大規模言語モデル）統合により、ワイン名に基づいてワイン情報（生産者、ブドウ品種、テイスティングノート）を LLM から取得し、取得したデータを自動的にデータベースに保存できるようになります。

## アーキテクチャ

### ディレクトリ構成

```
cellar-app-backend/
├── llm/
│   ├── types.go              # LLM データ構造とインターフェース
│   ├── provider.go            # プロバイダーファクトリと初期化
│   ├── gemini_provider.go     # Gemini API 実装
│   └── [openai_provider.go]   # 将来: OpenAI 実装
├── model/
│   └── model.go               # Grape、WineGrape、LLMWineInfo を追加
├── repository/
│   ├── grape.go               # Grape リポジトリ
│   ├── wine_grape.go          # WineGrape リポジトリ
│   └── wine.go                # （変更なし）
├── service/
│   ├── service.go             # GrapeRepo、WineGrapeRepo を追加
│   ├── wine_llm.go            # LLM 関連のサービスメソッド
│   └── wine.go                # （変更なし）
├── handler/
│   ├── wine_llm.go            # 新しい LLM ハンドラ: GetWineLLMInfo
│   └── wine.go                # （変更なし）
├── main.go                    # マイグレーション設定を追加
├── router.go                  # LLM エンドポイントを追加
└── go.mod                     # Gemini API 依存関係を追加
```

## モデル

### Grape

```go
type Grape struct {
    ID   uint   // プライマリキー
    Name string // 一意なインデックス、正規化されたブドウ品種名
}
```

**データベーステーブル**: `grapes`

### WineGrape

Wine と Grape 間の明示的な多対多リレーション。

```go
type WineGrape struct {
    WineID       uint      // wines へのフォーキンキー
    GrapeID      uint      // grapes へのフォーキンキー
    Percentage   *float64  // オプション: ブレンド内のブドウ比率
    DisplayOrder int       // 表示順序
    Grape        Grape     // リレーション
    Wine         *Wine     // リレーション
}
```

**データベーステーブル**: `wine_grapes`

### Wine モデルの更新

既存の `Wine` モデルが更新されてリレーションが追加されました:

```go
type Wine struct {
    // ... 既存フィールド ...
    WineGrapes []WineGrape `gorm:"foreignKey:WineID" json:"wine_grapes,omitempty"`
}
```

## API エンドポイント

### ワイン情報を LLM から取得

**エンドポイント**: `GET /wines/:id/llm-info`

**説明**: ワイン名に基づいて LLM からワイン情報を取得し、データベースに保存します。

**パスパラメータ**:
- `id` (uint): ワイン ID

**レスポンス**: 
```json
{
  "producer": "Domaine Leroy",
  "grapes": [
    {
      "name": "Pinot Noir"
    },
    {
      "name": "Chardonnay"
    }
  ],
  "tasting_note": "チェリーとオークの香りが優雅で芳香的..."
}
```

**ステータスコード**:
- `200 OK`: ワイン情報の取得に成功
- `400 Bad Request`: ワイン ID が無効
- `404 Not Found`: ワインが見つからない
- `500 Internal Server Error`: LLM プロバイダーが設定されていない、または API エラー

## LLM プロバイダーインターフェース

### プロバイダーインターフェース

```go
type WineInfoProvider interface {
    FetchWineInfo(ctx context.Context, wineName string) (*WineInfoResult, error)
}
```

### サポート対象のプロバイダー

#### Gemini（現在）

Google の Gemini API を使用してワイン情報を取得します。

**環境変数**:
```env
GEMINI_API_KEY=your-api-key-here
LLM_PROVIDER=gemini  # または省略（デフォルト: gemini）
```

**機能**:
- JSON のみのレスポンス形式を強制
- ブドウ品種名の自動正規化
- ストリーミング対応（SDK が内部処理）

#### OpenAI（将来対応）

OpenAI サポートを追加するには、`llm/openai_provider.go` を作成します:

```go
type OpenAIProvider struct {
    apiKey string
}

func NewOpenAIProvider(apiKey string) *OpenAIProvider {
    return &OpenAIProvider{apiKey: apiKey}
}

func (op *OpenAIProvider) FetchWineInfo(ctx context.Context, wineName string) (*WineInfoResult, error) {
    // 実装
}
```

その後、`llm/provider.go` を更新:

```go
case "openai":
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
    }
    return NewOpenAIProvider(apiKey), nil
```

## データベースマイグレーション

### マイグレーション戦略

新しいテーブルのみ GORM AutoMigrate で管理されます:

```go
db.AutoMigrate(
    &Grape{},
    &WineGrape{},
)
```

**重要**: 既存テーブル（wines、bottles、countries など）は変更されません。レガシーデータを保護するためです。

### スタートアップ時のマイグレーション

マイグレーションはアプリケーションスタートアップ時に `main.go` で自動実行されます:

```go
if err := db.AutoMigrate(
    &model.Grape{},
    &model.WineGrape{},
); err != nil {
    log.Fatalf("Migration failed: %v", err)
}
```

## サービスレイヤー

### 主要メソッド

#### FetchAndSaveWineInfo

```go
func (s *Service) FetchAndSaveWineInfo(
    ctx context.Context,
    wine *model.Wine,
    provider llm.WineInfoProvider,
) (*llm.WineInfoResult, error)
```

LLM からワイン情報を取得してデータベースに保存します:
1. LLM プロバイダーをワイン名で呼び出す
2. Wine.Producer が空の場合のみ更新
3. Grape レコードを作成/取得
4. WineGrape リレーションを作成
5. パースされた LLM レスポンスを返す

#### GetWineWithGrapes

```go
func (s *Service) GetWineWithGrapes(wineID uint) (*model.Wine, error)
```

ワインに関連するすべてのブドウ情報を取得します。

## ブドウ品種名の正規化

ブドウ品種名は一貫性を確保するために正規化されます:

```
入力例                    → 正規化後
"Pinot Noir"           → "pinot noir"
"pinot-noir"           → "pinot noir"
"PINOT  NOIR"          → "pinot noir"
"ピノノワール"          → "ピノノワール" (日本語テキストは保持、trim + 小文字化)
```

**正規化ステップ**:
1. 前後のホワイトスペースを削除
2. 小文字に変換
3. ハイフンをスペースに置換
4. 複数スペースを単一スペースに統合

## エラーハンドリング

### プロバイダー初期化エラー

`LLM_PROVIDER` が設定されておらず、`GEMINI_API_KEY` がない場合:

```
Error: GEMINI_API_KEY environment variable not set
```

### LLM リクエストエラー

LLM API 呼び出しが失敗した場合:

```json
{
  "error": "failed to fetch wine info"
}
```

### JSON パースエラー

LLM レスポンスが有効な JSON でない場合:

```
Error: failed to parse JSON response: <error details>, raw response: <response>
```

## 設定

### 環境変数

```env
# データベース
DB_HOST=localhost
DB_PORT=5432
DB_USER=cellar_app
DB_PASSWORD=
DB_NAME=cellar_app

# LLM 設定
LLM_PROVIDER=gemini  # または openai (将来対応)
GEMINI_API_KEY=      # Gemini に必須
OPENAI_API_KEY=      # OpenAI に必須 (将来対応)

# SSL（オプション）
SSL_CERT_FILE=
SSL_KEY_FILE=
```

## 使用例

### API でワイン情報を取得

```bash
curl -X GET "http://localhost:8080/wines/1/llm-info"
```

### レスポンス

```json
{
  "producer": "Burgundy Vineyard",
  "grapes": [
    {
      "name": "pinot noir"
    }
  ],
  "tasting_note": "上品で優雅な香りが特徴のワイン..."
}
```

システムは自動的に以下を実行します:
1. Grape レコードを作成/取得
2. WineGrape リレーションを作成
3. Wine の producer が空の場合のみ更新
4. パースされた情報を返す

## 将来の拡張

### 1. TastingNote ストレージ

現在、テイスティングノートは返されますが、保持されません。ストレージを追加するには:

1. Wine モデルに `TastingNote` テキストフィールドを追加
2. サービスレイヤーを更新してテイスティングノートを保存
3. マイグレーションでスキーマを更新

### 2. 信頼度スコア

LLM レスポンスの信頼度を追跡:

```go
type WineGrape struct {
    // ... 既存フィールド ...
    Confidence float64 // 0.0-1.0 の信頼度スコア
    Source     string  // "llm"、"user"、"import"
}
```

### 3. 複数の LLM プロバイダー

現在は Gemini のみサポート。新しいプロバイダーを追加するには:

1. 新しいプロバイダーファイルを作成（例: `llm/llama_provider.go`）
2. `WineInfoProvider` インターフェースを実装
3. `llm/provider.go` のファクトリを更新
4. 設定オプションを追加

### 4. キャッシング

API 呼び出しを削減するため LLM レスポンスをキャッシュ:

```go
type WineInfoCache interface {
    Get(wineName string) (*WineInfoResult, error)
    Set(wineName string, info *WineInfoResult) error
}
```

### 5. バッチ操作

複数ワインの LLM 情報を取得:

```go
func (s *Service) FetchAndSaveMultipleWinesInfo(
    ctx context.Context,
    wineIDs []uint,
    provider llm.WineInfoProvider,
) (map[uint]*llm.WineInfoResult, error)
```

### 6. バリデーションルール

LLM レスポンスのバリデーションを追加:

```go
type LLMValidator interface {
    Validate(result *WineInfoResult) error
}
```

## 実装上の注意

### Context 使用方法

すべての LLM 操作は `context.Context` を使用します:
- リクエストのタイムアウト管理
- キャンセルの伝播
- デッドライン強制

### エラーハンドリング

- Panic を使用しない - すべてのエラーは明示的に処理
- `fmt.Errorf` でコンテキストとともにエラーをラップ
- デバッグのための詳細なログ出力
- API レスポンスではユーザーフレンドリーなエラーメッセージ

### トランザクション安全性

Wine producer の更新とブドウの作成は別の操作です。将来のデータ一貫性向上のため:

```go
// 将来の拡張: トランザクション内でラップ
tx := db.BeginTx(ctx, nil)
// ... wine を更新 ...
// ... ブドウを保存 ...
tx.Commit()
```

## テスト

### ユニットテストの例

```go
func TestNormalizeGrapeName(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"Pinot Noir", "pinot noir"},
        {"pinot-noir", "pinot noir"},
        {"  PINOT  NOIR  ", "pinot noir"},
    }
    
    for _, tt := range tests {
        result := normalizeGrapeName(tt.input)
        if result != tt.expected {
            t.Errorf("got %q, want %q", result, tt.expected)
        }
    }
}
```

### 統合テストの例

```go
func TestGetWineLLMInfo(t *testing.T) {
    // テストデータベースをセットアップ
    // テストワインを作成
    // エンドポイントを呼び出し
    // レスポンスを検証
    // データベースの状態を確認
}
```

## トラブルシューティング

### 問題: "GEMINI_API_KEY environment variable not set"

**解決策**: 環境変数を設定:

```bash
export GEMINI_API_KEY="your-key-here"
```

または環境ローダーを使用する場合は `.env` ファイルに追加。

### 問題: LLM からの無効な JSON レスポンス

**解決策**: 

1. `gemini_provider.go` の LLM プロンプトが JSON 形式を強制していることを確認
2. API が有効なレスポンスを返していることを確認
3. ログで生のレスポンスコンテンツを確認

### 問題: ワイン ID が見つからない

**解決策**: 

1. ワインがデータベースに存在することを確認
2. ワイン ID が正しいことを確認
3. 最初に GET /wines/:id でワインが存在することを確認

## 依存関係

`go.mod` に追加された新しい依存関係:

- `github.com/google/generative-ai-go v0.4.0` - Gemini API SDK
- `google.golang.org/api v0.152.0` - Google API クライアントライブラリ

インストール:

```bash
go mod tidy
```

## 参考資料

- [Gemini API ドキュメント](https://ai.google.dev/)
- [Google Generative AI Go SDK](https://github.com/google/generative-ai-go)
