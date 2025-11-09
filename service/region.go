package service

import (
	"cellar-app/model"
	"context"
	"fmt"
)

func (s *Service) ListRegions() ([]model.Region, error) {
	rows, err := s.Pool.Query(context.Background(), `SELECT id, name, country_id, parent_id FROM regions ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	regions := []model.Region{}
	for rows.Next() {
		var r model.Region
		if err := rows.Scan(&r.ID, &r.Name, &r.CountryID, &r.ParentID); err != nil {
			return nil, err
		}
		regions = append(regions, r)
	}
	return regions, nil
}

func (s *Service) GetRegion(id int) (*model.Region, error) {
	row := s.Pool.QueryRow(context.Background(), `SELECT id, name, country_id FROM regions WHERE id = $1`, id)
	var r model.Region
	if err := row.Scan(&r.ID, &r.Name, &r.CountryID); err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Service) CreateRegion(region *model.Region) error {
	var (
		sql  string
		args []interface{}
	)
	if region.ParentID != nil {
		sql = `INSERT INTO regions (name, country_id, parent_id) VALUES ($1, $2, $3) RETURNING id`
		args = []interface{}{region.Name, region.CountryID, region.ParentID}
	} else {
		sql = `INSERT INTO regions (name, country_id, parent_id) VALUES ($1, $2, NULL) RETURNING id`
		args = []interface{}{region.Name, region.CountryID}
	}
	return s.Pool.QueryRow(context.Background(), sql, args...).Scan(&region.ID)
}

func (s *Service) UpdateRegion(region *model.Region) error {
	cmd, err := s.Pool.Exec(context.Background(),
		`UPDATE regions SET name = $1, country_id = $2 WHERE id = $3`, region.Name, region.CountryID, region.ID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func (s *Service) DeleteRegion(id int) error {
	cmd, err := s.Pool.Exec(context.Background(), `DELETE FROM regions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}
