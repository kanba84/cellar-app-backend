package database

import (
	"cellar-app/model"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Country{},
		&model.WineType{},
		&model.DesignationType{},
		&model.Region{},
		&model.Appellation{},
		&model.Grape{},
		&model.Wine{},
		&model.Bottle{},
		&model.WineGrape{},
		&model.InventoryDailySnapshot{},
		&model.InventoryDailySnapshotDetail{},
	)
}
