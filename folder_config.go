package main

import (
	"errors"
)

type FolderFormat int

const (
	YearThenQuarters FolderFormat = iota
	DayThenHours
)

var stateName = map[FolderFormat]string{
	YearThenQuarters: "year-then-quarters",
	DayThenHours:     "day-then-hours",
}

var reverseStateName = map[string]FolderFormat{
	"year-then-quarters": YearThenQuarters,
	"año-luego-cuartos":  YearThenQuarters,
	"day-then-hours":     DayThenHours,
	"dia-luego-horas":    DayThenHours,
}

func (ss FolderFormat) String() string {
	return stateName[ss]
}

func ParseFolderFormat(input string) (FolderFormat, error) {
	if format, ok := reverseStateName[input]; ok {
		return format, nil
	}
	return 0, errors.New("invalid FolderFormat: " + input)
}
