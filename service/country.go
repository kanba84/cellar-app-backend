package service

import (
	"cellar-app/model"
	"fmt"
)

func (s *Service) ListCountries() ([]model.Country, error) {
	countries, err := s.CountryRepo.List()
	if err != nil {
		fmt.Printf("Error listing countries: %v\n", err)
		return nil, err
	}
	return countries, nil
}

func (s *Service) GetCountry(id uint) (*model.Country, error) {
	country, err := s.CountryRepo.GetByID(id)
	if err != nil {
		fmt.Printf("Error retrieving country with ID %d: %v\n", id, err)
		return nil, err
	}
	return country, nil
}

func (s *Service) CreateCountry(country *model.Country) error {
	err := s.CountryRepo.Create(country)
	if err != nil {
		fmt.Printf("Error creating country: %v\n", err)
		return err
	}
	fmt.Printf("Country created: %+v\n", country)
	return nil
}

func (s *Service) UpdateCountry(country *model.Country) error {
	err := s.CountryRepo.Update(country)
	if err != nil {
		fmt.Printf("Error updating country with ID %d: %v\n", country.ID, err)
		return err
	}
	fmt.Printf("Country updated: %+v\n", country)
	return nil
}

func (s *Service) DeleteCountry(id uint) error {
	err := s.CountryRepo.Delete(id)
	if err != nil {
		fmt.Printf("Error deleting country with ID %d: %v\n", id, err)
		return err
	}
	fmt.Printf("Country with ID %d deleted\n", id)
	return nil
}
