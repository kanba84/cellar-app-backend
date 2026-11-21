package repository

import (
	"cellar-app/model"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrIdempotencyConflict  = errors.New("idempotency key was already used for a different request")
	ErrCommitOutcomeUnknown = errors.New("registration commit outcome is unknown")
)

type WineSearch struct {
	Producer  string
	Name      string
	Vintage   *int
	CountryID *uint
	Limit     int
}

type WineRepository struct {
	db *gorm.DB
}

func NewWineRepository(db *gorm.DB) *WineRepository {
	return &WineRepository{db: db}
}

func (r *WineRepository) List() ([]model.Wine, error) {
	var wines []model.Wine

	err := r.db.
		Preload("WineType").
		Preload("Country").
		Preload("Region").
		Preload("Appellation").
		Preload("Appellation.DesignationType").
		Find(&wines).Error

	return wines, err
}

func (r *WineRepository) Search(query WineSearch) ([]model.Wine, error) {
	var wines []model.Wine
	err := buildWineSearchQuery(r.db, query).Find(&wines).Error
	return wines, err
}

func buildWineSearchQuery(db *gorm.DB, query WineSearch) *gorm.DB {
	db = db.Model(&model.Wine{}).
		Preload("WineType").
		Preload("Country").
		Preload("Region").
		Order("id").
		Limit(query.Limit)
	if query.Producer != "" {
		db = db.Where("LOWER(COALESCE(producer, '')) LIKE ?", "%"+escapeLike(strings.ToLower(query.Producer))+"%")
	}
	if query.Name != "" {
		db = db.Where("LOWER(name) LIKE ?", "%"+escapeLike(strings.ToLower(query.Name))+"%")
	}
	if query.Vintage != nil {
		db = db.Where("vintage = ?", *query.Vintage)
	}
	if query.CountryID != nil {
		db = db.Where("country_id = ?", *query.CountryID)
	}
	return db
}

func escapeLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `%`, `\%`)
	return strings.ReplaceAll(value, `_`, `\_`)
}

func (r *WineRepository) GetByID(id uint) (*model.Wine, error) {
	var wine model.Wine

	err := r.db.
		Preload("WineType").
		Preload("Country").
		Preload("Region").
		Preload("Appellation").
		Preload("Appellation.DesignationType").
		Preload("WineGrapes").
		Preload("WineGrapes.Grape").
		First(&wine, id).Error

	if err != nil {
		return nil, err
	}

	return &wine, nil
}

func (r *WineRepository) Create(wine *model.Wine) error {
	return r.db.Create(wine).Error
}

func (r *WineRepository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

func (r *WineRepository) Update(wine *model.Wine) error {
	return r.db.Save(wine).Error
}

func (r *WineRepository) Patch(id uint, updates map[string]interface{}) error {
	return r.db.Model(&model.Wine{}).Where("id = ?", id).Updates(updates).Error
}

func (r *WineRepository) Delete(id uint) error {
	return r.db.Delete(&model.Wine{}, id).Error
}

func (r *WineRepository) CreateWithBottle(ctx context.Context, wine *model.Wine, bottle *model.Bottle) (*model.Wine, *model.Bottle, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return createWineAndBottleTx(tx, wine, bottle, nil, false)
	})
	if err != nil {
		return nil, nil, err
	}

	return wine, bottle, nil
}

func (r *WineRepository) FindMCPRegistration(key string) (*model.MCPRegistration, error) {
	var registration model.MCPRegistration
	if err := r.db.First(&registration, "idempotency_key = ?", key).Error; err != nil {
		return nil, err
	}
	return &registration, nil
}

func (r *WineRepository) CreateMCPRegistration(ctx context.Context, wine *model.Wine, bottle *model.Bottle, key, requestHash string, grapeNames []string) (bool, error) {
	transactionBodySucceeded := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if key != "" {
			var existing model.MCPRegistration
			err := tx.First(&existing, "idempotency_key = ?", key).Error
			if err == nil {
				if existing.RequestHash != requestHash {
					return ErrIdempotencyConflict
				}
				wine.ID = existing.WineID
				bottle.ID = existing.BottleID
				bottle.WineID = existing.WineID
				return errRegistrationExists
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		if wine.ID == 0 {
			if err := createWineAndBottleTx(tx, wine, bottle, grapeNames, true); err != nil {
				return err
			}
		} else if err := createWineAndBottleTx(tx, wine, bottle, nil, false); err != nil {
			return err
		}
		if key != "" {
			if err := tx.Create(&model.MCPRegistration{
				IdempotencyKey: key,
				RequestHash:    requestHash,
				WineID:         wine.ID,
				BottleID:       bottle.ID,
			}).Error; err != nil {
				return err
			}
		}
		transactionBodySucceeded = true
		return nil
	})
	if errors.Is(err, errRegistrationExists) {
		return true, nil
	}
	if err != nil && key != "" && IsUniqueViolation(err) {
		existing, lookupErr := r.FindMCPRegistration(key)
		if lookupErr == nil {
			if existing.RequestHash != requestHash {
				return false, ErrIdempotencyConflict
			}
			wine.ID = existing.WineID
			bottle.ID = existing.BottleID
			bottle.WineID = existing.WineID
			return true, nil
		}
	}
	if err != nil && transactionBodySucceeded {
		return false, fmt.Errorf("%w: %v", ErrCommitOutcomeUnknown, err)
	}
	return false, err
}

var errRegistrationExists = errors.New("registration already exists")

func createWineAndBottleTx(tx *gorm.DB, wine *model.Wine, bottle *model.Bottle, grapeNames []string, omitWineGrapes bool) error {
	createWine := tx
	if omitWineGrapes {
		createWine = createWine.Omit("WineGrapes")
	}
	if wine.ID == 0 {
		if err := createWine.Create(wine).Error; err != nil {
			return err
		}
	}
	if len(grapeNames) > 0 {
		createdWineGrapes := make([]model.WineGrape, 0, len(grapeNames))
		for order, name := range grapeNames {
			var grape model.Grape
			err := tx.Where("name = ?", name).First(&grape).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				grape.Name = name
				if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&grape).Error; err != nil {
					return err
				}
				if err := tx.Where("name = ?", name).First(&grape).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
			displayOrder := order
			if err := tx.Create(&model.WineGrape{
				WineID:       wine.ID,
				GrapeID:      grape.ID,
				DisplayOrder: &displayOrder,
			}).Error; err != nil {
				return err
			}
			createdWineGrapes = append(createdWineGrapes, model.WineGrape{
				WineID:       wine.ID,
				GrapeID:      grape.ID,
				Grape:        grape,
				DisplayOrder: &displayOrder,
			})
		}
		wine.WineGrapes = createdWineGrapes
	}
	bottle.WineID = wine.ID
	return tx.Create(bottle).Error
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
