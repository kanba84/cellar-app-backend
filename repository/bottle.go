package repository

import (
	"cellar-app/model"
	"time"

	"gorm.io/gorm"
)

type BottleRepository struct {
	db *gorm.DB
}

func NewBottleRepository(db *gorm.DB) *BottleRepository {
	return &BottleRepository{db: db}
}

func (r *BottleRepository) List() ([]model.Bottle, error) {
	var bottles []model.Bottle

	err := r.db.
		Preload("Wine").
		Preload("Wine.WineType").
		Preload("Wine.Country").
		Preload("Wine.Region").
		Find(&bottles).Error

	return bottles, err
}

func (r *BottleRepository) GetByID(id uint) (*model.Bottle, error) {
	var bottle model.Bottle

	err := r.db.
		Preload("Wine").
		Preload("Wine.WineType").
		Preload("Wine.Country").
		Preload("Wine.Region").
		First(&bottle, id).Error

	if err != nil {
		return nil, err
	}

	return &bottle, nil
}

func (r *BottleRepository) Create(bottle *model.Bottle) error {
	if bottle.AddedAt == nil {
		now := time.Now()
		bottle.AddedAt = &now
	}
	return r.db.Create(bottle).Error
}

func (r *BottleRepository) Update(bottle *model.Bottle) error {
	return r.db.Save(bottle).Error
}

func (r *BottleRepository) Patch(id uint, updates map[string]interface{}) error {
	return r.db.Model(&model.Bottle{}).Where("id = ?", id).Updates(updates).Error
}

func (r *BottleRepository) Delete(id uint) error {
	return r.db.Delete(&model.Bottle{}, id).Error
}
