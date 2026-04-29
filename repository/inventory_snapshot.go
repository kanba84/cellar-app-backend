package repository

import (
	"cellar-app/model"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InventorySnapshotRepo struct {
	db *gorm.DB
}

func NewInventorySnapshotRepo(db *gorm.DB) *InventorySnapshotRepo {
	return &InventorySnapshotRepo{db: db}
}

// SaveSnapshot: スナップショットを保存または更新（GORM Upsert + トランザクション）
// トランザクション内で以下を実行：
// ① 親テーブル（inventory_daily_snapshots）をUPSERT
// ② 子テーブル（inventory_daily_snapshot_details）をDELETE
// ③ 子テーブル（inventory_daily_snapshot_details）をINSERT
func (r *InventorySnapshotRepo) SaveSnapshot(snapshot *model.InventoryDailySnapshot) error {
	// トランザクション内で処理（① → ② → ③ の順序を保証）
	return r.db.Transaction(func(tx *gorm.DB) error {
		// ① 親テーブルをUPSERT（GORM Clauses使用）
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "snapshot_date"}},
			UpdateAll: true,
		}).Create(snapshot).Error; err != nil {
			return err
		}

		// ② 既存の子テーブルレコードをDELETE
		if err := tx.Where("snapshot_date = ?", snapshot.SnapshotDate.Format("2006-01-02")).
			Delete(&model.InventoryDailySnapshotDetail{}).Error; err != nil {
			return err
		}

		// ③ 新しい子テーブルレコードをINSERT（CreateInBatchesで効率的に処理）
		if len(snapshot.Details) > 0 {
			if err := tx.CreateInBatches(snapshot.Details, 100).Error; err != nil {
				return err
			}
		}

		return nil
	})
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

// GetSnapshotsByDateRange: 指定期間のスナップショット一覧を取得（日付昇順）
func (r *InventorySnapshotRepo) GetSnapshotsByDateRange(startDate, endDate time.Time) ([]model.InventoryDailySnapshot, error) {
	var snapshots []model.InventoryDailySnapshot
	err := r.db.Where("snapshot_date >= ? AND snapshot_date <= ?",
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02")).
		Order("snapshot_date ASC").
		Find(&snapshots).Error
	return snapshots, err
}
