package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cellar-app/dto"
	"cellar-app/model"
	"cellar-app/repository"

	"gorm.io/gorm"
)

var (
	ErrInvalidRegistration  = errors.New("invalid registration")
	ErrWineNotFound         = errors.New("wine was not found")
	ErrIdempotencyConflict  = repository.ErrIdempotencyConflict
	ErrCommitOutcomeUnknown = repository.ErrCommitOutcomeUnknown
)

func (s *Service) SearchWines(producer, name string, vintage *int, countryID *uint, limit int) ([]dto.WineDTO, error) {
	if strings.TrimSpace(producer) == "" && strings.TrimSpace(name) == "" && vintage == nil {
		return nil, ErrInvalidRegistration
	}
	if len([]rune(producer)) > 100 || len([]rune(name)) > 100 || limit < 1 || limit > 20 {
		return nil, ErrInvalidRegistration
	}
	if vintage != nil && (*vintage < 1800 || *vintage > time.Now().Year()) {
		return nil, ErrInvalidRegistration
	}
	if countryID != nil && (*countryID == 0 || *countryID > 2147483647) {
		return nil, ErrInvalidRegistration
	}
	wines, err := s.WineRepo.Search(repository.WineSearch{
		Producer:  producer,
		Name:      name,
		Vintage:   vintage,
		CountryID: countryID,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(wines))
	for _, wine := range wines {
		ids = append(ids, wine.ID)
	}
	counts, err := s.BottleRepo.CountByWineIDs(ids)
	if err != nil {
		return nil, err
	}
	result := make([]dto.WineDTO, 0, len(wines))
	for i := range wines {
		wine := convertWineToDTO(&wines[i])
		wine.StockCount = counts[wines[i].ID]
		wine.HasStock = wine.StockCount > 0
		result = append(result, wine)
	}
	return result, nil
}

func (s *Service) RegisterBottle(ctx context.Context, wineID uint, bottle model.Bottle, idempotencyKey, requestHash string) (model.Bottle, error) {
	if err := validateRegistrationIdempotency(idempotencyKey, requestHash); err != nil {
		return model.Bottle{}, err
	}
	if wineID == 0 || wineID > 2147483647 || bottle.RowNumber == nil || *bottle.RowNumber < 1 || *bottle.RowNumber > 10 ||
		bottle.ColumnNumber == nil || *bottle.ColumnNumber < 1 || *bottle.ColumnNumber > 2147483647 {
		return model.Bottle{}, ErrInvalidRegistration
	}
	if bottle.Note != nil && len([]rune(*bottle.Note)) > 2000 {
		return model.Bottle{}, ErrInvalidRegistration
	}
	if _, err := s.WineRepo.GetByID(wineID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Bottle{}, ErrWineNotFound
		}
		return model.Bottle{}, err
	}
	bottle.WineID = wineID
	opened := bottle.IsOpened != nil && *bottle.IsOpened
	bottle.IsOpened = &opened
	wine := model.Wine{ID: wineID}
	registered, _, err := s.createRegistration(ctx, &wine, &bottle, idempotencyKey, requestHash)
	return registered, err
}

func (s *Service) RegisterWineAndBottle(ctx context.Context, request dto.CreateWineWithBottleRequest, idempotencyKey, requestHash string) (model.Wine, model.Bottle, error) {
	if err := validateRegistrationIdempotency(idempotencyKey, requestHash); err != nil {
		return model.Wine{}, model.Bottle{}, err
	}
	wine := request.Wine.ToModel()
	bottle := request.Bottle.ToModel()
	if err := s.validateRegistrationWine(&wine); err != nil {
		return model.Wine{}, model.Bottle{}, err
	}
	if bottle.RowNumber == nil || *bottle.RowNumber < 1 || *bottle.RowNumber > 10 ||
		bottle.ColumnNumber == nil || *bottle.ColumnNumber < 1 || *bottle.ColumnNumber > 2147483647 {
		return model.Wine{}, model.Bottle{}, ErrInvalidRegistration
	}
	if bottle.Note != nil && len([]rune(*bottle.Note)) > 2000 {
		return model.Wine{}, model.Bottle{}, ErrInvalidRegistration
	}

	if idempotencyKey != "" {
		existing, err := s.WineRepo.FindMCPRegistration(idempotencyKey)
		if err == nil {
			if existing.RequestHash != requestHash {
				return model.Wine{}, model.Bottle{}, ErrIdempotencyConflict
			}
			return s.loadRegisteredWineAndBottle(existing)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Wine{}, model.Bottle{}, err
		}
	}

	registeredBottle, replayed, err := s.createRegistration(ctx, &wine, &bottle, idempotencyKey, requestHash)
	if err != nil {
		return model.Wine{}, model.Bottle{}, err
	}
	if replayed {
		return s.loadRegisteredWineAndBottle(&model.MCPRegistration{WineID: wine.ID, BottleID: registeredBottle.ID})
	}
	return wine, bottle, nil
}

