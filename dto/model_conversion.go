package dto

import "cellar-app/model"

func (d WineRequest) ToModel() model.Wine {
	wine := model.Wine{
		ID:                  d.ID,
		Name:                d.Name,
		CountryID:           d.CountryID,
		WineTypeID:          d.WineTypeID,
		Vintage:             d.Vintage,
		DrinkingWindowStart: d.DrinkingWindowStart,
		DrinkingWindowEnd:   d.DrinkingWindowEnd,
		RegionID:            d.RegionID,
		Producer:            d.Producer,
		AppellationID:       d.AppellationID,
		LabelImageURL:       d.LabelImageURL,
		ReferencePrice:      d.ReferencePrice,
	}
	if d.Country.ID != 0 || d.Country.Name != "" || d.Country.ISOCode != "" {
		wine.Country = d.Country.ToModel()
	}
	if d.WineType.ID != 0 || d.WineType.Name != "" {
		wine.WineType = model.WineType{ID: d.WineType.ID, Name: d.WineType.Name}
	}
	if d.Region != nil {
		region := d.Region.ToModel()
		wine.Region = &region
	}
	if d.Appellation != nil {
		appellation := d.Appellation.ToModel()
		wine.Appellation = &appellation
	}
	for _, grape := range d.WineGrapes {
		displayOrder := grape.DisplayOrder
		wineGrape := model.WineGrape{
			WineID:       grape.WineID,
			GrapeID:      grape.GrapeID,
			Percentage:   grape.Percentage,
			DisplayOrder: &displayOrder,
			Grape:        model.Grape{ID: grape.Grape.ID, Name: grape.Grape.Name},
		}
		wine.WineGrapes = append(wine.WineGrapes, wineGrape)
	}
	return wine
}

func WineFromModel(wine model.Wine) WineRequest {
	result := WineRequest{
		ID:                  wine.ID,
		Name:                wine.Name,
		CountryID:           wine.CountryID,
		Country:             CountryFromModel(wine.Country),
		WineTypeID:          wine.WineTypeID,
		WineType:            WineTypeFromModel(wine.WineType),
		Vintage:             wine.Vintage,
		DrinkingWindowStart: wine.DrinkingWindowStart,
		DrinkingWindowEnd:   wine.DrinkingWindowEnd,
		RegionID:            wine.RegionID,
		Producer:            wine.Producer,
		AppellationID:       wine.AppellationID,
		LabelImageURL:       wine.LabelImageURL,
		ReferencePrice:      wine.ReferencePrice,
	}
	if wine.Region != nil {
		region := RegionFromModel(*wine.Region)
		result.Region = &region
	}
	if wine.Appellation != nil {
		appellation := AppellationFromModel(*wine.Appellation)
		result.Appellation = &appellation
	}
	for _, grape := range wine.WineGrapes {
		displayOrder := 0
		if grape.DisplayOrder != nil {
			displayOrder = *grape.DisplayOrder
		}
		wineGrape := WineGrapeInput{
			WineID:       grape.WineID,
			GrapeID:      grape.GrapeID,
			Percentage:   grape.Percentage,
			DisplayOrder: displayOrder,
			Grape:        GrapeDTO{ID: grape.Grape.ID, Name: grape.Grape.Name},
		}
		result.WineGrapes = append(result.WineGrapes, wineGrape)
	}
	return result
}

func (d BottleRequest) ToModel() model.Bottle {
	isOpened := d.IsOpened
	bottle := model.Bottle{
		ID:           d.ID,
		WineID:       d.WineID,
		IsOpened:     &isOpened,
		AddedAt:      d.AddedAt,
		RemovedAt:    d.RemovedAt,
		RowNumber:    d.RowNumber,
		ColumnNumber: d.ColumnNumber,
		Note:         d.Note,
	}
	if d.Wine.ID != 0 {
		wine := d.Wine.ToModel()
		bottle.Wine = wine
	}
	return bottle
}

func BottleFromModel(bottle model.Bottle) BottleRequest {
	isOpened := false
	if bottle.IsOpened != nil {
		isOpened = *bottle.IsOpened
	}
	result := BottleRequest{
		ID:           bottle.ID,
		WineID:       bottle.WineID,
		IsOpened:     isOpened,
		AddedAt:      bottle.AddedAt,
		RemovedAt:    bottle.RemovedAt,
		RowNumber:    bottle.RowNumber,
		ColumnNumber: bottle.ColumnNumber,
		Note:         bottle.Note,
	}
	if bottle.Wine.ID != 0 {
		result.Wine = WineFromModel(bottle.Wine)
	}
	return result
}

