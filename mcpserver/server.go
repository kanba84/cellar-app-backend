package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"cellar-app/dto"
	"cellar-app/model"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
)

const (
	maxRecordID = uint(2147483647)
	maxDays     = 3650
	defaultDays = 30
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

func NewHTTPHandler(reader ReadOnlyService) http.Handler {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "cellar-app-backend",
		Version: "1.0.0",
	}, nil)
	registerTools(server, reader)

	return mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server },
		nil,
	)
}

func registerTools(server *mcp.Server, reader ReadOnlyService) {
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
	return &mcp.Tool{
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: boolPointer(false)},
	}
}

func safeReadError(err error) error {
	log.Printf("MCP read operation failed: %v", err)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("No matching record was found.")
	}
	return errors.New("The requested data could not be retrieved.")
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
