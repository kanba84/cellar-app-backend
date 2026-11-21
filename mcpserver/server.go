package mcpserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"cellar-app/dto"
	"cellar-app/model"
	"cellar-app/service"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
)

const (
	maxRecordID            = uint(2147483647)
	maxDays                = 3650
	defaultDays            = 30
	maxMCPRequestBodyBytes = 8 * 1024 * 1024
)

type ReadOnlyService interface {
	ListWines() ([]dto.WineDTO, error)
	GetWine(uint) (*dto.WineDTO, error)
	ListBottles() ([]dto.BottleWithWineDTO, error)
	GetBottle(uint) (*dto.BottleWithWineDTO, error)
	ListCountries() ([]model.Country, error)
	GetCountry(uint) (*model.Country, error)
	ListRegions() ([]model.Region, error)
	GetRegion(uint) (*model.Region, error)
	ListWineTypes() ([]model.WineType, error)
	GetWineType(uint) (*model.WineType, error)
	ListAppellations() ([]model.Appellation, error)
	GetAppellation(uint) (*model.Appellation, error)
	ListDesignationTypes() ([]model.DesignationType, error)
	GetDesignationType(uint) (*model.DesignationType, error)
	GetStats(int) (*dto.Stats, error)
	GetWineTypeStats() ([]dto.WineTypeStats, error)
	GetCountryStats() ([]dto.CountryStats, error)
	GetVintageStats() ([]dto.VintageStats, error)
	GetInventoryTrend(int) ([]dto.InventoryTrendDataPoint, error)
	GetDailySnapshot(time.Time) (*model.InventoryDailySnapshot, error)
}

type WriteService interface {
	SearchWines(string, string, *int, *uint, int) ([]dto.WineDTO, error)
	RegisterBottle(context.Context, uint, model.Bottle, string, string) (model.Bottle, error)
	RegisterWineAndBottle(context.Context, dto.CreateWineWithBottleRequest, string, string) (model.Wine, model.Bottle, error)
}

type emptyInput struct{}

type idInput struct {
	ID uint `json:"id" jsonschema:"positive database record ID"`
}

type daysInput struct {
	Days *int `json:"days,omitempty" jsonschema:"number of days to include; 0 uses the 30-day default"`
}

type dateInput struct {
	Date string `json:"date" jsonschema:"snapshot date in YYYY-MM-DD format"`
}

type searchWinesInput struct {
	Producer  string `json:"producer,omitempty"`
	Name      string `json:"name,omitempty"`
	Vintage   *int   `json:"vintage,omitempty"`
	CountryID *uint  `json:"country_id,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

type registerBottleInput struct {
	WineID         uint   `json:"wine_id"`
	RowNumber      int    `json:"row_number"`
	ColumnNumber   int    `json:"column_number"`
	IsOpened       bool   `json:"is_opened,omitempty"`
	Note           string `json:"note,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type registrationWineInput struct {
	Name                string   `json:"name"`
	CountryID           uint     `json:"country_id"`
	WineTypeID          uint     `json:"wine_type_id"`
	Vintage             *int     `json:"vintage,omitempty"`
	Producer            string   `json:"producer,omitempty"`
	RegionID            *uint    `json:"region_id,omitempty"`
	AppellationID       *uint    `json:"appellation_id,omitempty"`
	DrinkingWindowStart *int     `json:"drinking_window_start,omitempty"`
	DrinkingWindowEnd   *int     `json:"drinking_window_end,omitempty"`
	GrapeNames          []string `json:"grape_names,omitempty"`
}

type registrationBottleInput struct {
	RowNumber    int    `json:"row_number"`
	ColumnNumber int    `json:"column_number"`
	IsOpened     bool   `json:"is_opened,omitempty"`
	Note         string `json:"note,omitempty"`
}

type registerWineAndBottleInput struct {
	Wine           registrationWineInput   `json:"wine"`
	Bottle         registrationBottleInput `json:"bottle"`
	IdempotencyKey string                  `json:"idempotency_key,omitempty"`
}

