package service

import (
	"cellar-app/model"
	"fmt"
)

// --- DesignationType CRUD ---

func (s *Service) ListDesignationTypes() ([]model.DesignationType, error) {
	designationTypes, err := s.DesignationTypeRepo.List()
	if err != nil {
		fmt.Printf("Error listing designation types: %v\n", err)
		return nil, err
	}
	return designationTypes, nil
}

func (s *Service) GetDesignationType(id uint) (*model.DesignationType, error) {
	designationType, err := s.DesignationTypeRepo.GetByID(id)
	if err != nil {
		fmt.Printf("Error retrieving designation type with ID %d: %v\n", id, err)
		return nil, err
	}
	return designationType, nil
}

func (s *Service) CreateDesignationType(dt *model.DesignationType) error {
	err := s.DesignationTypeRepo.Create(dt)
	if err != nil {
		fmt.Printf("Error creating designation type: %v\n", err)
		return err
	}
	fmt.Printf("Designation type created: %+v\n", dt)
	return nil
}

func (s *Service) UpdateDesignationType(dt *model.DesignationType) error {
	err := s.DesignationTypeRepo.Update(dt)
	if err != nil {
		fmt.Printf("Error updating designation type with ID %d: %v\n", dt.ID, err)
		return err
	}
	fmt.Printf("Designation type updated: %+v\n", dt)
	return nil
}

func (s *Service) DeleteDesignationType(id uint) error {
	err := s.DesignationTypeRepo.Delete(id)
	if err != nil {
		fmt.Printf("Error deleting designation type with ID %d: %v\n", id, err)
		return err
	}
	fmt.Printf("Designation type with ID %d deleted\n", id)
	return nil
}
