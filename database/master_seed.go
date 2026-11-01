package database

import (
	"errors"
	"fmt"

	"cellar-app/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type countrySeed struct {
	code string
	name string
}

type regionSeed struct {
	code        string
	name        string
	countryCode string
	parentCode  string
	legacyNames []string
}

type designationTypeSeed struct {
	code        string
	name        string
	countryCode string
	rank        int
}

type appellationSeed struct {
	code                string
	name                string
	countryCode         string
	designationTypeCode string
	regionCodes         []string
	legacyNames         []string
}

var masterCountries = []countrySeed{
	{"FR", "France"}, {"IT", "Italy"}, {"ES", "Spain"}, {"US", "United States"},
	{"DE", "Germany"}, {"AU", "Australia"}, {"NZ", "New Zealand"}, {"CL", "Chile"},
	{"AR", "Argentina"}, {"JP", "Japan"}, {"RO", "Romania"}, {"HU", "Hungary"},
	{"PT", "Portugal"}, {"GE", "Georgia"}, {"AT", "Austria"}, {"CH", "Switzerland"},
	{"GR", "Greece"}, {"ZA", "South Africa"}, {"CA", "Canada"},
}

var masterRegions = []regionSeed{
	{"FR:BORDEAUX", "Bordeaux", "FR", "", nil},
	{"FR:BOURGOGNE", "Bourgogne", "FR", "", nil},
	{"FR:MEDOC", "Médoc", "FR", "FR:BORDEAUX", []string{"Medoc"}},
	{"FR:MARGAUX", "Margaux", "FR", "FR:MEDOC", nil},
	{"US:CALIFORNIA", "California", "US", "", nil},
	{"FR:ALSACE", "Alsace", "FR", "", nil},
	{"FR:COTES-DU-RHONE", "Côtes du Rhône", "FR", "", []string{"Cotes de Rhone"}},
	{"FR:PROVENCE", "Provence", "FR", "", nil},
	{"FR:COTES-DE-PROVENCE", "Côtes de Provence", "FR", "FR:PROVENCE", []string{"Cotes de Provence"}},
	{"FR:CHABLIS", "Chablis", "FR", "FR:BOURGOGNE", nil},
	{"FR:COTES-DE-NUITS", "Côtes de Nuits", "FR", "FR:BOURGOGNE", []string{"Cotes de Nuits"}},
	{"FR:VOSNE-ROMANEE", "Vosne-Romanée", "FR", "FR:COTES-DE-NUITS", []string{"Vosne Romanee"}},
	{"FR:CHAMPAGNE", "Champagne", "FR", "", nil},
	{"FR:LANGUEDOC-ROUSSILLON", "Languedoc-Roussillon", "FR", "", []string{"Languedoc Roussillon"}},
	{"FR:LISTRAC-MEDOC", "Listrac-Médoc", "FR", "FR:MEDOC", []string{"Listrac-médoc"}},
	{"IT:PIEMONTE", "Piemonte", "IT", "", nil},
	{"IT:LOMBARDIA", "Lombardia", "IT", "", nil},
	{"IT:TOSCANA", "Toscana", "IT", "", nil},
	{"IT:LUGANA", "Lugana", "IT", "IT:LOMBARDIA", nil},
	{"CL:CENTRAL-VALLEY", "Central Valley", "CL", "", nil},
	{"ES:RIOJA", "Rioja", "ES", "", nil},
	{"ES:NAVARRA", "Navarra", "ES", "", nil},
	{"ES:ARAGON", "Aragón", "ES", "", nil},
	{"ES:PAIS-VASCO", "País Vasco", "ES", "", nil},
	{"ES:GALICIA", "Galicia", "ES", "", nil},
	{"ES:CATALUNA", "Cataluña", "ES", "", nil},
	{"ES:VALENCIA", "Valencia", "ES", "", nil},
	{"ES:MURCIA", "Murcia", "ES", "", []string{"Murucia"}},
	{"ES:CASTILLA-Y-LEON", "Castilla y León", "ES", "", nil},
	{"ES:MADRID", "Madrid", "ES", "", nil},
	{"ES:CASTILLA-LA-MANCHA", "Castilla-La Mancha", "ES", "", []string{"Castilla La Mancha"}},
	{"ES:EXTREMADURA", "Extremadura", "ES", "", nil},
	{"ES:ANDALUCIA", "Andalucía", "ES", "", []string{"Andalucia"}},
	{"ES:ISLAS-BALEARES", "Islas Baleares", "ES", "", nil},
	{"ES:ISLAS-CANARIAS", "Islas Canarias", "ES", "", nil},
	{"DE:MOSEL", "Mosel", "DE", "", nil},
	{"FR:LOIRE", "Loire", "FR", "", nil},
	{"IT:VENETO", "Veneto", "IT", "", nil},
	{"IT:SICILIA", "Sicilia", "IT", "", nil},
	{"IT:PUGLIA", "Puglia", "IT", "", nil},
	{"IT:FRIULI-VENEZIA-GIULIA", "Friuli-Venezia Giulia", "IT", "", nil},
	{"ES:RIBERA-DEL-DUERO", "Ribera del Duero", "ES", "ES:CASTILLA-Y-LEON", nil},
	{"ES:PRIORAT", "Priorat", "ES", "ES:CATALUNA", nil},
	{"ES:RIAS-BAIXAS", "Rías Baixas", "ES", "ES:GALICIA", nil},
	{"PT:DOURO", "Douro", "PT", "", nil},
	{"PT:DAO", "Dão", "PT", "", nil},
	{"PT:ALENTEJO", "Alentejo", "PT", "", nil},
	{"PT:VINHO-VERDE", "Vinho Verde", "PT", "", nil},
	{"PT:MADEIRA", "Madeira", "PT", "", nil},
	{"DE:RHEINGAU", "Rheingau", "DE", "", nil},
	{"DE:PFALZ", "Pfalz", "DE", "", nil},
	{"DE:NAHE", "Nahe", "DE", "", nil},
	{"US:OREGON", "Oregon", "US", "", nil},
	{"US:WASHINGTON", "Washington", "US", "", nil},
	{"US:NEW-YORK", "New York", "US", "", nil},
	{"CL:MAIPO-VALLEY", "Maipo Valley", "CL", "CL:CENTRAL-VALLEY", nil},
	{"CL:COLCHAGUA", "Colchagua", "CL", "CL:CENTRAL-VALLEY", nil},
	{"AR:MENDOZA", "Mendoza", "AR", "", nil},
	{"AU:SOUTH-AUSTRALIA", "South Australia", "AU", "", nil},
	{"AU:WESTERN-AUSTRALIA", "Western Australia", "AU", "", nil},
	{"AU:NEW-SOUTH-WALES", "New South Wales", "AU", "", nil},
	{"AU:VICTORIA", "Victoria", "AU", "", nil},
	{"AU:BAROSSA-VALLEY", "Barossa Valley", "AU", "AU:SOUTH-AUSTRALIA", nil},
	{"AU:MARGARET-RIVER", "Margaret River", "AU", "AU:WESTERN-AUSTRALIA", nil},
	{"AU:YARRA-VALLEY", "Yarra Valley", "AU", "AU:VICTORIA", nil},
	{"NZ:NORTH-ISLAND", "North Island", "NZ", "", nil},
	{"NZ:SOUTH-ISLAND", "South Island", "NZ", "", nil},
	{"NZ:MARLBOROUGH", "Marlborough", "NZ", "NZ:SOUTH-ISLAND", nil},
	{"NZ:CENTRAL-OTAGO", "Central Otago", "NZ", "NZ:SOUTH-ISLAND", nil},
	{"ZA:COASTAL-REGION", "Coastal Region", "ZA", "", nil},
	{"ZA:STELLENBOSCH", "Stellenbosch", "ZA", "ZA:COASTAL-REGION", nil},
	{"ZA:SWARTLAND", "Swartland", "ZA", "ZA:COASTAL-REGION", nil},
	{"JP:YAMANASHI", "Yamanashi", "JP", "", nil},
	{"JP:NAGANO", "Nagano", "JP", "", nil},
	{"JP:HOKKAIDO", "Hokkaido", "JP", "", nil},
	{"FR:HAUT-MEDOC", "Haut-Médoc", "FR", "FR:MEDOC", nil},
	{"FR:PAUILLAC", "Pauillac", "FR", "FR:HAUT-MEDOC", nil},
	{"FR:SAINT-JULIEN", "Saint-Julien", "FR", "FR:HAUT-MEDOC", nil},
	{"FR:SAINT-ESTEPHE", "Saint-Estèphe", "FR", "FR:HAUT-MEDOC", nil},
	{"FR:SAINT-EMILION", "Saint-Émilion", "FR", "FR:BORDEAUX", nil},
	{"FR:POMEROL", "Pomerol", "FR", "FR:BORDEAUX", nil},
	{"FR:GRAVES", "Graves", "FR", "FR:BORDEAUX", nil},
	{"FR:PESSAC-LEOGNAN", "Pessac-Léognan", "FR", "FR:GRAVES", nil},
	{"FR:SAUTERNES", "Sauternes", "FR", "FR:GRAVES", nil},
	{"FR:GEVREY-CHAMBERTIN", "Gevrey-Chambertin", "FR", "FR:COTES-DE-NUITS", nil},
	{"FR:CHAMBOLLE-MUSIGNY", "Chambolle-Musigny", "FR", "FR:COTES-DE-NUITS", nil},
	{"FR:COTE-DE-BEAUNE", "Côte de Beaune", "FR", "FR:BOURGOGNE", nil},
	{"FR:MEURSAULT", "Meursault", "FR", "FR:COTE-DE-BEAUNE", nil},
	{"FR:PULIGNY-MONTRACHET", "Puligny-Montrachet", "FR", "FR:COTE-DE-BEAUNE", nil},
	{"FR:COTE-ROTIE", "Côte-Rôtie", "FR", "FR:COTES-DU-RHONE", nil},
	{"FR:HERMITAGE", "Hermitage", "FR", "FR:COTES-DU-RHONE", nil},
	{"FR:CHATEAUNEUF-DU-PAPE", "Châteauneuf-du-Pape", "FR", "FR:COTES-DU-RHONE", nil},
	{"FR:SANCERRE", "Sancerre", "FR", "FR:LOIRE", nil},
	{"IT:LANGHE", "Langhe", "IT", "IT:PIEMONTE", nil},
	{"IT:BAROLO", "Barolo", "IT", "IT:LANGHE", nil},
	{"IT:BARBARESCO", "Barbaresco", "IT", "IT:LANGHE", nil},
	{"IT:MONTALCINO", "Montalcino", "IT", "IT:TOSCANA", nil},
	{"IT:CHIANTI", "Chianti", "IT", "IT:TOSCANA", nil},
	{"IT:VALPOLICELLA", "Valpolicella", "IT", "IT:VENETO", nil},
	{"IT:SOAVE", "Soave", "IT", "IT:VENETO", nil},
	{"IT:FRANCIACORTA", "Franciacorta", "IT", "IT:LOMBARDIA", nil},
	{"IT:ETNA", "Etna", "IT", "IT:SICILIA", nil},
	{"IT:CAMPANIA", "Campania", "IT", "", nil},
	{"IT:TAURASI", "Taurasi", "IT", "IT:CAMPANIA", nil},
	{"ES:JEREZ", "Jerez", "ES", "ES:ANDALUCIA", nil},
	{"DE:RHEINHESSEN", "Rheinhessen", "DE", "", nil},
	{"DE:BADEN", "Baden", "DE", "", nil},
	{"US:NAPA-VALLEY", "Napa Valley", "US", "US:CALIFORNIA", nil},
	{"US:SONOMA-VALLEY", "Sonoma Valley", "US", "US:CALIFORNIA", nil},
	{"US:WILLAMETTE-VALLEY", "Willamette Valley", "US", "US:OREGON", nil},
	{"US:COLUMBIA-VALLEY", "Columbia Valley", "US", "US:WASHINGTON", nil},
	{"US:FINGER-LAKES", "Finger Lakes", "US", "US:NEW-YORK", nil},
	{"AU:HUNTER-VALLEY", "Hunter Valley", "AU", "AU:NEW-SOUTH-WALES", nil},
	{"AU:MCLAREN-VALE", "McLaren Vale", "AU", "AU:SOUTH-AUSTRALIA", nil},
	{"NZ:HAWKES-BAY", "Hawke’s Bay", "NZ", "NZ:NORTH-ISLAND", nil},
	{"NZ:WAIRARAPA", "Wairarapa", "NZ", "NZ:NORTH-ISLAND", nil},
	{"NZ:NELSON", "Nelson", "NZ", "NZ:SOUTH-ISLAND", nil},
	{"CL:ACONCAGUA", "Aconcagua", "CL", "", nil},
	{"CL:CASABLANCA-VALLEY", "Valle de Casablanca", "CL", "CL:ACONCAGUA", nil},
	{"CL:MAULE-VALLEY", "Valle del Maule", "CL", "CL:CENTRAL-VALLEY", nil},
	{"CL:ITATA-VALLEY", "Valle de Itata", "CL", "CL:CENTRAL-VALLEY", nil},
	{"AR:SALTA", "Salta", "AR", "", nil},
	{"AR:CAFAYATE", "Cafayate", "AR", "AR:SALTA", nil},
	{"AR:UCO-VALLEY", "Valle de Uco", "AR", "AR:MENDOZA", nil},
	{"AR:LUJAN-DE-CUYO", "Luján de Cuyo", "AR", "AR:MENDOZA", nil},
	{"JP:OSAKA", "Osaka", "JP", "", nil},
	{"JP:YAMAGATA", "Yamagata", "JP", "", nil},
	{"RO:TRANSYLVANIA", "Transylvania", "RO", "", nil},
	{"RO:MUNTENIA", "Muntenia", "RO", "", nil},
	{"RO:MOLDAVIA", "Moldavia", "RO", "", nil},
	{"RO:DOBROGEA", "Dobrogea", "RO", "", nil},
	{"RO:BANAT", "Banat", "RO", "", nil},
	{"RO:TARNAVE", "Târnave", "RO", "RO:TRANSYLVANIA", nil},
	{"RO:DEALU-MARE", "Dealu Mare", "RO", "RO:MUNTENIA", nil},
	{"RO:COTNARI", "Cotnari", "RO", "RO:MOLDAVIA", nil},
	{"RO:MURFATLAR", "Murfatlar", "RO", "RO:DOBROGEA", nil},
	{"RO:RECAS", "Recaș", "RO", "RO:BANAT", nil},
	{"HU:TOKAJ", "Tokaj", "HU", "", nil},
	{"HU:EGER", "Eger", "HU", "", nil},
	{"HU:VILLANY", "Villány", "HU", "", nil},
	{"HU:SZEKSZARD", "Szekszárd", "HU", "", nil},
	{"HU:SOMLO", "Somló", "HU", "", nil},
	{"GE:KAKHETI", "Kakheti", "GE", "", nil},
	{"GE:RACHA-LECHKHUMI", "Racha-Lechkhumi", "GE", "", nil},
	{"AT:NIEDEROSTERREICH", "Niederösterreich", "AT", "", nil},
	{"AT:STEIERMARK", "Steiermark", "AT", "", nil},
	{"AT:KAMPTAL", "Kamptal", "AT", "AT:NIEDEROSTERREICH", nil},
	{"AT:KREMSTAL", "Kremstal", "AT", "AT:NIEDEROSTERREICH", nil},
	{"AT:WAGRAM", "Wagram", "AT", "AT:NIEDEROSTERREICH", nil},
	{"AT:WEINVIERTEL", "Weinviertel", "AT", "AT:NIEDEROSTERREICH", nil},
	{"AT:SUDSTEIERMARK", "Südsteiermark", "AT", "AT:STEIERMARK", nil},
	{"CH:VAUD", "Vaud", "CH", "", nil},
	{"CH:VALAIS", "Valais", "CH", "", nil},
	{"CH:GENEVA", "Genève", "CH", "", nil},
	{"CH:LAVAUX", "Lavaux", "CH", "CH:VAUD", nil},
	{"CH:LA-COTE", "La Côte", "CH", "CH:VAUD", nil},
	{"CH:CHABLAIS", "Chablais", "CH", "CH:VAUD", nil},
	{"GR:CYCLADES", "Cyclades", "GR", "", nil},
	{"GR:PELOPONNESE", "Peloponnese", "GR", "", nil},
	{"GR:MACEDONIA", "Macedonia", "GR", "", nil},
	{"GR:SANTORINI", "Santorini", "GR", "GR:CYCLADES", nil},
	{"GR:NEMEA", "Nemea", "GR", "GR:PELOPONNESE", nil},
	{"GR:NAOUSSA", "Naoussa", "GR", "GR:MACEDONIA", nil},
	{"GR:AMYNDEON", "Amyndeon", "GR", "GR:MACEDONIA", nil},
	{"GR:MANTINIA", "Mantinia", "GR", "GR:PELOPONNESE", nil},
	{"ZA:BREEDE-RIVER-VALLEY", "Breede River Valley", "ZA", "", nil},
	{"ZA:CAPE-SOUTH-COAST", "Cape South Coast", "ZA", "", nil},
	{"ZA:PAARL", "Paarl", "ZA", "ZA:COASTAL-REGION", nil},
	{"ZA:ROBERTSON", "Robertson", "ZA", "ZA:BREEDE-RIVER-VALLEY", nil},
	{"ZA:WALKER-BAY", "Walker Bay", "ZA", "ZA:CAPE-SOUTH-COAST", nil},
	{"CA:ONTARIO", "Ontario", "CA", "", nil},
	{"CA:BRITISH-COLUMBIA", "British Columbia", "CA", "", nil},
	{"CA:NIAGARA-PENINSULA", "Niagara Peninsula", "CA", "CA:ONTARIO", nil},
	{"CA:PRINCE-EDWARD-COUNTY", "Prince Edward County", "CA", "CA:ONTARIO", nil},
	{"CA:LAKE-ERIE-NORTH-SHORE", "Lake Erie North Shore", "CA", "CA:ONTARIO", nil},
	{"CA:OKANAGAN-VALLEY", "Okanagan Valley", "CA", "CA:BRITISH-COLUMBIA", nil},
	{"CA:SIMILKAMEEN-VALLEY", "Similkameen Valley", "CA", "CA:BRITISH-COLUMBIA", nil},
}

