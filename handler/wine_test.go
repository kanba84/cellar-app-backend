package handler

import (
	"testing"
)

func intPointer(value int) *int {
	return &value
}

func TestValidateDrinkingWindow(t *testing.T) {
	tests := []struct {
		name  string
		start *int
		end   *int
		want  bool
	}{
		{name: "both unset", want: true},
		{name: "start only", start: intPointer(2027), want: true},
		{name: "end only", end: intPointer(2035), want: true},
		{name: "valid range", start: intPointer(2027), end: intPointer(2035), want: true},
		{name: "same year", start: intPointer(2030), end: intPointer(2030), want: true},
		{name: "reversed range", start: intPointer(2035), end: intPointer(2027), want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := validateDrinkingWindow(test.start, test.end) == nil
			if got != test.want {
				t.Fatalf("validateDrinkingWindow() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestValidateDrinkingWindowUpdates(t *testing.T) {
	if err := validateDrinkingWindowUpdates(map[string]interface{}{
		"drinking_window_start": float64(2027),
		"drinking_window_end":   float64(2035),
	}); err != nil {
		t.Fatalf("valid JSON values should pass: %v", err)
	}

	if err := validateDrinkingWindowUpdates(map[string]interface{}{
		"drinking_window_start": "2035",
		"drinking_window_end":   "2027",
	}); err == nil {
		t.Fatal("reversed multipart values should fail")
	}

	if err := validateDrinkingWindowUpdates(map[string]interface{}{
		"drinking_window_start": nil,
		"drinking_window_end":   float64(2035),
	}); err != nil {
		t.Fatalf("partial nullable values should pass: %v", err)
	}
}
