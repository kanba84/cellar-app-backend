# データディクショナリ

## 読み方と注意

- この一覧は `cellar-app-backend/model/model.go` の GORM モデル定義に基づきます。2026-10-13 に対象DBのpublic schema（PostgreSQL 17.6）の列・型・NULL・default、インデックス、制約を照会し、モデル定義との差分のみを本書末尾に記載しています。差分として挙げていない確認対象は、その DB・確認時点でモデル定義と整合していました。この結果は他環境や将来のスキーマを保証するものではありません。
- 「宣言型」は明示された GORM `type`／`size` と Go のフィールド型を示します。実 DB で差異が確認された物理型は末尾に記載します。
- nullable は pointer／GORM `not null` 等から読めるモデル上の意図です。主キーは NOT NULL です。タグを持たない値型についても、物理 DB の NULL 可否を断定せず「タグ上未指定」と示します。
- `default` は GORM タグ上の default です。Go の zero value や DTO の zero value が DB default と同じ意味とは限りません。
- テーブルごとの列はモデルから導かれる構造を一覧化します。実 DB にモデル外の列やオブジェクトが存在しないという意味ではありません。
- 産地マスターの追加列・indexはコードに実装済みですが、対応SQL移行は稼働DBに適用していません。

## `wines`

ワインそのものの属性を管理します。個別ボトルは `bottles`、品種構成は `wine_grapes` に分かれています。

| カラム | 宣言型 | NULL | default | 制約／意味・補足 |
|---|---|---|---|---|
| `id` | Go `uint`; GORM `integer` | 不可 | 明示タグなし | 主キー。モデルで `type:integer;primaryKey` |
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
| `id` | Go `uint`; GORM `integer` | 不可 | 明示タグなし | 主キー |
| `wine_id` | Go `uint`; GORM `integer` | 不可 | タグ指定なし | `wines.id` 参照、明示 `ON DELETE CASCADE` |
| `is_opened` | `*bool`; GORM `bool` | 可 | `false` | 開封状態。DB default 指定あり |
| `added_at` | `*time.Time`; GORM `timestamp` | 可 | `CURRENT_TIMESTAMP` | セラー追加日時 |
| `removed_at` | `*time.Time`; GORM の型推論 | 可 | タグ指定なし | NULL は現行在庫、値ありは在庫から除外した状態 |
| `row_number` | `*int`; GORM `integer` | 可 | タグ指定なし | セラーの行位置。`column_number` と複合 `unique_position` |
| `column_number` | `*int`; GORM `integer` | 可 | タグ指定なし | セラーの列位置。`row_number` と複合 `unique_position` |
| `note` | `*string`; GORM `text` | 可 | タグ指定なし | ボトル個別メモ |
| `Wine` | 関連フィールド | — | — | `wine_id` からの GORM 関連。列ではない |

`unique_position` は `row_number` と `column_number` の複合一意インデックスです。PostgreSQL の通常の UNIQUE では NULL を含む組が複数許可され得るため、NULL を含む棚位置の一意性は保証されません。部分条件はモデルに定義されていないため、除外済みボトルもインデックスの対象です。

## `countries`

| カラム | 宣言型 | NULL | default | 制約／意味・補足 |
|---|---|---|---|---|
| `id` | Go `uint`; GORM `integer` | 不可 | 明示タグなし | 主キー |
| `name` | string / `varchar(255)` | 不可 | タグ指定なし | 国名。一意インデックス |
| `iso_code` | `*string`; `varchar(2)` | 可 | タグ指定なし | ISO 3166-1 alpha-2想定。今回のSQL migrationで一意indexを追加。NULLは移行期に許容 |

## `wine_types`

| カラム | 宣言型 | NULL | default | 制約／意味・補足 |
|---|---|---|---|---|
| `id` | Go `uint`; GORM `integer` | 不可 | 明示タグなし | 主キー |
| `name` | string / `varchar(255)` | 不可 | タグ指定なし | ワイン種別名。一意インデックス |

