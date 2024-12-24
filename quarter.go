package main

// quarterInfoForMonth returns the quarter number (1–4) and a localized label
// like "jan-feb-mar" (English) or "ene-feb-mar" (Spanish).
func quarterInfoForMonth(m int, lang string) (int, string) {
	var quarter int
	switch {
	case m >= 1 && m <= 3:
		quarter = 1
	case m >= 4 && m <= 6:
		quarter = 2
	case m >= 7 && m <= 9:
		quarter = 3
	case m >= 10 && m <= 12:
		quarter = 4
	default:
		return 0, ""
	}
	return quarter, translateQuarterLabel(quarter, lang)
}

// translateQuarterLabel returns the localized string for the given quarter.
// If the language or quarter is unknown, default to English.
func translateQuarterLabel(q int, lang string) string {
	quarterLabels := map[int]map[string]string{
		1: {
			"en": "JAN-FEB-MAR",
			"es": "ENE-FEB-MAR",
		},
		2: {
			"en": "APR-MAY-JUN",
			"es": "ABR-MAY-JUN",
		},
		3: {
			"en": "JUL-AUG-SEP",
			"es": "JUL-AGO-SEP",
		},
		4: {
			"en": "OCT-NOV-DEC",
			"es": "OCT-NOV-DIC",
		},
	}
	qMap, ok := quarterLabels[q]
	if !ok {
		return "unknown-quarter"
	}
	if label, ok := qMap[lang]; ok {
		return label
	}
	// Fallback to English if not found
	return qMap["en"]
}
