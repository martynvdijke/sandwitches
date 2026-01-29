import re

# Define a set of common units for better parsing
COMMON_UNITS = {
    "cup",
    "cups",
    "oz",
    "ounce",
    "ounces",
    "g",
    "gram",
    "grams",
    "kg",
    "kilogram",
    "kilograms",
    "lb",
    "lbs",
    "pound",
    "pounds",
    "ml",
    "milliliter",
    "milliliters",
    "l",
    "liter",
    "liters",
    "tsp",
    "teaspoon",
    "teaspoons",
    "tbsp",
    "tablespoon",
    "tablespoons",
    "pinch",
    "pinches",
    "slice",
    "slices",
    "clove",
    "cloves",
    "large",
    "medium",
    "small",
    "can",
    "cans",
    "package",
    "packages",
    "piece",
    "pieces",
    "dash",
    "dashes",
    "sprig",
    "sprigs",
    "to taste",
}

# Simple in-memory DB for order status tracking
ORDER_DB = {}


def parse_ingredient_line(line):
    """
    Parses a single ingredient line to extract quantity, unit, and ingredient name.
    This is a heuristic approach and may not cover all possible formats.
    Returns a dictionary with 'quantity', 'unit', 'name', and 'original_line'.
    Quantity will be a float, unit and name strings.
    If parsing fails, returns None for quantity/unit but includes original_line.
    """
    original_line = line.strip()

    quantity = None
    unit = None
    name = original_line

    # Regex to capture quantity at the beginning:
    # 1. Whole number with fraction (e.g., "1 1/2", "2 3/4")
    # 2. Simple fraction (e.g., "1/2", "3/4")
    # 3. Decimal number (e.g., "0.5", "2.0", ".75")
    # 4. Whole number (e.g., "1", "100")
    # The quantity is optional to handle lines like "Salt to taste".
    #
    # This pattern specifically tries to get the quantity part.
    # It covers:
    # - `1 1/2` (whole number followed by space and fraction)
    # - `1/2` (fraction)
    # - `1.5` (decimal)
    # - `1` (integer)
    quantity_match = re.match(
        r"^\s*(\d+\s+\d/\d|\d/\d|\d+\.?\d*|\.\d+)\s*(.*)", original_line
    )

    remaining_line = original_line

    if quantity_match:
        quantity_str = quantity_match.group(1)
        remaining_line = quantity_match.group(2).strip()

        try:
            # Handle fractions like "1 1/2" or "1/2"
            if " " in quantity_str and "/" in quantity_str:
                parts = quantity_str.split(" ", 1)
                whole = float(parts[0])
                fraction_parts = parts[1].split("/")
                numerator = float(fraction_parts[0])
                denominator = float(fraction_parts[1])
                quantity = whole + (numerator / denominator)
            elif "/" in quantity_str:
                parts = quantity_str.split("/")
                numerator = float(parts[0])
                denominator = float(parts[1])
                quantity = numerator / denominator
            else:
                quantity = float(quantity_str)
        except ValueError:
            quantity = None  # Parsing quantity failed

    # Now, try to find a unit in the remaining_line
    # We iterate through common units to find the longest match first
    found_unit = None
    for common_unit in sorted(list(COMMON_UNITS), key=len, reverse=True):
        # Check if the unit is at the beginning of the remaining_line, followed by a space or end of string
        # using word boundary to avoid partial matches (e.g., "cup" in "cupholder")
        unit_regex = r"^\s*" + re.escape(common_unit) + r"(\b|\s|$)"  # ty:ignore[invalid-argument-type]
        unit_match = re.match(unit_regex, remaining_line, re.IGNORECASE)
        if unit_match:
            found_unit = remaining_line[: unit_match.end(1)].strip()
            remaining_line = remaining_line[unit_match.end(1) :].strip()
            break  # Found the unit, stop searching

    if found_unit:
        unit = found_unit

    # The rest of the line is the name
    name = (
        remaining_line
        if remaining_line
        else original_line
        if quantity is None and unit is None
        else ""
    )

    # If quantity was not parsed, and no unit was found, then the whole line is the name
    if quantity is None and unit is None:
        name = original_line
    elif name == "" and (quantity is not None or unit is not None):
        # If we parsed a quantity or unit but no name, it means the name was part of the unit search
        # or it was implicitly part of the quantity_match that took everything.
        # This is a fallback to ensure something is in name if it's not a unit.
        if unit and original_line.lower().startswith(
            f"{quantity_str.lower() if quantity_str else ''} {unit.lower()}".strip()
        ):
            name_start_index = original_line.lower().find(unit.lower()) + len(unit)
            potential_name = original_line[name_start_index:].strip()
            if potential_name:
                name = potential_name
        elif quantity_match:
            # Fallback if name is empty and quantity was matched, it means quantity was everything
            # or the unit was consumed.
            if not unit:  # If no unit, the remaining_line should be the name
                name = remaining_line

    return {
        "quantity": quantity,
        "unit": unit,
        "name": name.strip(),
        "original_line": original_line,
    }


def scale_ingredient(parsed_ingredient, current_servings, target_servings):
    """
    Scales a parsed ingredient based on current and target servings.
    """
    if parsed_ingredient["quantity"] is None or current_servings == 0:
        # Cannot scale if quantity is unknown or current servings is zero
        return parsed_ingredient

    scale_factor = target_servings / current_servings
    scaled_quantity = parsed_ingredient["quantity"] * scale_factor

    return {
        "quantity": scaled_quantity,
        "unit": parsed_ingredient["unit"],
        "name": parsed_ingredient["name"],
        "original_line": parsed_ingredient[
            "original_line"
        ],  # Keep original line reference
    }


def format_scaled_ingredient(scaled_ingredient):
    """
    Formats a scaled ingredient back into a human-readable string.
    This will try to reconstruct the string based on the parsed components.
    """
    quantity = scaled_ingredient["quantity"]
    unit = scaled_ingredient["unit"]
    name = scaled_ingredient["name"]

    if quantity is None:
        return scaled_ingredient[
            "original_line"
        ]  # Fallback to original if quantity was not parsed

    # Simple formatting for now. Could add fraction handling later.
    if quantity == int(quantity):
        quantity_str = str(int(quantity))
    else:
        # Format to 2 decimal places, remove trailing zeros, and remove trailing dot if integer
        quantity_str = f"{quantity:.2f}".rstrip("0").rstrip(".")
        if not quantity_str:  # Handle cases like 0.0 -> ""
            quantity_str = "0"

    parts = []
    if quantity_str:
        parts.append(quantity_str)

    # Check for units that usually don't have a space before them
    if unit:
        if unit.lower() in ["g", "mg"]:  # Only g and mg are typically appended directly
            parts[-1] += unit  # Append unit directly to the last part (quantity)
        else:
            parts.append(unit)

    if name:
        parts.append(name)

    return " ".join(parts).strip()