## `regions`

| カラム | 宣言型 | NULL | default | 制約／意味・補足 |
|---|---|---|---|---|
| `id` | Go `uint`; GORM `integer` | 不可 | 明示タグなし | 主キー |
| `name` | string / `varchar(255)` | 不可 | タグ指定なし | 地域名。`parent_id` と複合一意 `regions_name_parent_id_key` |
| `code` | `*string`; `varchar(64)` | 可 | タグ指定なし | 国コードを含むSeed識別子。一意制約。既存行は移行後Seedで補完 |
| `country_id` | Go `uint`; GORM `integer` | 不可 | タグ指定なし | `countries.id` 参照、明示 `NO ACTION` |
| `parent_id` | `*uint`; GORM `bigint` | 可 | タグ指定なし | 親地域 ID。自己FK `regions_parent_id_fkey`、削除時 `NO ACTION` |
| `sort_order` | `int`; GORM `integer` | 不可 | `0` | 表示順。IDから独立 |
| `Country` | 関連フィールド | — | — | `country_id` からの関連。列ではない |
| `Parent` | 関連フィールド | — | — | `parent_id` からの自己参照。列ではない |

PostgreSQL の UNIQUE と NULL の扱いにより、`parent_id IS NULL` の同名行が複数許可される場合があります。国を含めた一意性の範囲は現モデルのインデックス指定ではありません。

## `appellations`

| カラム | 宣言型 | NULL | default | 制約／意味・補足 |
|---|---|---|---|---|
| `id` | Go `uint`; GORM `integer` | 不可 | 明示タグなし | 主キー |
| `name` | string / `varchar(255)` | 不可 | タグ指定なし | アペラシオン／呼称名 |
| `code` | `*string`; `varchar(100)` | 可 | タグ指定なし | 国・制度の名前空間を含むSeed識別子。一意制約 |
| `country_id` | `*uint`; GORM `integer` | 可 | タグ指定なし | `countries.id` 参照。既存Regionから移行時に補完 |
| `designation_type_id` | Go `*uint`; GORM `integer` | 可 | タグ指定なし | `designation_types.id` 参照、削除時 `SET NULL` |
| `region_id` | Go `*uint`; GORM `integer` | 可 | タグ指定なし | 既存API互換用の `regions.id` 参照、削除時 `RESTRICT` |
| `sort_order` | `int`; GORM `integer` | 不可 | `0` | 表示順。IDから独立 |
| `Regions` | 関連slice | — | — | `appellation_regions` を介す多対多関連。複数地域にまたがる呼称を表現 |
| `Country` | 関連フィールド | — | — | `country_id` からの関連。列ではない |
| `DesignationType` | 関連フィールド | — | — | 種別関連。列ではない |
| `Region` | 関連フィールド | — | — | 互換用 `region_id` 関連。列ではない |

`appellation_regions` は `(appellation_id, region_id)` を複合主キーとするJoin tableです。SQL移行は既存 `region_id` を関連へコピーし、legacy列を保ったまま段階移行します。

## `appellation_regions`

多対多関連用テーブルです。作成SQLはAppellation／Region削除時にJoin行をCASCADE削除します。

| カラム | 宣言型 | NULL | default | 制約／意味・補足 |
|---|---|---|---|---|
| `appellation_id` | Go `uint`; GORM `integer` | 不可 | タグ指定なし | `appellations.id` 参照。複合主キー |
| `region_id` | Go `uint`; GORM `integer` | 不可 | タグ指定なし | `regions.id` 参照。複合主キー、検索index |

## `designation_types`

