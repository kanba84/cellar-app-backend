# バックエンド技術スタック

この文書は、バックエンドのコードと設定から確認できる技術スタックをまとめたものです。バージョンは主に `go.mod` と `Dockerfile` の記載に基づきます。依存ファイルに含まれるだけのライブラリと、アプリケーションの実装で利用している技術は区別しています。

## アプリケーション概要

ワイン、ボトル、関連マスターデータ、在庫統計を扱う Go 製 HTTP API です。PostgreSQL にデータを保存し、ワイン情報の取得には Google Gemini API を利用します。ボトル画像のラベル画像を EXIF 補正・縮小して保存する処理と、日次在庫スナップショットを作成する処理も含みます。

## 技術スタック

### アプリケーション実装技術

| 技術 | バージョン | 用途 | 確認できる根拠 |
|---|---|---|---|
| Go | 1.25.0（`go.mod` の `go` 指定） | API サーバーおよびバックエンド処理の実装 | `go.mod`、ルートパッケージの `main.go` |
| Gin | 1.10.1 | HTTP ルーティング、ハンドラー実行、JSON 応答、HTTP/TLS サーバー起動 | `go.mod`、`router.go` の `gin.Default()` とルート登録、`main.go` の `Run`／`RunTLS` |
| GORM | 1.31.1 | PostgreSQL を使った ORM によるデータアクセス | `go.mod`、`database/` の接続・`AutoMigrate`、`repository/` |
| GORM PostgreSQL ドライバー | 1.6.0 | GORM から PostgreSQL へ接続 | `go.mod`、`main.go` の `gorm.io/driver/postgres` |
| pgx/v5 | 5.8.0 | PostgreSQL 接続ドライバー。GORM PostgreSQL ドライバーの接続基盤として依存 | `go.mod` に直接記載。アプリケーションの repository は GORM 経由で DB を操作 |
| Google Generative AI Go SDK | 0.20.1 | Gemini によるワイン情報の生成 | `go.mod`、`llm/gemini_provider.go` の `genai.NewClient`／`GenerateContent` |
| Google API Go クライアント | 0.186.0 | Gemini クライアントへの API キー指定、Google API エラーの扱い | `go.mod`、`llm/gemini_provider.go` の `option.WithAPIKey`、`llm/retry.go` |
| robfig/cron | 3.0.1 | 毎日 0 時に在庫スナップショットを作成するスケジューラー | `go.mod`、`service/snapshot.go` |
| imaging | 1.6.2 | アップロードされたラベル画像の回転・リサイズ | `go.mod`、`service/label_image.go` |
| goexif | `v0.0.0-20190401172101-9e8deecbddbd` | 画像の EXIF Orientation 読み取り | `go.mod`、`service/label_image.go` |

`go.mod` には上記以外にも多くの間接依存関係がありますが、この表ではアプリケーション機能として利用しているもの、または主要な接続基盤に絞っています。

### インフラ・運用技術

| 技術 | バージョン／指定 | 用途 | 確認できる根拠 |
|---|---|---|---|
| PostgreSQL | 17.6（cellar-app 上での確認値） | アプリケーションデータの保存先 | 接続先 DB の `version()` |
| Docker | Dockerfile あり。Docker Engine のバージョン指定なし | Linux 向けバイナリをビルドし、コンテナイメージを構成 | `Dockerfile` |
| Go builder image | `golang:1.25` | Docker のビルドステージ | `Dockerfile` の `FROM golang:1.25 AS builder` |
| Alpine Linux | `alpine:3.20` | Docker の実行ステージ | `Dockerfile` の実行ステージ |
| Linux ARM | `GOOS=linux GOARCH=arm GOARM=6` | Raspberry Pi 向けバイナリのビルド | `build_and_deploy.sh`、`launch-on-raspberry-pi.md` |
| systemd | バージョン未確認 | Raspberry Pi 上でのサービス登録・起動例 | `launch-on-raspberry-pi.md`、`build_and_deploy.sh` の `systemctl` 呼び出し |

このバックエンドのディレクトリには Docker Compose ファイルはありません。周辺の配置環境に Compose 設定があっても、このリポジトリのデプロイ構成としてはここでは扱いません。

## アーキテクチャ概要

```text
HTTP クライアント
    |
    v
router.go (Gin のルート登録)
    |
    v
handler/ (HTTP 入力・応答)
    |
    v
service/ (アプリケーション処理)
    |                         \
    v                          v
repository/ (GORM)         llm/ (Gemini API)
    |
    v
PostgreSQL
```

`main.go` が DB 接続、モデルのマイグレーション、LLM プロバイダーと service の生成を行い、`router.go` で定義した Gin サーバーを起動します。証明書と秘密鍵の両方が設定されている場合は HTTPS（8443）、それ以外は HTTP（8080）で待ち受けます。service は毎日 0 時のスナップショット作成ジョブも開始します。

