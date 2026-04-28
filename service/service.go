package service

import (
	"cellar-app/repository"

	"gorm.io/gorm"
)

type Service struct {
	WineRepo            *repository.WineRepository
	AppellationRepo     *repository.AppellationRepository
	BottleRepo          *repository.BottleRepository
	CountryRepo         *repository.CountryRepository
	RegionRepo          *repository.RegionRepository
	WineTypeRepo        *repository.WineTypeRepository
	DesignationTypeRepo *repository.DesignationTypeRepository
	SnapshotRepo        *repository.InventorySnapshotRepo
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		WineRepo:            repository.NewWineRepository(db),
		AppellationRepo:     repository.NewAppellationRepository(db),
		BottleRepo:          repository.NewBottleRepository(db),
		CountryRepo:         repository.NewCountryRepository(db),
		RegionRepo:          repository.NewRegionRepository(db),
		WineTypeRepo:        repository.NewWineTypeRepository(db),
		DesignationTypeRepo: repository.NewDesignationTypeRepository(db),
		SnapshotRepo:        repository.NewInventorySnapshotRepo(db),
	}
}
