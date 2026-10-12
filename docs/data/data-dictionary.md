# データディクショナリ

## 読み方と注意

- この一覧は `cellar-app-backend/model/model.go` の GORM モデル定義に基づく**コード上の定義**です。実 DB の列・型・NULL・default・索引・制約を接続して調査したものではありません。
- 「宣言型」は明示された GORM `type`／`size` と Go のフィールド型を示します。明示 DB 型のない Go 型は PostgreSQL Dialector が生成する実型の確認が必要です。Go `int`・`uint`・`float64` を DB の `integer` 等と同一視しません。
- nullable は pointer／GORM `not null` 等から読めるモデル上の意図です。主キーは NOT NULL です。タグを持たない値型についても、物理 DB の NULL 可否を断定せず「タグ上未指定」と示します。
- `default` は GORM タグ上の default です。Go の zero value や DTO の zero value が DB default と同じ意味とは限りません。
- テーブルごとの列はモデルから導かれる構造を一覧化します。実 DB にモデル外の列やオブジェクトが存在しないという意味ではありません。

## `wines`

ワインそのものの属性を管理します。個別ボトルは `bottles`、品種構成は `wine_grapes` に分かれています。

| カラム | 宣言型 | NULL | default | 制約／意味・補足 |
|---|---|---|---|---|
| `id` | Go `uint`; GORM `integer` | 不可 | 明示タグなし。GORM の整数主キーは自動採番扱い（実 DDL 未確認） | 主キー。モデルで `type:integer;primaryKey` |
| `name` | string / `varchar(255)` | 不可 | タグ指定なし | ワイン表示名 |
| `country_id` | Go `uint`; GORM `integer` | 不可 | タグ指定なし | `countries.id` 参照、明示 `NO ACTION` |
| `wine_type_id` | Go `uint`; GORM `integer` | 不可 | タグ指定なし | `wine_types.id` 参照、明示 `NO ACTION` |
| `vintage` | `*int`; CHECK 式で列型指定なし | 可 | タグ指定なし | 1800以上かつ現在年以下。未設定可 |
| `drinking_window_start` | `*int`; 型タグなし | 可 | タグ指定なし | 飲み頃開始年。モデル上の範囲 CHECK なし |
| `drinking_window_end` | `*int`; 型タグなし | 可 | タグ指定なし | 飲み頃終了年。handler は開始年との順序を検証 |
| `region_id` | Go `*uint`; GORM `integer` | 可 | タグ指定なし | `regions.id` 参照、明示 `NO ACTION` |
| `producer` | `*string`; `size:255` | 可 | タグ指定なし | 生産者名 |
| `appellation_id` | Go `*uint`; GORM `integer` | 可 | タグ指定なし | `appellations.id` 参照、削除時 `SET NULL` |
| `label_image_url` | `*string`; `size:512` | 可 | タグ指定なし | ラベル画像 URL／パス。保存先仕様は別途確認 |
| `reference_price` | `*float64`; 型タグなし | 可 | タグ指定なし | 参考価格。通貨・スケール・丸め規則は未定義 |
| `Country` | 関連フィールド | — | — | `country_id` からの GORM 関連。列ではない |
| `WineType` | 関連フィールド | — | — | `wine_type_id` からの GORM 関連。列ではない |
| `Region` | 関連フィールド | — | — | `region_id` からの GORM 関連。列ではない |
| `Appellation` | 関連フィールド | — | — | `appellation_id` からの GORM 関連。列ではない |
| `WineGrapes` | 関連 slice | — | — | 品種関連。`wine_grapes` に保存 |

## `bottles`

在庫の単位となる個々のボトルを管理します。GORM の既定テーブル名は `bottles` です。モデルの外部キー制約名は既存名 `cellar_bottles_wine_id_fkey` と明示されていますが、テーブル名自体は `Bottle` モデルのスキーマテストで `bottles` と確認されています。

