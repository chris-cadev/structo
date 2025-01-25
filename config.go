package main

import (
	"fmt"
	"os"
	"time"

	"github.com/alexflint/go-arg"
)

type CommandLineArguments struct {
	Input             string  `arg:"--input,required" help:"Path to the input folder (required)."`
	Output            string  `arg:"--output" help:"Path to the output folder (defaults to input folder)."`
	Lang              string  `arg:"--lang" help:"Language to use (e.g., 'en' for English or 'es' for Spanish; defaults to 'en')."`
	PreserveStructure bool    `arg:"--preserve-structure" help:"Preserve subfolder structure under the quarter folder."`
	Before            *string `arg:"--before" help:"Date in YYYY-MM-DD format; files before this date will be processed."`
	NoDryRun          *bool   `arg:"--no-dry-run" help:"This will make the changes happen."`
	FolderFormat      *string `arg:"--folder-format" help:"The folder format to use when creating files and directories"`
}

type FilesMoveConfiguration struct {
	InputFolder       string
	OutputFolder      string
	Language          string
	PreserveStructure bool
	DryRun            bool
	Before            *string
	Logger            *os.File
	FolderFormat      FolderFormat
}

func parseArgs() (FilesMoveConfiguration, error) {
	var args CommandLineArguments
	arg.MustParse(&args)

	if args.Input == "" {
		return FilesMoveConfiguration{}, fmt.Errorf("invalid folders: input=%q, output=%q", args.Input, args.Output)
	}

	if args.Output == "" {
		args.Output = args.Input
	}

	if args.Lang == "" {
		args.Lang = "en"
	}

	var before *string
	if args.Before != nil {
		parsedDate, err := validateDate(*args.Before)
		if err != nil {
			return FilesMoveConfiguration{}, fmt.Errorf("invalid date format for 'before': %v", err)
		}
		before = &parsedDate
	}

	noDryRun := false
	if args.NoDryRun != nil {
		noDryRun = *args.NoDryRun
	}

	folderFormat := YearThenQuarters
	var err error = nil
	if args.FolderFormat != nil {
		folderFormat, err = ParseFolderFormat(*args.FolderFormat)
		if err != nil {
			return FilesMoveConfiguration{}, fmt.Errorf("invalid folder format: %v", err)
		}
	}

	return FilesMoveConfiguration{
		InputFolder:       args.Input,
		OutputFolder:      args.Output,
		Language:          args.Lang,
		PreserveStructure: args.PreserveStructure,
		DryRun:            !noDryRun,
		Before:            before,
		FolderFormat:      folderFormat,
	}, nil
}

func validateDate(dateStr string) (string, error) {
	const layout = "2006-01-02"
	_, err := time.Parse(layout, dateStr)
	if err != nil {
		return "", fmt.Errorf("expected format is YYYY-MM-DD")
	}
	return dateStr, nil
}