var masterDesignationTypes = []designationTypeSeed{
	{"AOC", "Appellation d'Origine Contrôlée", "FR", 1},
	{"D.O.C.G.", "Denominazione di Origine Controllata e Garantita", "IT", 1},
	{"D.O.C.", "Denominazione di Origine Controllata", "IT", 2},
	{"DO", "Denominación de Origen (DO/DOP)", "ES", 1},
	{"DOCa", "Denominación de Origen Calificada (DOCa)", "ES", 1},
	{"AVA", "American Viticultural Area (AVA)", "US", 1},
	{"PDO", "Protected Designation of Origin (Qualitätswein)", "DE", 1},
	{"GI", "Geographical Indication (GI)", "AU", 1},
	{"GI", "Geographical Indication (GI)", "NZ", 1},
	{"D.O.", "Denominación de Origen", "CL", 1},
	{"IG", "Indicación Geográfica (IG)", "AR", 1},
	{"DOC", "Denominación de Origen Controlada (DOC)", "AR", 1},
	{"GI", "Geographical Indication (GI)", "JP", 1},
	{"DOC", "Denumire de Origine Controlată (DOC/PDO)", "RO", 1},
	{"PDO", "Protected Designation of Origin (OEM)", "HU", 1},
	{"DOC", "Denominação de Origem Controlada (DOC/DOP)", "PT", 1},
	{"PDO", "Protected Designation of Origin (PDO)", "GE", 1},
	{"DAC", "Districtus Austriae Controllatus (DAC)", "AT", 1},
	{"AOC", "Appellation d'origine contrôlée (AOC)", "CH", 1},
	{"PDO", "Protected Designation of Origin (PDO)", "GR", 1},
	{"WO", "Wine of Origin (WO)", "ZA", 1},
	{"VQA", "Vintners Quality Alliance (VQA)", "CA", 1},
}

