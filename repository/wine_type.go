package repository

import (
	"cellar-app/model"

	"gorm.io/gorm"
)

type WineTypeRepository struct {
	db *gorm.DB
}

func NewWineTypeRepository(db *gorm.DB) *WineTypeRepository {
	return &WineTypeRepository{db: db}
}

func (r *WineTypeRepository) List() ([]model.WineType, error) {
	var wineTypes []model.WineType

	err := r.db.Order("id").Find(&wineTypes).Error

	return wineTypes, err
}

func (r *WineTypeRepository) GetByID(id uint) (*model.WineType, error) {
	var wineType model.WineType

	err := r.db.First(&wineType, id).Error

	if err != nil {
		return nil, err
	}

	return &wineType, nil
}

func (r *WineTypeRepository) Create(wineType *model.WineType) error {
	return r.db.Create(wineType).Error
}

func (r *WineTypeRepository) Update(wineType *model.WineType) error {
	return r.db.Save(wineType).Error
}

func (r *WineTypeRepository) Delete(id uint) error {
	return r.db.Delete(&model.WineType{}, id).Error
}
