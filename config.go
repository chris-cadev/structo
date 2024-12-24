package main

import "fmt"

type Args struct {
	Input             string `arg:"--input,required" help:"Path to the input folder"`
	Output            string `arg:"--output,required" help:"Path to the output folder"`
	Lang              string `arg:"--lang" help:"Language to use (e.g. 'en' or 'es')"`
	PreserveStructure bool   `arg:"--preserve-structure" help:"Preserve subfolder structure under the quarter folder"`
}

// Config stores the parsed input folder, output folder, and chosen language.
type Config struct {
	InputFolder       string
	OutputFolder      string
	Language          string
	PreserveStructure bool
}

func parseArgs(args Args) (Config, error) {
	if args.Input == "" || args.Output == "" {
		return Config{}, fmt.Errorf("invalid folders: input=%q, output=%q", args.Input, args.Output)
	}

	// Default language to English if not provided
	lang := args.Lang
	if lang == "" {
		lang = "en"
	}

	return Config{
		InputFolder:       args.Input,
		OutputFolder:      args.Output,
		Language:          lang,
		PreserveStructure: args.PreserveStructure,
	}, nil
}