## 主要なディレクトリ・ファイル

| パス | 役割 |
|---|---|
| `main.go` | アプリケーションのエントリーポイント。DB 接続、AutoMigrate、依存関係の生成、スケジューラーと HTTP サーバーの起動 |
| `router.go` | Gin のルートと各 handler の対応付け |
| `handler/` | HTTP リクエストの受け取り、入力検証、service 呼び出し、HTTP 応答の生成 |
| `service/` | ワイン・ボトル・統計・スナップショット・画像・LLM 連携などのアプリケーション処理 |
| `repository/` | GORM を使ったデータ取得・更新 |
| `model/` | GORM による永続化対象の DB モデル |
| `dto/` | API リクエスト・レスポンス用 DTO と DB モデル間の変換 |
| `database/` | PostgreSQL 接続と、管理対象を列挙した `AutoMigrate` |
| `llm/` | LLM プロバイダー、Gemini 実装、プロンプト、レスポンス解析、ブドウ品種名の正規化、リトライ処理 |
| `go.mod` / `go.sum` | Go モジュールと依存関係のバージョン・チェックサム |
| `Dockerfile` | Docker のビルド・実行ステージ |
| `build_and_deploy.sh` | Linux ARM バイナリの作成と、SSH/SCP/systemd による配備手順 |

## データベースとデータアクセス

- DB は PostgreSQL で、GORM と `gorm.io/driver/postgres` を通じてアクセスします。
- モデルにはワイン、ボトル、国、地域、ワイン種別、アペラシオン、ブドウ品種、在庫スナップショットなどがあります。永続化構造は `model/model.go`、API 用構造は `dto/dto.go` を参照してください。
- `main.go` はDB接続後に `database.Migrate` を呼び出し、`database/migrate.go` に列挙された GORM モデルだけを `AutoMigrate` に渡します。DTO は管理対象に含めません。
- このリポジトリには SQL マイグレーションファイルや初期データ投入スクリプトはありません。2026-10-12 に PostgreSQL 17.6 の本番 DB で AutoMigrate を実行し、外部キー列7個を `bigint` から `integer` に変更しました。対象11テーブルの行数は移行前後で一致しています。
- 適用直前の custom-format バックアップを取得し、PostgreSQL 17.9 の隔離環境へ復元できることを確認しました。復元コピーで AutoMigrate を実行し、読み取り API 20件と DTO を使う作成・更新・削除 API を確認しました。
- 移行前から、同じ外部キー関係に対する重複制約が7組ありました。アペラシオン削除の重複制約は既存の `ON DELETE SET NULL` と競合し、地域削除の重複制約は既存の `ON DELETE CASCADE` と競合していました。隔離コピーで削除動作を検証した後、2026-10-12 に適用直前バックアップを取得し、GORM名の重複制約7個だけを本番から削除しました。既存の正式制約7個と削除動作は維持され、対象テーブルの行数は変更前後で一致し、同一参照先・列への重複外部キーは残っていません。
- モデル管理外のテーブル（例: `auth_*`、`django_*`、`cellars`、`wine_search_caches`）と `wines` の追加列は維持してください。バックエンドとフロントエンドのソース検索では直接参照を確認できませんでしたが、他アプリや運用処理からの利用は未確認です。不要と判断できるまでは削除しないでください。
- DTO を使った隔離 API の作成テストで、ネストされたCountryを省略したWineリクエストが空のCountry行を作る問題を検出し、変換処理を修正しました。アペラシオン削除後の取得が誤ってHTTP 500になる問題も修正し、404を返すようにしました。検証済みの Linux ARM バイナリを2026-10-12に本番へ配備し、サービス稼働と主要な読み取り API 12件のHTTP 200を確認しました。
- DB 接続設定は `database/db.go` が `DB_HOST`、`DB_USER`、`DB_PASSWORD`、`DB_PORT`、`DB_NAME` から読み込みます。
- 移行前バックアップは PostgreSQL 17.9 の隔離環境へ復元して確認済みです。外部アプリ・運用処理から管理外テーブルや列が使われていないかは未確認です。継続的なバックアップ・リストア運用手順も未整備です。

## 外部 API・外部サービス

- **Google Gemini API:** `llm/provider.go` が `LLM_PROVIDER` を読み、現在は `gemini` のみを選択可能です。`llm/gemini_provider.go` が Gemini クライアントを作成し、ワイン名等を含むプロンプトを送信して構造化されたワイン情報を取得します。モデル名は `GEMINI_MODEL` で指定でき、未指定時の既定値はコード上 `gemini-2.5-flash-lite` です。API キーは `GEMINI_API_KEY` から読み込みます。
- **PostgreSQL:** 外部データベースとして接続します。ホストや認証情報の指定方法は「設定管理」を参照してください。
- `SERP_API_KEY` や OpenAI に関する記載が既存ファイルにありますが、現行コードでそれらの API を呼び出す実装は確認できません。利用中の外部連携としては記載していません。

