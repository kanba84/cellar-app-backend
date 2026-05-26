package llm

import "fmt"

func buildWineInfoPrompt(key WineLookupKey) string {
	// Build wine description with available information
	wineDescription := key.Name
	if key.Vintage != nil {
		wineDescription = fmt.Sprintf("%s (%d)", wineDescription, *key.Vintage)
	}
	if key.Producer != nil && *key.Producer != "" {
		wineDescription = fmt.Sprintf("%s, produced by %s", wineDescription, *key.Producer)
	}
	if key.CountryName != nil && *key.CountryName != "" {
		wineDescription = fmt.Sprintf("%s, from %s", wineDescription, *key.CountryName)
	}
	if key.RegionName != nil && *key.RegionName != "" {
		wineDescription = fmt.Sprintf("%s, %s", wineDescription, *key.RegionName)
	}

	if key.IncludeTastingNote {
		return fmt.Sprintf(
			`Return JSON only. No markdown. Schema: {"producer":string|null,"grapes":[{"name":string,"percentage":number|null}],"tasting_note":string|null,"reference_price_jpy":number|null}. If grape percentage is unknown, set it to null. If the reference price in JPY is unknown, set it to null. Wine: %s`,
			wineDescription,
		)
	}

	return fmt.Sprintf(
		`Return JSON only. No markdown. Schema: {"producer":string|null,"grapes":[{"name":string,"percentage":number|null}],"reference_price_jpy":number|null}. If grape percentage is unknown, set it to null. If the reference price in JPY is unknown, set it to null. Wine: %s`,
		wineDescription,
	)
}
