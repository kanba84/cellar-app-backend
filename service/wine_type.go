package service

import (
	"cellar-app/model"
	"fmt"
)

func (s *Service) ListWineTypes() ([]model.WineType, error) {
	wineTypes, err := s.WineTypeRepo.List()
	if err != nil {
		fmt.Printf("Error listing wine types: %v\n", err)
		return nil, err
	}
	return wineTypes, nil
}

func (s *Service) GetWineType(id uint) (*model.WineType, error) {
	wineType, err := s.WineTypeRepo.GetByID(id)
	if err != nil {
		fmt.Printf("Error retrieving wine type with ID %d: %v\n", id, err)
		return nil, err
	}
	return wineType, nil
}

func (s *Service) CreateWineType(wt *model.WineType) error {
	err := s.WineTypeRepo.Create(wt)
	if err != nil {
		fmt.Printf("Error creating wine type: %v\n", err)
		return err
	}
	fmt.Printf("Wine type created: %+v\n", wt)
	return nil
}

func (s *Service) UpdateWineType(wt *model.WineType) error {
	err := s.WineTypeRepo.Update(wt)
	if err != nil {
		fmt.Printf("Error updating wine type with ID %d: %v\n", wt.ID, err)
		return err
	}
	fmt.Printf("Wine type updated: %+v\n", wt)
	return nil
}

func (s *Service) DeleteWineType(id uint) error {
	err := s.WineTypeRepo.Delete(id)
	if err != nil {
		fmt.Printf("Error deleting wine type with ID %d: %v\n", id, err)
		return err
	}
	fmt.Printf("Wine type with ID %d deleted\n", id)
	return nil
}
