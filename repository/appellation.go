package repository

import (
	"cellar-app/model"

	"gorm.io/gorm"
)

type AppellationRepository struct {
	db *gorm.DB
}

func NewAppellationRepository(db *gorm.DB) *AppellationRepository {
	return &AppellationRepository{db: db}
}

func (r *AppellationRepository) List() ([]model.Appellation, error) {
	var appellations []model.Appellation

	err := r.db.
		Preload("DesignationType").
		Preload("Region").
		Find(&appellations).Error

	return appellations, err
}

func (r *AppellationRepository) GetByID(id uint) (*model.Appellation, error) {
	var appellation model.Appellation

	err := r.db.
		Preload("DesignationType").
		Preload("Region").
		First(&appellation, id).Error

	if err != nil {
		return nil, err
	}

	return &appellation, nil
}

func (r *AppellationRepository) Create(appellation *model.Appellation) error {
	return r.db.Create(appellation).Error
}

func (r *AppellationRepository) Update(appellation *model.Appellation) error {
	return r.db.Save(appellation).Error
}

func (r *AppellationRepository) Delete(id uint) error {
	return r.db.Delete(&model.Appellation{}, id).Error
}