func (s *Service) createRegistration(ctx context.Context, wine *model.Wine, bottle *model.Bottle, key, requestHash string) (model.Bottle, bool, error) {
	grapeNames := make([]string, 0, len(wine.WineGrapes))
	for _, grape := range wine.WineGrapes {
		grapeNames = append(grapeNames, grape.Grape.Name)
	}
	replayed, err := s.WineRepo.CreateMCPRegistration(ctx, wine, bottle, key, requestHash, grapeNames)
	if err != nil {
		if errors.Is(err, ErrIdempotencyConflict) {
			return model.Bottle{}, false, err
		}
		if IsUniqueViolation(err) {
			return model.Bottle{}, false, ErrPositionOccupied
		}
		return model.Bottle{}, false, err
	}
	if replayed {
		registeredBottle, err := s.BottleRepo.GetByID(bottle.ID)
		if err != nil {
			return model.Bottle{}, true, err
		}
		return *registeredBottle, true, nil
	}
	return *bottle, false, nil
}

func validateRegistrationIdempotency(key, requestHash string) error {
	if key == "" {
		return nil
	}
	if len(key) > 128 || strings.TrimSpace(key) != key || requestHash == "" {
		return ErrInvalidRegistration
	}
	for _, char := range key {
		if !(char >= 'a' && char <= 'z') && !(char >= 'A' && char <= 'Z') &&
			!(char >= '0' && char <= '9') && !strings.ContainsRune("._:-", char) {
			return ErrInvalidRegistration
		}
	}
	return nil
}

func (s *Service) loadRegisteredWineAndBottle(registration *model.MCPRegistration) (model.Wine, model.Bottle, error) {
	wine, err := s.WineRepo.GetByID(registration.WineID)
	if err != nil {
		return model.Wine{}, model.Bottle{}, err
	}
	bottle, err := s.BottleRepo.GetByID(registration.BottleID)
	if err != nil {
		return model.Wine{}, model.Bottle{}, err
	}
	return *wine, *bottle, nil
}

func (s *Service) validateRegistrationWine(wine *model.Wine) error {
	if strings.TrimSpace(wine.Name) == "" || len([]rune(wine.Name)) > 255 || wine.CountryID == 0 || wine.CountryID > 2147483647 ||
		wine.WineTypeID == 0 || wine.WineTypeID > 2147483647 {
		return ErrInvalidRegistration
	}
	if wine.Producer != nil && len([]rune(*wine.Producer)) > 255 {
		return ErrInvalidRegistration
	}
	grapes := make(map[string]struct{}, len(wine.WineGrapes))
	if len(wine.WineGrapes) > 12 {
		return ErrInvalidRegistration
	}
	for _, grape := range wine.WineGrapes {
		name := strings.TrimSpace(grape.Grape.Name)
		key := strings.ToLower(name)
		if name == "" || len([]rune(name)) > 80 {
			return ErrInvalidRegistration
		}
		if _, exists := grapes[key]; exists {
			return ErrInvalidRegistration
		}
		grapes[key] = struct{}{}
	}
	if wine.Vintage != nil && (*wine.Vintage < 1800 || *wine.Vintage > time.Now().Year()) {
		return ErrInvalidRegistration
	}
	if wine.DrinkingWindowStart != nil && wine.DrinkingWindowEnd != nil && *wine.DrinkingWindowStart > *wine.DrinkingWindowEnd {
		return ErrInvalidRegistration
	}
	for _, windowYear := range []*int{wine.DrinkingWindowStart, wine.DrinkingWindowEnd} {
		if windowYear != nil && (*windowYear < 1800 || *windowYear > 2200) {
			return ErrInvalidRegistration
		}
	}
	country, err := s.CountryRepo.GetByID(wine.CountryID)
	if err != nil {
		return fmt.Errorf("%w: country_id", ErrInvalidRegistration)
	}
	if _, err := s.WineTypeRepo.GetByID(wine.WineTypeID); err != nil {
		return fmt.Errorf("%w: wine_type_id", ErrInvalidRegistration)
	}
	if wine.RegionID != nil {
		if *wine.RegionID == 0 || *wine.RegionID > 2147483647 {
			return fmt.Errorf("%w: region_id", ErrInvalidRegistration)
		}
		region, err := s.RegionRepo.GetByID(*wine.RegionID)
		if err != nil {
			return fmt.Errorf("%w: region_id", ErrInvalidRegistration)
		}
		if region.CountryID != country.ID {
			return fmt.Errorf("%w: region_id", ErrInvalidRegistration)
		}
	}
	if wine.AppellationID != nil {
		if *wine.AppellationID == 0 || *wine.AppellationID > 2147483647 {
			return fmt.Errorf("%w: appellation_id", ErrInvalidRegistration)
		}
		appellation, err := s.AppellationRepo.GetByID(*wine.AppellationID)
		if err != nil {
			return fmt.Errorf("%w: appellation_id", ErrInvalidRegistration)
		}
		if appellation.CountryID != nil && *appellation.CountryID != country.ID {
			return fmt.Errorf("%w: appellation_id", ErrInvalidRegistration)
		}
	}
	return nil
}
