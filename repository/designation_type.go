package repository

import (
	"cellar-app/model"

	"gorm.io/gorm"
)

type DesignationTypeRepository struct {
	db *gorm.DB
}

func NewDesignationTypeRepository(db *gorm.DB) *DesignationTypeRepository {
	return &DesignationTypeRepository{db: db}
}

func (r *DesignationTypeRepository) List() ([]model.DesignationType, error) {
	var designationTypes []model.DesignationType

	err := r.db.Find(&designationTypes).Error

	return designationTypes, err
}

func (r *DesignationTypeRepository) GetByID(id uint) (*model.DesignationType, error) {
	var designationType model.DesignationType

	err := r.db.First(&designationType, id).Error

	if err != nil {
		return nil, err
	}

	return &designationType, nil
}

func (r *DesignationTypeRepository) Create(designationType *model.DesignationType) error {
	return r.db.Create(designationType).Error
}

func (r *DesignationTypeRepository) Update(designationType *model.DesignationType) error {
	return r.db.Save(designationType).Error
}

func (r *DesignationTypeRepository) Delete(id uint) error {
	return r.db.Delete(&model.DesignationType{}, id).Error
}
