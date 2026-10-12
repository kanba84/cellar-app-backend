# DB運用手順

この手順は PostgreSQL と GORM を使う Cellar App 向けの運用ガイドです。破壊的操作・本番復元は必ず作業承認と復旧計画のもとで行ってください。本書のコマンド例はテンプレートであり、実DBへ実行したものではありません。

## 現行構成と確認できた範囲

- バックエンドの `database.Connect()` は `DB_HOST`、`DB_PORT`、`DB_USER`、`DB_PASSWORD`、`DB_NAME` から接続先を作り、`sslmode=disable` を指定します。既定値はホスト `localhost`、ポート `5432`、ユーザー／DB 名 `cellar_app` です。
- Docker Compose の DB サービスは `postgres:17`、永続領域は named volume `pgdata` です。初期化ディレクトリ `db/init` を `/docker-entrypoint-initdb.d` にマウントし、ホスト側 5432 番を公開しています。
- Compose は DB コンテナーに `POSTGRES_USER`／`POSTGRES_DB` を設定しますが、API コンテナーへ明示する DB 接続値は `DB_HOST` と `DB_PASSWORD` のみです。API は未設定時に `DB_USER`／`DB_NAME` を `cellar_app`、`DB_PORT` を `5432` とするため、実環境の PostgreSQL 初期値と一致するか起動前に照合してください。
- 同ディレクトリの `restore.sh` は `POSTGRES_USER`／`POSTGRES_DB` を使って `cellar_app.dump` を `pg_restore --no-owner` で復元します。公式 PostgreSQL コンテナーの init hook はデータディレクトリ初期化時に実行されるため、既存 `pgdata` の通常再起動でこの dump が再適用される仕組みではありません。
- Raspberry Pi 向け資料・スクリプトでは `cellar-app` という SSH 接続先エイリアス、`cellar-app.service` という API の systemd サービスが使われます。API のホストと DB は別設定で、Pi 上の PostgreSQL のサービス名、実 DB 名・ユーザー名・現在の接続先バージョン、バックアップ保存先、バックアップスケジュールはコードから確定できません。
- バックエンド既存資料には2026-10-12に本番 PostgreSQL 17.6 を扱った過去記録がありますが、今回ライブ接続はしていません。現DBが現在も同じバージョン・構成であるとは断定しません。
- `pg_dump` を定期実行するスクリプトや復元訓練の継続的な運用記録は確認できません。バックアップ自動化・保持期間は運用者が別途決定し、定期化する必要があります。

## 接続と読み取り専用の確認

接続値は安全な運用環境から設定します。パスワードをコマンド引数、シェル履歴、資料、ログに埋め込まないでください。`pg_dump`／`psql` は `--password` で対話入力するか、権限を制限した `.pgpass`／`PGPASSFILE` を使用します。

```sh
export DB_HOST='<db-host>'
export DB_PORT='5432'
export DB_USER='<db-user>'
export DB_NAME='<db-name>'

psql --host "$DB_HOST" --port "$DB_PORT" --username "$DB_USER" \
  --password --dbname "$DB_NAME" --command 'SELECT version();'
psql --host "$DB_HOST" --port "$DB_PORT" --username "$DB_USER" \
  --password --dbname "$DB_NAME" --command '\dt'
```

実際の列・索引・制約は `psql` の `\d+ <table>` や PostgreSQL カタログを読み取り専用で確認し、[データディクショナリ](./data-dictionary.md)のコード由来の期待値と比較します。`\dt` の結果が GORM モデルに列挙したテーブル以外を含んでも、管理外テーブルの可能性があるため削除しません。

## 初期化と AutoMigrate のタイミング

1. Docker 構成では、PostgreSQL コンテナーが空のデータディレクトリで初期化される際、`db/init/restore.sh` が dump を復元します。既存 volume を消して初期化し直すことはデータを失うため禁止です。
2. Go API の `main()` は DB 接続後、`database.Migrate()` を実行します。`Migrate()` は `database/migrate.go` の11モデルで `AutoMigrate` を行い、その完了後に LLM とサービスを初期化して API サーバーを起動します。
3. DB が接続不能または AutoMigrate が失敗すると API の起動は継続しません。Compose の `depends_on` は DB のヘルスチェック完了を意味する設定ではありません。

