package service

import (
	"cellar-app/llm"
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
	GrapeRepo           *repository.GrapeRepository
	WineGrapeRepo       *repository.WineGrapeRepository
	LLMProvider         llm.WineInfoProvider
}

func NewService(db *gorm.DB, provider llm.WineInfoProvider) *Service {
	return &Service{
		WineRepo:            repository.NewWineRepository(db),
		AppellationRepo:     repository.NewAppellationRepository(db),
		BottleRepo:          repository.NewBottleRepository(db),
		CountryRepo:         repository.NewCountryRepository(db),
		RegionRepo:          repository.NewRegionRepository(db),
		WineTypeRepo:        repository.NewWineTypeRepository(db),
		DesignationTypeRepo: repository.NewDesignationTypeRepository(db),
		SnapshotRepo:        repository.NewInventorySnapshotRepo(db),
		GrapeRepo:           repository.NewGrapeRepository(db),
		WineGrapeRepo:       repository.NewWineGrapeRepository(db),
		LLMProvider:         provider,
	}
}
