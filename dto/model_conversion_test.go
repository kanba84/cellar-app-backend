package dto

import (
	"cellar-app/model"
	"encoding/json"
	"strings"
	"testing"
)

func TestWineRequestModelConversionPreservesAPIFields(t *testing.T) {
	vintage := 2020
	regionID := uint(4)
	labelURL := "/labels/test.jpg"
	request := WineRequest{
		ID:            9,
		Name:          "Example",
		CountryID:     1,
		WineTypeID:    2,
		Vintage:       &vintage,
		RegionID:      &regionID,
		LabelImageURL: &labelURL,
	}

	persisted := request.ToModel()
	if persisted.ID != request.ID || persisted.Name != request.Name || persisted.CountryID != request.CountryID ||
		persisted.WineTypeID != request.WineTypeID || persisted.Vintage == nil || *persisted.Vintage != vintage ||
		persisted.RegionID == nil || *persisted.RegionID != regionID || persisted.LabelImageURL == nil || *persisted.LabelImageURL != labelURL {
		t.Fatalf("request conversion lost a field: %#v", persisted)
	}
	if persisted.Country.ID != 0 || persisted.Country.Name != "" || persisted.Country.ISOCode != nil {
		t.Fatalf("omitted country association should not be persisted: %#v", persisted.Country)
	}
	if persisted.WineType.ID != 0 || persisted.WineType.Name != "" {
		t.Fatalf("omitted wine type association should not be persisted: %#v", persisted.WineType)
	}

	response := WineFromModel(persisted)
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response DTO: %v", err)
	}
	for _, field := range []string{`"id":9`, `"country_id":1`, `"wine_type_id":2`, `"vintage":2020`, `"region_id":4`, `"label_image_url":"/labels/test.jpg"`} {
		if !strings.Contains(string(encoded), field) {
			t.Errorf("response DTO JSON %s does not contain %s", encoded, field)
		}
	}
}

func TestRegionRequestOmitsUnprovidedCountryAssociation(t *testing.T) {
	request := RegionDTO{ID: 3, Name: "Example region", CountryID: 1}
	persisted := request.ToModel()
	if persisted.Country.ID != 0 || persisted.Country.Name != "" || persisted.Country.ISOCode != nil {
		t.Fatalf("omitted country association should not be persisted: %#v", persisted.Country)
	}
}

func TestAppellationDTOConvertsCountryAndMultipleRegions(t *testing.T) {
	code := "FR:AOC:EXAMPLE"
	countryID := uint(1)
	regionID := uint(3)
	parentID := uint(1)
	request := AppellationDTO{
		ID:        9,
		Name:      "Example AOC",
		Code:      &code,
		CountryID: &countryID,
		RegionID:  &regionID,
		Regions: []RegionDTO{
			{ID: 3, Name: "Médoc", CountryID: 1, ParentID: &parentID},
		},
		SortOrder: 4,
	}

	persisted := request.ToModel()
	if persisted.Code == nil || *persisted.Code != code || persisted.CountryID == nil ||
		*persisted.CountryID != countryID || len(persisted.Regions) != 1 ||
		persisted.Regions[0].ID != 3 || persisted.SortOrder != 4 {
		t.Fatalf("request conversion lost appellation fields: %#v", persisted)
	}

	response := AppellationFromModel(persisted)
	if len(response.Regions) != 1 || response.Regions[0].Name != "Médoc" ||
		response.Code == nil || *response.Code != code || response.CountryID == nil ||
		*response.CountryID != countryID {
		t.Fatalf("response conversion lost appellation fields: %#v", response)
	}
}

func TestSnapshotDTOConversion(t *testing.T) {
	var snapshot model.InventoryDailySnapshot
	snapshot.Details = []model.InventoryDailySnapshotDetail{{ID: 1, CategoryType: "country", Count: 2}}

	got := SnapshotFromModel(snapshot)
	if len(got.Details) != 1 || got.Details[0].ID != 1 || got.Details[0].CategoryType != "country" || got.Details[0].Count != 2 {
		t.Fatalf("snapshot conversion lost details: %#v", got)
	}
}
