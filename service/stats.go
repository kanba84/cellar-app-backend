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

// GetInventoryTrend: 在庫数推移を集計
// days: 過去何日分のデータを取得するか（デフォルト30日）
// 戻り値は日付昇順でソート済み
func (s *Service) GetInventoryTrend(days int) ([]model.InventoryTrendDataPoint, error) {
	if days <= 0 {
		days = 30
	}

	bottles, err := s.BottleRepo.List()
	if err != nil {
		fmt.Printf("Error listing bottles for inventory trend: %v\n", err)
		return nil, err
	}

	// 日付ごとのボトル数をカウント
	// 日付キー（YYYY-MM-DD）-> そのボトル数
	inventoryByDate := make(map[string]int)
	dateSet := make(map[string]bool)

	// 過去daysの日付範囲を生成
	now := time.Now()
	endDate := now
	startDate := now.AddDate(0, 0, -days)

	// 対象期間の全日付を初期化
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		inventoryByDate[dateStr] = 0
		dateSet[dateStr] = true
	}

	// ボトルの追加・削除日付を処理して在庫数を計算
	for _, bottle := range bottles {
		if bottle.AddedAt == nil {
			continue
		}

		addedDate := bottle.AddedAt.Format("2006-01-02")
		removedDate := ""
		if bottle.RemovedAt != nil {
			removedDate = bottle.RemovedAt.Format("2006-01-02")
		}

		// 対象期間の日付について在庫数を計算
		for dateStr := range dateSet {
			bottleDate, _ := time.Parse("2006-01-02", dateStr)
			addedTime, _ := time.Parse("2006-01-02", addedDate)

			// ボトルが追加された日時点でカウント
			if !bottleDate.Before(addedTime) {
				// 削除日付が設定されている場合、削除前の日付のみカウント
				if removedDate != "" {
					removedTime, _ := time.Parse("2006-01-02", removedDate)
					if bottleDate.Before(removedTime) {
						inventoryByDate[dateStr]++
					}
				} else {
					// 削除されていないボトルはカウント
					inventoryByDate[dateStr]++
				}
			}
		}
	}

	// 日付順にソートしたデータを構築
	var data []model.InventoryTrendDataPoint
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		data = append(data, model.InventoryTrendDataPoint{
			Date:  dateStr,
			Count: inventoryByDate[dateStr],
		})
	}

	// 日付昇順でソート（既に昇順だが、念のため）
	sort.Slice(data, func(i, j int) bool {
		return data[i].Date < data[j].Date
	})

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
