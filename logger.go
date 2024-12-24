package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// setupLogger opens a log file in the output folder and configures Go's logger to write there.
// The log file name includes a timestamp for traceability, e.g. ".organizer_2024-12-31_15-04-05.log".
func setupLogger(outputFolder string) (*os.File, error) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFilename := filepath.Join(outputFolder, fmt.Sprintf(".organizer_%s.log", timestamp))

	logFile, err := os.OpenFile(logFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %q: %w", logFilename, err)
	}

	// Configure the default logger to write to this file
	log.SetOutput(logFile)
	// Include date/time, source file, and line number for traceability
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	return logFile, nil
}
