package llm

import "strings"

// NormalizeGrapeName returns a lowercase, trimmed name for internal comparison.
func NormalizeGrapeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "-", " ")
	return strings.Join(strings.Fields(name), " ")
}

// ApplyGrapeNormalization sets NormalizedName while preserving display Name.
func ApplyGrapeNormalization(grapes []GrapeInfo) {
	for i := range grapes {
		grapes[i].Name = strings.TrimSpace(grapes[i].Name)
		grapes[i].NormalizedName = NormalizeGrapeName(grapes[i].Name)
	}
}
