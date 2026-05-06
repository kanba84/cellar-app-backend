package model

import "time"

type Wine struct {
	ID            uint         `gorm:"primaryKey" json:"id"`
	Name          string       `gorm:"size:255;not null" json:"name"`
	CountryID     uint         `gorm:"not null" json:"country_id"`
	Country       Country      `gorm:"foreignKey:CountryID" json:"country,omitempty"`
	WineTypeID    uint         `gorm:"not null" json:"wine_type_id"`
	WineType      WineType     `gorm:"foreignKey:WineTypeID" json:"wine_type,omitempty"`
	Vintage       *int         `json:"vintage"`
	RegionID      *uint        `json:"region_id"`
	Region        *Region      `gorm:"foreignKey:RegionID" json:"region,omitempty"`
	Producer      *string      `gorm:"size:255" json:"producer"`
	AppellationID *uint        `json:"appellation_id"`
	Appellation   *Appellation `gorm:"foreignKey:AppellationID" json:"appellation,omitempty"`
	LabelImageURL *string      `gorm:"size:512" json:"label_image_url"`
}

// WineDTOは、APIレスポンス用のWineデータを表す構造体です。
// Wine構造体をベースに、必要なフィールドのみを含めています。
type WineDTO struct {
	ID                  int     `json:"id"`
	Name                string  `json:"name"`
	CountryID           int     `json:"country_id"`
	CountryName         string  `json:"country_name"`
	CountryISOCode      string  `json:"country_iso_code"`
	WineTypeID          int     `json:"wine_type_id"`
	WinTypeName         string  `json:"wine_type_name"`
	Vintage             *int    `json:"vintage"`
	RegionID            *int    `json:"region_id"`
	RegionName          *string `json:"region_name"`
	Producer            *string `json:"producer"`
	AppellationID       *int    `json:"appellation_id"`
	AppellationName     *string `json:"appellation_name"`
	DesignationTypeID   *int    `json:"designation_type_id"`
	DesignationTypeName *string `json:"designation_type_name"`
	LabelImageURL       *string `json:"label_image_url"`

	HasStock   bool  `json:"has_stock"`
	StockCount int64 `json:"stock_count"`
}

type Bottle struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	WineID       uint       `gorm:"not null" json:"wine_id"`
	Wine         Wine       `gorm:"foreignKey:WineID" json:"wine,omitempty"`
	IsOpened     bool       `json:"is_opened"`
	AddedAt      *time.Time `json:"added_at"`
	RemovedAt    *time.Time `json:"removed_at"`
	RowNumber    *int       `json:"row_number"`
	ColumnNumber *int       `json:"column_number"`
	Note         *string    `gorm:"size:512" json:"note"`
}

type BottleWithWineDTO struct {
	ID           int        `json:"id"`
	IsOpened     bool       `json:"is_opened"`
	AddedAt      *time.Time `json:"added_at"`
	RemovedAt    *time.Time `json:"removed_at"`
	RowNumber    *int       `json:"row_number"`
	ColumnNumber *int       `json:"column_number"`
	Note         *string    `json:"note"`

	Wine WineDTO `json:"wine"`
}

type WineType struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"size:255;not null" json:"name"`
}

type Country struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Name    string `gorm:"size:255;not null" json:"name"`
	ISOCode string `gorm:"size:2" json:"iso_code"` // ISO 3166-1 alpha-2 コード
}

type Region struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	Name      string  `gorm:"size:255;not null" json:"name"`
	CountryID uint    `gorm:"not null" json:"country_id"`
	Country   Country `gorm:"foreignKey:CountryID" json:"country,omitempty"`
	ParentID  *uint   `json:"parent_id"` // 親地域のID
}

type Appellation struct {
	ID                uint             `gorm:"primaryKey" json:"id"`
	Name              string           `gorm:"size:255;not null" json:"name"`
	DesignationTypeID *uint            `json:"designation_type_id"`
	DesignationType   *DesignationType `gorm:"foreignKey:DesignationTypeID" json:"designation_type,omitempty"`
	RegionID          *uint            `json:"region_id"`
	Region            *Region          `gorm:"foreignKey:RegionID" json:"region,omitempty"`
}

type DesignationType struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Name      string `gorm:"size:255;not null" json:"name"`
	Code      string `gorm:"size:50" json:"code"`
	Rank      int    `json:"rank"`
	CountryID uint   `json:"country_id"`
}

type CreateWineWithBottleRequest struct {
	Wine   Wine   `json:"wine"`
	Bottle Bottle `json:"bottle"`
}

// 統計情報のレスポンス構造体

// WineTypeStats: ワインタイプ別の在庫構成比
type WineTypeStats struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// CountryStats: 生産国別の在庫構成比
type CountryStats struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// InventoryTrendDataPoint: 在庫数推移のデータポイント
type InventoryTrendDataPoint struct {
	Date  string `json:"date"` // YYYY-MM-DD format
	Count int    `json:"count"`
}

// VintageStats: ワインのビンテージ別の在庫構成比
type VintageStats struct {
	Vintage int `json:"vintage"`
	Count   int `json:"count"`
}

// Stats: すべての統計情報
type Stats struct {
	WineTypes      []WineTypeStats           `json:"wineTypes"`
	Countries      []CountryStats            `json:"countries"`
	Vintages       []VintageStats            `json:"vintages"`
	InventoryTrend []InventoryTrendDataPoint `json:"inventoryTrend"`
}

// InventoryDailySnapshot: 日次在庫スナップショット
type InventoryDailySnapshot struct {
	SnapshotDate time.Time                      `gorm:"primaryKey;column:snapshot_date" json:"snapshot_date"`
	TotalCount   int                            `gorm:"column:total_count" json:"total_count"`
	CreatedAt    time.Time                      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time                      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	Details      []InventoryDailySnapshotDetail `gorm:"foreignKey:SnapshotDate;references:SnapshotDate" json:"details,omitempty"`
}

// InventoryDailySnapshotDetail: 日次在庫スナップショット詳細
type InventoryDailySnapshotDetail struct {
	ID           uint      `gorm:"primaryKey;column:id" json:"id"`
	SnapshotDate time.Time `gorm:"column:snapshot_date;index" json:"snapshot_date"`
	CategoryType string    `gorm:"column:category_type" json:"category_type"` // wine_type, country, vintage
	CategoryID   *int      `gorm:"column:category_id" json:"category_id"`
	CategoryName *string   `gorm:"column:category_name" json:"category_name"`
	Count        int       `gorm:"column:count" json:"count"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}
