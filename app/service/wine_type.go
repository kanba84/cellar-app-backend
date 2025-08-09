package service

import (
	"cellar-app/model"
	"context"
	"fmt"
)

func (s *Service) ListWineTypes() ([]model.WineType, error) {
	rows, err := s.Pool.Query(context.Background(), `SELECT id, name FROM wine_types ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	wineTypes := []model.WineType{}
	for rows.Next() {
		var wt model.WineType
		if err := rows.Scan(&wt.ID, &wt.Name); err != nil {
			return nil, err
		}
		wineTypes = append(wineTypes, wt)
	}
	return wineTypes, nil
}

func (s *Service) GetWineType(id int) (*model.WineType, error) {
	row := s.Pool.QueryRow(context.Background(), `SELECT id, name FROM wine_types WHERE id = $1`, id)
	var wt model.WineType
	if err := row.Scan(&wt.ID, &wt.Name); err != nil {
		return nil, err
	}
	return &wt, nil
}

func (s *Service) CreateWineType(wt *model.WineType) error {
	err := s.Pool.QueryRow(context.Background(),
		`INSERT INTO wine_types (name) VALUES ($1) RETURNING id`, wt.Name).Scan(&wt.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) UpdateWineType(wt *model.WineType) error {
	cmd, err := s.Pool.Exec(context.Background(),
		`UPDATE wine_types SET name = $1 WHERE id = $2`, wt.Name, wt.ID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func (s *Service) DeleteWineType(id int) error {
	cmd, err := s.Pool.Exec(context.Background(), `DELETE FROM wine_types WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}
