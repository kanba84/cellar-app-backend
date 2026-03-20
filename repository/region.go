package repository

import (
	"cellar-app/model"

	"gorm.io/gorm"
)

type RegionRepository struct {
	db *gorm.DB
}

func NewRegionRepository(db *gorm.DB) *RegionRepository {
	return &RegionRepository{db: db}
}

func (r *RegionRepository) List() ([]model.Region, error) {
	var regions []model.Region

	err := r.db.
		Preload("Country").
		Order("id").
		Find(&regions).Error

	return regions, err
}

func (r *RegionRepository) GetByID(id uint) (*model.Region, error) {
	var region model.Region

	err := r.db.
		Preload("Country").
		First(&region, id).Error

	if err != nil {
		return nil, err
	}

	return &region, nil
}

func (r *RegionRepository) Create(region *model.Region) error {
	return r.db.Create(region).Error
}

func (r *RegionRepository) Update(region *model.Region) error {
	return r.db.Save(region).Error
}

func (r *RegionRepository) Delete(id uint) error {
	return r.db.Delete(&model.Region{}, id).Error
}
