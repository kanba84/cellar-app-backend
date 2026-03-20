package service

import (
	"cellar-app/model"
	"fmt"
)

// --- Appellation CRUD ---

func (s *Service) ListAppellations() ([]model.Appellation, error) {
	appellations, err := s.AppellationRepo.List()
	if err != nil {
		fmt.Printf("Error listing appellations: %v\n", err)
		return nil, err
	}
	return appellations, nil
}

func (s *Service) GetAppellation(id uint) (*model.Appellation, error) {
	appellation, err := s.AppellationRepo.GetByID(id)
	if err != nil {
		fmt.Printf("Error retrieving appellation with ID %d: %v\n", id, err)
		return nil, err
	}
	return appellation, nil
}

func (s *Service) CreateAppellation(app *model.Appellation) error {
	err := s.AppellationRepo.Create(app)
	if err != nil {
		fmt.Printf("Error creating appellation: %v\n", err)
		return err
	}
	fmt.Printf("Appellation created: %+v\n", app)
	return nil
}

func (s *Service) UpdateAppellation(app *model.Appellation) error {
	err := s.AppellationRepo.Update(app)
	if err != nil {
		fmt.Printf("Error updating appellation with ID %d: %v\n", app.ID, err)
		return err
	}
	fmt.Printf("Appellation updated: %+v\n", app)
	return nil
}

func (s *Service) DeleteAppellation(id uint) error {
	err := s.AppellationRepo.Delete(id)
	if err != nil {
		fmt.Printf("Error deleting appellation with ID %d: %v\n", id, err)
		return err
	}
	fmt.Printf("Appellation with ID %d deleted\n", id)
	return nil
}
