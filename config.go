package main

import "fmt"

// Args holds the command-line arguments.
// Example usage in a console:
//
//	go run main.go --input "C:/in" --output "C:/out" --lang "es"
type Args struct {
	Input  string `arg:"--input,required" help:"Path to the input folder"`
	Output string `arg:"--output,required" help:"Path to the output folder"`
	Lang   string `arg:"--lang" help:"Language to use (e.g. 'en' or 'es')"`
}

// Config stores the parsed input folder, output folder, and chosen language.
type Config struct {
	InputFolder  string
	OutputFolder string
	Language     string
}

// parseArgs creates a Config from the user-provided Args.
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
		InputFolder:  args.Input,
		OutputFolder: args.Output,
		Language:     lang,
	}, nil
}
