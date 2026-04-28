package repository

import (
	"cellar-app/model"
	"time"

	"gorm.io/gorm"
)

type InventorySnapshotRepo struct {
	db *gorm.DB
}

func NewInventorySnapshotRepo(db *gorm.DB) *InventorySnapshotRepo {
	return &InventorySnapshotRepo{db: db}
}

// SaveSnapshot: スナップショットを保存または更新
func (r *InventorySnapshotRepo) SaveSnapshot(snapshot *model.InventoryDailySnapshot) error {
	return r.db.Save(snapshot).Error
}

// GetSnapshot: 指定日付のスナップショットを取得
func (r *InventorySnapshotRepo) GetSnapshot(snapshotDate time.Time) (*model.InventoryDailySnapshot, error) {
	var snapshot model.InventoryDailySnapshot
	err := r.db.Preload("Details").
		Where("snapshot_date = ?", snapshotDate.Format("2006-01-02")).
		First(&snapshot).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // スナップショットが存在しない場合は nil を返す
		}
		return nil, err
	}
	return &snapshot, nil
}

// GetLatestSnapshot: 最新のスナップショットを取得
func (r *InventorySnapshotRepo) GetLatestSnapshot() (*model.InventoryDailySnapshot, error) {
	var snapshot model.InventoryDailySnapshot
	err := r.db.Preload("Details").
		Order("snapshot_date DESC").
		First(&snapshot).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &snapshot, nil
}

// DeleteSnapshotDetails: スナップショット詳細を削除
func (r *InventorySnapshotRepo) DeleteSnapshotDetails(snapshotDate time.Time) error {
	return r.db.Where("snapshot_date = ?", snapshotDate.Format("2006-01-02")).
		Delete(&model.InventoryDailySnapshotDetail{}).Error
}
