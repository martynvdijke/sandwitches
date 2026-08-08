package utils

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

var commonUnits = map[string]bool{
	"cup": true, "cups": true, "oz": true, "ounce": true, "ounces": true,
	"g": true, "gram": true, "grams": true, "kg": true, "kilogram": true, "kilograms": true,
	"lb": true, "lbs": true, "pound": true, "pounds": true,
	"ml": true, "milliliter": true, "milliliters": true,
	"l": true, "liter": true, "liters": true,
	"tsp": true, "teaspoon": true, "teaspoons": true,
	"tbsp": true, "tablespoon": true, "tablespoons": true,
	"pinch": true, "pinches": true, "slice": true, "slices": true,
	"clove": true, "cloves": true, "large": true, "medium": true, "small": true,
	"can": true, "cans": true, "package": true, "packages": true,
	"piece": true, "pieces": true, "dash": true, "dashes": true,
	"sprig": true, "sprigs": true, "to taste": true,
}

type ParsedIngredient struct {
	Quantity     float64
	Unit         string
	Name         string
	OriginalLine string
}

var quantityRe = regexp.MustCompile(`^\s*(\d+\s+\d+/\d+|\d+/\d+|\d+\.?\d*|\.\d+)\s*(.*)`)

func ParseIngredientLines(text string) []ParsedIngredient {
	lines := strings.Split(text, "\n")
	var result []ParsedIngredient
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		result = append(result, parseLine(line))
	}
	return result
}

func parseLine(original string) ParsedIngredient {
	result := ParsedIngredient{OriginalLine: original, Name: original}
	remaining := original

	matches := quantityRe.FindStringSubmatch(original)
	if matches != nil {
		qty := parseFraction(matches[1])
		if qty != 0 {
			result.Quantity = qty
		}
		remaining = strings.TrimSpace(matches[2])
	}

	if result.Quantity != 0 || remaining != original {
		unit := findUnit(remaining)
		if unit != "" {
			result.Unit = unit
			remaining = strings.TrimSpace(strings.TrimPrefix(strings.ToLower(remaining), strings.ToLower(unit)))
		}
	}

	if remaining != "" {
		result.Name = remaining
	} else if result.Quantity == 0 && result.Unit == "" {
		result.Name = original
	}

	return result
}

func parseFraction(s string) float64 {
	s = strings.TrimSpace(s)
	if strings.Contains(s, " ") && strings.Contains(s, "/") {
		parts := strings.SplitN(s, " ", 2)
		whole, _ := parseNum(parts[0])
		fracParts := strings.SplitN(parts[1], "/", 2)
		num, _ := parseNum(fracParts[0])
		den, _ := parseNum(fracParts[1])
		if den != 0 {
			return whole + num/den
		}
	}
	if strings.Contains(s, "/") {
		parts := strings.SplitN(s, "/", 2)
		num, _ := parseNum(parts[0])
		den, _ := parseNum(parts[1])
		if den != 0 {
			return num / den
		}
	}
	v, _ := parseNum(s)
	return v
}

func parseNum(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(strings.TrimSpace(s), "%f", &f)
	return f, err
}

func findUnit(s string) string {
	s = strings.ToLower(s)
	words := strings.Fields(s)
	if len(words) >= 2 && commonUnits[words[0]+" "+words[1]] {
		return words[0] + " " + words[1]
	}
	if len(words) >= 1 && commonUnits[words[0]] {
		return words[0]
	}
	return ""
}

func ScaleIngredient(p ParsedIngredient, current, target float64) ParsedIngredient {
	if p.Quantity == 0 || current == 0 {
		return p
	}
	factor := target / current
	p.Quantity = p.Quantity * factor
	return p
}

func FormatScaledIngredient(p ParsedIngredient) string {
	if p.Quantity == 0 {
		return p.OriginalLine
	}

	qtyStr := formatQuantity(p.Quantity)

	if p.Unit == "" {
		if p.Name != "" {
			return qtyStr + " " + p.Name
		}
		return qtyStr
	}

	if strings.HasPrefix(strings.ToLower(p.Unit), "g") {
		return qtyStr + p.Unit + " " + p.Name
	}
	return qtyStr + " " + p.Unit + " " + p.Name
}

func formatQuantity(v float64) string {
	if v == math.Trunc(v) {
		return fmt.Sprintf("%d", int(v))
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", v), "0"), ".")
}
