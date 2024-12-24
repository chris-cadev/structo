package main

import (
	"log"
	"os"
	"time"

	"github.com/alexflint/go-arg"
)

func main() {
	var a Args
	arg.MustParse(&a)

	// Build our config from the arguments
	cfg, err := parseArgs(a)
	if err != nil {
		// We'll temporarily log to stderr, then exit
		log.Fatalf("Error parsing config: %v", err)
	}

	// Ensure the output folder exists (or create it).
	if err := os.MkdirAll(cfg.OutputFolder, 0755); err != nil {
		log.Fatalf("Failed to create output folder: %v", err)
	}

	// Set up our logger to write to a file in the output folder
	logFile, err := setupLogger(cfg.OutputFolder)
	if err != nil {
		log.Fatalf("Could not set up logger: %v", err)
	}
	// Ensure we close the file when finished
	defer logFile.Close()

	// Initial logs (program start)
	log.Printf(locMsg("start_organizer", cfg.Language), time.Now().Format(time.RFC3339))
	log.Printf(locMsg("input_folder", cfg.Language), cfg.InputFolder)
	log.Printf(locMsg("output_folder", cfg.Language), cfg.OutputFolder)

	// Check if the input folder is valid
	if err := checkFolderExists(cfg.InputFolder); err != nil {
		log.Fatalf(locMsg("input_folder_invalid", cfg.Language)+": %v", err)
	}

	// Organize files
	if err := organizeFiles(cfg); err != nil {
		log.Fatalf(locMsg("error_organizing", cfg.Language)+": %v", err)
	}

	log.Println(locMsg("file_org_complete", cfg.Language))
	log.Printf(locMsg("finished", cfg.Language)+"\n", time.Now().Format(time.RFC3339))
}
