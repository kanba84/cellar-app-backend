package dto

import "time"

type WineDTO struct {
	ID                  int            `json:"id"`
	Name                string         `json:"name"`
	CountryID           int            `json:"country_id"`
	CountryName         string         `json:"country_name"`
	CountryISOCode      string         `json:"country_iso_code"`
	WineTypeID          int            `json:"wine_type_id"`
	WinTypeName         string         `json:"wine_type_name"`
	Vintage             *int           `json:"vintage"`
	DrinkingWindowStart *int           `json:"drinking_window_start"`
	DrinkingWindowEnd   *int           `json:"drinking_window_end"`
	RegionID            *int           `json:"region_id"`
	RegionName          *string        `json:"region_name"`
	Producer            *string        `json:"producer"`
	AppellationID       *int           `json:"appellation_id"`
	AppellationName     *string        `json:"appellation_name"`
	DesignationTypeID   *int           `json:"designation_type_id"`
	DesignationTypeName *string        `json:"designation_type_name"`
	LabelImageURL       *string        `json:"label_image_url"`
	ReferencePrice      *float64       `json:"reference_price"`
	WineGrapes          []WineGrapeDTO `json:"wine_grapes,omitempty"`
	HasStock            bool           `json:"has_stock"`
	StockCount          int64          `json:"stock_count"`
}

type WineGrapeDTO struct {
	Percentage   *float64 `json:"percentage"`
	DisplayOrder int      `json:"display_order"`
	Name         string   `json:"name"`
}

type WineGrapeInput struct {
	WineID       uint         `json:"wine_id"`
	GrapeID      uint         `json:"grape_id"`
	Percentage   *float64     `json:"percentage"`
	DisplayOrder int          `json:"display_order"`
	Grape        GrapeDTO     `json:"grape,omitempty"`
	Wine         *WineRequest `json:"wine,omitempty"`
}

type WineRequest struct {
	ID                  uint             `json:"id"`
	Name                string           `json:"name"`
	CountryID           uint             `json:"country_id"`
	Country             CountryDTO       `json:"country,omitempty"`
	WineTypeID          uint             `json:"wine_type_id"`
	WineType            WineTypeDTO      `json:"wine_type,omitempty"`
	Vintage             *int             `json:"vintage"`
	DrinkingWindowStart *int             `json:"drinking_window_start"`
	DrinkingWindowEnd   *int             `json:"drinking_window_end"`
	RegionID            *uint            `json:"region_id"`
	Region              *RegionDTO       `json:"region,omitempty"`
	Producer            *string          `json:"producer"`
	AppellationID       *uint            `json:"appellation_id"`
	Appellation         *AppellationDTO  `json:"appellation,omitempty"`
	LabelImageURL       *string          `json:"label_image_url"`
	ReferencePrice      *float64         `json:"reference_price"`
	WineGrapes          []WineGrapeInput `json:"wine_grapes,omitempty"`
}

type BottleRequest struct {
	ID           uint        `json:"id"`
	WineID       uint        `json:"wine_id"`
	Wine         WineRequest `json:"wine,omitempty"`
	IsOpened     bool        `json:"is_opened"`
	AddedAt      *time.Time  `json:"added_at"`
	RemovedAt    *time.Time  `json:"removed_at"`
	RowNumber    *int        `json:"row_number"`
	ColumnNumber *int        `json:"column_number"`
	Note         *string     `json:"note"`
}

type BottleWithWineDTO struct {
	ID           int        `json:"id"`
	IsOpened     bool       `json:"is_opened"`
	AddedAt      *time.Time `json:"added_at"`
	RemovedAt    *time.Time `json:"removed_at"`
	RowNumber    *int       `json:"row_number"`
	ColumnNumber *int       `json:"column_number"`
	Note         *string    `json:"note"`
	Wine         WineDTO    `json:"wine"`
}

type CreateWineWithBottleRequest struct {
	Wine   WineRequest   `json:"wine"`
	Bottle BottleRequest `json:"bottle"`
}

type WineTypeDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type CountryDTO struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	ISOCode string `json:"iso_code"`
}

type RegionDTO struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	CountryID uint       `json:"country_id"`
	Country   CountryDTO `json:"country,omitempty"`
	ParentID  *uint      `json:"parent_id"`
}

type AppellationDTO struct {
	ID                uint                `json:"id"`
	Name              string              `json:"name"`
	DesignationTypeID *uint               `json:"designation_type_id"`
	DesignationType   *DesignationTypeDTO `json:"designation_type,omitempty"`
	RegionID          *uint               `json:"region_id"`
	Region            *RegionDTO          `json:"region,omitempty"`
}

type DesignationTypeDTO struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	Rank      int    `json:"rank"`
	CountryID uint   `json:"country_id"`
}

type WineTypeStats struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type CountryStats struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type InventoryTrendDataPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type VintageStats struct {
	Vintage int `json:"vintage"`
	Count   int `json:"count"`
}

type Stats struct {
	WineTypes      []WineTypeStats           `json:"wineTypes"`
	Countries      []CountryStats            `json:"countries"`
	Vintages       []VintageStats            `json:"vintages"`
	InventoryTrend []InventoryTrendDataPoint `json:"inventoryTrend"`
}

type InventoryDailySnapshotDTO struct {
	SnapshotDate time.Time                         `json:"snapshot_date"`
	TotalCount   int                               `json:"total_count"`
	CreatedAt    time.Time                         `json:"created_at"`
	UpdatedAt    time.Time                         `json:"updated_at"`
	Details      []InventoryDailySnapshotDetailDTO `json:"details,omitempty"`
}

type InventoryDailySnapshotDetailDTO struct {
	ID           uint      `json:"id"`
	SnapshotDate time.Time `json:"snapshot_date"`
	CategoryType string    `json:"category_type"`
	CategoryID   *int      `json:"category_id"`
	CategoryName *string   `json:"category_name"`
	Count        int       `json:"count"`
	CreatedAt    time.Time `json:"created_at"`
}

type GrapeDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}
