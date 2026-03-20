# cellar-app-go

cellar-app-goは、ワインセラー管理アプリ用のGo製バックエンドAPIサーバーです。

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

## ライセンス

MIT License