## 設定管理とセキュリティ上の注意点

| 環境変数 | 用途 | 既定値／動作 |
|---|---|---|
| `DB_HOST` | PostgreSQL ホスト | `localhost` |
| `DB_USER` | PostgreSQL ユーザー | `cellar_app` |
| `DB_PASSWORD` | PostgreSQL パスワード | コード内の既定値なし |
| `DB_PORT` | PostgreSQL ポート | `5432` |
| `DB_NAME` | データベース名 | `cellar_app` |
| `SSL_CERT_FILE` | HTTPS 証明書ファイルのパス | 未指定時は HTTP（8080） |
| `SSL_KEY_FILE` | HTTPS 秘密鍵ファイルのパス | 証明書・鍵の両方が設定された場合に HTTPS（8443） |
| `LLM_PROVIDER` | LLM プロバイダー | `gemini`。その他の値は未対応 |
| `GEMINI_API_KEY` | Gemini API キー | 必須。未設定時はプロバイダー初期化エラー |
| `GEMINI_MODEL` | Gemini モデル名 | `gemini-2.5-flash-lite` |
| `IMAGE_OUTPUT_DIR` | 画像出力先 | `./` |

注意事項:

- `main.go` で組み立てる PostgreSQL DSN は `sslmode=disable` です。本番環境では DB 通信経路の保護方法を別途確認してください。
- 証明書と秘密鍵のどちらかが欠けると、サーバーは HTTPS ではなく HTTP で起動します。公開環境では TLS 終端やネットワーク制御を確認してください。
- ルーターには認証ミドルウェアの登録が見当たりません。外部に公開する場合は、認証・アクセス制御を別途設計してください。
- Gemini 連携ではワインの識別情報が外部 API に送信されます。送信するデータと外部サービスの利用条件を運用前に確認してください。
- API キー、DB パスワード、TLS 秘密鍵などの実値は、文書・ソース・リポジトリに記載せず、適切な環境変数や秘密情報管理手段で供給してください。環境設定用スクリプトを扱う場合も、秘密値が含まれていないことを確認してください。

## 開発・テスト

バックエンドのルートディレクトリで実行します。

```sh
go run .
```

Go のテストファイルは `handler/` と `llm/` にあります。テスト実行には Go 標準のコマンドを利用できます。

```sh
go test ./...
```

`.golangci.yml`、`staticcheck.conf`、Makefile などの静的解析・タスクランナー設定は確認できません。Go の標準フォーマッターは `gofmt` ですが、リポジトリ固有のフォーマット設定はありません。

ローカル実行には PostgreSQL 接続設定が必要です。また、現行の起動処理は Gemini プロバイダーを初期化するため、`GEMINI_API_KEY` が必要です。環境変数を設定する具体的なローカル用サンプル設定ファイルは確認できません。

## デプロイ・実行環境

- **Docker:** `Dockerfile` は `golang:1.25` で Linux 向けバイナリをビルドし、`alpine:3.20` を実行ステージとして使います。コンテナはポート 8443 を `EXPOSE` します。一方、アプリケーションコードは TLS 設定がない場合に 8080 を使うため、起動環境の証明書設定とポート公開を合わせる必要があります。Docker Compose による起動手順はこのリポジトリ内では確認できません。
- **Raspberry Pi:** `build_and_deploy.sh` は `GOOS=linux GOARCH=arm GOARM=6` でビルドし、SSH/SCP と systemd で既存サービスを更新します。`launch-on-raspberry-pi.md` に手動セットアップ例があります。対象 Raspberry Pi の具体的な OS・機種・互換性は未確認です。
- **待受ポート:** TLS 証明書と鍵の両方が設定された場合は HTTPS 8443、そうでない場合は HTTP 8080 です。
- **実行時依存:** Dockerfile には PostgreSQL が含まれないため、実行時に接続可能な PostgreSQL と必要な環境変数が別途必要です。

## 関連ドキュメント

- プロジェクト概要・既存の起動案内: [README](../README.md)
- LLM 機能の説明: [LLM_FEATURE.md](../LLM_FEATURE.md)
- Raspberry Pi の起動例: [launch-on-raspberry-pi.md](../launch-on-raspberry-pi.md)

既存ドキュメントには現在の実装と一致しない構成・依存関係の記述があります。技術バージョンや実行動作の確認では、コード、`go.mod`、`Dockerfile` を優先してください。
