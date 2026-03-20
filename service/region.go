package service

import (
	"cellar-app/model"
	"fmt"
)

func (s *Service) ListRegions() ([]model.Region, error) {
	regions, err := s.RegionRepo.List()
	if err != nil {
		fmt.Printf("Error listing regions: %v\n", err)
		return nil, err
	}
	return regions, nil
}

func (s *Service) GetRegion(id uint) (*model.Region, error) {
	region, err := s.RegionRepo.GetByID(id)
	if err != nil {
		fmt.Printf("Error retrieving region with ID %d: %v\n", id, err)
		return nil, err
	}
	return region, nil
}

func (s *Service) CreateRegion(region *model.Region) error {
	err := s.RegionRepo.Create(region)
	if err != nil {
		fmt.Printf("Error creating region: %v\n", err)
		return err
	}
	fmt.Printf("Region created: %+v\n", region)
	return nil
}

func (s *Service) UpdateRegion(region *model.Region) error {
	err := s.RegionRepo.Update(region)
	if err != nil {
		fmt.Printf("Error updating region with ID %d: %v\n", region.ID, err)
		return err
	}
	fmt.Printf("Region updated: %+v\n", region)
	return nil
}

func (s *Service) DeleteRegion(id uint) error {
	err := s.RegionRepo.Delete(id)
	if err != nil {
		fmt.Printf("Error deleting region with ID %d: %v\n", id, err)
		return err
	}
	fmt.Printf("Region with ID %d deleted\n", id)
	return nil
}