func NewHTTPHandler(reader ReadOnlyService, writers ...WriteService) http.Handler {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "cellar-app-backend",
		Version: "1.0.0",
	}, nil)
	registerTools(server, reader, writers...)

	return mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server },
		// Cloudflare Tunnel connects to localhost while preserving the public
		// Host header. Cloudflare Access JWT validation is enforced separately
		// by the /mcp route middleware.
		&mcp.StreamableHTTPOptions{
			DisableLocalhostProtection: true,
			MaxRequestBodyBytes:        maxMCPRequestBodyBytes,
		},
	)
}

func registerTools(server *mcp.Server, reader ReadOnlyService, writers ...WriteService) {
	addListTool(server, "list_wines", "List every wine with its current stock count. Returns all records; no pagination or truncation is applied.", nonNilCall(reader.ListWines))
	addIDTool(server, "get_wine", "Get one wine and its current stock count by positive database ID.", reader.GetWine)
	addListTool(server, "list_bottles", "List every bottle with its associated wine. Returns all records; no pagination or truncation is applied.", nonNilCall(reader.ListBottles))
	addIDTool(server, "get_bottle", "Get one bottle with its associated wine by positive database ID.", reader.GetBottle)

	addListTool(server, "list_countries", "List every country. Returns all records; no pagination or truncation is applied.", func() ([]dto.CountryDTO, error) {
		countries, err := reader.ListCountries()
		if err != nil {
			return nil, err
		}
		result := make([]dto.CountryDTO, 0, len(countries))
		for _, country := range countries {
			result = append(result, dto.CountryFromModel(country))
		}
		return result, nil
	})
	addIDTool(server, "get_country", "Get one country by positive database ID.", func(id uint) (*dto.CountryDTO, error) {
		country, err := reader.GetCountry(id)
		if err != nil {
			return nil, err
		}
		result := dto.CountryFromModel(*country)
		return &result, nil
	})

	addListTool(server, "list_regions", "List every region with its country. Returns all records; no pagination or truncation is applied.", func() ([]dto.RegionDTO, error) {
		regions, err := reader.ListRegions()
		if err != nil {
			return nil, err
		}
		result := make([]dto.RegionDTO, 0, len(regions))
		for _, region := range regions {
			result = append(result, dto.RegionFromModel(region))
		}
		return result, nil
	})
	addIDTool(server, "get_region", "Get one region with its country by positive database ID.", func(id uint) (*dto.RegionDTO, error) {
		region, err := reader.GetRegion(id)
		if err != nil {
			return nil, err
		}
		result := dto.RegionFromModel(*region)
		return &result, nil
	})

	addListTool(server, "list_wine_types", "List every wine type. Returns all records; no pagination or truncation is applied.", func() ([]dto.WineTypeDTO, error) {
		wineTypes, err := reader.ListWineTypes()
		if err != nil {
			return nil, err
		}
		result := make([]dto.WineTypeDTO, 0, len(wineTypes))
		for _, wineType := range wineTypes {
			result = append(result, dto.WineTypeFromModel(wineType))
		}
		return result, nil
	})
	addIDTool(server, "get_wine_type", "Get one wine type by positive database ID.", func(id uint) (*dto.WineTypeDTO, error) {
		wineType, err := reader.GetWineType(id)
		if err != nil {
			return nil, err
		}
		result := dto.WineTypeFromModel(*wineType)
		return &result, nil
	})

	addListTool(server, "list_appellations", "List every appellation with its existing related country, designation type, and regions. Returns all records; no pagination or truncation is applied.", func() ([]dto.AppellationDTO, error) {
		appellations, err := reader.ListAppellations()
		if err != nil {
			return nil, err
		}
		result := make([]dto.AppellationDTO, 0, len(appellations))
		for _, appellation := range appellations {
			result = append(result, dto.AppellationFromModel(appellation))
		}
		return result, nil
	})
	addIDTool(server, "get_appellation", "Get one appellation with its existing related country, designation type, and regions by positive database ID.", func(id uint) (*dto.AppellationDTO, error) {
		appellation, err := reader.GetAppellation(id)
		if err != nil {
			return nil, err
		}
		result := dto.AppellationFromModel(*appellation)
		return &result, nil
	})

	addListTool(server, "list_designation_types", "List every designation type. Returns all records; no pagination or truncation is applied.", func() ([]dto.DesignationTypeDTO, error) {
		designationTypes, err := reader.ListDesignationTypes()
		if err != nil {
			return nil, err
		}
		result := make([]dto.DesignationTypeDTO, 0, len(designationTypes))
		for _, designationType := range designationTypes {
			result = append(result, dto.DesignationTypeFromModel(designationType))
		}
		return result, nil
	})
	addIDTool(server, "get_designation_type", "Get one designation type by positive database ID.", func(id uint) (*dto.DesignationTypeDTO, error) {
		designationType, err := reader.GetDesignationType(id)
		if err != nil {
			return nil, err
		}
		result := dto.DesignationTypeFromModel(*designationType)
		return &result, nil
	})

	addDaysTool(server, "get_stats", "Get inventory statistics and a daily inventory trend. days may be 0 (the existing 30-day default) or 1 through 3650.", func(days int) (*dto.Stats, error) {
		stats, err := reader.GetStats(days)
		if err != nil {
			return nil, err
		}
		stats.WineTypes = nonNil(stats.WineTypes)
		stats.Countries = nonNil(stats.Countries)
		stats.Vintages = nonNil(stats.Vintages)
		stats.InventoryTrend = nonNil(stats.InventoryTrend)
		return stats, nil
	})
	addListTool(server, "get_wine_type_stats", "Get the current inventory count grouped by wine type.", nonNilCall(reader.GetWineTypeStats))
	addListTool(server, "get_country_stats", "Get the current inventory count grouped by country.", nonNilCall(reader.GetCountryStats))
	addListTool(server, "get_vintage_stats", "Get the current inventory count grouped by vintage; wines without a vintage are excluded.", nonNilCall(reader.GetVintageStats))
	addDaysTool(server, "get_inventory_trend", "Get the daily inventory trend. days may be 0 (the existing 30-day default) or 1 through 3650.", nonNilCallWithArg(reader.GetInventoryTrend))
	addDateTool(server, "get_daily_snapshot", "Get the saved daily inventory snapshot for a calendar date in YYYY-MM-DD format.", func(date time.Time) (*dto.InventoryDailySnapshotDTO, error) {
		snapshot, err := reader.GetDailySnapshot(date)
		if err != nil {
			return nil, err
		}
		if snapshot == nil {
			return nil, errors.New("No daily snapshot exists for the requested date.")
		}
		result := dto.SnapshotFromModel(*snapshot)
		return &result, nil
	})

	if len(writers) == 0 || writers[0] == nil {
		return
	}
	writer := writers[0]
	mcp.AddTool(server, tool("search_wines", "Searches existing cellar wines using database facts. Provide at least one of producer, name, or vintage. Results are partial-match candidates, not identity confirmation; always compare them and ask the user when ambiguous.", searchWinesObjectSchema), func(_ context.Context, _ *mcp.CallToolRequest, input searchWinesInput) (*mcp.CallToolResult, []dto.WineDTO, error) {
		if err := validateWineSearch(input); err != nil {
			return nil, nil, err
		}
		limit := input.Limit
		if limit == 0 {
			limit = 10
		}
		result, err := writer.SearchWines(input.Producer, input.Name, input.Vintage, input.CountryID, limit)
		if err != nil {
			return nil, nil, safeReadError(err)
		}
		return nil, nonNil(result), nil
	})
	mcp.AddTool(server, toolWithHints("register_bottle", "Adds one physical bottle to an existing wine. This changes cellar inventory. Call only once after presenting the exact wine and bottle/shelf details and receiving the user's explicit affirmative confirmation; this tool cannot verify that confirmation.", registerBottleObjectSchema, false), func(ctx context.Context, _ *mcp.CallToolRequest, input registerBottleInput) (*mcp.CallToolResult, map[string]any, error) {
		if err := validateRegisterBottle(input); err != nil {
			return nil, nil, err
		}
		hash, err := requestHash(input)
		if err != nil {
			return nil, nil, safeWriteError(err)
		}
		bottle := model.Bottle{
			RowNumber:    &input.RowNumber,
			ColumnNumber: &input.ColumnNumber,
			Note:         optionalString(input.Note),
		}
		opened := input.IsOpened
		bottle.IsOpened = &opened
		created, err := writer.RegisterBottle(ctx, input.WineID, bottle, input.IdempotencyKey, hash)
		if err != nil {
			return nil, nil, safeWriteError(err)
		}
		return nil, map[string]any{
			"id":            created.ID,
			"wine_id":       created.WineID,
			"is_opened":     created.IsOpened != nil && *created.IsOpened,
			"added_at":      created.AddedAt,
			"row_number":    created.RowNumber,
			"column_number": created.ColumnNumber,
			"note":          created.Note,
		}, nil
	})
	mcp.AddTool(server, toolWithHints("register_wine_and_bottle", "Atomically creates a new wine and its first physical bottle without sending or storing a label image. This changes cellar records. Call only once after presenting all wine/bottle details and receiving the user's explicit affirmative confirmation; this tool cannot verify that confirmation. Search the cellar first; do not use for ambiguous matches.", registerWineAndBottleObjectSchema, false), func(ctx context.Context, _ *mcp.CallToolRequest, input registerWineAndBottleInput) (*mcp.CallToolResult, map[string]any, error) {
		if err := validateRegisterWineAndBottle(input); err != nil {
			return nil, nil, err
		}
		hash, err := requestHash(input)
		if err != nil {
			return nil, nil, safeWriteError(err)
		}
		grapes := make([]dto.WineGrapeInput, 0, len(input.Wine.GrapeNames))
		for order, name := range input.Wine.GrapeNames {
			grapes = append(grapes, dto.WineGrapeInput{
				DisplayOrder: order,
				Grape:        dto.GrapeDTO{Name: name},
			})
		}
		request := dto.CreateWineWithBottleRequest{
			Wine: dto.WineRequest{
				Name:                input.Wine.Name,
				CountryID:           input.Wine.CountryID,
				WineTypeID:          input.Wine.WineTypeID,
				Vintage:             input.Wine.Vintage,
				Producer:            optionalString(input.Wine.Producer),
				RegionID:            input.Wine.RegionID,
				AppellationID:       input.Wine.AppellationID,
				DrinkingWindowStart: input.Wine.DrinkingWindowStart,
				DrinkingWindowEnd:   input.Wine.DrinkingWindowEnd,
				WineGrapes:          grapes,
			},
			Bottle: dto.BottleRequest{
				RowNumber:    &input.Bottle.RowNumber,
				ColumnNumber: &input.Bottle.ColumnNumber,
				IsOpened:     input.Bottle.IsOpened,
				Note:         optionalString(input.Bottle.Note),
			},
		}
		wine, bottle, err := writer.RegisterWineAndBottle(ctx, request, input.IdempotencyKey, hash)
		if err != nil {
			return nil, nil, safeWriteError(err)
		}
		return nil, map[string]any{"wine": dto.WineFromModel(wine), "bottle": dto.BottleFromModel(bottle)}, nil
	})
}

