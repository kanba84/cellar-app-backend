package service

import (
	"cellar-app/model"
	"context"
	"fmt"

	"gorm.io/gorm"
)

func (s *Service) ListWines() ([]model.WineDTO, error) {
	wines, err := s.WineRepo.List()
	if err != nil {
		return nil, err
	}

	wineIDs := make([]uint, 0, len(wines))
	for i := range wines {
		wineIDs = append(wineIDs, wines[i].ID)
	}
	countsByWineID, err := s.BottleRepo.CountByWineIDs(wineIDs)
	if err != nil {
		return nil, err
	}

	var result []model.WineDTO

	for _, w := range wines {

		// --- ポインタ変換 ---
		var regionID *int
		if w.RegionID != nil {
			v := int(*w.RegionID)
			regionID = &v
		}

		var appellationID *int
		if w.AppellationID != nil {
			v := int(*w.AppellationID)
			appellationID = &v
		}

		dto := model.WineDTO{
			ID:            int(w.ID),
			Name:          w.Name,
			Vintage:       w.Vintage,
			WineTypeID:    int(w.WineTypeID),
			CountryID:     int(w.CountryID),
			RegionID:      regionID,
			Producer:      w.Producer,
			LabelImageURL: w.LabelImageURL,
			AppellationID: appellationID,

			// ★ ここは基本そのままでOK（Preload前提）
			WinTypeName:    w.WineType.Name,
			CountryName:    w.Country.Name,
			CountryISOCode: w.Country.ISOCode,
		}

		stockCount := countsByWineID[w.ID]
		dto.StockCount = stockCount
		dto.HasStock = stockCount > 0

		// --- nullable系だけチェック ---
		if w.Region != nil {
			dto.RegionName = &w.Region.Name
		}

		if w.Appellation != nil {
			dto.AppellationName = &w.Appellation.Name

			if w.Appellation.DesignationType != nil {
				// ★ ポインタ型に合わせる
				id := int(w.Appellation.DesignationType.ID)
				dto.DesignationTypeID = &id
				dto.DesignationTypeName = &w.Appellation.DesignationType.Name
			}
		}

		result = append(result, dto)
	}

	return result, nil
}

func (s *Service) GetWine(id uint) (*model.WineDTO, error) {
	wine, err := s.WineRepo.GetByID(id)
	if err != nil {
		fmt.Printf("Error retrieving wine with ID %d: %v\n", id, err)
		return nil, err
	}

	// WineをWineDTOに変換
	dto := convertWineToDTO(wine)

	stockCount, err := s.BottleRepo.CountByWineID(wine.ID)
	if err != nil {
		return nil, err
	}
	dto.StockCount = stockCount
	dto.HasStock = stockCount > 0

	return &dto, nil
}

func (s *Service) CreateWine(wine *model.Wine) error {
	// name, country_idは必須
	if wine.Name == "" || wine.CountryID == 0 {
		return fmt.Errorf("name and country_id are required")
	}

	// ラベル画像URLが未設定の場合、デフォルト値を設定
	if wine.LabelImageURL == nil {
		if wine.WineTypeID == 1 { // 赤ワインの場合
			defaultURL := "https://cellar-app.local/labels/sample_thumbnail.png"
			wine.LabelImageURL = &defaultURL
		} else { // 白ワイン/スパークリングの場合
			defaultURL := "https://cellar-app.local/labels/sample_thumbnail2.png"
			wine.LabelImageURL = &defaultURL
		}
	}

	err := s.WineRepo.Create(wine)
	if err != nil {
		fmt.Printf("Error creating wine: %v\n", err)
		return err
	}
	fmt.Printf("Wine created: %+v\n", wine)
	return nil
}

func (s *Service) CreateWineWithBottle(ctx context.Context, req model.CreateWineWithBottleRequest) (model.Wine, model.Bottle, error) {
	// ラベル画像URLが未設定の場合、デフォルト値を設定
	if req.Wine.LabelImageURL == nil {
		if req.Wine.WineTypeID == 1 {
			defaultURL := "https://cellar-app.local/labels/sample_thumbnail.png"
			req.Wine.LabelImageURL = &defaultURL
		} else {
			defaultURL := "https://cellar-app.local/labels/sample_thumbnail2.png"
			req.Wine.LabelImageURL = &defaultURL
		}
	}

	wine, bottle, err := s.WineRepo.CreateWithBottle(ctx, &req.Wine, &req.Bottle)
	if err != nil {
		if IsUniqueViolation(err) {
			return model.Wine{}, model.Bottle{}, ErrPositionOccupied
		}
		return model.Wine{}, model.Bottle{}, err
	}

	return *wine, *bottle, nil
}

