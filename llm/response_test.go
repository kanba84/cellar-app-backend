package llm

import "testing"

func TestCleanJSONResponse(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "plain json",
			in:   ` {"producer":"Chateau"} `,
			want: `{"producer":"Chateau"}`,
		},
		{
			name: "json code fence",
			in:   "```json\n{\"grapes\":[]}\n```",
			want: `{"grapes":[]}`,
		},
		{
			name: "generic code fence",
			in:   "```\n{\"grapes\":[]}\n```",
			want: `{"grapes":[]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanJSONResponse(tt.in); got != tt.want {
				t.Fatalf("cleanJSONResponse() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseWineInfoResultPreservesDisplayName(t *testing.T) {
	raw := `{"producer":"Domaine X","grapes":[{"name":"Pinot Noir","percentage":75.0}],"tasting_note":null}`
	result, err := parseWineInfoResult(raw)
	if err != nil {
		t.Fatalf("parseWineInfoResult() error = %v", err)
	}
	if result.Grapes[0].Name != "Pinot Noir" {
		t.Fatalf("display name = %q, want Pinot Noir", result.Grapes[0].Name)
	}
	if result.Grapes[0].NormalizedName != "pinot noir" {
		t.Fatalf("normalized name = %q, want pinot noir", result.Grapes[0].NormalizedName)
	}
	if result.Grapes[0].Percentage == nil || *result.Grapes[0].Percentage != 75.0 {
		t.Fatalf("percentage = %v, want 75.0", result.Grapes[0].Percentage)
	}
}

func TestParseWineInfoResultAllowsMissingPercentage(t *testing.T) {
	raw := `{"producer":"Domaine Y","grapes":[{"name":"Merlot"}],"tasting_note":null}`
	result, err := parseWineInfoResult(raw)
	if err != nil {
		t.Fatalf("parseWineInfoResult() error = %v", err)
	}
	if result.Grapes[0].Name != "Merlot" {
		t.Fatalf("display name = %q, want Merlot", result.Grapes[0].Name)
	}
	if result.Grapes[0].Percentage != nil {
		t.Fatalf("percentage should be nil when missing, got %v", result.Grapes[0].Percentage)
	}
}
