package service

import (
	"cellar-app/model"
	"context"
	"fmt"
)

func (s *Service) ListCountries() ([]model.Country, error) {
	rows, err := s.Pool.Query(context.Background(), `SELECT id, name FROM countries ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	countries := []model.Country{}
	for rows.Next() {
		var c model.Country
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		countries = append(countries, c)
	}
	return countries, nil
}

func (s *Service) GetCountry(id int) (*model.Country, error) {
	row := s.Pool.QueryRow(context.Background(), `SELECT id, name FROM countries WHERE id = $1`, id)
	var c model.Country
	if err := row.Scan(&c.ID, &c.Name); err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Service) CreateCountry(country *model.Country) error {
	err := s.Pool.QueryRow(context.Background(),
		`INSERT INTO countries (name) VALUES ($1) RETURNING id`, country.Name).Scan(&country.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) UpdateCountry(country *model.Country) error {
	cmd, err := s.Pool.Exec(context.Background(),
		`UPDATE countries SET name = $1 WHERE id = $2`, country.Name, country.ID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func (s *Service) DeleteCountry(id int) error {
	cmd, err := s.Pool.Exec(context.Background(), `DELETE FROM countries WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}
