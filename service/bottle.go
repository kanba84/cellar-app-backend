package service

import (
	"cellar-app/model"
	"fmt"
)

func (s *Service) ListBottles() ([]model.Bottle, error) {
	bottles, err := s.BottleRepo.List()
	if err != nil {
		fmt.Printf("Error listing bottles: %v\n", err)
		return nil, err
	}
	return bottles, nil
}

func (s *Service) GetBottle(id uint) (*model.Bottle, error) {
	bottle, err := s.BottleRepo.GetByID(id)
	if err != nil {
		fmt.Printf("Error retrieving bottle with ID %d: %v\n", id, err)
		return nil, err
	}
	return bottle, nil
}

func (s *Service) CreateBottle(bottle *model.Bottle) error {
	err := s.BottleRepo.Create(bottle)
	if err != nil {
		fmt.Printf("Error creating bottle: %v\n", err)
		return err
	}
	fmt.Printf("Bottle created: %+v\n", bottle)
	return nil
}

func (s *Service) DeleteBottle(id uint) error {
	err := s.BottleRepo.Delete(id)
	if err != nil {
		fmt.Printf("Error deleting bottle with ID %d: %v\n", id, err)
		return err
	}
	fmt.Printf("Bottle with ID %d deleted\n", id)
	return nil
}

func (s *Service) UpdateBottle(bottle *model.Bottle) error {
	err := s.BottleRepo.Update(bottle)
	if err != nil {
		fmt.Printf("Error updating bottle with ID %d: %v\n", bottle.ID, err)
		return err
	}
	fmt.Printf("Bottle updated: %+v\n", bottle)
	return nil
}

func (s *Service) PatchBottle(id uint, updates map[string]interface{}) error {
	// GORMではMapを使用してパッチ更新
	err := s.BottleRepo.Patch(id, updates)
	if err != nil {
		fmt.Printf("Error patching bottle with ID %d: %v\n", id, err)
		return err
	}
	fmt.Printf("Bottle with ID %d patched successfully\n", id)
	return nil
}
