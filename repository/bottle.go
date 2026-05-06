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

func (r *BottleRepository) CountByWineID(wineID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Bottle{}).Where("wine_id = ?", wineID).Count(&count).Error
	return count, err
}

type wineBottleCountRow struct {
	WineID uint  `gorm:"column:wine_id"`
	Count  int64 `gorm:"column:count"`
}

// CountByWineIDs は wine_id ごとのボトル数をまとめて取得します（N+1回避用）。
func (r *BottleRepository) CountByWineIDs(wineIDs []uint) (map[uint]int64, error) {
	if len(wineIDs) == 0 {
		return map[uint]int64{}, nil
	}

	var rows []wineBottleCountRow
	err := r.db.
		Model(&model.Bottle{}).
		Select("wine_id, COUNT(*) as count").
		Where("wine_id IN ?", wineIDs).
		Group("wine_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint]int64, len(rows))
	for _, row := range rows {
		result[row.WineID] = row.Count
	}

	return result, nil
}

func (r *BottleRepository) List() ([]model.Bottle, error) {
	var bottles []model.Bottle

	err := r.db.
		Preload("Wine").
		Preload("Wine.WineType").
		Preload("Wine.Country").
		Preload("Wine.Region").
		Preload("Wine.Appellation").
		Preload("Wine.Appellation.DesignationType").
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
		Preload("Wine.Appellation").
		Preload("Wine.Appellation.DesignationType").
		First(&bottle, id).Error

	if err != nil {
		return nil, err
	}

	return &bottle, nil
}

func (r *BottleRepository) Create(bottle *model.Bottle) error {
	// AddedAt が指定されていない場合、現在時刻を自動設定
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
