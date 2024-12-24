package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
)

type CommandLineArguments struct {
	Input             string `arg:"--input,required" help:"Path to the input folder"`
	Output            string `arg:"--output" help:"Path to the output folder"`
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

func parseArgs() (MovementConfiguration, error) {

	var args CommandLineArguments
	arg.MustParse(&args)
	if args.Input == "" {
		return MovementConfiguration{}, fmt.Errorf("invalid folders: input=%q, output=%q", args.Input, args.Output)
	}

	if args.Output == "" {
		args.Output = args.Input
	}

	if args.Lang == "" {
		args.Lang = "en"
	}

	return MovementConfiguration{
		InputFolder:       args.Input,
		OutputFolder:      args.Output,
		Language:          args.Lang,
		PreserveStructure: args.PreserveStructure,
	}, nil
}