| カラム | 宣言型 | NULL | default | 制約／意味・補足 |
|---|---|---|---|---|
| `id` | Go `uint`; GORM `integer` | 不可 | 明示タグなし。GORM の整数主キーは自動採番扱い（実 DDL 未確認） | 主キー |
| `wine_id` | Go `uint`; GORM `integer` | 不可 | タグ指定なし | `wines.id` 参照、明示 `ON DELETE CASCADE` |
| `is_opened` | `*bool`; GORM `bool` | 可 | `false` | 開封状態。DB default 指定あり |
| `added_at` | `*time.Time`; GORM `timestamp` | 可 | `CURRENT_TIMESTAMP` | セラー追加日時。タイムゾーン型ではないタグ |
| `removed_at` | `*time.Time`; GORM の型推論 | 可 | タグ指定なし | NULL は現行在庫、値ありは在庫から除外した状態 |
| `row_number` | `*int`; GORM `integer` | 可 | タグ指定なし | セラーの行位置。`column_number` と複合 `unique_position` |
| `column_number` | `*int`; GORM `integer` | 可 | タグ指定なし | セラーの列位置。`row_number` と複合 `unique_position` |
| `note` | `*string`; GORM `text` | 可 | タグ指定なし | ボトル個別メモ |
| `Wine` | 関連フィールド | — | — | `wine_id` からの GORM 関連。列ではない |

棚位置が NULL を含む場合の一意性や、除外済みボトルも同じ位置制約の対象となるかは実 DB の索引定義を確認してください。モデルに部分一意インデックスの条件はありません。

## `countries`

| カラム | 宣言型 | NULL | default | 制約／意味・補足 |
|---|---|---|---|---|
| `id` | Go `uint`; GORM `integer` | 不可 | 明示タグなし。GORM の整数主キーは自動採番扱い（実 DDL 未確認） | 主キー |
| `name` | string / `varchar(255)` | 不可 | タグ指定なし | 国名。一意インデックス |
| `iso_code` | `*string`; `varchar(2)` | 可 | タグ指定なし | ISO コード。2文字上限。形式や大文字化は未指定 |

## `wine_types`

| カラム | 宣言型 | NULL | default | 制約／意味・補足 |
|---|---|---|---|---|
| `id` | Go `uint`; GORM `integer` | 不可 | 明示タグなし。GORM の整数主キーは自動採番扱い（実 DDL 未確認） | 主キー |
| `name` | string / `varchar(255)` | 不可 | タグ指定なし | ワイン種別名。一意インデックス |

## `regions`

| カラム | 宣言型 | NULL | default | 制約／意味・補足 |
|---|---|---|---|---|
| `id` | Go `uint`; GORM `integer` | 不可 | 明示タグなし。GORM の整数主キーは自動採番扱い（実 DDL 未確認） | 主キー |
| `name` | string / `varchar(255)` | 不可 | タグ指定なし | 地域名。`parent_id` と複合一意 `regions_name_parent_id_key` |
| `country_id` | Go `uint`; GORM `integer` | 不可 | タグ指定なし | `countries.id` 参照、明示 `NO ACTION` |
| `parent_id` | `*uint`; 型タグなし | 可 | タグ指定なし | 親地域 ID 値。自己 FK はモデルに定義されていない |
| `Country` | 関連フィールド | — | — | `country_id` からの関連。列ではない |

PostgreSQL の UNIQUE と NULL の扱いにより、`parent_id IS NULL` の同名行が複数許可される場合があります。国を含めた一意性の範囲は現モデルのインデックス指定ではありません。

## `appellations`

| カラム | 宣言型 | NULL | default | 制約／意味・補足 |
|---|---|---|---|---|
| `id` | Go `uint`; GORM `integer` | 不可 | タグ指定なし | 主キー |
| `name` | string / `varchar(255)` | 不可 | タグ指定なし | アペラシオン／呼称名 |
| `designation_type_id` | Go `*uint`; GORM `integer` | 可 | タグ指定なし | `designation_types.id` 参照、削除時 `SET NULL` |
| `region_id` | Go `*uint`; GORM `integer` | 可 | タグ指定なし | `regions.id` 参照、削除時 `CASCADE` |
| `DesignationType` | 関連フィールド | — | — | 種別関連。列ではない |
| `Region` | 関連フィールド | — | — | 地域関連。列ではない |

## `designation_types`

| カラム | 宣言型 | NULL | default | 制約／意味・補足 |
|---|---|---|---|---|
| `id` | Go `uint`; 型タグなし | 不可 | 明示タグなし。GORM の整数主キーは自動採番扱い（実 DDL 未確認） | 主キー。実 DB 型は未確認 |
| `name` | string / `varchar(255)` | GORM の必須タグなし | タグ指定なし | 呼称種別名。モデルには `not null` 指定なし |
| `code` | string / `varchar(50)` | GORM の必須タグなし | タグ指定なし | 外部呼称コード等。意味・一意性は未指定 |
| `rank` | Go `int`; 型タグなし | GORM の必須タグなし | タグ指定なし | 順位。単位・値域は未指定 |
| `country_id` | `*uint`; 型タグなし | 可 | タグ指定なし | 国に関係する ID 値。モデルに Country 関連・FK なし |

