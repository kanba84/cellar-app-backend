package llm

import "fmt"

func buildWineInfoPrompt(key WineLookupKey) string {
	wineDescription := key.Name
	if key.Vintage != nil {
		wineDescription = fmt.Sprintf("%s (%d)", key.Name, *key.Vintage)
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
