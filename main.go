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

// Args holds the environment variable for input/output folders.
// Example:
//
//	AutoOrganizerDaemon="C:/in;C:/out"
type Args struct {
	AutoOrganizerDaemon string `arg:"env:AutoOrganizerDaemon"`
}

// Config stores the parsed input and output folder paths.
type Config struct {
	InputFolder  string
	OutputFolder string
}

func main() {
	var a Args
	arg.MustParse(&a)

	// Build our config from the environment variable
	cfg, err := parseConfig(a.AutoOrganizerDaemon)
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

	// Initial log (program start)
	log.Printf("=== Started File Organizer at %s ===", time.Now().Format(time.RFC3339))
	log.Printf("Input folder: %s", cfg.InputFolder)
	log.Printf("Output folder: %s", cfg.OutputFolder)

	// Check if the input folder is valid
	if err := checkFolderExists(cfg.InputFolder); err != nil {
		log.Fatalf("Input folder check failed: %v", err)
	}

	// Organize files
	if err := organizeFiles(cfg); err != nil {
		log.Fatalf("Error organizing files: %v", err)
	}

	log.Println("File organization complete.")
	log.Printf("=== Finished at %s ===\n", time.Now().Format(time.RFC3339))
}

// parseConfig splits the env var (e.g. "C:/in;C:/out") into a Config struct.
func parseConfig(envVal string) (Config, error) {
	parts := strings.Split(envVal, ";")
	if len(parts) < 2 {
		return Config{}, fmt.Errorf("expected two folders separated by a semicolon, got %q", envVal)
	}
	inputFolder := parts[0]
	outputFolder := parts[1]

	if inputFolder == "" || outputFolder == "" {
		return Config{}, fmt.Errorf("invalid folders: input=%q, output=%q", inputFolder, outputFolder)
	}
	return Config{
		InputFolder:  inputFolder,
		OutputFolder: outputFolder,
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
		// Skip directories but keep walking into them
		if info.IsDir() {
			return nil
		}

		// Build the target directory path (year + quarter)
		targetDir, err := buildQuarterFolder(cfg.OutputFolder, info.ModTime())
		if err != nil {
			return err
		}

		// Ensure the target directory exists
		if mkErr := os.MkdirAll(targetDir, 0755); mkErr != nil {
			return fmt.Errorf("failed to create target directory %q: %w", targetDir, mkErr)
		}

		// Move or copy the file
		targetPath := filepath.Join(targetDir, info.Name())
		if err := moveFile(path, targetPath, info); err != nil {
			log.Printf("Error moving file %q to %q: %v", path, targetPath, err)
			return err
		}

		// Log success
		log.Printf("Moved: %q => %q", path, targetPath)
		return nil
	})
}

// buildQuarterFolder constructs a directory path like:
//
//	<outputRoot>/YYYY/Q<number>_monthRange
func buildQuarterFolder(outputRoot string, modTime time.Time) (string, error) {
	year := modTime.Year()
	quarterNum, quarterLabel := quarterInfoForMonth(int(modTime.Month()))
	if quarterNum == 0 {
		return "", fmt.Errorf("invalid month %d in modTime %v", modTime.Month(), modTime)
	}
	qFolder := fmt.Sprintf("Q%d_%s", quarterNum, quarterLabel)
	return filepath.Join(outputRoot, fmt.Sprintf("%d", year), qFolder), nil
}

// quarterInfoForMonth returns the quarter number (1–4) and a label like "jan-feb-mar".
func quarterInfoForMonth(m int) (int, string) {
	switch {
	case m >= 1 && m <= 3:
		return 1, "jan-feb-mar"
	case m >= 4 && m <= 6:
		return 2, "apr-may-jun"
	case m >= 7 && m <= 9:
		return 3, "jul-aug-sep"
	case m >= 10 && m <= 12:
		return 4, "oct-nov-dec"
	default:
		return 0, ""
	}
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
// The log file name includes a timestamp for traceability, e.g. "organizer_2024-12-31_15-04-05.log".
func setupLogger(outputFolder string) (*os.File, error) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFilename := filepath.Join(outputFolder, fmt.Sprintf("organizer_%s.log", timestamp))
	fmt.Println((logFilename))

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
