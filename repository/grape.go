package repository

import (
	"cellar-app/model"

	"gorm.io/gorm"
)

// GrapeRepository: Grape のリポジトリ
type GrapeRepository struct {
	db *gorm.DB
}

// NewGrapeRepository: GrapeRepository を作成します
func NewGrapeRepository(db *gorm.DB) *GrapeRepository {
	return &GrapeRepository{db: db}
}

// GetOrCreateByName: ブドウ品種を名前で取得、存在しない場合は作成します
func (r *GrapeRepository) GetOrCreateByName(name string) (*model.Grape, error) {
	var grape model.Grape
	err := r.db.Where("name = ?", name).First(&grape).Error
	
	if err == gorm.ErrRecordNotFound {
		// 存在しないため、新規作成
		grape = model.Grape{Name: name}
		if err := r.db.Create(&grape).Error; err != nil {
			return nil, err
		}
		return &grape, nil
	} else if err != nil {
		return nil, err
	}
	
	return &grape, nil
}

func (r *GrapeRepository) GetOrCreateByNameTx(tx *gorm.DB, name string) (*model.Grape, error) {
	var grape model.Grape
	err := tx.Where("name = ?", name).First(&grape).Error
	if err == gorm.ErrRecordNotFound {
		grape = model.Grape{Name: name}
		if err := tx.Create(&grape).Error; err != nil {
			return nil, err
		}
		return &grape, nil
	} else if err != nil {
		return nil, err
	}
	return &grape, nil
}

// GetByID: ID でブドウ品種を取得します
func (r *GrapeRepository) GetByID(id uint) (*model.Grape, error) {
	var grape model.Grape
	err := r.db.First(&grape, id).Error
	if err != nil {
		return nil, err
	}
	return &grape, nil
}

// List: すべてのブドウ品種を取得します
func (r *GrapeRepository) List() ([]model.Grape, error) {
	var grapes []model.Grape
	err := r.db.Find(&grapes).Error
	return grapes, err
}
