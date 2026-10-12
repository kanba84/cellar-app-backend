package model

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestPersistentModelSchemas(t *testing.T) {
	tests := []struct {
		name       string
		model      interface{}
		table      string
		primaryKey []string
	}{
		{name: "wine", model: &Wine{}, table: "wines", primaryKey: []string{"id"}},
		{name: "bottle", model: &Bottle{}, table: "bottles", primaryKey: []string{"id"}},
		{name: "wine type", model: &WineType{}, table: "wine_types", primaryKey: []string{"id"}},
		{name: "country", model: &Country{}, table: "countries", primaryKey: []string{"id"}},
		{name: "region", model: &Region{}, table: "regions", primaryKey: []string{"id"}},
		{name: "appellation", model: &Appellation{}, table: "appellations", primaryKey: []string{"id"}},
		{name: "designation type", model: &DesignationType{}, table: "designation_types", primaryKey: []string{"id"}},
		{name: "snapshot", model: &InventoryDailySnapshot{}, table: "inventory_daily_snapshots", primaryKey: []string{"snapshot_date"}},
		{name: "snapshot detail", model: &InventoryDailySnapshotDetail{}, table: "inventory_daily_snapshot_details", primaryKey: []string{"id"}},
		{name: "grape", model: &Grape{}, table: "grapes", primaryKey: []string{"id"}},
		{name: "wine grape", model: &WineGrape{}, table: "wine_grapes", primaryKey: []string{"wine_id", "grape_id"}},
	}

	cache := &sync.Map{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := schema.Parse(test.model, cache, schema.NamingStrategy{})
			if err != nil {
				t.Fatalf("parse model schema: %v", err)
			}
			if parsed.Table != test.table {
				t.Fatalf("table = %q, want %q", parsed.Table, test.table)
			}
			if len(parsed.PrimaryFields) != len(test.primaryKey) {
				t.Fatalf("primary key fields = %d, want %d", len(parsed.PrimaryFields), len(test.primaryKey))
			}
			for i, field := range parsed.PrimaryFields {
				if field.DBName != test.primaryKey[i] {
					t.Errorf("primary key field %d = %q, want %q", i, field.DBName, test.primaryKey[i])
				}
			}
		})
	}
}

func TestPersistentModelColumnTypesAndNullability(t *testing.T) {
	tests := []struct {
		name     string
		model    interface{}
		field    string
		dataType string
		notNull  bool
		size     int
	}{
		{name: "wine integer id", model: &Wine{}, field: "ID", dataType: "integer", notNull: true},
		{name: "wine integer country id", model: &Wine{}, field: "CountryID", dataType: "integer", notNull: true},
		{name: "wine integer type id", model: &Wine{}, field: "WineTypeID", dataType: "integer", notNull: true},
		{name: "nullable wine integer region id", model: &Wine{}, field: "RegionID", dataType: "integer", notNull: false},
		{name: "nullable wine integer appellation id", model: &Wine{}, field: "AppellationID", dataType: "integer", notNull: false},
		{name: "bottle integer id", model: &Bottle{}, field: "ID", dataType: "integer", notNull: true},
		{name: "bottle integer wine id", model: &Bottle{}, field: "WineID", dataType: "integer", notNull: true},
		{name: "bottle integer row", model: &Bottle{}, field: "RowNumber", dataType: "integer", notNull: false},
		{name: "bottle integer column", model: &Bottle{}, field: "ColumnNumber", dataType: "integer", notNull: false},
		{name: "nullable bottle opened", model: &Bottle{}, field: "IsOpened", dataType: "bool", notNull: false},
		{name: "timestamp without timezone bottle added", model: &Bottle{}, field: "AddedAt", dataType: "timestamp", notNull: false},
		{name: "timestamp without timezone bottle removed", model: &Bottle{}, field: "RemovedAt", dataType: "timestamp", notNull: false},
		{name: "text bottle note", model: &Bottle{}, field: "Note", dataType: "text", notNull: false},
		{name: "nullable country ISO", model: &Country{}, field: "ISOCode", dataType: "string", notNull: false, size: 2},
		{name: "nullable designation country", model: &DesignationType{}, field: "CountryID", dataType: "uint", notNull: false},
		{name: "nullable appellation integer designation id", model: &Appellation{}, field: "DesignationTypeID", dataType: "integer", notNull: false},
		{name: "nullable appellation integer region id", model: &Appellation{}, field: "RegionID", dataType: "integer", notNull: false},
		{name: "region integer country id", model: &Region{}, field: "CountryID", dataType: "integer", notNull: true},
		{name: "snapshot date", model: &InventoryDailySnapshot{}, field: "SnapshotDate", dataType: "date", notNull: true},
		{name: "snapshot detail date", model: &InventoryDailySnapshotDetail{}, field: "SnapshotDate", dataType: "date", notNull: true},
		{name: "snapshot category text", model: &InventoryDailySnapshotDetail{}, field: "CategoryType", dataType: "text", notNull: true},
		{name: "snapshot category name text", model: &InventoryDailySnapshotDetail{}, field: "CategoryName", dataType: "text", notNull: false},
		{name: "nullable grape order", model: &WineGrape{}, field: "DisplayOrder", dataType: "int", notNull: false},
	}

	cache := &sync.Map{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := schema.Parse(test.model, cache, schema.NamingStrategy{})
			if err != nil {
				t.Fatalf("parse model schema: %v", err)
			}
			field := parsed.LookUpField(test.field)
			if field == nil {
				t.Fatalf("field %q was not parsed", test.field)
			}
			if string(field.DataType) != test.dataType {
				t.Errorf("data type = %q, want %q", field.DataType, test.dataType)
			}
			if (field.NotNull || field.PrimaryKey) != test.notNull {
				t.Errorf("not null = %v, want %v", field.NotNull || field.PrimaryKey, test.notNull)
			}
			if test.size != 0 && field.Size != test.size {
				t.Errorf("size = %d, want %d", field.Size, test.size)
			}
		})
	}
}

