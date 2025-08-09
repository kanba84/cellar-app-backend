package service

import (
	"cellar-app/model"
	"context"
	"fmt"
	"time"
)

func (s *Service) ListBottles() ([]model.Bottle, error) {
	query := `
        SELECT 
            b.id, b.wine_id, b.is_opened, b.added_at, b.removed_at, b.row_number, b.column_number, b.note,
            w.id, w.name, w.vintage, w.wine_type_id, wt.name as wine_type_name, w.country_id, c.name as country_name, w.region_id, r.name as region_name, w.producer
        FROM bottles b
        JOIN wines w ON b.wine_id = w.id
        LEFT JOIN wine_types wt ON w.wine_type_id = wt.id
        LEFT JOIN countries c ON w.country_id = c.id
        LEFT JOIN regions r ON w.region_id = r.id
    `
	rows, err := s.Pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bottles := []model.Bottle{}
	for rows.Next() {
		var bottle model.Bottle
		var wine model.WineDTO
		// Bottle と Wine のフィールドをスキャン
		if err := rows.Scan(
			&bottle.ID, &bottle.WineID, &bottle.IsOpened, &bottle.AddedAt, &bottle.RemovedAt, &bottle.RowNumber, &bottle.ColumnNumber, &bottle.Note,
			&wine.ID, &wine.Name, &wine.Vintage, &wine.WineTypeID, &wine.WinTypeName, &wine.CountryID, &wine.CountryName, &wine.RegionID, &wine.RegionName, &wine.Producer,
		); err != nil {
			return nil, err
		}
		bottle.Wine = wine // Bottle に Wine を埋め込む
		bottles = append(bottles, bottle)
	}

	return bottles, nil
}

func (s *Service) GetBottle(id int) (*model.Bottle, error) {
	query := `
        SELECT 
            b.id, b.wine_id, b.is_opened, b.added_at, b.removed_at, b.row_number, b.column_number, b.note,
            w.id, w.name, w.vintage, w.wine_type_id, wt.name as wine_type_name, w.country_id, c.name as country_name, w.region_id, r.name as region_name, w.producer
        FROM bottles b
        JOIN wines w ON b.wine_id = w.id
        LEFT JOIN wine_types wt ON w.wine_type_id = wt.id
        LEFT JOIN countries c ON w.country_id = c.id
        LEFT JOIN regions r ON w.region_id = r.id
        WHERE b.id = $1
    `
	row := s.Pool.QueryRow(context.Background(), query, id)

	var bottle model.Bottle
	var wine model.WineDTO
	if err := row.Scan(
		&bottle.ID, &bottle.WineID, &bottle.IsOpened, &bottle.AddedAt, &bottle.RemovedAt, &bottle.RowNumber, &bottle.ColumnNumber, &bottle.Note,
		&wine.ID, &wine.Name, &wine.Vintage, &wine.WineTypeID, &wine.WinTypeName, &wine.CountryID, &wine.CountryName, &wine.RegionID, &wine.RegionName, &wine.Producer,
	); err != nil {
		return nil, err
	}
	bottle.Wine = wine // Bottle に Wine を埋め込む

	return &bottle, nil
}

func (s *Service) CreateBottle(bottle *model.Bottle) error {
	if bottle.AddedAt == nil {
		now := time.Now()
		bottle.AddedAt = &now
	}
	_, err := s.Pool.Exec(context.Background(),
		"INSERT INTO bottles (wine_id, is_opened, added_at, removed_at, row_number, column_number, note) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		bottle.WineID, bottle.IsOpened, bottle.AddedAt, bottle.RemovedAt, bottle.RowNumber, bottle.ColumnNumber, bottle.Note)
	if err != nil {
		fmt.Printf("Error creating bottle: %v\n", err)
		return err
	}
	fmt.Printf("Bottle created: %+v\n", bottle)
	return nil
}

func (s *Service) DeleteBottle(id int) error {
	_, err := s.Pool.Exec(context.Background(), "DELETE FROM bottles WHERE id = $1", id)
	if err != nil {
		fmt.Printf("Error deleting bottle with ID %d: %v\n", id, err)
		return err
	}
	fmt.Printf("Bottle with ID %d deleted\n", id)
	return nil
}

func (s *Service) UpdateBottle(bottle *model.Bottle) error {
	_, err := s.Pool.Exec(context.Background(),
		"UPDATE bottles SET wine_id = $1, is_opened = $2, added_at = $3, removed_at = $4, row_number = $5, column_number = $6, note = $7 WHERE id = $8",
		bottle.WineID, bottle.IsOpened, bottle.AddedAt, bottle.RemovedAt, bottle.RowNumber, bottle.ColumnNumber, bottle.Note, bottle.ID)
	if err != nil {
		fmt.Printf("Error updating bottle with ID %d: %v\n", bottle.ID, err)
		return err
	}
	fmt.Printf("Bottle updated: %+v\n", bottle)
	return nil
}

func (s *Service) PatchBottle(id int, updates map[string]interface{}) error {
	query := "UPDATE bottles SET "
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
		fmt.Printf("Error patching bottle with ID %d: %v\n", id, err)
		return err
	}
	fmt.Printf("Bottle with ID %d patched successfully\n", id)
	return nil
}
