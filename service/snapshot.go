package service

import (
	"cellar-app/model"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

var snapshotScheduler *cron.Cron

// CreateOrUpdateDailySnapshot: 指定日付のスナップショットを作成または更新
func (s *Service) CreateOrUpdateDailySnapshot(snapshotDate time.Time) error {
	// 統計情報を取得
	stats, err := s.GetStats(30)
	if err != nil {
		fmt.Printf("Error getting stats for snapshot: %v\n", err)
		return err
	}

	// スナップショット日付を YYYY-MM-DD にフォーマット
	snapshotDateOnly := time.Date(snapshotDate.Year(), snapshotDate.Month(), snapshotDate.Day(), 0, 0, 0, 0, snapshotDate.Location())

	// スナップショットを作成
	snapshot := &model.InventoryDailySnapshot{
		SnapshotDate: snapshotDateOnly,
		TotalCount:   s.calculateTotalCount(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Details:      []model.InventoryDailySnapshotDetail{},
	}

	// 既存のスナップショット詳細を削除
	if err := s.SnapshotRepo.DeleteSnapshotDetails(snapshotDateOnly); err != nil {
		fmt.Printf("Error deleting old snapshot details: %v\n", err)
		return err
	}

	// ワインタイプ別の詳細を追加
	for _, stat := range stats.WineTypes {
		detail := model.InventoryDailySnapshotDetail{
			SnapshotDate: snapshotDateOnly,
			CategoryType: "wine_type",
			CategoryName: &stat.Name,
			Count:        stat.Count,
			CreatedAt:    time.Now(),
		}
		snapshot.Details = append(snapshot.Details, detail)
	}

	// 生産国別の詳細を追加
	for _, stat := range stats.Countries {
		detail := model.InventoryDailySnapshotDetail{
			SnapshotDate: snapshotDateOnly,
			CategoryType: "country",
			CategoryName: &stat.Name,
			Count:        stat.Count,
			CreatedAt:    time.Now(),
		}
		snapshot.Details = append(snapshot.Details, detail)
	}

	// ビンテージ別の詳細を追加
	for _, stat := range stats.Vintages {
		vintageID := stat.Vintage
		vintageStr := fmt.Sprintf("%d", stat.Vintage)
		detail := model.InventoryDailySnapshotDetail{
			SnapshotDate: snapshotDateOnly,
			CategoryType: "vintage",
			CategoryID:   &vintageID,
			CategoryName: &vintageStr,
			Count:        stat.Count,
			CreatedAt:    time.Now(),
		}
		snapshot.Details = append(snapshot.Details, detail)
	}

	// スナップショットを保存
	if err := s.SnapshotRepo.SaveSnapshot(snapshot); err != nil {
		fmt.Printf("Error saving snapshot: %v\n", err)
		return err
	}

	fmt.Printf("Snapshot created/updated for date: %s\n", snapshotDateOnly.Format("2006-01-02"))
	return nil
}

// UpdateSnapshotIfNeeded: ボトル変更時にスナップショットを更新
// 本日のスナップショットが存在する場合のみ更新（スケジュール済みの場合のみ）
func (s *Service) UpdateSnapshotIfNeeded() error {
	today := time.Now()
	todayOnly := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	// 本日のスナップショットが存在するか確認
	existing, err := s.SnapshotRepo.GetSnapshot(todayOnly)
	if err != nil {
		fmt.Printf("Error checking existing snapshot: %v\n", err)
		return err
	}

	// スナップショットが存在する場合のみ更新
	if existing != nil {
		if err := s.CreateOrUpdateDailySnapshot(todayOnly); err != nil {
			fmt.Printf("Error updating snapshot: %v\n", err)
			return err
		}
	}

	return nil
}

// calculateTotalCount: 現在の総ボトル数を計算
func (s *Service) calculateTotalCount() int {
	bottles, err := s.BottleRepo.List()
	if err != nil {
		fmt.Printf("Error counting bottles: %v\n", err)
		return 0
	}

	// RemovedAtがnullのボトルのみカウント
	count := 0
	for _, bottle := range bottles {
		if bottle.RemovedAt == nil {
			count++
		}
	}

	return count
}

// GetDailySnapshot: 指定日付のスナップショットを取得
func (s *Service) GetDailySnapshot(snapshotDate time.Time) (*model.InventoryDailySnapshot, error) {
	snapshotDateOnly := time.Date(snapshotDate.Year(), snapshotDate.Month(), snapshotDate.Day(), 0, 0, 0, 0, snapshotDate.Location())
	return s.SnapshotRepo.GetSnapshot(snapshotDateOnly)
}

// StartDailySnapshotScheduler: 毎日0:00にスナップショットを作成するスケジューラーを開始
func (s *Service) StartDailySnapshotScheduler() error {
	if snapshotScheduler != nil {
		fmt.Println("[Scheduler] Snapshot scheduler is already running")
		return nil
	}

	snapshotScheduler = cron.New()

	// 毎日0:00にスナップショットを作成 (CRON式: 0 0 * * *)
	_, err := snapshotScheduler.AddFunc("0 0 * * *", func() {
		today := time.Now()
		todayOnly := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

		fmt.Printf("[Scheduler] Creating daily snapshot for %s\n", todayOnly.Format("2006-01-02"))

		if err := s.CreateOrUpdateDailySnapshot(todayOnly); err != nil {
			fmt.Printf("[Scheduler] Error creating daily snapshot: %v\n", err)
		}
	})

	if err != nil {
		fmt.Printf("[Scheduler] Error registering snapshot job: %v\n", err)
		return err
	}

	snapshotScheduler.Start()
	fmt.Println("[Scheduler] Daily snapshot scheduler started (0:00 UTC)")
	return nil
}

// StopDailySnapshotScheduler: スケジューラーを停止
func (s *Service) StopDailySnapshotScheduler() {
	if snapshotScheduler != nil {
		snapshotScheduler.Stop()
		snapshotScheduler = nil
		fmt.Println("[Scheduler] Daily snapshot scheduler stopped")
	}
}
