package service

import (
	"cellar-app/model"
	"fmt"
)

func (s *Service) ListBottles() ([]model.BottleWithWineDTO, error) {
	bottles, err := s.BottleRepo.List()
	if err != nil {
		fmt.Printf("Error listing bottles: %v\n", err)
		return nil, err
	}
	dtos := s.toBottleWithWineDTOs(bottles)
	return dtos, nil
}

func (s *Service) GetBottle(id uint) (*model.BottleWithWineDTO, error) {
	bottle, err := s.BottleRepo.GetByID(id)
	if err != nil {
		fmt.Printf("Error retrieving bottle with ID %d: %v\n", id, err)
		return nil, err
	}
	dto := s.toBottleWithWineDTO(bottle)
	return &dto, nil
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

func (s *Service) toBottleWithWineDTO(b *model.Bottle) model.BottleWithWineDTO {
	var wineDTO model.WineDTO

	if b.Wine.ID != 0 {
		wineDTO = convertWineToDTO(&b.Wine)
	}

	return model.BottleWithWineDTO{
		ID:           int(b.ID),
		IsOpened:     b.IsOpened,
		AddedAt:      b.AddedAt,
		RemovedAt:    b.RemovedAt,
		RowNumber:    b.RowNumber,
		ColumnNumber: b.ColumnNumber,
		Note:         b.Note,
		Wine:         wineDTO,
	}
}

func (s *Service) toBottleWithWineDTOs(bottles []model.Bottle) []model.BottleWithWineDTO {
	result := make([]model.BottleWithWineDTO, 0, len(bottles))

	for i := range bottles {
		result = append(result, s.toBottleWithWineDTO(&bottles[i]))
	}

	return result
}