// ヘルパー関数
// convertWineToDTO: Wine構造体をWineDTO構造体に変換
func convertWineToDTO(wine *model.Wine) model.WineDTO {
	dto := model.WineDTO{
		ID:             int(wine.ID),
		Name:           wine.Name,
		Vintage:        wine.Vintage,
		WineTypeID:     int(wine.WineTypeID),
		CountryID:      int(wine.CountryID),
		Producer:       wine.Producer,
		LabelImageURL:  wine.LabelImageURL,
		ReferencePrice: wine.ReferencePrice,
		WineGrapes:     convertWineGrapesToDTO(wine.WineGrapes),
		WinTypeName:    wine.WineType.Name,
		CountryName:    wine.Country.Name,
		CountryISOCode: wine.Country.ISOCode,
	}

	if wine.RegionID != nil {
		id := int(*wine.RegionID)
		dto.RegionID = &id
		if wine.Region != nil {
			dto.RegionName = &wine.Region.Name
		}
	}

	if wine.AppellationID != nil {
		id := int(*wine.AppellationID)
		dto.AppellationID = &id
		if wine.Appellation != nil {
			dto.AppellationName = &wine.Appellation.Name

			if wine.Appellation.DesignationType != nil {
				designationID := int(wine.Appellation.DesignationType.ID)
				dto.DesignationTypeID = &designationID
				dto.DesignationTypeName = &wine.Appellation.DesignationType.Name
			}
		}
	}

	return dto
}

func convertWineGrapesToDTO(wineGrapes []model.WineGrape) []model.WineGrapeDTO {
	result := make([]model.WineGrapeDTO, 0, len(wineGrapes))
	for _, wg := range wineGrapes {
		name := ""
		if wg.Grape.ID != 0 {
			name = wg.Grape.Name
		}
		result = append(result, model.WineGrapeDTO{
			Percentage:   wg.Percentage,
			DisplayOrder: wg.DisplayOrder,
			Name:         name,
		})
	}
	return result
}

func (s *Service) DeleteWine(id uint) error {
	err := s.WineRepo.Delete(id)
	if err != nil {
		fmt.Printf("Error deleting wine with ID %d: %v\n", id, err)
		return err
	}
	fmt.Printf("Wine with ID %d deleted\n", id)
	return nil
}

func (s *Service) UpdateWine(wine *model.Wine) error {
	err := s.WineRepo.Update(wine)
	if err != nil {
		fmt.Printf("Error updating wine with ID %d: %v\n", wine.ID, err)
		return err
	}
	fmt.Printf("Wine updated: %+v\n", wine)
	return nil
}

func (s *Service) PatchWine(id uint, updates map[string]interface{}) error {
	err := s.WineRepo.Patch(id, updates)
	if err != nil {
		fmt.Printf("Error patching wine with ID %d: %v\n", id, err)
		return err
	}
	fmt.Printf("Wine with ID %d patched successfully\n", id)
	return nil
}

// UpdateWineWithGrapes: ワイン情報とWineGrapesを更新します
// grapes が空でない場合は既存の WineGrapes を削除して新しく作成します
// 新しいブドウが必要な場合は自動作成します
func (s *Service) UpdateWineWithGrapes(ctx context.Context, wineID uint, updates map[string]interface{}, grapes []model.WineGrapeDTO) error {
	if wineID == 0 {
		return fmt.Errorf("wine id is required")
	}

	// トランザクション内で実行
	err := s.WineRepo.Transaction(ctx, func(tx *gorm.DB) error {
		// ワイン情報を更新
		if len(updates) > 0 {
			if err := tx.Model(&model.Wine{}).
				Where("id = ?", wineID).
				Updates(updates).Error; err != nil {
				return fmt.Errorf("failed to update wine: %w", err)
			}
		}

		// WineGrapes が提供されていない場合は終了
		if len(grapes) == 0 {
			return nil
		}

		// 既存の WineGrapes を削除
		if err := s.WineGrapeRepo.DeleteByWineIDTx(tx, wineID); err != nil {
			return fmt.Errorf("failed to delete existing wine grapes: %w", err)
		}

		// 新しい WineGrapes を作成
		wineGrapes := make([]model.WineGrape, 0, len(grapes))
		for order, grapeDTO := range grapes {
			// ブドウ名から既存ブドウを検索または作成
			// WineGrapeDTO.Name はノーマライズされたブドウ名を使用
			normalizedName := grapeDTO.Name
			if normalizedName == "" {
				return fmt.Errorf("grape name is required at position %d", order)
			}

			grape, err := s.GrapeRepo.GetOrCreateByNameTx(tx, normalizedName)
			if err != nil {
				return fmt.Errorf("failed to get or create grape '%s': %w", normalizedName, err)
			}

			wineGrape := model.WineGrape{
				WineID:       wineID,
				GrapeID:      grape.ID,
				Percentage:   grapeDTO.Percentage,
				DisplayOrder: order,
			}
			wineGrapes = append(wineGrapes, wineGrape)
		}

		// 新しい WineGrapes をバッチ作成
		if err := s.WineGrapeRepo.CreateBatchTx(tx, wineGrapes); err != nil {
			return fmt.Errorf("failed to save wine grapes: %w", err)
		}

		return nil
	})

	return err
}