func addListTool[Out any](server *mcp.Server, name, description string, call func() (Out, error)) {
	mcp.AddTool(server, tool(name, description, emptyObjectSchema), func(context.Context, *mcp.CallToolRequest, emptyInput) (*mcp.CallToolResult, Out, error) {
		output, err := call()
		if err != nil {
			var zero Out
			return nil, zero, safeReadError(err)
		}
		return nil, output, nil
	})
}

func addIDTool[Out any](server *mcp.Server, name, description string, call func(uint) (Out, error)) {
	mcp.AddTool(server, tool(name, description, idObjectSchema), func(_ context.Context, _ *mcp.CallToolRequest, input idInput) (*mcp.CallToolResult, Out, error) {
		if input.ID == 0 || input.ID > maxRecordID {
			var zero Out
			return nil, zero, fmt.Errorf("id must be an integer from 1 to %d", maxRecordID)
		}
		output, err := call(input.ID)
		if err != nil {
			var zero Out
			return nil, zero, safeReadError(err)
		}
		return nil, output, nil
	})
}

func addDaysTool[Out any](server *mcp.Server, name, description string, call func(int) (Out, error)) {
	mcp.AddTool(server, tool(name, description, daysObjectSchema), func(_ context.Context, _ *mcp.CallToolRequest, input daysInput) (*mcp.CallToolResult, Out, error) {
		days := defaultDays
		if input.Days != nil {
			if *input.Days < 0 || *input.Days > maxDays {
				var zero Out
				return nil, zero, fmt.Errorf("days must be 0 to %d; 0 uses the 30-day default", maxDays)
			}
			if *input.Days != 0 {
				days = *input.Days
			}
		}
		output, err := call(days)
		if err != nil {
			var zero Out
			return nil, zero, safeReadError(err)
		}
		return nil, output, nil
	})
}

