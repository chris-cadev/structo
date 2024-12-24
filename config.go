package main

import (
	"fmt"
	"os"
)

type CommandLineArguments struct {
	Input             string `arg:"--input,required" help:"Path to the input folder"`
	Output            string `arg:"--output,required" help:"Path to the output folder"`
	Lang              string `arg:"--lang" help:"Language to use (e.g. 'en' or 'es')"`
	PreserveStructure bool   `arg:"--preserve-structure" help:"Preserve subfolder structure under the quarter folder"`
}

// MovementConfiguration stores the parsed input folder, output folder, and chosen language.
type MovementConfiguration struct {
	InputFolder       string
	OutputFolder      string
	Language          string
	PreserveStructure bool
	Logger            *os.File
}

func parseArgs(args CommandLineArguments) (MovementConfiguration, error) {
	if args.Input == "" || args.Output == "" {
		return MovementConfiguration{}, fmt.Errorf("invalid folders: input=%q, output=%q", args.Input, args.Output)
	}

	// Default language to English if not provided
	lang := args.Lang
	if lang == "" {
		lang = "en"
	}

	return MovementConfiguration{
		InputFolder:       args.Input,
		OutputFolder:      args.Output,
		Language:          lang,
		PreserveStructure: args.PreserveStructure,
	}, nil
}