var masterAppellations = []appellationSeed{
	{"FR:AOC:BORDEAUX", "A.C. Bordeaux", "FR", "AOC", []string{"FR:BORDEAUX"}, nil},
	{"FR:AOC:COTE-DE-NUITS-VILLAGES", "A.C. Côte de Nuits-Villages", "FR", "AOC", []string{"FR:COTES-DE-NUITS"}, nil},
	{"FR:AOC:MARGAUX", "A.C. Margaux", "FR", "AOC", []string{"FR:MARGAUX"}, nil},
	{"FR:AOC:LISTRAC-MEDOC", "A.C. Listrac-Médoc", "FR", "AOC", []string{"FR:LISTRAC-MEDOC"}, []string{"A.C. Listrac-médoc"}},
	{"FR:AOC:MEDOC", "A.C. Médoc", "FR", "AOC", []string{"FR:MEDOC"}, nil},
	{"FR:AOC:COTES-DE-PROVENCE", "A.C. Côtes de Provence", "FR", "AOC", []string{"FR:COTES-DE-PROVENCE"}, nil},
	{"FR:AOC:VOSNE-ROMANEE", "A.C. Vosne-Romanée", "FR", "AOC", []string{"FR:VOSNE-ROMANEE"}, nil},
	{"FR:AOC:CHAMPAGNE", "A.C. Champagne", "FR", "AOC", []string{"FR:CHAMPAGNE"}, nil},
	{"IT:DOC:LUGANA", "D.O.C. Lugana", "IT", "D.O.C.", []string{"IT:LUGANA"}, nil},
	{"CL:DO:CENTRAL-VALLEY", "D.O. Central Valley", "CL", "D.O.", []string{"CL:CENTRAL-VALLEY"}, nil},
	{"FR:AOC:PAUILLAC", "A.C. Pauillac", "FR", "AOC", []string{"FR:BORDEAUX", "FR:MEDOC", "FR:PAUILLAC"}, nil},
	{"FR:AOC:SAINT-JULIEN", "A.C. Saint-Julien", "FR", "AOC", []string{"FR:SAINT-JULIEN"}, nil},
	{"FR:AOC:SAINT-EMILION-GRAND-CRU", "A.C. Saint-Émilion Grand Cru", "FR", "AOC", []string{"FR:SAINT-EMILION"}, nil},
	{"FR:AOC:SAINT-ESTEPHE", "A.C. Saint-Estèphe", "FR", "AOC", []string{"FR:SAINT-ESTEPHE"}, nil},
	{"FR:AOC:POMEROL", "A.C. Pomerol", "FR", "AOC", []string{"FR:POMEROL"}, nil},
	{"FR:AOC:GEVREY-CHAMBERTIN", "A.C. Gevrey-Chambertin", "FR", "AOC", []string{"FR:GEVREY-CHAMBERTIN"}, nil},
	{"FR:AOC:CHAMBOLLE-MUSIGNY", "A.C. Chambolle-Musigny", "FR", "AOC", []string{"FR:CHAMBOLLE-MUSIGNY"}, nil},
	{"FR:AOC:MEURSAULT", "A.C. Meursault", "FR", "AOC", []string{"FR:MEURSAULT"}, nil},
	{"FR:AOC:PULIGNY-MONTRACHET", "A.C. Puligny-Montrachet", "FR", "AOC", []string{"FR:PULIGNY-MONTRACHET"}, nil},
	{"FR:AOC:CHATEAUNEUF-DU-PAPE", "A.C. Châteauneuf-du-Pape", "FR", "AOC", []string{"FR:CHATEAUNEUF-DU-PAPE"}, nil},
	{"FR:AOC:COTE-ROTIE", "A.C. Côte-Rôtie", "FR", "AOC", []string{"FR:COTE-ROTIE"}, nil},
	{"FR:AOC:HERMITAGE", "A.C. Hermitage", "FR", "AOC", []string{"FR:HERMITAGE"}, nil},
	{"FR:AOC:SANCERRE", "A.C. Sancerre", "FR", "AOC", []string{"FR:SANCERRE"}, nil},
	{"IT:DOCG:BAROLO", "D.O.C.G. Barolo", "IT", "D.O.C.G.", []string{"IT:BAROLO"}, nil},
	{"IT:DOCG:BARBARESCO", "D.O.C.G. Barbaresco", "IT", "D.O.C.G.", []string{"IT:BARBARESCO"}, nil},
	{"IT:DOCG:BRUNELLO-DI-MONTALCINO", "D.O.C.G. Brunello di Montalcino", "IT", "D.O.C.G.", []string{"IT:MONTALCINO"}, nil},
	{"IT:DOCG:CHIANTI-CLASSICO", "D.O.C.G. Chianti Classico", "IT", "D.O.C.G.", []string{"IT:CHIANTI"}, nil},
	{"IT:DOC:PROSECCO", "D.O.C. Prosecco", "IT", "D.O.C.", []string{"IT:VENETO", "IT:FRIULI-VENEZIA-GIULIA"}, nil},
	{"IT:DOCG:AMARONE-DELLA-VALPOLICELLA", "D.O.C.G. Amarone della Valpolicella", "IT", "D.O.C.G.", []string{"IT:VALPOLICELLA"}, nil},
	{"IT:DOC:SOAVE", "D.O.C. Soave", "IT", "D.O.C.", []string{"IT:SOAVE"}, nil},
	{"IT:DOCG:FRANCIACORTA", "D.O.C.G. Franciacorta", "IT", "D.O.C.G.", []string{"IT:FRANCIACORTA"}, nil},
	{"IT:DOC:ETNA", "D.O.C. Etna", "IT", "D.O.C.", []string{"IT:ETNA"}, nil},
	{"IT:DOCG:TAURASI", "D.O.C.G. Taurasi", "IT", "D.O.C.G.", []string{"IT:TAURASI"}, nil},
	{"ES:DOCA:RIOJA", "D.O.Ca Rioja", "ES", "DOCa", []string{"ES:RIOJA"}, nil},
	{"ES:DOQ:PRIORAT", "D.O.Q. Priorat", "ES", "DOCa", []string{"ES:PRIORAT"}, nil},
	{"ES:DO:RIBERA-DEL-DUERO", "D.O. Ribera del Duero", "ES", "DO", []string{"ES:RIBERA-DEL-DUERO"}, nil},
	{"ES:DO:RIAS-BAIXAS", "D.O. Rías Baixas", "ES", "DO", []string{"ES:RIAS-BAIXAS"}, nil},
	{"ES:DO:JEREZ-XERES-SHERRY", "D.O. Jerez-Xérès-Sherry", "ES", "DO", []string{"ES:JEREZ"}, nil},
	{"ES:DO:RUEDA", "D.O. Rueda", "ES", "DO", []string{"ES:CASTILLA-Y-LEON"}, nil},
	{"ES:DO:CAVA", "D.O. Cava", "ES", "DO", []string{"ES:CATALUNA", "ES:ARAGON", "ES:VALENCIA", "ES:NAVARRA", "ES:PAIS-VASCO", "ES:RIOJA", "ES:EXTREMADURA"}, nil},
	{"ES:DO:BIERZO", "D.O. Bierzo", "ES", "DO", []string{"ES:CASTILLA-Y-LEON"}, nil},
	{"US:AVA:NAPA-VALLEY", "Napa Valley AVA", "US", "AVA", []string{"US:NAPA-VALLEY"}, nil},
	{"US:AVA:SONOMA-VALLEY", "Sonoma Valley AVA", "US", "AVA", []string{"US:SONOMA-VALLEY"}, nil},
	{"US:AVA:WILLAMETTE-VALLEY", "Willamette Valley AVA", "US", "AVA", []string{"US:WILLAMETTE-VALLEY"}, nil},
	{"US:AVA:COLUMBIA-VALLEY", "Columbia Valley AVA", "US", "AVA", []string{"US:COLUMBIA-VALLEY", "US:OREGON"}, nil},
	{"US:AVA:FINGER-LAKES", "Finger Lakes AVA", "US", "AVA", []string{"US:FINGER-LAKES"}, nil},
	{"DE:PDO:MOSEL", "Mosel Qualitätswein", "DE", "PDO", []string{"DE:MOSEL"}, nil},
	{"DE:PDO:RHEINGAU", "Rheingau Qualitätswein", "DE", "PDO", []string{"DE:RHEINGAU"}, nil},
	{"DE:PDO:PFALZ", "Pfalz Qualitätswein", "DE", "PDO", []string{"DE:PFALZ"}, nil},
	{"DE:PDO:RHEINHESSEN", "Rheinhessen Qualitätswein", "DE", "PDO", []string{"DE:RHEINHESSEN"}, nil},
	{"DE:PDO:BADEN", "Baden Qualitätswein", "DE", "PDO", []string{"DE:BADEN"}, nil},
	{"AU:GI:BAROSSA-VALLEY", "Barossa Valley GI", "AU", "GI", []string{"AU:BAROSSA-VALLEY"}, nil},
	{"AU:GI:MARGARET-RIVER", "Margaret River GI", "AU", "GI", []string{"AU:MARGARET-RIVER"}, nil},
	{"AU:GI:HUNTER-VALLEY", "Hunter Valley GI", "AU", "GI", []string{"AU:HUNTER-VALLEY"}, nil},
	{"AU:GI:MCLAREN-VALE", "McLaren Vale GI", "AU", "GI", []string{"AU:MCLAREN-VALE"}, nil},
	{"AU:GI:YARRA-VALLEY", "Yarra Valley GI", "AU", "GI", []string{"AU:YARRA-VALLEY"}, nil},
	{"NZ:GI:MARLBOROUGH", "Marlborough GI", "NZ", "GI", []string{"NZ:MARLBOROUGH"}, nil},
	{"NZ:GI:CENTRAL-OTAGO", "Central Otago GI", "NZ", "GI", []string{"NZ:CENTRAL-OTAGO"}, nil},
	{"NZ:GI:HAWKES-BAY", "Hawke’s Bay GI", "NZ", "GI", []string{"NZ:HAWKES-BAY"}, nil},
	{"NZ:GI:WAIRARAPA", "Wairarapa GI", "NZ", "GI", []string{"NZ:WAIRARAPA"}, nil},
	{"NZ:GI:NELSON", "Nelson GI", "NZ", "GI", []string{"NZ:NELSON"}, nil},
	{"CL:DO:MAIPO", "D.O. Valle del Maipo", "CL", "D.O.", []string{"CL:MAIPO-VALLEY"}, nil},
	{"CL:DO:CASABLANCA", "D.O. Valle de Casablanca", "CL", "D.O.", []string{"CL:CASABLANCA-VALLEY"}, nil},
	{"CL:DO:COLCHAGUA", "D.O. Valle de Colchagua", "CL", "D.O.", []string{"CL:COLCHAGUA"}, nil},
	{"CL:DO:MAULE", "D.O. Valle del Maule", "CL", "D.O.", []string{"CL:MAULE-VALLEY"}, nil},
	{"CL:DO:ITATA", "D.O. Valle de Itata", "CL", "D.O.", []string{"CL:ITATA-VALLEY"}, nil},
	{"AR:IG:MENDOZA", "IG Mendoza", "AR", "IG", []string{"AR:MENDOZA"}, nil},
	{"AR:IG:SALTA", "IG Salta", "AR", "IG", []string{"AR:SALTA"}, nil},
	{"AR:IG:CAFAYATE", "IG Cafayate", "AR", "IG", []string{"AR:CAFAYATE"}, nil},
	{"AR:IG:UCO-VALLEY", "IG Valle de Uco", "AR", "IG", []string{"AR:UCO-VALLEY"}, nil},
	{"AR:DOC:LUJAN-DE-CUYO", "DOC Luján de Cuyo", "AR", "DOC", []string{"AR:LUJAN-DE-CUYO"}, nil},
	{"JP:GI:YAMANASHI", "GI Yamanashi", "JP", "GI", []string{"JP:YAMANASHI"}, nil},
	{"JP:GI:HOKKAIDO", "GI Hokkaido", "JP", "GI", []string{"JP:HOKKAIDO"}, nil},
	{"JP:GI:NAGANO", "GI Nagano", "JP", "GI", []string{"JP:NAGANO"}, nil},
	{"JP:GI:OSAKA", "GI Osaka", "JP", "GI", []string{"JP:OSAKA"}, nil},
	{"JP:GI:YAMAGATA", "GI Yamagata", "JP", "GI", []string{"JP:YAMAGATA"}, nil},
	{"RO:DOC:TARNAVE", "DOC Târnave", "RO", "DOC", []string{"RO:TARNAVE"}, nil},
	{"RO:DOC:DEALU-MARE", "DOC Dealu Mare", "RO", "DOC", []string{"RO:DEALU-MARE"}, nil},
	{"RO:DOC:COTNARI", "DOC Cotnari", "RO", "DOC", []string{"RO:COTNARI"}, nil},
	{"RO:DOC:MURFATLAR", "DOC Murfatlar", "RO", "DOC", []string{"RO:MURFATLAR"}, nil},
	{"RO:DOC:RECAS", "DOC Recaș", "RO", "DOC", []string{"RO:RECAS"}, nil},
	{"HU:PDO:TOKAJ", "OEM Tokaj", "HU", "PDO", []string{"HU:TOKAJ"}, nil},
	{"HU:PDO:EGER", "OEM Eger", "HU", "PDO", []string{"HU:EGER"}, nil},
	{"HU:PDO:VILLANY", "OEM Villány", "HU", "PDO", []string{"HU:VILLANY"}, nil},
	{"HU:PDO:SZEKSZARD", "OEM Szekszárd", "HU", "PDO", []string{"HU:SZEKSZARD"}, nil},
	{"HU:PDO:SOMLO", "OEM Somló", "HU", "PDO", []string{"HU:SOMLO"}, nil},
	{"PT:DOC:DOURO", "DOC Douro", "PT", "DOC", []string{"PT:DOURO"}, nil},
	{"PT:DOC:VINHO-VERDE", "DOC Vinho Verde", "PT", "DOC", []string{"PT:VINHO-VERDE"}, nil},
	{"PT:DOC:DAO", "DOC Dão", "PT", "DOC", []string{"PT:DAO"}, nil},
	{"PT:DOC:ALENTEJO", "DOC Alentejo", "PT", "DOC", []string{"PT:ALENTEJO"}, nil},
	{"PT:DOC:MADEIRA", "DOC Madeira", "PT", "DOC", []string{"PT:MADEIRA"}, nil},
	{"GE:PDO:KAKHETI", "PDO Kakheti", "GE", "PDO", []string{"GE:KAKHETI"}, nil},
	{"GE:PDO:MUKUZANI", "PDO Mukuzani", "GE", "PDO", []string{"GE:KAKHETI"}, nil},
	{"GE:PDO:KINDZMARAULI", "PDO Kindzmarauli", "GE", "PDO", []string{"GE:KAKHETI"}, nil},
	{"GE:PDO:TSINANDALI", "PDO Tsinandali", "GE", "PDO", []string{"GE:KAKHETI"}, nil},
	{"GE:PDO:KHAVANCHKARA", "PDO Khvanchkara", "GE", "PDO", []string{"GE:RACHA-LECHKHUMI"}, nil},
	{"AT:DAC:KAMPTAL", "Kamptal DAC", "AT", "DAC", []string{"AT:KAMPTAL"}, nil},
	{"AT:DAC:KREMSTAL", "Kremstal DAC", "AT", "DAC", []string{"AT:KREMSTAL"}, nil},
	{"AT:DAC:WAGRAM", "Wagram DAC", "AT", "DAC", []string{"AT:WAGRAM"}, nil},
	{"AT:DAC:WEINVIERTEL", "Weinviertel DAC", "AT", "DAC", []string{"AT:WEINVIERTEL"}, nil},
	{"AT:DAC:SUDSTEIERMARK", "Südsteiermark DAC", "AT", "DAC", []string{"AT:SUDSTEIERMARK"}, nil},
	{"CH:AOC:LAVAUX", "AOC Lavaux", "CH", "AOC", []string{"CH:LAVAUX"}, nil},
	{"CH:AOC:LA-COTE", "AOC La Côte", "CH", "AOC", []string{"CH:LA-COTE"}, nil},
	{"CH:AOC:CHABLAIS", "AOC Chablais", "CH", "AOC", []string{"CH:CHABLAIS"}, nil},
	{"CH:AOC:VALAIS", "AOC Valais", "CH", "AOC", []string{"CH:VALAIS"}, nil},
	{"CH:AOC:GENEVE", "AOC Genève", "CH", "AOC", []string{"CH:GENEVA"}, nil},
	{"GR:PDO:SANTORINI", "PDO Santorini", "GR", "PDO", []string{"GR:SANTORINI"}, nil},
	{"GR:PDO:NEMEA", "PDO Nemea", "GR", "PDO", []string{"GR:NEMEA"}, nil},
	{"GR:PDO:NAOUSSA", "PDO Naoussa", "GR", "PDO", []string{"GR:NAOUSSA"}, nil},
	{"GR:PDO:AMYNDEON", "PDO Amyndeon", "GR", "PDO", []string{"GR:AMYNDEON"}, nil},
	{"GR:PDO:MANTINIA", "PDO Mantinia", "GR", "PDO", []string{"GR:MANTINIA"}, nil},
	{"ZA:WO:STELLENBOSCH", "WO Stellenbosch", "ZA", "WO", []string{"ZA:STELLENBOSCH"}, nil},
	{"ZA:WO:PAARL", "WO Paarl", "ZA", "WO", []string{"ZA:PAARL"}, nil},
	{"ZA:WO:SWARTLAND", "WO Swartland", "ZA", "WO", []string{"ZA:SWARTLAND"}, nil},
	{"ZA:WO:ROBERTSON", "WO Robertson", "ZA", "WO", []string{"ZA:ROBERTSON"}, nil},
	{"ZA:WO:WALKER-BAY", "WO Walker Bay", "ZA", "WO", []string{"ZA:WALKER-BAY"}, nil},
	{"CA:VQA:NIAGARA-PENINSULA", "VQA Niagara Peninsula", "CA", "VQA", []string{"CA:NIAGARA-PENINSULA"}, nil},
	{"CA:VQA:PRINCE-EDWARD-COUNTY", "VQA Prince Edward County", "CA", "VQA", []string{"CA:PRINCE-EDWARD-COUNTY"}, nil},
	{"CA:VQA:LAKE-ERIE-NORTH-SHORE", "VQA Lake Erie North Shore", "CA", "VQA", []string{"CA:LAKE-ERIE-NORTH-SHORE"}, nil},
	{"CA:VQA:OKANAGAN-VALLEY", "VQA Okanagan Valley", "CA", "VQA", []string{"CA:OKANAGAN-VALLEY"}, nil},
	{"CA:VQA:SIMILKAMEEN-VALLEY", "VQA Similkameen Valley", "CA", "VQA", []string{"CA:SIMILKAMEEN-VALLEY"}, nil},
}

