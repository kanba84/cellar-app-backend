package model

import "time"

type Wine struct {
	ID                  uint     `gorm:"type:integer;primaryKey"`
	Name                string   `gorm:"size:255;not null"`
	CountryID           uint     `gorm:"type:integer;not null"`
	Country             Country  `gorm:"foreignKey:CountryID;constraint:wines_country_id_fkey,OnDelete:NO ACTION"`
	WineTypeID          uint     `gorm:"type:integer;not null"`
	WineType            WineType `gorm:"foreignKey:WineTypeID;constraint:wines_wine_type_id_fkey,OnDelete:NO ACTION"`
	Vintage             *int     `gorm:"check:wines_vintage_check,vintage >= 1800 AND vintage <= date_part('year', CURRENT_DATE)"`
	DrinkingWindowStart *int
	DrinkingWindowEnd   *int
	RegionID            *uint        `gorm:"type:integer"`
	Region              *Region      `gorm:"foreignKey:RegionID;constraint:wines_region_id_fkey,OnDelete:NO ACTION"`
	Producer            *string      `gorm:"size:255"`
	AppellationID       *uint        `gorm:"type:integer"`
	Appellation         *Appellation `gorm:"foreignKey:AppellationID;constraint:wines_appellation_id_fkey,OnDelete:SET NULL"`
	LabelImageURL       *string      `gorm:"size:512"`
	ReferencePrice      *float64
	WineGrapes          []WineGrape `gorm:"foreignKey:WineID"`
}

type Bottle struct {
	ID           uint       `gorm:"type:integer;primaryKey"`
	WineID       uint       `gorm:"type:integer;not null"`
	Wine         Wine       `gorm:"foreignKey:WineID;constraint:cellar_bottles_wine_id_fkey,OnDelete:CASCADE"`
	IsOpened     *bool      `gorm:"default:false"`
	AddedAt      *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	RemovedAt    *time.Time `gorm:"type:timestamp"`
	RowNumber    *int       `gorm:"type:integer;uniqueIndex:unique_position"`
	ColumnNumber *int       `gorm:"type:integer;uniqueIndex:unique_position"`
	Note         *string    `gorm:"type:text"`
}

type WineType struct {
	ID   uint   `gorm:"type:integer;primaryKey"`
	Name string `gorm:"size:255;not null;uniqueIndex"`
}

type Country struct {
	ID      uint    `gorm:"type:integer;primaryKey"`
	Name    string  `gorm:"size:255;not null;uniqueIndex"`
	ISOCode *string `gorm:"size:2"`
}

type Region struct {
	ID        uint    `gorm:"type:integer;primaryKey"`
	Name      string  `gorm:"size:255;not null;uniqueIndex:regions_name_parent_id_key"`
	CountryID uint    `gorm:"type:integer;not null"`
	Country   Country `gorm:"foreignKey:CountryID;constraint:regions_country_id_fkey,OnDelete:NO ACTION"`
	ParentID  *uint   `gorm:"uniqueIndex:regions_name_parent_id_key"`
}

type Appellation struct {
	ID                uint             `gorm:"type:integer;primaryKey"`
	Name              string           `gorm:"size:255;not null"`
	DesignationTypeID *uint            `gorm:"type:integer"`
	DesignationType   *DesignationType `gorm:"foreignKey:DesignationTypeID;constraint:appellations_designation_type_id_fkey,OnDelete:SET NULL"`
	RegionID          *uint            `gorm:"type:integer"`
	Region            *Region          `gorm:"foreignKey:RegionID;constraint:appellations_region_id_fkey,OnDelete:CASCADE"`
}

type DesignationType struct {
	ID        uint   `gorm:"type:integer;primaryKey"`
	Name      string `gorm:"size:255;not null"`
	Code      string `gorm:"size:50"`
	Rank      int
	CountryID *uint
}

type InventoryDailySnapshot struct {
	SnapshotDate time.Time                      `gorm:"type:date;primaryKey;column:snapshot_date"`
	TotalCount   int                            `gorm:"type:integer;not null;column:total_count"`
	CreatedAt    time.Time                      `gorm:"not null;default:CURRENT_TIMESTAMP;column:created_at;autoCreateTime"`
	UpdatedAt    time.Time                      `gorm:"not null;default:CURRENT_TIMESTAMP;column:updated_at;autoUpdateTime"`
	Details      []InventoryDailySnapshotDetail `gorm:"foreignKey:SnapshotDate;references:SnapshotDate;constraint:fk_snapshot_date,OnDelete:CASCADE"`
}

type InventoryDailySnapshotDetail struct {
	ID           uint      `gorm:"primaryKey;column:id"`
	SnapshotDate time.Time `gorm:"type:date;not null;column:snapshot_date;index:idx_snapshot_details_date"`
	CategoryType string    `gorm:"type:text;not null;column:category_type;index:idx_snapshot_details_category;check:inventory_daily_snapshot_details_category_type_check,category_type IN ('wine_type','country','vintage')"`
	CategoryID   *int      `gorm:"type:integer;column:category_id;index:idx_snapshot_details_category"`
	CategoryName *string   `gorm:"type:text;column:category_name"`
	Count        int       `gorm:"type:integer;not null;column:count"`
	CreatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;column:created_at;autoCreateTime"`
}

type Grape struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:255;not null;uniqueIndex"`
}

type WineGrape struct {
	WineID       uint     `gorm:"primaryKey;column:wine_id"`
	GrapeID      uint     `gorm:"primaryKey;column:grape_id"`
	Percentage   *float64 `gorm:"column:percentage"`
	DisplayOrder *int     `gorm:"column:display_order"`
	Grape        Grape    `gorm:"foreignKey:GrapeID"`
}

func (WineGrape) TableName() string {
	return "wine_grapes"
}
