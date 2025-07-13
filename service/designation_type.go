package service

import (
	"cellar-app/model"
	"context"
)

// --- DesignationType CRUD ---

func (s *Service) ListDesignationTypes() ([]model.DesignationType, error) {
	rows, err := s.Pool.Query(context.Background(), `SELECT id, name FROM designation_types`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []model.DesignationType{}
	for rows.Next() {
		var dt model.DesignationType
		if err := rows.Scan(&dt.ID, &dt.Name); err != nil {
			return nil, err
		}
		result = append(result, dt)
	}
	return result, nil
}

func (s *Service) GetDesignationType(id int) (*model.DesignationType, error) {
	row := s.Pool.QueryRow(context.Background(), `SELECT id, name FROM designation_types WHERE id = $1`, id)
	var dt model.DesignationType
	if err := row.Scan(&dt.ID, &dt.Name); err != nil {
		return nil, err
	}
	return &dt, nil
}

func (s *Service) CreateDesignationType(dt *model.DesignationType) error {
	return s.Pool.QueryRow(
		context.Background(),
		`INSERT INTO designation_types (name, code, rank, country_id) VALUES ($1, $2, $3, $4) RETURNING id`,
		dt.Name, dt.Code, dt.Rank, dt.CountryID,
	).Scan(&dt.ID)
}

func (s *Service) UpdateDesignationType(dt *model.DesignationType) error {
	_, err := s.Pool.Exec(
		context.Background(),
		`UPDATE designation_types SET name=$1 WHERE id=$2`,
		dt.Name, dt.ID,
	)
	return err
}

func (s *Service) DeleteDesignationType(id int) error {
	_, err := s.Pool.Exec(context.Background(), "DELETE FROM designation_types WHERE id = $1", id)
	return err
}
