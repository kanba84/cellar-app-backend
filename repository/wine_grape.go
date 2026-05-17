package repository

import (
	"cellar-app/model"

	"gorm.io/gorm"
)

// WineGrapeRepository: WineGrape のリポジトリ
type WineGrapeRepository struct {
	db *gorm.DB
}

// NewWineGrapeRepository: WineGrapeRepository を作成します
func NewWineGrapeRepository(db *gorm.DB) *WineGrapeRepository {
	return &WineGrapeRepository{db: db}
}

// GetByWineID: Wine ID でブドウ品種を取得します
func (r *WineGrapeRepository) GetByWineID(wineID uint) ([]model.WineGrape, error) {
	var wineGrapes []model.WineGrape
	err := r.db.
		Where("wine_id = ?", wineID).
		Order("display_order ASC").
		Preload("Grape").
		Find(&wineGrapes).Error
	return wineGrapes, err
}

// Create: WineGrape を作成します
func (r *WineGrapeRepository) Create(wineGrape *model.WineGrape) error {
	return r.db.Create(wineGrape).Error
}

func (r *WineGrapeRepository) CreateBatchTx(tx *gorm.DB, wineGrapes []model.WineGrape) error {
	if len(wineGrapes) == 0 {
		return nil
	}
	return tx.CreateInBatches(wineGrapes, 100).Error
}

// DeleteByWineID: Wine ID に紐づくすべての WineGrape を削除します
func (r *WineGrapeRepository) DeleteByWineID(wineID uint) error {
	return r.db.Where("wine_id = ?", wineID).Delete(&model.WineGrape{}).Error
}

func (r *WineGrapeRepository) DeleteByWineIDTx(tx *gorm.DB, wineID uint) error {
	return tx.Where("wine_id = ?", wineID).Delete(&model.WineGrape{}).Error
}

// CreateBatch: 複数の WineGrape を一度に作成します
func (r *WineGrapeRepository) CreateBatch(wineGrapes []model.WineGrape) error {
	if len(wineGrapes) == 0 {
		return nil
	}
	return r.db.CreateInBatches(wineGrapes, 100).Error
}