func SeedMasterData(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		countries, err := seedCountries(tx)
		if err != nil {
			return err
		}
		regions, err := seedRegions(tx, countries)
		if err != nil {
			return err
		}
		designationTypes, err := seedDesignationTypes(tx, countries)
		if err != nil {
			return err
		}
		return seedAppellations(tx, countries, regions, designationTypes)
	})
}

func seedCountries(tx *gorm.DB) (map[string]model.Country, error) {
	countries := make(map[string]model.Country, len(masterCountries))
	for _, seed := range masterCountries {
		country := model.Country{Name: seed.name, ISOCode: &seed.code}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "iso_code"}},
			DoUpdates: clause.AssignmentColumns([]string{"name"}),
		}).Create(&country).Error; err != nil {
			return nil, fmt.Errorf("upsert country %s: %w", seed.code, err)
		}
		if err := tx.Where("iso_code = ?", seed.code).First(&country).Error; err != nil {
			return nil, fmt.Errorf("load country %s after upsert: %w", seed.code, err)
		}
		countries[seed.code] = country
	}
	return countries, nil
}

func seedRegions(tx *gorm.DB, countries map[string]model.Country) (map[string]model.Region, error) {
	regions := make(map[string]model.Region, len(masterRegions))
	sortOrders := make(map[string]int)
	for _, seed := range masterRegions {
		country, ok := countries[seed.countryCode]
		if !ok {
			return nil, fmt.Errorf("unknown country code %s for region %s", seed.countryCode, seed.code)
		}
		sortOrders[seed.countryCode]++
		region := model.Region{
			Name:      seed.name,
			CountryID: country.ID,
			SortOrder: sortOrders[seed.countryCode],
		}
		if seed.parentCode != "" {
			parent, ok := regions[seed.parentCode]
			if !ok {
				return nil, fmt.Errorf("parent region %s must be seeded before %s", seed.parentCode, seed.code)
			}
			region.ParentID = &parent.ID
		}

		var existing model.Region
		err := tx.Where("code = ?", seed.code).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			names := append([]string{seed.name}, seed.legacyNames...)
			query := tx.Where("country_id = ? AND name IN ?", country.ID, names)
			if region.ParentID == nil {
				query = query.Where("parent_id IS NULL")
			} else {
				query = query.Where("parent_id = ?", *region.ParentID)
			}
			err = query.First(&existing).Error
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find region %s: %w", seed.code, err)
		}
		region.Code = &seed.code
		if err == nil {
			region.ID = existing.ID
			if err := tx.Omit("Country", "Parent").Save(&region).Error; err != nil {
				return nil, fmt.Errorf("update region %s: %w", seed.code, err)
			}
		} else if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "code"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "country_id", "parent_id", "sort_order"}),
		}).Omit("Country", "Parent").Create(&region).Error; err != nil {
			return nil, fmt.Errorf("insert region %s: %w", seed.code, err)
		}
		if err := tx.Where("code = ?", seed.code).First(&region).Error; err != nil {
			return nil, fmt.Errorf("load region %s after upsert: %w", seed.code, err)
		}
		regions[seed.code] = region
	}
	return regions, nil
}

