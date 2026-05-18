package service

import (
	"cellar-app/llm"
	"cellar-app/model"
	"context"
	"fmt"

	"gorm.io/gorm"
)

// FetchWineInfoOnly: ワイン名から LLM を使用してワイン情報を取得します（DB保存なし）
// フロントエンドで表示し、ユーザーの操作により別途保存する想定です
func (s *Service) FetchWineInfoOnly(ctx context.Context, wine *model.Wine) (*llm.WineInfoResult, error) {
	if wine == nil {
		return nil, fmt.Errorf("wine is nil")
	}

	if s.LLMProvider == nil {
		return nil, fmt.Errorf("LLM provider is not configured")
	}

	// LLM から情報を取得
	lookupKey := llm.WineLookupKey{Name: wine.Name}
	if wine.Vintage != nil {
		lookupKey.Vintage = wine.Vintage
	}
	info, err := s.LLMProvider.FetchWineInfo(ctx, lookupKey)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch wine info from LLM: %w", err)
	}

	if info == nil {
		return nil, fmt.Errorf("no wine info returned from LLM")
	}

	return info, nil
}

// FetchAndSaveWineInfo: ワイン名から LLM を使用してワイン情報を取得し、DBに保存します
func (s *Service) FetchAndSaveWineInfo(ctx context.Context, wine *model.Wine) (*llm.WineInfoResult, error) {
	if wine == nil {
		return nil, fmt.Errorf("wine is nil")
	}

	if s.LLMProvider == nil {
		return nil, fmt.Errorf("LLM provider is not configured")
	}

	// LLM から情報を取得
	lookupKey := llm.WineLookupKey{Name: wine.Name}
	if wine.Vintage != nil {
		lookupKey.Vintage = wine.Vintage
	}
	info, err := s.LLMProvider.FetchWineInfo(ctx, lookupKey)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch wine info from LLM: %w", err)
	}

	if info == nil {
		return nil, fmt.Errorf("no wine info returned from LLM")
	}

	err = s.WineRepo.Transaction(ctx, func(tx *gorm.DB) error {
		if info.Producer != nil && (wine.Producer == nil || *wine.Producer == "") {
			updates := map[string]interface{}{"producer": info.Producer}
			if err := tx.Model(&model.Wine{}).
				Where("id = ?", wine.ID).
				Updates(updates).Error; err != nil {
				return fmt.Errorf("failed to update wine producer: %w", err)
			}
		}

		if len(info.Grapes) == 0 {
			return nil
		}

		if err := s.WineGrapeRepo.DeleteByWineIDTx(tx, wine.ID); err != nil {
			return fmt.Errorf("failed to delete existing wine grapes: %w", err)
		}

		wineGrapes := make([]model.WineGrape, 0, len(info.Grapes))
		for order, grapeInfo := range info.Grapes {
			normalizedName := grapeInfo.NormalizedName
			if normalizedName == "" {
				normalizedName = llm.NormalizeGrapeName(grapeInfo.Name)
			}

			grape, err := s.GrapeRepo.GetOrCreateByNameTx(tx, normalizedName)
			if err != nil {
				return fmt.Errorf("failed to get or create grape '%s': %w", normalizedName, err)
			}

			wineGrape := model.WineGrape{
				WineID:       wine.ID,
				GrapeID:      grape.ID,
				Percentage:   grapeInfo.Percentage,
				DisplayOrder: order,
			}
			wineGrapes = append(wineGrapes, wineGrape)
		}

		if err := s.WineGrapeRepo.CreateBatchTx(tx, wineGrapes); err != nil {
			return fmt.Errorf("failed to save wine grapes: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to persist wine llm data: %w", err)
	}

	return info, nil
}