func CountryFromModel(country model.Country) CountryDTO {
	isoCode := ""
	if country.ISOCode != nil {
		isoCode = *country.ISOCode
	}
	return CountryDTO{ID: country.ID, Name: country.Name, ISOCode: isoCode}
}

func RegionFromModel(region model.Region) RegionDTO {
	return RegionDTO{
		ID:        region.ID,
		Name:      region.Name,
		CountryID: region.CountryID,
		Country:   CountryFromModel(region.Country),
		ParentID:  region.ParentID,
	}
}

func (d RegionDTO) ToModel() model.Region {
	region := model.Region{
		ID:        d.ID,
		Name:      d.Name,
		CountryID: d.CountryID,
		ParentID:  d.ParentID,
	}
	if d.Country.ID != 0 || d.Country.Name != "" || d.Country.ISOCode != "" {
		region.Country = d.Country.ToModel()
	}
	return region
}

func AppellationFromModel(appellation model.Appellation) AppellationDTO {
	result := AppellationDTO{
		ID:                appellation.ID,
		Name:              appellation.Name,
		DesignationTypeID: appellation.DesignationTypeID,
		RegionID:          appellation.RegionID,
	}
	if appellation.DesignationType != nil {
		designationType := DesignationTypeFromModel(*appellation.DesignationType)
		result.DesignationType = &designationType
	}
	if appellation.Region != nil {
		region := RegionFromModel(*appellation.Region)
		result.Region = &region
	}
	return result
}

func (d AppellationDTO) ToModel() model.Appellation {
	result := model.Appellation{
		ID:                d.ID,
		Name:              d.Name,
		DesignationTypeID: d.DesignationTypeID,
		RegionID:          d.RegionID,
	}
	if d.DesignationType != nil {
		designationType := d.DesignationType.ToModel()
		result.DesignationType = &designationType
	}
	if d.Region != nil {
		region := d.Region.ToModel()
		result.Region = &region
	}
	return result
}

func DesignationTypeFromModel(designationType model.DesignationType) DesignationTypeDTO {
	return DesignationTypeDTO{
		ID:        designationType.ID,
		Name:      designationType.Name,
		Code:      designationType.Code,
		Rank:      designationType.Rank,
		CountryID: uintValue(designationType.CountryID),
	}
}

func (d DesignationTypeDTO) ToModel() model.DesignationType {
	return model.DesignationType{
		ID:        d.ID,
		Name:      d.Name,
		Code:      d.Code,
		Rank:      d.Rank,
		CountryID: uintPointer(d.CountryID),
	}
}

func (d CountryDTO) ToModel() model.Country {
	return model.Country{ID: d.ID, Name: d.Name, ISOCode: &d.ISOCode}
}

func uintValue(value *uint) uint {
	if value == nil {
		return 0
	}
	return *value
}

func uintPointer(value uint) *uint {
	return &value
}

func WineTypeFromModel(wineType model.WineType) WineTypeDTO {
	return WineTypeDTO{ID: wineType.ID, Name: wineType.Name}
}

func (d WineTypeDTO) ToModel() model.WineType {
	return model.WineType{ID: d.ID, Name: d.Name}
}

func SnapshotFromModel(snapshot model.InventoryDailySnapshot) InventoryDailySnapshotDTO {
	result := InventoryDailySnapshotDTO{
		SnapshotDate: snapshot.SnapshotDate,
		TotalCount:   snapshot.TotalCount,
		CreatedAt:    snapshot.CreatedAt,
		UpdatedAt:    snapshot.UpdatedAt,
	}
	for _, detail := range snapshot.Details {
		result.Details = append(result.Details, InventoryDailySnapshotDetailDTO{
			ID:           detail.ID,
			SnapshotDate: detail.SnapshotDate,
			CategoryType: detail.CategoryType,
			CategoryID:   detail.CategoryID,
			CategoryName: detail.CategoryName,
			Count:        detail.Count,
			CreatedAt:    detail.CreatedAt,
		})
	}
	return result
}
