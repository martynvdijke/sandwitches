package utils

import "strings"

type Translations map[string]map[string]string

var messages = Translations{
	"en": {},
	"nl": {
		"Home":                    "Startpagina",
		"Favorites":               "Favorieten",
		"Community":               "Gemeenschap",
		"API Docs":                "API Documentatie",
		"RSS Feed":                "RSS Feed",
		"Admin Dashboard":         "Beheerderspaneel",
		"Login":                   "Inloggen",
		"Sign up":                 "Registreren",
		"Logout":                  "Uitloggen",
		"Settings":                "Instellingen",
		"Admin":                   "Beheer",
		"Your Favorites":          "Jouw Favorieten",
		"The recipes you love the most.": "De recepten waar je het meest van houdt.",
		"Submit a Recipe":         "Voeg een Recept Toe",
		"Title":                   "Titel",
		"Description":             "Beschrijving",
		"Ingredients":             "Ingrediënten",
		"Instructions":            "Instructies",
		"Tags (comma separated)":  "Tags (komma gescheiden)",
		"Servings":                "Porties",
		"Price":                   "Prijs",
		"Submit Recipe":           "Recept Indienen",
		"Community Recipes":       "Recepten uit de Gemeenschap",
		"Back to all":             "Terug naar alles",
		"Tags":                    "Tags",
		"Rating":                  "Beoordeling",
		"vote":                    "stem",
		"votes":                   "stemmen",
		"Your rating":             "Jouw beoordeling",
		"Add a comment (optional)":"Voeg een opmerking toe (optioneel)",
		"Comment":                 "Opmerking",
		"Submit Rating":           "Beoordeling Indienen",
		"Log in":                  "Inloggen",
		"to rate this recipe.":    "om dit recept te beoordelen.",
		"No description yet.":      "Nog geen beschrijving.",
		"Your Profile":            "Jouw Profiel",
		"Edit your profile":       "Bewerk je profiel",
		"First name":              "Voornaam",
		"Last name":               "Achternaam",
		"Email":                   "E-mail",
		"Bio":                     "Bio",
		"Save changes":            "Wijzigingen opslaan",
		"Order History":           "Bestelgeschiedenis",
		"All Statuses":            "Alle Statussen",
		"Filter by Status":        "Filter op Status",
		"Sort by":                 "Sorteer op",
		"Newest First":            "Nieuwste Eerst",
		"Oldest First":            "Oudste Eerst",
		"Pending":                 "In Afwachting",
		"Preparing":               "In Bereiding",
		"Made":                    "Gemaakt",
		"Shipped":                 "Verzonden",
		"Completed":               "Voltooid",
		"Cancelled":               "Geannuleerd",
		"Your Shopping Cart":      "Jouw Winkelwagen",
		"Item":                    "Item",
		"Quantity":                "Aantal",
		"Summary":                 "Overzicht",
		"Total":                   "Totaal",
		"Checkout":                "Afrekenen",
		"Continue Shopping":       "Verder Winkelen",
		"Your cart is empty.":     "Je winkelwagen is leeg.",
		"Start Shopping":          "Begin met Winkelen",
		"User Settings":           "Gebruikersinstellingen",
		"Language":                "Taal",
		"Theme":                   "Thema",
		"Light":                   "Licht",
		"Dark":                    "Donker",
		"Save Settings":           "Instellingen Opslaan",
		"Recipes":                 "Recepten",
		"Users":                   "Gebruikers",
		"Recent Recipes":          "Recente Recepten",
		"Pending Approvals":       "In Afwachting van Goedkeuring",
		"Approve":                 "Goedkeuren",
		"Add Recipe":              "Recept Toevoegen",
		"Edit Recipe":             "Recept Bewerken",
		"Delete":                  "Verwijderen",
		"Edit":                    "Bewerken",
		"Add Tag":                 "Tag Toevoegen",
		"Edit Tag":                "Tag Bewerken",
		"Orders":                  "Bestellingen",
		"Order Id":               "Bestelnummer",
		"Status":                  "Status",
		"Actions":                 "Acties",
		"Ratings":                 "Beoordelingen",
		"Site Settings":           "Site Instellingen",
	},
}

func T(lang, key string) string {
	if lang == "" {
		lang = "en"
	}
	if msgs, ok := messages[lang]; ok {
		if translated, ok := msgs[key]; ok {
			return translated
		}
	}
	return key
}

func NormalizeLang(acceptLang string) string {
	for _, part := range strings.Split(acceptLang, ",") {
		lang := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		if len(lang) >= 2 {
			code := strings.ToLower(lang[:2])
			if code == "nl" || code == "en" {
				return code
			}
		}
	}
	return "en"
}
