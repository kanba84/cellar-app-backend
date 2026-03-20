package repository

import (
	"cellar-app/model"

	"gorm.io/gorm"
)

type CountryRepository struct {
	db *gorm.DB
}

func NewCountryRepository(db *gorm.DB) *CountryRepository {
	return &CountryRepository{db: db}
}

func (r *CountryRepository) List() ([]model.Country, error) {
	var countries []model.Country

	err := r.db.Order("id").Find(&countries).Error

	return countries, err
}

func (r *CountryRepository) GetByID(id uint) (*model.Country, error) {
	var country model.Country

	err := r.db.First(&country, id).Error

	if err != nil {
		return nil, err
	}

	return &country, nil
}

func (r *CountryRepository) Create(country *model.Country) error {
	return r.db.Create(country).Error
}

func (r *CountryRepository) Update(country *model.Country) error {
	return r.db.Save(country).Error
}

func (r *CountryRepository) Delete(id uint) error {
	return r.db.Delete(&model.Country{}, id).Error
}
