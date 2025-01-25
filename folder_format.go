package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"
)

type FolderFormat int

const (
	YearThenQuarters FolderFormat = iota
	DayThenHours
)

const (
	FormatYearQuarters = "year-then-quarters"
	FormatDayHours     = "day-then-hours"
	LangSpanishYear    = "a\u00f1o-luego-cuartos"
	LangSpanishDay     = "dia-luego-horas"
)

var stateName = map[FolderFormat]string{
	YearThenQuarters: FormatYearQuarters,
	DayThenHours:     FormatDayHours,
}

var reverseStateName = map[string]FolderFormat{
	FormatYearQuarters: YearThenQuarters,
	LangSpanishYear:    YearThenQuarters,
	FormatDayHours:     DayThenHours,
	LangSpanishDay:     DayThenHours,
}

// String returns the string representation of FolderFormat.
func (ss FolderFormat) String() string {
	return stateName[ss]
}

// ParseFolderFormat parses a string into a FolderFormat.
func ParseFolderFormat(input string) (FolderFormat, error) {
	if format, ok := reverseStateName[input]; ok {
		return format, nil
	}
	return 0, fmt.Errorf("invalid FolderFormat: %s", input)
}

// createFolderFormatDirectory constructs a directory path based on the given FolderFormat.
func createFolderFormatDirectory(outputRoot string, modTime time.Time, cfg FilesMoveConfiguration) (string, error) {
	switch cfg.FolderFormat {
	case YearThenQuarters:
		return createYearThenQuartersFolder(outputRoot, modTime, cfg.Language)
	case DayThenHours:
		return createDayThenHoursFolder(outputRoot, modTime)
	default:
		return "", errors.New("unsupported FolderFormat")
	}
}

// createYearThenQuartersFolder constructs a directory path like <outputRoot>/YYYY/Q<number>_monthRange.
func createYearThenQuartersFolder(outputRoot string, modTime time.Time, lang string) (string, error) {
	year := modTime.Year()
	quarterNum, quarterLabel := quarterInfoForMonth(int(modTime.Month()), lang)
	if quarterNum == 0 {
		return "", fmt.Errorf("invalid month %d in modTime %v", modTime.Month(), modTime)
	}
	qFolder := formatQuarterFolder(quarterNum, quarterLabel)
	return filepath.Join(outputRoot, fmt.Sprintf("%d", year), qFolder), nil
}

// createDayThenHoursFolder constructs a directory path like <outputFolder>/YYYY-MM-dd/HHa.
func createDayThenHoursFolder(outputFolder string, modTime time.Time) (string, error) {
	year, month, day := modTime.Date()
	hourLabel := modTime.Format("03PM")

	if !isValidDate(year, month, day) {
		return "", fmt.Errorf("invalid date in modTime: %v", modTime)
	}

	dayFolder := fmt.Sprintf("%d-%02d-%02d", year, month, day)
	return filepath.Join(outputFolder, dayFolder, hourLabel), nil
}

// quarterInfoForMonth returns the quarter number and label based on the month and language.
func quarterInfoForMonth(month int, lang string) (int, string) {
	quarters := map[string][]string{
		"en": {"Jan-Mar", "Apr-Jun", "Jul-Sep", "Oct-Dec"},
		"es": {"Ene-Mar", "Abr-Jun", "Jul-Sep", "Oct-Dic"},
	}
	if month < 1 || month > 12 {
		return 0, ""
	}
	quarterNum := (month-1)/3 + 1
	quarterLabels := quarters[lang]
	if len(quarterLabels) == 0 {
		quarterLabels = quarters["en"]
	}
	return quarterNum, quarterLabels[quarterNum-1]
}

// formatQuarterFolder formats the quarter folder name based on quarter number and label.
func formatQuarterFolder(quarterNum int, quarterLabel string) string {
	return fmt.Sprintf("Q%d_%s", quarterNum, quarterLabel)
}

// isValidDate checks if the provided date components form a valid date.
func isValidDate(year int, month time.Month, day int) bool {
	return year > 0 && month >= 1 && month <= 12 && day >= 1 && day <= 31
}