func seedDesignationTypes(tx *gorm.DB, countries map[string]model.Country) (map[string]model.DesignationType, error) {
	result := make(map[string]model.DesignationType, len(masterDesignationTypes))
	for _, seed := range masterDesignationTypes {
		country, ok := countries[seed.countryCode]
		if !ok {
			return nil, fmt.Errorf("unknown country code %s for designation type %s", seed.countryCode, seed.code)
		}
		countryID := country.ID
		designationType := model.DesignationType{
			Name:      seed.name,
			Code:      seed.code,
			Rank:      seed.rank,
			CountryID: &countryID,
		}
		err := tx.Where("country_id = ? AND code = ?", country.ID, seed.code).First(&designationType).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "country_id"}, {Name: "code"}},
				DoUpdates: clause.AssignmentColumns([]string{"name", "rank"}),
			}).Create(&designationType).Error
		} else if err == nil {
			designationType.Name = seed.name
			designationType.Code = seed.code
			designationType.Rank = seed.rank
			err = tx.Omit("Country").Save(&designationType).Error
		}
		if err != nil {
			return nil, fmt.Errorf("upsert designation type %s/%s: %w", seed.countryCode, seed.code, err)
		}
		if err := tx.Where("country_id = ? AND code = ?", country.ID, seed.code).First(&designationType).Error; err != nil {
			return nil, fmt.Errorf("load designation type %s after upsert: %w", seed.code, err)
		}
		result[seed.countryCode+":"+seed.code] = designationType
	}
	return result, nil
}