func TestPersistentModelUsesExistingForeignKeyNames(t *testing.T) {
	tests := []struct {
		name         string
		model        interface{}
		relationship string
		constraint   string
		onDelete     string
	}{
		{name: "bottle wine", model: &Bottle{}, relationship: "Wine", constraint: "cellar_bottles_wine_id_fkey", onDelete: "CASCADE"},
		{name: "snapshot details", model: &InventoryDailySnapshot{}, relationship: "Details", constraint: "fk_snapshot_date", onDelete: "CASCADE"},
		{name: "wine country", model: &Wine{}, relationship: "Country", constraint: "wines_country_id_fkey", onDelete: "NO ACTION"},
		{name: "wine type", model: &Wine{}, relationship: "WineType", constraint: "wines_wine_type_id_fkey", onDelete: "NO ACTION"},
		{name: "wine region", model: &Wine{}, relationship: "Region", constraint: "wines_region_id_fkey", onDelete: "NO ACTION"},
		{name: "wine appellation", model: &Wine{}, relationship: "Appellation", constraint: "wines_appellation_id_fkey", onDelete: "SET NULL"},
		{name: "region country", model: &Region{}, relationship: "Country", constraint: "regions_country_id_fkey", onDelete: "NO ACTION"},
		{name: "appellation designation type", model: &Appellation{}, relationship: "DesignationType", constraint: "appellations_designation_type_id_fkey", onDelete: "SET NULL"},
		{name: "appellation region", model: &Appellation{}, relationship: "Region", constraint: "appellations_region_id_fkey", onDelete: "CASCADE"},
	}

	cache := &sync.Map{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := schema.Parse(test.model, cache, schema.NamingStrategy{})
			if err != nil {
				t.Fatalf("parse model schema: %v", err)
			}
			relationship := parsed.Relationships.Relations[test.relationship]
			if relationship == nil {
				t.Fatalf("relationship %q was not parsed", test.relationship)
			}
			constraint := relationship.ParseConstraint()
			if constraint == nil {
				t.Fatalf("constraint for relationship %q was not parsed", test.relationship)
			}
			if constraint.Name != test.constraint {
				t.Errorf("constraint name = %q, want %q", constraint.Name, test.constraint)
			}
			if constraint.OnDelete != test.onDelete {
				t.Errorf("on delete = %q, want %q", constraint.OnDelete, test.onDelete)
			}
		})
	}
}
