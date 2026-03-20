package repository

import (
	"cellar-app/model"
	"context"

	"gorm.io/gorm"
)

type WineRepository struct {
	db *gorm.DB
}

func NewWineRepository(db *gorm.DB) *WineRepository {
	return &WineRepository{db: db}
}

func (r *WineRepository) List() ([]model.Wine, error) {
	var wines []model.Wine

	err := r.db.
		Preload("WineType").
		Preload("Country").
		Preload("Region").
		Preload("Appellation").
		Preload("Appellation.DesignationType").
		Find(&wines).Error

	return wines, err
}

func (r *WineRepository) GetByID(id uint) (*model.Wine, error) {
	var wine model.Wine

	err := r.db.
		Preload("WineType").
		Preload("Country").
		Preload("Region").
		Preload("Appellation").
		Preload("Appellation.DesignationType").
		First(&wine, id).Error

	if err != nil {
		return nil, err
	}

	return &wine, nil
}

func (r *WineRepository) Create(wine *model.Wine) error {
	return r.db.Create(wine).Error
}

func (r *WineRepository) Update(wine *model.Wine) error {
	return r.db.Save(wine).Error
}

func (r *WineRepository) Patch(id uint, updates map[string]interface{}) error {
	return r.db.Model(&model.Wine{}).Where("id = ?", id).Updates(updates).Error
}

func (r *WineRepository) Delete(id uint) error {
	return r.db.Delete(&model.Wine{}, id).Error
}

func (r *WineRepository) CreateWithBottle(ctx context.Context, wine *model.Wine, bottle *model.Bottle) (*model.Wine, *model.Bottle, error) {
	tx := r.db.WithContext(ctx).Begin()

	// ワイン作成
	if err := tx.Create(wine).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	// ボトル作成（wine.IDはワイン作成後に自動割り当てされる）
	bottle.WineID = wine.ID
	if err := tx.Create(bottle).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, nil, err
	}

	return wine, bottle, nil
}
