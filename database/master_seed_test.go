package database

import "testing"

func TestMasterSeedReferencesAreConsistent(t *testing.T) {
	countries := make(map[string]struct{}, len(masterCountries))
	for _, country := range masterCountries {
		if len(country.code) != 2 {
			t.Errorf("country code %q must be ISO alpha-2", country.code)
		}
		if _, exists := countries[country.code]; exists {
			t.Errorf("duplicate country code %q", country.code)
		}
		countries[country.code] = struct{}{}
	}

	regions := make(map[string]regionSeed, len(masterRegions))
	for _, region := range masterRegions {
		if len(region.code) > 64 {
			t.Errorf("region code %q exceeds database column size", region.code)
		}
		if _, exists := regions[region.code]; exists {
			t.Errorf("duplicate region code %q", region.code)
		}
		if _, exists := countries[region.countryCode]; !exists {
			t.Errorf("region %q references unknown country %q", region.code, region.countryCode)
		}
		if region.parentCode != "" {
			parent, exists := regions[region.parentCode]
			if !exists {
				t.Errorf("region %q parent %q must be seeded before the child", region.code, region.parentCode)
			} else if parent.countryCode != region.countryCode {
				t.Errorf("region %q and its parent belong to different countries", region.code)
			}
		}
		regions[region.code] = region
	}

	designationTypes := make(map[string]struct{}, len(masterDesignationTypes))
	for _, designationType := range masterDesignationTypes {
		key := designationType.countryCode + ":" + designationType.code
		if len(designationType.code) > 50 {
			t.Errorf("designation type code %q exceeds database column size", designationType.code)
		}
		if _, exists := designationTypes[key]; exists {
			t.Errorf("duplicate designation type %q", key)
		}
		if _, exists := countries[designationType.countryCode]; !exists {
			t.Errorf("designation type %q references unknown country", key)
		}
		designationTypes[key] = struct{}{}
	}

	appellations := make(map[string]struct{}, len(masterAppellations))
	appellationCounts := make(map[string]int, len(masterCountries))
	for _, appellation := range masterAppellations {
		if len(appellation.code) > 100 {
			t.Errorf("appellation code %q exceeds database column size", appellation.code)
		}
		if _, exists := appellations[appellation.code]; exists {
			t.Errorf("duplicate appellation code %q", appellation.code)
		}
		appellations[appellation.code] = struct{}{}
		appellationCounts[appellation.countryCode]++
		if _, exists := countries[appellation.countryCode]; !exists {
			t.Errorf("appellation %q references unknown country %q", appellation.code, appellation.countryCode)
		}
		if _, exists := designationTypes[appellation.countryCode+":"+appellation.designationTypeCode]; !exists {
			t.Errorf("appellation %q references unknown designation type %q", appellation.code, appellation.designationTypeCode)
		}
		if len(appellation.regionCodes) == 0 {
			t.Errorf("appellation %q must reference at least one region", appellation.code)
		}
		for _, regionCode := range appellation.regionCodes {
			region, exists := regions[regionCode]
			if !exists {
				t.Errorf("appellation %q references unknown region %q", appellation.code, regionCode)
				continue
			}
			if region.countryCode != appellation.countryCode {
				t.Errorf("appellation %q and region %q belong to different countries", appellation.code, regionCode)
			}
		}
	}
	for _, country := range masterCountries {
		if count := appellationCounts[country.code]; count < 5 {
			t.Errorf("country %q has %d seeded appellations; want at least 5", country.code, count)
		}
	}
}

func TestCanonicalRegionNamesPreserveAccents(t *testing.T) {
	canonicalNames := map[string]string{
		"FR:MEDOC":                "Médoc",
		"FR:COTES-DE-PROVENCE":    "Côtes de Provence",
		"FR:COTES-DE-NUITS":       "Côtes de Nuits",
		"FR:VOSNE-ROMANEE":        "Vosne-Romanée",
		"FR:LANGUEDOC-ROUSSILLON": "Languedoc-Roussillon",
		"ES:ANDALUCIA":            "Andalucía",
		"ES:RIAS-BAIXAS":          "Rías Baixas",
		"PT:DAO":                  "Dão",
		"FR:SAINT-EMILION":        "Saint-Émilion",
		"FR:COTE-ROTIE":           "Côte-Rôtie",
		"RO:TARNAVE":              "Târnave",
		"HU:VILLANY":              "Villány",
		"CH:GENEVA":               "Genève",
	}
	for code, expected := range canonicalNames {
		found := false
		for _, region := range masterRegions {
			if region.code == code {
				found = true
				if region.name != expected {
					t.Errorf("region %q name = %q, want %q", code, region.name, expected)
				}
				break
			}
		}
		if !found {
			t.Errorf("canonical region %q is missing from seed data", code)
		}
	}
}

func TestRepresentativeHierarchyAndAppellationsExist(t *testing.T) {
	regionParents := make(map[string]string, len(masterRegions))
	for _, region := range masterRegions {
		regionParents[region.code] = region.parentCode
	}
	for code, expectedParent := range map[string]string{
		"FR:MEDOC":         "FR:BORDEAUX",
		"FR:HAUT-MEDOC":    "FR:MEDOC",
		"FR:PAUILLAC":      "FR:HAUT-MEDOC",
		"FR:SAINT-JULIEN":  "FR:HAUT-MEDOC",
		"FR:SAINT-EMILION": "FR:BORDEAUX",
	} {
		if parent := regionParents[code]; parent != expectedParent {
			t.Errorf("region %q parent = %q, want %q", code, parent, expectedParent)
		}
	}

	appellationCodes := make(map[string]struct{}, len(masterAppellations))
	appellationRegions := make(map[string]map[string]struct{}, len(masterAppellations))
	for _, appellation := range masterAppellations {
		appellationCodes[appellation.code] = struct{}{}
		appellationRegions[appellation.code] = make(map[string]struct{}, len(appellation.regionCodes))
		for _, regionCode := range appellation.regionCodes {
			appellationRegions[appellation.code][regionCode] = struct{}{}
		}
	}
	for _, code := range []string{
		"FR:AOC:SAINT-JULIEN",
		"FR:AOC:SAINT-EMILION-GRAND-CRU",
		"IT:DOCG:BAROLO",
		"ES:DOCA:RIOJA",
		"US:AVA:NAPA-VALLEY",
		"JP:GI:YAMANASHI",
		"CA:VQA:NIAGARA-PENINSULA",
	} {
		if _, exists := appellationCodes[code]; !exists {
			t.Errorf("representative appellation %q is missing from seed data", code)
		}
	}
	for _, relation := range [][2]string{
		{"FR:AOC:PAUILLAC", "FR:PAUILLAC"},
		{"ES:DO:CAVA", "ES:CATALUNA"},
		{"ES:DO:CAVA", "ES:ARAGON"},
		{"IT:DOC:PROSECCO", "IT:FRIULI-VENEZIA-GIULIA"},
		{"US:AVA:COLUMBIA-VALLEY", "US:OREGON"},
	} {
		if _, exists := appellationRegions[relation[0]][relation[1]]; !exists {
			t.Errorf("appellation %q is not associated with region %q", relation[0], relation[1])
		}
	}
}