開発者は起動により対象 DB のスキーマが変わる可能性を理解し、共有 DB や本番 DB に未レビューのバイナリを接続しないでください。

## スキーマ変更前のバックアップ

### 事前条件

- 適用対象 DB・接続ユーザー・PostgreSQL バージョンを確認します。
- 書き込み量、空き容量、適用時間、停止要否、復旧責任者を確認します。
- バックアップを DB サーバーと異なる安全な保存先へ置き、アクセス権と暗号化・保持期間を確認します。
- パスワードは対話入力または安全なパスワードファイルから供給します。シェルの `set -x` を有効にせず、秘密を表示・記録しません。

### custom format のバックアップ例

クライアント側に `pg_dump` があり、DB へのネットワーク接続が許可されている環境で実行します。接続情報は実環境の値で置き換えてください。

```sh
export DB_HOST='<db-host>'
export DB_PORT='5432'
export DB_USER='<db-user>'
export DB_NAME='<db-name>'
BACKUP_FILE='<secure-backup-dir>/<db-name>-<YYYYMMDD-HHMMSS>.dump'

umask 077
pg_dump --host "$DB_HOST" --port "$DB_PORT" --username "$DB_USER" \
  --password --format=custom --file "$BACKUP_FILE" "$DB_NAME"
chmod 600 "$BACKUP_FILE"
pg_restore --list "$BACKUP_FILE" > "$BACKUP_FILE.toc"
```

`pg_dump --format=custom` は既存初期化スクリプトの `pg_restore` と互換な形式です。TOC 一覧が読めても完全復元の証明ではありません。別 DB／隔離 PostgreSQL への復元試験を定期実施してください。Pi 上で SSH 実行する場合は、実際の接続先と保存先が確認できてから、DB の同一ホスト上に一時ファイルを作成し、転送後に checksum と権限を確認して不要な一時ファイルを削除します。

## スキーマ変更の作業フロー

詳細な設計ルールは[スキーマ管理方針](./schema-management.md)を参照してください。

### 新しいテーブル／nullable カラム

1. `model/` のモデルを追加・更新し、必要なら DTO と変換処理も更新します。
2. 新テーブルなら `database/migrate.go` の対象一覧に明示追加します。既存テーブルの列追加なら対象モデルの既存列定義を更新します。
3. 主キー、NULL、default、FK 削除動作、一意性、インデックス、互換性をレビューし、モデルテストを更新します。
4. `go test ./...` を実行します。テスト成功だけでは実 DB DDL が検証されないので、隔離 PostgreSQL に現行データの安全なコピーを復元し、実バイナリで AutoMigrate を試します。
5. 前後のスキーマ、制約、インデックス、行数、代表データ、アプリ読み書きを確認します。
6. 本番バックアップと復元試験が確認できてから、承認済みバイナリを配備して起動します。起動ログにマイグレーション成功があること、主要 API、DB の状態を確認します。

### 破壊的変更

カラム削除・リネーム、型変更、NOT NULL・一意制約追加、既存値変換、インデックス置換は AutoMigrate だけで適用しません。

1. 参照コード、API クライアント、他サービス、DB 内の全依存とデータ影響を確認します。
2. 変更手順を段階化し、必要ならレビュー済み SQL を別作業として準備します。移行ルールと逆操作／復旧可能点を明記します。
3. 本番と同系統の PostgreSQL にバックアップを復元し、変換前後件数・値・重複・NULL・FK 整合性を確認します。
4. 変更直前に本番バックアップを新規取得し、ダンプの TOC 確認と復元可能な保存を確認します。
5. 影響に応じて API 書き込みを停止／制限し、合意された SQL と互換性のあるアプリ版を適用します。
6. 変更後のカタログ、件数、代表レコード、読み書き、ログを検証します。問題があれば安易に逆 SQL を適用せず、バックアップと承認済み復旧手順に従います。

### API／サービスの再起動

Pi への API 配備スクリプトは `cellar-app.service` を停止し、新しいバイナリを転送して restart します。DB のメンテナンスと API 配備を混同せず、データベースの再起動・復元が必要な場合は別途 DB サービス名と停止影響を管理者が確認してください。

```sh
ssh <ssh-alias> "sudo systemctl status cellar-app.service"
ssh <ssh-alias> "sudo systemctl restart cellar-app.service"
ssh <ssh-alias> "sudo systemctl status cellar-app.service"
ssh <ssh-alias> "sudo journalctl -u cellar-app.service --since '<start-time>' --no-pager"
```