func addDateTool[Out any](server *mcp.Server, name, description string, call func(time.Time) (Out, error)) {
	mcp.AddTool(server, tool(name, description, dateObjectSchema), func(_ context.Context, _ *mcp.CallToolRequest, input dateInput) (*mcp.CallToolResult, Out, error) {
		date, err := time.Parse("2006-01-02", input.Date)
		if err != nil {
			var zero Out
			return nil, zero, errors.New("date must be a valid calendar date in YYYY-MM-DD format")
		}
		output, err := call(date)
		if err != nil {
			var zero Out
			return nil, zero, safeReadError(err)
		}
		return nil, output, nil
	})
}

func tool(name, description string, inputSchema map[string]any) *mcp.Tool {
	return toolWithHints(name, description, inputSchema, true)
}

func toolWithHints(name, description string, inputSchema map[string]any, readOnly bool) *mcp.Tool {
	return &mcp.Tool{
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: readOnly, DestructiveHint: boolPointer(false), OpenWorldHint: boolPointer(false)},
	}
}

func safeReadError(err error) error {
	log.Printf("MCP read operation failed: %v", err)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("No matching record was found.")
	}
	if errors.Is(err, service.ErrInvalidRegistration) {
		return errors.New("Invalid wine search input.")
	}
	return errors.New("The requested data could not be retrieved.")
}

