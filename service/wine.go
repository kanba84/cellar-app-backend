package service

import (
	"cellar-app/model"
	"context"
	"fmt"
)

func (s *Service) ListWines() ([]model.WineDTO, error) {
	rows, err := s.Pool.Query(context.Background(), `
        SELECT 
            w.id, w.name, w.vintage, 
            w.wine_type_id, wt.name AS wine_type_name, 
            w.country_id, c.name AS country_name, 
            w.region_id, r.name AS region_name, 
            w.producer, w.label_image_url,
            w.appellation_id, a.name AS appellation_name,
            a.designation_type_id, dt.name AS designation_type_name
        FROM wines w
        LEFT JOIN wine_types wt ON w.wine_type_id = wt.id
        LEFT JOIN countries c ON w.country_id = c.id
        LEFT JOIN regions r ON w.region_id = r.id
        LEFT JOIN appellations a ON w.appellation_id = a.id
        LEFT JOIN designation_types dt ON a.designation_type_id = dt.id
		`)
	if err != nil {
		fmt.Printf("Error querying wines: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	wines := []model.WineDTO{}
	for rows.Next() {
		var w model.WineDTO

		if err := rows.Scan(
			&w.ID, &w.Name, &w.Vintage,
			&w.WineTypeID, &w.WinTypeName,
			&w.CountryID, &w.CountryName,
			&w.RegionID, &w.RegionName,
			&w.Producer, &w.LabelImageURL,
			&w.AppellationID, &w.AppellationName,
			&w.DesignationTypeID, &w.DesignationTypeName,
		); err != nil {
			return nil, err
		}
		wines = append(wines, w)
	}
	return wines, nil
}

func (s *Service) GetWine(id int) (*model.WineDTO, error) {
	row := s.Pool.QueryRow(context.Background(), `
        SELECT 
            w.id, w.name, w.vintage, 
            w.wine_type_id, wt.name AS wine_type_name, 
            w.country_id, c.name AS country_name, 
            w.region_id, r.name AS region_name, 
            w.producer, w.label_image_url,
            w.appellation_id, a.name AS appellation_name,
            a.designation_type_id, dt.name AS designation_type_name
        FROM wines w
        LEFT JOIN wine_types wt ON w.wine_type_id = wt.id
        LEFT JOIN countries c ON w.country_id = c.id
        LEFT JOIN regions r ON w.region_id = r.id
        LEFT JOIN appellations a ON w.appellation_id = a.id
        LEFT JOIN designation_types dt ON a.designation_type_id = dt.id
        WHERE w.id = $1
    `, id)
	var w model.WineDTO

	if err := row.Scan(
		&w.ID, &w.Name, &w.Vintage,
		&w.WineTypeID, &w.WinTypeName,
		&w.CountryID, &w.CountryName,
		&w.RegionID, &w.RegionName,
		&w.Producer, &w.LabelImageURL,
		&w.AppellationID, &w.AppellationName,
		&w.DesignationTypeID, &w.DesignationTypeName,
	); err != nil {
		fmt.Printf("Error retrieving wine with ID %d: %v\n", id, err)
		return nil, err
	}

	return &w, nil
}

func (s *Service) CreateWine(wine *model.Wine) error {
	// name, country_idは必須
	if wine.Name == "" || wine.CountryID == 0 {
		return fmt.Errorf("name and country_id are required")
	}

	// 可変長のカラムと値を構築
	columns := []string{"name", "country_id"}
	values := []interface{}{wine.Name, wine.CountryID}
	placeholders := []string{"$1", "$2"}
	i := 3

	if wine.Vintage != nil {
		columns = append(columns, "vintage")
		values = append(values, wine.Vintage)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		i++
	}
	if wine.WineTypeID != 0 {
		columns = append(columns, "wine_type_id")
		values = append(values, wine.WineTypeID)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		i++
	}
	if wine.RegionID != nil {
		columns = append(columns, "region_id")
		values = append(values, wine.RegionID)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		i++
	}
	if wine.Producer != nil {
		columns = append(columns, "producer")
		values = append(values, wine.Producer)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		i++
	}
	if wine.AppellationID != nil {
		columns = append(columns, "appellation_id")
		values = append(values, wine.AppellationID)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		i++
	}
	if wine.LabelImageURL != nil {
		columns = append(columns, "label_image_url")
		values = append(values, wine.LabelImageURL)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		i++
	} else if wine.WineTypeID == 1 { // 赤ワインの場合、デフォルトのラベル画像URLを設定
		defaultURL := "https://cellar-app.local/labels/sample_thumbnail.png"
		columns = append(columns, "label_image_url")
		values = append(values, defaultURL)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		i++
	} else { // 白ワイン/スパークリングの場合、デフォルトのラベル画像その２のURLを設定
		defaultURL := "https://cellar-app.local/labels/sample_thumbnail2.png"
		columns = append(columns, "label_image_url")
		values = append(values, defaultURL)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		i++
	}
	sql := fmt.Sprintf(
		"INSERT INTO wines (%s) VALUES (%s)",
		join(columns, ", "),
		join(placeholders, ", "),
	)

	_, err := s.Pool.Exec(context.Background(), sql, values...)
	if err != nil {
		fmt.Printf("Error creating wine: %v\n", err)
		return err
	}
	fmt.Printf("Wine created: %+v\n", wine)
	return nil
}

func (s *Service) CreateWineWithBottle(ctx context.Context, req model.CreateWineWithBottleRequest) (model.Wine, model.Bottle, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return model.Wine{}, model.Bottle{}, err
	}
	defer tx.Rollback(ctx)

	var wineID int
	if req.Wine.WineTypeID == 1 {
		*req.Wine.LabelImageURL = "https://cellar-app.local/labels/sample_thumbnail.png"
	} else {
		*req.Wine.LabelImageURL = "https://cellar-app.local/labels/sample_thumbnail2.png"
	}
	err = tx.QueryRow(ctx, `
        INSERT INTO wines (name, vintage, wine_type_id, country_id, region_id, producer, label_image_url)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id
    `, req.Wine.Name, req.Wine.Vintage, req.Wine.WineTypeID, req.Wine.CountryID, req.Wine.RegionID, req.Wine.Producer, req.Wine.LabelImageURL).Scan(&wineID)
	if err != nil {
		return model.Wine{}, model.Bottle{}, err
	}

	var bottleID int
	err = tx.QueryRow(ctx, `
        INSERT INTO bottles (wine_id, row_number, column_number, note)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `, wineID, req.Bottle.RowNumber, req.Bottle.ColumnNumber, req.Bottle.Note).Scan(&bottleID)
	if err != nil {
		return model.Wine{}, model.Bottle{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Wine{}, model.Bottle{}, err
	}

	return model.Wine{
			ID:            wineID,
			Name:          req.Wine.Name,
			Vintage:       req.Wine.Vintage,
			WineTypeID:    req.Wine.WineTypeID,
			CountryID:     req.Wine.CountryID,
			RegionID:      req.Wine.RegionID,
			Producer:      req.Wine.Producer,
			LabelImageURL: req.Wine.LabelImageURL,
		}, model.Bottle{
			ID:           bottleID,
			WineID:       wineID,
			RowNumber:    req.Bottle.RowNumber,
			ColumnNumber: req.Bottle.ColumnNumber,
			Note:         req.Bottle.Note,
		}, nil
}

// ヘルパー関数
func join(arr []string, sep string) string {
	result := ""
	for i, s := range arr {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

func (s *Service) DeleteWine(id int) error {
	_, err := s.Pool.Exec(context.Background(), "DELETE FROM wines WHERE id = $1", id)
	if err != nil {
		fmt.Printf("Error deleting wine with ID %d: %v\n", id, err)
		return err
	}
	fmt.Printf("Wine with ID %d deleted\n", id)
	return nil
}

func (s *Service) UpdateWine(wine *model.Wine) error {
	_, err := s.Pool.Exec(context.Background(),
		"UPDATE wines SET name = $1, vintage = $2, wine_type_id = $3, country_id = $4, region_id = $5, producer = $6 , appellation_id = $7 WHERE id = $8",
		wine.Name, wine.Vintage, wine.WineTypeID, wine.CountryID, wine.RegionID, wine.Producer, wine.AppellationID, wine.ID)
	if err != nil {
		fmt.Printf("Error updating wine with ID %d: %v\n", wine.ID, err)
		return err
	}
	fmt.Printf("Wine updated: %+v\n", wine)
	return nil
}

func (s *Service) PatchWine(id int, updates map[string]interface{}) error {
	query := "UPDATE wines SET "
	args := []interface{}{}
	i := 1

	for key, value := range updates {
		query += fmt.Sprintf("%s = $%d, ", key, i)
		args = append(args, value)
		i++
	}
	query = query[:len(query)-2] // Remove trailing comma and space
	query += " WHERE id = $" + fmt.Sprintf("%d", i)
	args = append(args, id)

	_, err := s.Pool.Exec(context.Background(), query, args...)
	if err != nil {
		fmt.Printf("Error patching wine with ID %d: %v\n", id, err)
		return err
	}
	fmt.Printf("Wine with ID %d patched successfully\n", id)
	return nil
}
