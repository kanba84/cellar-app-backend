package llm

import (
	"context"
	"fmt"
	"strings"
)

// WineLookupKey identifies a wine for LLM lookup and future cache keys.
type WineLookupKey struct {
	Name              string
	Vintage           *int
	Producer          *string
	CountryName       *string
	RegionName        *string
	IncludeTastingNote bool
}

// CacheKey returns a stable key for caching (wine name + optional vintage).
// Tasting note inclusion is not part of the cache key because it is an optional output flag.
func (k WineLookupKey) CacheKey() string {
	name := strings.TrimSpace(strings.ToLower(k.Name))
	if k.Vintage != nil {
		return fmt.Sprintf("%s|%d", name, *k.Vintage)
	}
	return name
}

// WineInfoResult is structured wine information returned by an LLM provider.
type WineInfoResult struct {
	Producer           *string     `json:"producer"`
	Grapes             []GrapeInfo `json:"grapes"`
	TastingNote        *string     `json:"tasting_note"`
	ReferencePriceJPY  *float64    `json:"reference_price_jpy"`
}

// GrapeInfo holds display and normalized grape names.
// LLM responses populate Name only; NormalizedName is set after parsing.
type GrapeInfo struct {
	Name           string   `json:"name"`
	Percentage     *float64 `json:"percentage,omitempty"`
	NormalizedName string   `json:"-"`
}

// WineInfoProvider fetches wine information from an LLM backend.
type WineInfoProvider interface {
	FetchWineInfo(ctx context.Context, key WineLookupKey) (*WineInfoResult, error)
}

// Closer is implemented by providers that hold external resources.
type Closer interface {
	Close() error
}