func safeWriteError(err error) error {
	log.Printf("MCP registration operation failed: %v", err)
	switch {
	case errors.Is(err, service.ErrInvalidRegistration):
		return errors.New("Invalid input. Check required fields, master-data IDs, vintage, and shelf position.")
	case errors.Is(err, service.ErrWineNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		return errors.New("No matching wine was found; search the cellar again before proceeding.")
	case errors.Is(err, service.ErrPositionOccupied):
		return errors.New("Shelf position conflict. Choose an available position and obtain confirmation again.")
	case errors.Is(err, service.ErrIdempotencyConflict):
		return errors.New("This idempotency key was already used for different content. Search the cellar and ask the user before retrying.")
	default:
		return errors.New("Temporary server failure or unknown outcome. Do not retry blindly; search the cellar first.")
	}
}

func validateWineSearch(input searchWinesInput) error {
	if len([]rune(input.Producer)) > 100 || len([]rune(input.Name)) > 100 ||
		input.Limit < 0 || input.Limit > 20 {
		return errors.New("Invalid search input: text is limited to 100 characters and limit to 1 through 20.")
	}
	if strings.TrimSpace(input.Producer) == "" && strings.TrimSpace(input.Name) == "" && input.Vintage == nil {
		return errors.New("Provide at least one non-empty producer, name, or vintage search condition.")
	}
	if input.Vintage != nil && (*input.Vintage < 1800 || *input.Vintage > time.Now().Year()) {
		return errors.New("Vintage must be between 1800 and the current year.")
	}
	if input.CountryID != nil && (*input.CountryID == 0 || *input.CountryID > maxRecordID) {
		return errors.New("country_id must be a positive database ID.")
	}
	return nil
}

func validateRegisterBottle(input registerBottleInput) error {
	if input.WineID == 0 || input.WineID > maxRecordID || input.RowNumber < 1 || input.RowNumber > 10 ||
		input.ColumnNumber < 1 || input.ColumnNumber > 2147483647 || len([]rune(input.Note)) > 2000 {
		return errors.New("Invalid bottle input. Use a valid wine ID, row 1-10, positive column, and note up to 2000 characters.")
	}
	return validateIdempotencyKey(input.IdempotencyKey)
}

func validateRegisterWineAndBottle(input registerWineAndBottleInput) error {
	wine := input.Wine
	bottle := input.Bottle
	if strings.TrimSpace(wine.Name) == "" || len([]rune(wine.Name)) > 255 || wine.CountryID == 0 || wine.CountryID > maxRecordID ||
		wine.WineTypeID == 0 || wine.WineTypeID > maxRecordID || len([]rune(wine.Producer)) > 255 ||
		(wine.Vintage != nil && (*wine.Vintage < 1800 || *wine.Vintage > time.Now().Year())) ||
		(wine.RegionID != nil && (*wine.RegionID == 0 || *wine.RegionID > maxRecordID)) ||
		(wine.AppellationID != nil && (*wine.AppellationID == 0 || *wine.AppellationID > maxRecordID)) ||
		(bottle.RowNumber < 1 || bottle.RowNumber > 10) ||
		(bottle.ColumnNumber < 1 || bottle.ColumnNumber > 2147483647) ||
		len([]rune(bottle.Note)) > 2000 {
		return errors.New("Invalid wine or bottle input. Check required IDs, field lengths, vintage, and shelf position.")
	}
	if wine.DrinkingWindowStart != nil && wine.DrinkingWindowEnd != nil && *wine.DrinkingWindowStart > *wine.DrinkingWindowEnd {
		return errors.New("drinking_window_start must not be after drinking_window_end.")
	}
	if len(wine.GrapeNames) > 12 {
		return errors.New("At most 12 grape names may be supplied.")
	}
	grapes := make(map[string]struct{}, len(wine.GrapeNames))
	for _, name := range wine.GrapeNames {
		normalized := strings.ToLower(strings.TrimSpace(name))
		if normalized == "" || len([]rune(name)) > 80 {
			return errors.New("Grape names must be non-empty and at most 80 characters.")
		}
		if _, exists := grapes[normalized]; exists {
			return errors.New("Grape names must not contain duplicates.")
		}
		grapes[normalized] = struct{}{}
	}
	return validateIdempotencyKey(input.IdempotencyKey)
}

func validateIdempotencyKey(key string) error {
	if len(key) > 128 || strings.TrimSpace(key) != key || strings.ContainsAny(key, "\r\n\t") {
		return errors.New("idempotency_key must be at most 128 characters without surrounding whitespace.")
	}
	return nil
}

func requestHash(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func nonNil[T any](values []T) []T {
	if values == nil {
		return []T{}
	}
	return values
}

func nonNilCall[T any](call func() ([]T, error)) func() ([]T, error) {
	return func() ([]T, error) {
		values, err := call()
		if err != nil {
			return nil, err
		}
		return nonNil(values), nil
	}
}

func nonNilCallWithArg[In any, Out any](call func(In) ([]Out, error)) func(In) ([]Out, error) {
	return func(input In) ([]Out, error) {
		values, err := call(input)
		if err != nil {
			return nil, err
		}
		return nonNil(values), nil
	}
}

func boolPointer(value bool) *bool {
	return &value
}

var (
	emptyObjectSchema = objectSchema(nil, nil)
	idObjectSchema    = objectSchema(map[string]any{
		"id": map[string]any{"type": "integer", "minimum": 1, "maximum": maxRecordID, "description": "Positive database record ID."},
	}, []string{"id"})
	daysObjectSchema = objectSchema(map[string]any{
		"days": map[string]any{"type": "integer", "minimum": 0, "maximum": maxDays, "description": "0 uses the existing 30-day default."},
	}, nil)
	dateObjectSchema = objectSchema(map[string]any{
		"date": map[string]any{"type": "string", "format": "date", "pattern": "^\\d{4}-\\d{2}-\\d{2}$", "description": "Calendar date in YYYY-MM-DD format."},
	}, []string{"date"})
	searchWinesObjectSchema = objectSchema(map[string]any{
		"producer":   map[string]any{"type": "string", "maxLength": 100},
		"name":       map[string]any{"type": "string", "maxLength": 100},
		"vintage":    map[string]any{"type": "integer", "minimum": 1800, "maximum": time.Now().Year()},
		"country_id": map[string]any{"type": "integer", "minimum": 1, "maximum": maxRecordID},
		"limit":      map[string]any{"type": "integer", "minimum": 1, "maximum": 20, "description": "Maximum candidates; defaults to 10."},
	}, nil)
	registerBottleObjectSchema = objectSchema(map[string]any{
		"wine_id":         map[string]any{"type": "integer", "minimum": 1, "maximum": maxRecordID},
		"row_number":      map[string]any{"type": "integer", "minimum": 1, "maximum": 10},
		"column_number":   map[string]any{"type": "integer", "minimum": 1, "maximum": 2147483647},
		"is_opened":       map[string]any{"type": "boolean"},
		"note":            map[string]any{"type": "string", "maxLength": 2000},
		"idempotency_key": idempotencyKeySchema,
	}, []string{"wine_id", "row_number", "column_number"})
	registerWineAndBottleObjectSchema = objectSchema(map[string]any{
		"wine": objectSchema(map[string]any{
			"name":                  map[string]any{"type": "string", "minLength": 1, "maxLength": 255},
			"country_id":            map[string]any{"type": "integer", "minimum": 1, "maximum": maxRecordID},
			"wine_type_id":          map[string]any{"type": "integer", "minimum": 1, "maximum": maxRecordID},
			"vintage":               map[string]any{"type": "integer", "minimum": 1800, "maximum": time.Now().Year()},
			"producer":              map[string]any{"type": "string", "maxLength": 255},
			"region_id":             map[string]any{"type": "integer", "minimum": 1, "maximum": maxRecordID},
			"appellation_id":        map[string]any{"type": "integer", "minimum": 1, "maximum": maxRecordID},
			"drinking_window_start": map[string]any{"type": "integer", "minimum": 1800, "maximum": 2200},
			"drinking_window_end":   map[string]any{"type": "integer", "minimum": 1800, "maximum": 2200},
			"grape_names":           map[string]any{"type": "array", "maxItems": 12, "items": map[string]any{"type": "string", "minLength": 1, "maxLength": 80}},
		}, []string{"name", "country_id", "wine_type_id"}),
		"bottle": objectSchema(map[string]any{
			"row_number":    map[string]any{"type": "integer", "minimum": 1, "maximum": 10},
			"column_number": map[string]any{"type": "integer", "minimum": 1, "maximum": 2147483647},
			"is_opened":     map[string]any{"type": "boolean"},
			"note":          map[string]any{"type": "string", "maxLength": 2000},
		}, []string{"row_number", "column_number"}),
		"idempotency_key": idempotencyKeySchema,
	}, []string{"wine", "bottle"})
	idempotencyKeySchema = map[string]any{"type": "string", "maxLength": 128, "pattern": "^[A-Za-z0-9._:-]*$"}
)

func objectSchema(properties map[string]any, required []string) map[string]any {
	if properties == nil {
		properties = map[string]any{}
	}
	schema := map[string]any{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}
