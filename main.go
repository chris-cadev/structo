package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
)

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

// checkFolderExists ensures the given folder is actually a directory.
func checkFolderExists(folderPath string) error {
	info, err := os.Stat(folderPath)
	if err != nil {
		return fmt.Errorf("folder does not exist: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %v", folderPath)
	}
	return nil
}

// organizeFiles walks the input folder, determines each file's year/quarter
// from its modification time, and moves it into a subfolder in the output folder.
func organizeFiles(cfg Config) error {
	return filepath.Walk(cfg.InputFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk error: %w", err)
		}
		if info.IsDir() {
			// Skip directories but continue walking
			return nil
		}

		// 1) Check if this file should be skipped
		skip, skipErr := skipIfInOutput(path, cfg.OutputFolder)
		if skipErr != nil {
			return skipErr
		}
		if skip {
			log.Printf(locMsg("skipping_file", cfg.Language), path)
			return nil
		}

		// 2) Build (and ensure) the target directory
		targetDir, dirErr := buildAndEnsureTargetDir(cfg.OutputFolder, info.ModTime(), cfg.Language)
		if dirErr != nil {
			return dirErr
		}

		// 3) Move (or copy) the file
		targetPath := filepath.Join(targetDir, info.Name())
		if moveErr := moveFile(path, targetPath, info); moveErr != nil {
			log.Printf(locMsg("move_error", cfg.Language), path, targetPath, moveErr)
			return moveErr
		}

		log.Printf(locMsg("moved_file", cfg.Language), path, targetPath)
		return nil
	})
}

// skipIfInOutput checks whether `path` is inside the `outputFolder`. If so, return `true` so we can skip it.
func skipIfInOutput(path, outputFolder string) (bool, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, fmt.Errorf("failed to get absolute path for %q: %w", path, err)
	}
	absOutput, err := filepath.Abs(outputFolder)
	if err != nil {
		return false, fmt.Errorf("failed to get absolute output path for %q: %w", outputFolder, err)
	}
	return strings.HasPrefix(absPath, absOutput), nil
}

// buildAndEnsureTargetDir determines the correct quarter/year folder, then creates
// the directory if necessary. It returns the final path where files should go.
func buildAndEnsureTargetDir(outputFolder string, modTime time.Time, lang string) (string, error) {
	dir, err := buildQuarterFolder(outputFolder, modTime, lang)
	if err != nil {
		return "", fmt.Errorf("failed to build quarter folder: %w", err)
	}
	if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
		return "", fmt.Errorf("failed to create target directory %q: %w", dir, mkErr)
	}
	return dir, nil
}

// buildQuarterFolder constructs a directory path like:
//
//	<outputRoot>/YYYY/Q<number>_monthRange
func buildQuarterFolder(outputRoot string, modTime time.Time, lang string) (string, error) {
	year := modTime.Year()
	quarterNum, quarterLabel := quarterInfoForMonth(int(modTime.Month()), lang)
	if quarterNum == 0 {
		return "", fmt.Errorf("invalid month %d in modTime %v", modTime.Month(), modTime)
	}
	qFolder := fmt.Sprintf("Q%d_%s", quarterNum, quarterLabel)
	return filepath.Join(outputRoot, fmt.Sprintf("%d", year), qFolder), nil
}

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

// moveFile attempts a rename to preserve metadata. If that fails (e.g., cross-drive move),
// it copies the file and sets mod times, then removes the original.
func moveFile(src, dst string, info os.FileInfo) error {
	err := os.Rename(src, dst)
	if err == nil {
		// Rename succeeded
		return nil
	}

	log.Printf("Rename failed, falling back to copy: %s => %s (err=%v)", src, dst, err)

	// Copy fallback
	if copyErr := copyFilePreserve(src, dst, info); copyErr != nil {
		return fmt.Errorf("copy fallback failed: %w", copyErr)
	}
	// Remove the original
	if rmErr := os.Remove(src); rmErr != nil {
		return fmt.Errorf("failed removing original %q: %w", src, rmErr)
	}
	return nil
}

// copyFilePreserve copies src into dst, then sets mod/acc times
// to match the original file.
func copyFilePreserve(src, dst string, info os.FileInfo) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Close to allow time changes
	srcFile.Close()
	dstFile.Close()

	// Preserve mod/access time
	modTime := info.ModTime()
	if err := os.Chtimes(dst, modTime, modTime); err != nil {
		return err
	}
	return nil
}

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

// Simple localization approach for top-level log messages
func locMsg(key, lang string) string {
	messages := map[string]map[string]string{
		"start_organizer": {
			"en": "=== Started File Organizer at %s ===",
			"es": "=== Iniciando el organizador de archivos en %s ===",
		},
		"input_folder": {
			"en": "Input folder: %s",
			"es": "Carpeta de entrada: %s",
		},
		"output_folder": {
			"en": "Output folder: %s",
			"es": "Carpeta de salida: %s",
		},
		"input_folder_invalid": {
			"en": "Input folder check failed",
			"es": "Error al verificar la carpeta de entrada",
		},
		"error_organizing": {
			"en": "Error organizing files",
			"es": "Error organizando archivos",
		},
		"file_org_complete": {
			"en": "File organization complete.",
			"es": "Organización de archivos completa.",
		},
		"finished": {
			"en": "=== Finished at %s ===",
			"es": "=== Finalizado a las %s ===",
		},
		"skipping_file": {
			"en": "Skipping file already in output folder: %s",
			"es": "Saltando archivo, ya se encuentra en carpeta de salida: %s",
		},
		"move_error": {
			"en": "Error moving file %q to %q: %v",
			"es": "Error al mover archivo %q a %q: %v",
		},
		"moved_file": {
			"en": "Moved: %q => %q",
			"es": "Movido: %q => %q",
		},
	}

	// Fallback logic: if the key or lang is missing, default to English
	if msgMap, ok := messages[key]; ok {
		if msg, ok := msgMap[lang]; ok {
			return msg
		}
		return msgMap["en"]
	}
	// If the key is unknown, fallback to a simple message in English
	return fmt.Sprintf("Missing translation for key=%q", key)
}
