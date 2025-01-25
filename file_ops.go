package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// organizeFiles walks the input folder, determines each file's year/quarter
// from its modification time, and moves it into a subfolder in the output folder.
func organizeFiles(cfg FilesMoveConfiguration) error {
	return filepath.Walk(cfg.InputFolder, func(path string, info os.FileInfo, err error) error {
		path = strings.TrimSpace(path)
		if err != nil {
			logError("error_organizing", cfg.Language, err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if skip, skipErr := applySkipFilters(path, info, cfg); skip || skipErr != nil {
			return skipErr
		}

		targetPath, dirErr := determineTargetPath(path, info, cfg)
		if dirErr != nil {
			return dirErr
		}

		if mkErr := ensureTargetDirectory(targetPath, cfg.DryRun); mkErr != nil {
			return mkErr
		}

		if moveErr := moveFile(path, targetPath, info, cfg.DryRun); moveErr != nil {
			logMoveError(path, targetPath, cfg.Language, moveErr)
			return moveErr
		}

		if !cfg.DryRun {
			logMovedFile(path, targetPath, cfg.Language)
		}
		return nil
	})
}

func logError(msgKey, language string, err error) {
	log.Println(locMsg(msgKey, language)+": %v", err)
}

func applySkipFilters(path string, info os.FileInfo, cfg FilesMoveConfiguration) (bool, error) {
	filters := []func(string, os.FileInfo, FilesMoveConfiguration) (bool, error){
		isPathAlreadyRelocatedFilter,
		isLoggerPathFilter,
		isFilterByBeforeConfiguration,
	}

	for _, filter := range filters {
		if skip, err := filter(path, info, cfg); skip || err != nil {
			return skip, err
		}
	}
	return false, nil
}

func isPathAlreadyRelocatedFilter(path string, info os.FileInfo, cfg FilesMoveConfiguration) (bool, error) {
	skip, skipErr := isPathAlreadyRelocated(path, determineTargetPathUnsafe(path, info, cfg))
	if skipErr != nil {
		return false, skipErr
	}
	if skip {
		log.Printf(locMsg("skipping_file", cfg.Language), path)
	}
	return skip, nil
}

func isLoggerPathFilter(path string, info os.FileInfo, cfg FilesMoveConfiguration) (bool, error) {
	if isPathTheLogger(path, cfg) {
		log.Printf(locMsg("skipping_file", cfg.Language), path)
		return true, nil
	}
	return false, nil
}

func isFilterByBeforeConfiguration(path string, info os.FileInfo, cfg FilesMoveConfiguration) (bool, error) {
	if cfg.Before == nil {
		return false, nil
	}
	beforeDate, parseErr := time.Parse("2006-01-02", *cfg.Before)
	if parseErr != nil {
		return false, fmt.Errorf("invalid 'before' date format: %w", parseErr)
	}
	isFiltered := info.ModTime().After(beforeDate)
	if isFiltered {
		log.Printf("[INFO] Skipping file: '%s'. Reason: Modified on '%s', which is after the specified 'before' date '%s'.", path, info.ModTime().Format("2006-01-02"), *cfg.Before)
	}
	return isFiltered, nil
}

func determineTargetPath(path string, info os.FileInfo, cfg FilesMoveConfiguration) (string, error) {
	quarterDir, dirErr := buildAndEnsureTargetDir(cfg.OutputFolder, info.ModTime(), cfg)
	if dirErr != nil {
		return "", dirErr
	}
	if !cfg.PreserveStructure {
		return filepath.Join(quarterDir, info.Name()), nil
	}
	relPath, relErr := filepath.Rel(cfg.InputFolder, path)
	if relErr != nil {
		return "", fmt.Errorf("failed to determine relative path: %w", relErr)
	}
	return filepath.Join(quarterDir, relPath), nil
}

func determineTargetPathUnsafe(path string, info os.FileInfo, cfg FilesMoveConfiguration) string {
	quarterDir, _ := buildAndEnsureTargetDir(cfg.OutputFolder, info.ModTime(), cfg)
	if !cfg.PreserveStructure {
		return filepath.Join(quarterDir, info.Name())
	}
	relPath, _ := filepath.Rel(cfg.InputFolder, path)
	return filepath.Join(quarterDir, relPath)
}

func ensureTargetDirectory(targetPath string, dryRun bool) error {
	if dryRun {
		return nil
	}
	dir := filepath.Dir(targetPath)

	if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
		return fmt.Errorf("failed to create target directory for %q: %w", targetPath, mkErr)
	}
	return nil
}

func logMoveError(path, targetPath, language string, err error) {
	log.Printf(locMsg("move_error", language), path, targetPath, err)
}

func logMovedFile(path, targetPath, language string) {
	log.Printf(locMsg("moved_file", language), path, targetPath)
}

func isPathTheLogger(path string, config FilesMoveConfiguration) bool {
	loggerPath := config.Logger.Name()
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Printf("Error getting absolute path for %s: %v", path, err)
		return false
	}

	absLoggerPath, err := filepath.Abs(loggerPath)
	if err != nil {
		log.Printf("Error getting absolute logger path for %s: %v", loggerPath, err)
		return false
	}

	return absPath == absLoggerPath
}