| カラム | 宣言型 | NULL | default | 制約／意味・補足 |
|---|---|---|---|---|
| `id` | Go `uint`; 型タグなし | 不可 | 明示タグなし | 主キー |
| `name` | string / `varchar(255)` | 不可 | タグ指定なし | 呼称種別名 |
| `code` | string / `varchar(50)` | タグ上未指定 | タグ指定なし | 国ごとの呼称制度コード。`country_id` と複合一意index |
| `rank` | Go `int`; 型タグなし | タグ上未指定 | タグ指定なし | 順位。単位・値域は未指定 |
| `country_id` | `*uint`; GORM `bigint` | 可 | タグ指定なし | `countries.id` 参照。明示的なCountry関連、削除時 `CASCADE` |
| `Country` | 関連フィールド | — | — | `country_id` からの関連。列ではない |

## `grapes`

| カラム | 宣言型 | NULL | default | 制約／意味・補足 |
|---|---|---|---|---|
| `id` | Go `uint`; 型タグなし | 不可 | 明示タグなし | 主キー |
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
| `id` | Go `uint`; 型タグなし | 不可 | 明示タグなし | 主キー |
| `snapshot_date` | `date` | 不可 | タグ指定なし | 親 `inventory_daily_snapshots.snapshot_date` 参照。日付インデックス |
| `category_type` | `text` | 不可 | タグ指定なし | CHECK: `wine_type` / `country` / `vintage`。カテゴリインデックス |
| `category_id` | `*int`; GORM `integer` | 可 | タグ指定なし | カテゴリ ID。カテゴリインデックス |
| `category_name` | `*string`; `text` | 可 | タグ指定なし | 集計時に表示するカテゴリ名 |
| `count` | `integer` | 不可 | タグ指定なし | カテゴリ内の在庫件数 |
| `created_at` | GORM 推論時刻型 | 不可 | `CURRENT_TIMESTAMP` | GORM `autoCreateTime` |

## 型・業務仕様の未確定事項

- `reference_price` の通貨・精度、品種 `percentage` の単位と合計ルール、`display_order` の基点、地域番号の意味・一意性はモデルから確認できません。
- `DesignationType.CountryID` と `Region.ParentID` はモデル上の関連・外部キー関係を定義しています。
- `DesignationType.Code` と `Rank` には `not null` タグがありません。空文字やゼロの業務上の扱いはモデルだけでは決まりません。
- API DTO の数値型・JSON 名・null 処理は DB 型とは別の契約です。詳細は `cellar-app-backend/dto/dto.go` を参照してください。

## 実 DB との差分（2026-10-13）

確認先は対象DBのpublic schema（PostgreSQL 17.6）です。次の差分以外の確認対象は、確認時点でモデル定義と整合していました。

| 対象 | 差分 |
|---|---|
| 産地マスターの移行状態 | 追加列・index・多対多Join tableを含む `20261101_wine_geography_master_data.sql` は未適用。 |
| `regions.parent_id` | 確認時の実 DB に `regions_parent_id_fkey` がある（`regions(id)`、`NO ACTION`）。現行モデルにも自己参照関連を定義済み。 |
| `designation_types.country_id` | 確認時の実 DB に `designation_types_country_id_fkey` がある（`countries(id)`、`CASCADE`）。現行モデルにもCountry関連を定義済み。 |
| `wines` のモデル外列 | `grape_varieties`、`notes`、`label_image_path`、`quantity`、`price`、`is_favorite`、`created_at`、`updated_at` が実 DB に存在する。既存列であり、未使用・削除対象とは推定しない。 |
| 物理型 | `wines.reference_price` と `wine_grapes.percentage` は実 DB で `numeric`（モデルは `*float64`、DB 型タグなし）。`designation_types.id` は実 DB で `integer`（モデルは型タグなしの `uint`）。 |

この記録は上記 DB の確認時点に限られます。別環境やスキーマ変更後は差分を再照合してください。

## 参照

- [データモデル](./data-model.md)
- [スキーマ管理方針](./schema-management.md)
- [DB運用手順](./database-operations.md)
