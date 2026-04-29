package service

import (
	"cellar-app/model"
	"fmt"
	"sort"
	"time"
)

// GetWineTypeStats: ワインタイプ別の在庫構成比を集計
// RemovedAtがnullのボトル（在庫にある）のみを対象
// countが多い順でソート
func (s *Service) GetWineTypeStats() ([]model.WineTypeStats, error) {
	bottles, err := s.BottleRepo.List()
	if err != nil {
		fmt.Printf("Error listing bottles for wine type stats: %v\n", err)
		return nil, err
	}

	// ワインタイプ別にカウント
	wineTypeMap := make(map[uint]string)
	wineTypeCounts := make(map[uint]int)

	for _, bottle := range bottles {
		// 在庫にあるボトルのみ対象（RemovedAtがnull）
		if bottle.RemovedAt != nil {
			continue
		}

		wineTypeID := bottle.Wine.WineTypeID
		wineTypeMap[wineTypeID] = bottle.Wine.WineType.Name
		wineTypeCounts[wineTypeID]++
	}

	// 結果をスライスに変換
	var stats []model.WineTypeStats
	for typeID, count := range wineTypeCounts {
		stats = append(stats, model.WineTypeStats{
			Name:  wineTypeMap[typeID],
			Count: count,
		})
	}

	// countが多い順でソート（降順）
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	return stats, nil
}

// GetCountryStats: 生産国別の在庫構成比を集計
// RemovedAtがnullのボトル（在庫にある）のみを対象
// countが多い順でソート
func (s *Service) GetCountryStats() ([]model.CountryStats, error) {
	bottles, err := s.BottleRepo.List()
	if err != nil {
		fmt.Printf("Error listing bottles for country stats: %v\n", err)
		return nil, err
	}

	// 生産国別にカウント
	countryMap := make(map[uint]string)
	countryCounts := make(map[uint]int)

	for _, bottle := range bottles {
		// 在庫にあるボトルのみ対象（RemovedAtがnull）
		if bottle.RemovedAt != nil {
			continue
		}

		countryID := bottle.Wine.CountryID
		countryMap[countryID] = bottle.Wine.Country.Name
		countryCounts[countryID]++
	}

	// 結果をスライスに変換
	var stats []model.CountryStats
	for countryID, count := range countryCounts {
		stats = append(stats, model.CountryStats{
			Name:  countryMap[countryID],
			Count: count,
		})
	}

	// countが多い順でソート（降順）
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	return stats, nil
}

// GetInventoryTrend: 在庫数推移を集計（スナップショットから取得）
// days: 過去何日分のデータを取得するか（デフォルト30日）
// 戻り値は日付昇順でソート済み
// スナップショットが存在しない日については、最も近い直前のスナップショットで補完
func (s *Service) GetInventoryTrend(days int) ([]model.InventoryTrendDataPoint, error) {
	if days <= 0 {
		days = 30
	}

	// 過去daysの日付範囲を生成
	now := time.Now()
	endDate := now
	startDate := now.AddDate(0, 0, -days)

	// 対象期間内のスナップショットを日付昇順で取得
	snapshots, err := s.SnapshotRepo.GetSnapshotsByDateRange(startDate, endDate)
	if err != nil {
		fmt.Printf("Error retrieving snapshots for inventory trend: %v\n", err)
		return nil, err
	}

	// 日付ごとのボトル数マップを作成
	snapshotMap := make(map[string]int)
	for _, snapshot := range snapshots {
		dateStr := snapshot.SnapshotDate.Format("2006-01-02")
		snapshotMap[dateStr] = snapshot.TotalCount
	}

	// 対象期間の全日付でデータを生成
	var data []model.InventoryTrendDataPoint
	lastKnownCount := 0
	lastSnapshotFound := false

	// 開始日より前のスナップショットから最新値を取得（初期値設定用）
	if len(snapshots) > 0 {
		lastKnownCount = snapshots[0].TotalCount
		lastSnapshotFound = true
	}

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")

		// スナップショットが存在する場合
		if count, exists := snapshotMap[dateStr]; exists {
			lastKnownCount = count
			lastSnapshotFound = true
			data = append(data, model.InventoryTrendDataPoint{
				Date:  dateStr,
				Count: count,
			})
		} else {
			// スナップショットが存在しない場合は、最後に見つかったスナップショットで補完
			if lastSnapshotFound {
				data = append(data, model.InventoryTrendDataPoint{
					Date:  dateStr,
					Count: lastKnownCount,
				})
			} else {
				// スナップショットがまったく存在しない場合は0
				data = append(data, model.InventoryTrendDataPoint{
					Date:  dateStr,
					Count: 0,
				})
			}
		}
	}

	return data, nil
}

// GetVintageStats: ワインのビンテージ別の在庫構成比を集計
// RemovedAtがnullのボトル（在庫にある）のみを対象
// Vintageがnullのワインは除外
// countが多い順でソート
func (s *Service) GetVintageStats() ([]model.VintageStats, error) {
	bottles, err := s.BottleRepo.List()
	if err != nil {
		fmt.Printf("Error listing bottles for vintage stats: %v\n", err)
		return nil, err
	}

	// ビンテージ別にカウント
	vintageCounts := make(map[int]int)

	for _, bottle := range bottles {
		// 在庫にあるボトルのみ対象（RemovedAtがnull）
		if bottle.RemovedAt != nil {
			continue
		}

		// Vintageがnullの場合はスキップ
		if bottle.Wine.Vintage == nil {
			continue
		}

		vintage := *bottle.Wine.Vintage
		vintageCounts[vintage]++
	}

	// 結果をスライスに変換
	var stats []model.VintageStats
	for vintage, count := range vintageCounts {
		stats = append(stats, model.VintageStats{
			Vintage: vintage,
			Count:   count,
		})
	}

	// vintageの昇順（古い順）でソート
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Vintage < stats[j].Vintage
	})

	return stats, nil
}

// GetStats: すべての統計情報を取得
func (s *Service) GetStats(days int) (*model.Stats, error) {
	wineTypeStats, err := s.GetWineTypeStats()
	if err != nil {
		return nil, err
	}

	countryStats, err := s.GetCountryStats()
	if err != nil {
		return nil, err
	}

	vintageStats, err := s.GetVintageStats()
	if err != nil {
		return nil, err
	}

	inventoryTrend, err := s.GetInventoryTrend(days)
	if err != nil {
		return nil, err
	}

	return &model.Stats{
		WineTypes:      wineTypeStats,
		Countries:      countryStats,
		Vintages:       vintageStats,
		InventoryTrend: inventoryTrend,
	}, nil
}
