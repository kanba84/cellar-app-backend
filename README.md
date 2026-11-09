# cellar-app-backend

cellar-app-backendは、ワインセラー管理アプリ用のGo製バックエンドAPIサーバーです。

## ディレクトリ構成

- `app/` : Goアプリケーション本体
  - `handler/` : 各種APIハンドラー
  - `middleware/` : ミドルウェア
  - `model/` : モデル定義
  - `service/` : ビジネスロジック
  - `main.go` : エントリーポイント
- `db/` : DB初期化用ファイル
  - `init/` : 初期データやリストアスクリプト
- `.env` : 環境変数設定
- `docker-compose.yml` : Docker Compose構成ファイル

## セットアップ

### 1. 必要なツール

- Go 1.20以降
- Docker, Docker Compose

### 2. 開発用サーバーの起動

#### Goで直接起動

```bash
cd app
# 依存パッケージのインストール
go mod tidy
# サーバー起動
go run .
```

#### バイナリをビルドして起動

```bash
cd app
# バイナリをビルド
go build -o cellar-app
# サーバー起動
./cellar-app
```

#### Docker Composeで起動

```bash
# 初回はビルドが必須
docker compose up --build -d

# 次回以降は以下のコマンドでOK
docker compose up -d
```

#### Raspberry Pi向けにビルドして起動

```bash
cd app
# Raspberry Pi用にバイナリをビルド
GOOS=linux GOARCH=arm GOARM=6 go build -o cellar-app
```

Raspberry Pi上でのサービス起動方法については[こちら](./launch-on-raspberry-pi.md)を参照

### 3. DB初期化

- `db/init/`配下のスクリプトやダンプファイルを利用

## データ設計・DB運用ドキュメント

DB のデータモデル、GORM `AutoMigrate` によるスキーマ管理、バックアップ／復元手順、データディクショナリは[`docs/data/`](./docs/data/README.md)を参照してください。データ設計に関する正本はこのリポジトリ内に置きます。

## MCP サーバー

backend と同じプロセス・HTTP サーバーで、公式 Go SDK の Streamable HTTP transport による MCP endpoint を常時提供します。専用の feature flag はありません。既存の待受ポートを使い、LAN 内の MCP クライアントから次の URL に接続してください。

- 通常の HTTP: `http://<Raspberry Pi の LAN アドレス>:8080/mcp`
- TLS 設定時: `https://<Raspberry Pi の LAN アドレス>:8443/mcp`

公開するのは読み取り tool のみです。

- ワイン・ボトル: `list_wines`, `get_wine`, `list_bottles`, `get_bottle`
- マスタ: `list_countries`, `get_country`, `list_regions`, `get_region`, `list_wine_types`, `get_wine_type`, `list_appellations`, `get_appellation`, `list_designation_types`, `get_designation_type`
- 統計・スナップショット: `get_stats`, `get_wine_type_stats`, `get_country_stats`, `get_vintage_stats`, `get_inventory_trend`, `get_daily_snapshot`

List tool は現在のサービス動作どおり全件を返し、暗黙の件数制限やページングはありません。ID は `1` から `2147483647`、日数は `0` または `1` から `3650` を指定できます。日数 `0` または省略時は既存サービスの 30 日既定値を使います。日次スナップショットの日付は `YYYY-MM-DD` 形式です。LLM 情報取得、作成・更新・削除、アップロード、スナップショット生成は公開しません。

MCP endpoint に専用の認証は追加していません。LAN から backend に直接到達できるクライアントを前提とし、接続可能なネットワークを適切に制限してください。

## ライセンス

MIT License