func seedAppellations(
	tx *gorm.DB,
	countries map[string]model.Country,
	regions map[string]model.Region,
	designationTypes map[string]model.DesignationType,
) error {
	for order, seed := range masterAppellations {
		country, ok := countries[seed.countryCode]
		if !ok {
			return fmt.Errorf("unknown country code %s for appellation %s", seed.countryCode, seed.code)
		}
		designationType, ok := designationTypes[seed.countryCode+":"+seed.designationTypeCode]
		if !ok {
			return fmt.Errorf("unknown designation type %s for appellation %s", seed.designationTypeCode, seed.code)
		}
		appellationRegions := make([]model.Region, 0, len(seed.regionCodes))
		for _, regionCode := range seed.regionCodes {
			region, ok := regions[regionCode]
			if !ok {
				return fmt.Errorf("unknown region code %s for appellation %s", regionCode, seed.code)
			}
			appellationRegions = append(appellationRegions, region)
		}
		if len(appellationRegions) == 0 {
			return fmt.Errorf("appellation %s must reference at least one region", seed.code)
		}

		countryID := country.ID
		designationTypeID := designationType.ID
		regionID := appellationRegions[0].ID
		appellation := model.Appellation{
			Name:              seed.name,
			Code:              &seed.code,
			CountryID:         &countryID,
			DesignationTypeID: &designationTypeID,
			RegionID:          &regionID,
			Regions:           appellationRegions,
			SortOrder:         order + 1,
		}
		err := tx.Where("code = ?", seed.code).First(&appellation).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			names := append([]string{seed.name}, seed.legacyNames...)
			err = tx.
				Where("name IN ? AND (country_id = ? OR country_id IS NULL)", names, country.ID).
				Where("region_id = ?", regionID).
				First(&appellation).Error
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("find appellation %s: %w", seed.code, err)
		}
		if err == nil {
			appellation.Name = seed.name
			appellation.Code = &seed.code
			appellation.CountryID = &countryID
			appellation.DesignationTypeID = &designationTypeID
			appellation.RegionID = &regionID
			appellation.SortOrder = order + 1
			if err := tx.Omit("Country", "DesignationType", "Region", "Regions").Save(&appellation).Error; err != nil {
				return fmt.Errorf("update appellation %s: %w", seed.code, err)
			}
		} else if err := tx.Omit("Country", "DesignationType", "Region", "Regions").Create(&appellation).Error; err != nil {
			return fmt.Errorf("insert appellation %s: %w", seed.code, err)
		}
		if err := tx.Where("code = ?", seed.code).First(&appellation).Error; err != nil {
			return fmt.Errorf("load appellation %s after upsert: %w", seed.code, err)
		}
		if err := tx.Model(&appellation).Association("Regions").Append(appellationRegions); err != nil {
			return fmt.Errorf("associate regions with appellation %s: %w", seed.code, err)
		}
	}
	return nil
}
