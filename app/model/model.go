package model

import "time"

type Wine struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	CountryID     int     `json:"country_id"`
	WineTypeID    int     `json:"wine_type_id"`
	Vintage       *int    `json:"vintage"`
	RegionID      *int    `json:"region_id"`
	Producer      *string `json:"producer"`
	AppellationID *int    `json:"appellation_id"`
}

// WineDTOは、APIレスポンス用のWineデータを表す構造体です。
// Wine構造体をベースに、必要なフィールドのみを含めています。
type WineDTO struct {
	ID                  int     `json:"id"`
	Name                string  `json:"name"`
	CountryID           int     `json:"country_id"`
	CountryName         string  `json:"country_name"`
	WineTypeID          int     `json:"wine_type_id"`
	WinTypeName         string  `json:"wine_type_name"`
	Vintage             *int    `json:"vintage"`
	RegionID            *int    `json:"region_id"`
	RegionName          *string `json:"region_name"`
	Producer            *string `json:"producer"`
	AppellationID       *string `json:"appellation_id"`
	AppellationName     *string `json:"appellation_name"`
	DesignationTypeID   *int    `json:"designation_type_id"`
	DesignationTypeName *string `json:"designation_type_name"`
}

type Bottle struct {
	ID           int        `json:"id"`
	Wine         WineDTO    `json:"wine,omitempty"`
	WineID       int        `json:"wine_id"`
	IsOpened     bool       `json:"is_opened"`
	AddedAt      *time.Time `json:"added_at"`
	RemovedAt    *time.Time `json:"removed_at"`
	RowNumber    *int       `json:"row_number"`
	ColumnNumber *int       `json:"column_number"`
	Note         *string    `json:"note"`
}

type WineType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Country struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Region struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CountryID int    `json:"country_id"`
	ParentID  *int   `json:"parent_id"` // 親地域のID
}

type Appellation struct {
	ID                  int    `json:"id"`
	Name                string `json:"name"`
	DesignationTypeName string `json:"designation_type_name"`
	DesignationTypeID   int    `json:"designation_type_id"`
	//CountryID        int             `json:"country_id"`
	RegionID *int `json:"region_id,omitempty"`
}

type DesignationType struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`       // 例: "AOC", "DOC", "DO", "DOP", "DOQ", "DOCa"
	Rank      int    `json:"rank"`       // ランク（1が最上位）
	CountryID int    `json:"country_id"` // この指定が属する国のID
}

type CreateWineWithBottleRequest struct {
	Wine   Wine   `json:"wine"`
	Bottle Bottle `json:"bottle"`
}