現行配備スクリプトの接続先は SSH エイリアス `cellar-app` ですが、上記では別環境でも使えるよう `<ssh-alias>` としています。

## バックアップ復元

復元先はまず**新しい隔離 DB**とし、本番 DB を直接上書きしません。稼働 DB への復元が必要な場合は、停止・切替・データ損失範囲を含む個別の復旧計画を承認してから行います。

```sh
export DB_HOST='<restore-host>'
export DB_PORT='5432'
export DB_USER='<restore-user>'
RESTORE_DB='<new-isolated-db-name>'
BACKUP_FILE='<secure-backup-dir>/<db-name>-<YYYYMMDD-HHMMSS>.dump'

createdb --host "$DB_HOST" --port "$DB_PORT" --username "$DB_USER" \
  --password --template template0 "$RESTORE_DB"
pg_restore --host "$DB_HOST" --port "$DB_PORT" --username "$DB_USER" \
  --password --exit-on-error --no-owner --dbname "$RESTORE_DB" "$BACKUP_FILE"
```

この例の実行ユーザーには、DB 作成権限と復元先への必要権限が必要です。バックアップの PostgreSQL バージョン、所有者、拡張機能、ロール、照合順序などに応じて調整してください。`--no-owner` は所有者を復元しない指定で、既存 `restore.sh` でも使われています。ロール／権限が必要な運用では別途管理された復旧手順が必要です。

復元後にカタログ、テーブル件数、代表データを検証し、隔離先に向けて API の起動・AutoMigrate・読み取りとテスト書き込みを確認します。合格するまで本番への切替や元 DB の削除をしません。

## 障害時の初動

1. 追加書き込みや再起動を繰り返さず、発生時刻、影響範囲、直近の配備・DB 変更を記録します。
2. API の `cellar-app.service` 状態・ログ、PostgreSQL サービス状態・ログ、空きディスク、ネットワーク／DNS、接続設定を読み取り中心に確認します。DB のサービス名や配置が未確認なら、Pi 管理者に確認します。
3. アプリ起動時の AutoMigrate エラー、DB 接続エラー、制約違反を分離します。安易にテーブルや制約を削除して起動させないでください。
4. データ破損または誤操作の可能性がある場合は、状態を保全し、現在のバックアップと復元可能性を確認します。承認なしに古いダンプを本番へ復元しないでください。
5. 切り戻しはアプリケーションバイナリのみで済むか、DB スキーマ・データも戻す必要があるかを区別し、責任者の判断で実施します。

## 環境ごとの注意

- **ローカル／検証:** 本番データではなく合成データまたはアクセス制御された匿名化コピーを使用します。起動時 AutoMigrate を確かめる場として利用し、破壊的変更は disposable DB から試します。
- **Docker Compose:** `pgdata` は永続 volume です。`docker compose down -v` 等 volume を削除する操作を本番・共有環境で実行しません。init dump は初回 DB 初期化時だけ復元します。
- **Raspberry Pi／本番:** 現在の DB サービス名、DB がローカルか別ホストか、バックアップ保存先・頻度・保持期間・暗号化・復元担当を現地で確認して記録します。アプリ接続 DSN が `sslmode=disable` なので、ネットワーク境界・トンネル・ファイアウォールで DB 通信経路が保護されているか確認します。
- **認証情報:** 実値をリポジトリ、Markdown、端末ログ、コマンド履歴に記録しません。設定ファイルや環境スクリプトは権限を限定し、リポジトリに置かない方法を採用します。既存の環境設定ファイルに秘密値が含まれる場合は、追跡状態を確認し、漏えい範囲に応じて資格情報を交換します。
- **調査で確認した要対応事項:** 調査環境のバックエンド `.env.sh` に実認証情報が設定されていました。このファイルは調査時点で Git ignore 対象かつ未追跡でしたが、ignore は秘密情報の保護を保証しません。本書には値を転記していません。ファイルのアクセス権・共有範囲を確認し、値が使用中または外部共有済みなら資格情報を交換してください。設定値をそのまま運用コマンドや資料へコピーしないでください。

関連資料：[データモデル](./data-model.md) · [スキーマ管理方針](./schema-management.md) · [データディクショナリ](./data-dictionary.md)