// buildAndEnsureTargetDir determines the correct quarter/year folder, then creates
// the directory if necessary. It returns the final path where files should go.
func buildAndEnsureTargetDir(outputFolder string, modTime time.Time, cfg FilesMoveConfiguration) (string, error) {
	var dir string
	var err error
	if cfg.FolderFormat == YearThenQuarters {
		dir, err = buildQuarterFolder(outputFolder, modTime, cfg.Language)
	} else if cfg.FolderFormat == DayThenHours {
		dir, err = buildHourFolder(outputFolder, modTime)
	}
	if err != nil {
		return "", fmt.Errorf("failed to build quarter folder: %w", err)
	}

	if cfg.DryRun {
		return dir, nil
	}

	if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
		return "", fmt.Errorf("failed to create target directory %q: %w", dir, mkErr)
	}
	return dir, nil
}

// buildHourFolder constructs a directory path like:
//
//	<outputFolder>/YYYY-MM-dd/HHa
func buildHourFolder(outputFolder string, modTime time.Time) (string, error) {
	// Extract year, month, and day
	year := modTime.Year()
	month := modTime.Month()
	day := modTime.Day()

	// Format hour with AM/PM
	hourLabel := modTime.Format("03PM")

	// Ensure the formatted values are valid
	if year == 0 || int(month) < 1 || int(month) > 12 || day < 1 || day > 31 {
		return "", fmt.Errorf("invalid date in modTime: %v", modTime)
	}

	// Construct the folder path
	dayFolder := fmt.Sprintf("%d-%02d-%02d", year, month, day)
	hourFolder := hourLabel

	return filepath.Join(outputFolder, dayFolder, hourFolder), nil
}

// ensureUniquePath checks if path already exists, and if so, appends (1), (2), etc.
// until we find a free name. Returns the final path that doesn't conflict.
func ensureUniquePath(path string) (string, error) {
	if !fileExists(path) {
		return path, nil
	}

	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]

	i := 1
	for {
		// e.g. "document(1).pdf", "document(2).pdf"
		newBase := fmt.Sprintf("%s(%d)%s", name, i, ext)
		newPath := filepath.Join(dir, newBase)
		if !fileExists(newPath) {
			return newPath, nil
		}
		i++
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// In your moveFile function, before actually renaming/copying:
func moveFile(src, dst string, info os.FileInfo, dryRun bool) error {
	uniqueDst, err := ensureUniquePath(dst)
	if err != nil {
		return fmt.Errorf("error ensuring unique path: %w", err)
	}

	if dryRun {
		log.Printf("[DRY RUN] Would move: %s => %s", src, uniqueDst)
		return nil
	}

	err = os.Rename(src, uniqueDst)
	if err == nil {
		// Rename succeeded
		return nil
	}

	log.Printf("Rename failed, falling back to copy: %s => %s (err=%v)", src, uniqueDst, err)

	// Copy fallback
	if copyErr := copyFilePreserve(src, uniqueDst, info, dryRun); copyErr != nil {
		return fmt.Errorf("copy fallback failed: %w", copyErr)
	}

	// Remove the original (only if not a dry run)
	if dryRun {
		log.Printf("[DRY RUN] Would remove original: %s", src)
	} else if rmErr := os.Remove(src); rmErr != nil {
		return fmt.Errorf("failed removing original %q: %w", src, rmErr)
	}

	return nil
}

// copyFilePreserve copies src into dst, then sets mod/acc times
// to match the original file.
func copyFilePreserve(src, dst string, info os.FileInfo, dryRun bool) error {
	if dryRun {
		log.Printf("[DRY RUN] Would copy: %s => %s", src, dst)
		return nil
	}

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

func isPathAlreadyRelocated(path, targetPath string) (bool, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, fmt.Errorf("failed to get absolute path for %q: %w", path, err)
	}
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return false, fmt.Errorf("failed to get absolute output path for %q: %w", targetPath, err)
	}
	return strings.Compare(absPath, absTarget) == 0, nil
}
