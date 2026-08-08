package utils

import (
	"math"
	"testing"
)

func TestParseIngredientLines(t *testing.T) {
	text := "2 cups flour\n3 eggs\n1/2 tsp salt\n1 1/2 oz butter"
	results := ParseIngredientLines(text)

	if len(results) != 4 {
		t.Fatalf("expected 4 ingredients, got %d", len(results))
	}

	if results[0].Quantity != 2 || results[0].Unit != "cups" || results[0].Name != "flour" {
		t.Errorf("bad parse: qty=%.1f unit=%q name=%q", results[0].Quantity, results[0].Unit, results[0].Name)
	}

	if results[1].Quantity != 3 || results[1].Unit != "" || results[1].Name != "eggs" {
		t.Errorf("bad parse: qty=%.1f unit=%q name=%q", results[1].Quantity, results[1].Unit, results[1].Name)
	}

	if results[2].Unit != "tsp" {
		t.Errorf("expected unit tsp, got %q", results[2].Unit)
	}

	if math.Abs(results[3].Quantity-1.5) > 0.01 {
		t.Errorf("expected qty 1.5, got %.2f", results[3].Quantity)
	}
}

func TestScaleIngredient(t *testing.T) {
	p := ParsedIngredient{Quantity: 2, Unit: "cups", Name: "flour", OriginalLine: "2 cups flour"}
	scaled := ScaleIngredient(p, 2, 4)

	if scaled.Quantity != 4 {
		t.Errorf("expected 4, got %.2f", scaled.Quantity)
	}
}

func TestFormatScaledIngredient(t *testing.T) {
	p := ParsedIngredient{Quantity: 2.5, Unit: "cups", Name: "flour", OriginalLine: "2 cups flour"}
	result := FormatScaledIngredient(p)
	expected := "2.5 cups flour"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}

	grams := ParsedIngredient{Quantity: 250, Unit: "g", Name: "sugar", OriginalLine: "250g sugar"}
	result = FormatScaledIngredient(grams)
	if result != "250g sugar" {
		t.Errorf("expected '250g sugar', got %q", result)
	}
}

func TestParseLineNoQuantity(t *testing.T) {
	result := parseLine("Salt to taste")
	if result.Quantity != 0 || result.Unit != "" || result.Name == "" {
		t.Errorf("bad parse for no-quantity line: %+v", result)
	}
}

func TestParseFraction(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"1/2", 0.5},
		{"1 1/2", 1.5},
		{"3/4", 0.75},
		{"2", 2.0},
		{"0.5", 0.5},
	}

	for _, tt := range tests {
		result := parseFraction(tt.input)
		if math.Abs(result-tt.expected) > 0.001 {
			t.Errorf("parseFraction(%q) = %.2f, want %.2f", tt.input, result, tt.expected)
		}
	}
}
