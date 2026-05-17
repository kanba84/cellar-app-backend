package llm

import "testing"

func TestNormalizeGrapeName(t *testing.T) {
	if got := NormalizeGrapeName("  Pinot-Noir  "); got != "pinot noir" {
		t.Fatalf("NormalizeGrapeName() = %q, want pinot noir", got)
	}
}

func TestApplyGrapeNormalization(t *testing.T) {
	grapes := []GrapeInfo{{Name: "Cabernet-Sauvignon"}}
	ApplyGrapeNormalization(grapes)
	if grapes[0].Name != "Cabernet-Sauvignon" {
		t.Fatalf("display name = %q", grapes[0].Name)
	}
	if grapes[0].NormalizedName != "cabernet sauvignon" {
		t.Fatalf("normalized name = %q", grapes[0].NormalizedName)
	}
}