## `grapes`

| カラム | 宣言型 | NULL | default | 制約／意味・補足 |
|---|---|---|---|---|
| `id` | Go `uint`; 型タグなし | 不可 | 明示タグなし。GORM の整数主キーは自動採番扱い（実 DDL 未確認） | 主キー。実 DB 型は未確認 |
| `name` | string / `varchar(255)` | 不可 | タグ指定なし | 品種名。一意インデックス |

## `wine_grapes`

| カラム | 宣言型 | NULL | default | 制約／意味・補足 |
|---|---|---|---|---|
| `wine_id` | Go `uint`; 型タグなし | 不可 | タグ指定なし | 複合主キーの一部。Wine との関連 FK 列 |
| `grape_id` | Go `uint`; 型タグなし | 不可 | タグ指定なし | 複合主キーの一部。Grape との関連 FK 列 |
| `percentage` | `*float64`; 型タグなし | 可 | タグ指定なし | 品種構成割合。単位・合計規則・精度未定義 |
| `display_order` | `*int`; 型タグなし | 可 | タグ指定なし | 表示順。既定値・値域未定義 |
| `Grape` | 関連フィールド | — | — | `grape_id` からの関連。列ではない |

## `inventory_daily_snapshots`

| カラム | 宣言型 | NULL | default | 制約／意味・補足 |
|---|---|---|---|---|
| `snapshot_date` | `date` | 不可 | タグ指定なし | 主キー。集計対象日 |
| `total_count` | `integer` | 不可 | タグ指定なし | その日の総在庫数 |
| `created_at` | GORM 推論時刻型 | 不可 | `CURRENT_TIMESTAMP` | 作成時刻、GORM `autoCreateTime` |
| `updated_at` | GORM 推論時刻型 | 不可 | `CURRENT_TIMESTAMP` | 更新時刻、GORM `autoUpdateTime` |
| `Details` | 関連 slice | — | — | `snapshot_date` を参照する詳細行。列ではない |

## `inventory_daily_snapshot_details`

| カラム | 宣言型 | NULL | default | 制約／意味・補足 |
|---|---|---|---|---|
| `id` | Go `uint`; 型タグなし | 不可 | 明示タグなし。GORM の整数主キーは自動採番扱い（実 DDL 未確認） | 主キー。実 DB 型は未確認 |
| `snapshot_date` | `date` | 不可 | タグ指定なし | 親 `inventory_daily_snapshots.snapshot_date` 参照。日付インデックス |
| `category_type` | `text` | 不可 | タグ指定なし | CHECK: `wine_type` / `country` / `vintage`。カテゴリインデックス |
| `category_id` | `*int`; GORM `integer` | 可 | タグ指定なし | カテゴリ ID。カテゴリインデックス |
| `category_name` | `*string`; `text` | 可 | タグ指定なし | 集計時に表示するカテゴリ名 |
| `count` | `integer` | 不可 | タグ指定なし | カテゴリ内の在庫件数 |
| `created_at` | GORM 推論時刻型 | 不可 | `CURRENT_TIMESTAMP` | GORM `autoCreateTime` |

`idx_snapshot_details_category` は `category_type` と `category_id` のタグで同じインデックス名を指定しています。実際の複合インデックス列順・定義は PostgreSQL カタログで確認してください。

## 型・業務仕様の未確定事項

- `reference_price` の通貨・精度、品種 `percentage` の単位と合計ルール、`display_order` の基点、地域番号の意味・一意性はモデルから確認できません。
- `DesignationType.CountryID` と `Region.ParentID` は ID 値として存在しますが、モデル上の外部キー関係はありません。
- pointer ではない `DesignationType.Name`、`Code`、`Rank` には `not null` タグがありません。空文字やゼロの扱いと実 DB NULL 制約は分けて確認してください。
- タグのない Go 数値型や時刻型の PostgreSQL 物理型、既存 DB の追加列・インデックス・制約は未確認です。
- API DTO の数値型・JSON 名・null 処理は DB 型とは別の契約です。詳細は `cellar-app-backend/dto/dto.go` を参照してください。

## 参照

- [データモデル](./data-model.md)
- [スキーマ管理方針](./schema-management.md)
- [DB運用手順](./database-operations.md)
