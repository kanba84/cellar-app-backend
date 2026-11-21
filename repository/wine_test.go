package repository

import (
	"strings"
	"testing"

	"cellar-app/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestWineSearchBuildsBoundedParameterizedPartialMatchQuery(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  "host=127.0.0.1 user=test dbname=test sslmode=disable",
		PreferSimpleProtocol: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}

	vintage := 2020
	countryID := uint(33)
	query := WineSearch{
		Producer:  `A_%`,
		Name:      "Chateau",
		Vintage:   &vintage,
		CountryID: &countryID,
		Limit:     12,
	}
	var wines []model.Wine
	result := buildWineSearchQuery(db, query).Find(&wines)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	statement := result.Statement
	sql := strings.ToLower(statement.SQL.String())
	for _, expected := range []string{"lower(coalesce(producer", "lower(name)", "vintage =", "country_id =", "order by", "limit"} {
		if !strings.Contains(sql, expected) {
			t.Errorf("search SQL missing %q: %s", expected, statement.SQL.String())
		}
	}
	if !strings.Contains(statement.SQL.String(), "LIMIT $5") || statement.Vars[4] != 12 {
		t.Errorf("search SQL has no bounded limit: %s", statement.SQL.String())
	}
	if got, want := statement.Vars[0], `%a\_\%%`; got != want {
		t.Errorf("producer pattern = %#v, want %#v", got, want)
	}
	if got, want := statement.Vars[1], "%chateau%"; got != want {
		t.Errorf("name pattern = %#v, want %#v", got, want)
	}
	if statement.Vars[2] != vintage || statement.Vars[3] != countryID {
		t.Errorf("numeric filters were not parameterized as expected: %#v", statement.Vars)
	}
}
