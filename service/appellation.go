package service

import (
	"cellar-app/model"
	"context"
)

// --- Appellation CRUD ---

func (s *Service) ListAppellations() ([]model.Appellation, error) {
	rows, err := s.Pool.Query(context.Background(), `
        SELECT 
            a.id, a.name, a.designation_type_id, a.region_id,
            dt.id, dt.name
        FROM appellations a
        JOIN designation_types dt ON a.designation_type_id = dt.id
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []model.Appellation{}
	for rows.Next() {
		var a model.Appellation
		if err := rows.Scan(
			&a.ID, &a.Name, &a.DesignationTypeID, &a.RegionID,
			&a.DesignationTypeID, &a.DesignationTypeName,
		); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, nil
}

func (s *Service) GetAppellation(id int) (*model.Appellation, error) {
	row := s.Pool.QueryRow(context.Background(), `
        SELECT 
            a.id, a.name, a.designation_type_id, a.region_id,
            dt.id, dt.name
        FROM appellations a
        JOIN designation_types dt ON a.designation_type_id = dt.id
        WHERE a.id = $1
    `, id)

	var a model.Appellation
	//var regionID *int
	if err := row.Scan(
		&a.ID, &a.Name, &a.DesignationTypeID, &a.RegionID,
		&a.DesignationTypeID, &a.DesignationTypeName,
	); err != nil {
		return nil, err
	}
	//a.RegionID = regionID
	//a.DesignationType = dt
	return &a, nil
}

func (s *Service) CreateAppellation(app *model.Appellation) error {
	err := s.Pool.QueryRow(
		context.Background(),
		`INSERT INTO appellations (name, designation_type_id, region_id)
         VALUES ($1, $2, $3) RETURNING id`,
		app.Name, app.DesignationTypeID, app.RegionID,
	).Scan(&app.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) UpdateAppellation(app *model.Appellation) error {
	_, err := s.Pool.Exec(
		context.Background(),
		`UPDATE appellations SET name=$1, designation_type_id=$2, region_id=$3 WHERE id=$3`,
		app.Name, app.DesignationTypeID, app.RegionID, app.ID,
	)
	return err
}

func (s *Service) DeleteAppellation(id int) error {
	_, err := s.Pool.Exec(context.Background(), "DELETE FROM appellations WHERE id = $1", id)
	return err
}
