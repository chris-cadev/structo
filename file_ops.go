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
// organizeFiles walks the input folder, determines each file's year/quarter
// from its modification time, and moves it into a subfolder in the output folder.
func organizeFiles(cfg MovementConfiguration) error {
	return filepath.Walk(cfg.InputFolder, func(path string, info os.FileInfo, err error) error {
		path = strings.TrimSpace(path)
		if err != nil {
			log.Println(locMsg("error_organizing", cfg.Language)+": %v", err)
			return nil
		}
		if info.IsDir() {
			return nil
		}

		quarterDir, dirErr := buildAndEnsureTargetDir(cfg.OutputFolder, info.ModTime(), cfg.Language)
		if dirErr != nil {
			return dirErr
		}

		var targetPath string
		if !cfg.PreserveStructure {
			// Just place it directly in the quarter folder
			targetPath = filepath.Join(quarterDir, info.Name())
		} else {
			relPath, relErr := filepath.Rel(cfg.InputFolder, path)
			if relErr != nil {
				return fmt.Errorf("failed to determine relative path: %w", relErr)
			}
			targetPath = filepath.Join(quarterDir, relPath)
		}

		skip, skipErr := isPathAlreadyRelocated(path, targetPath)
		if skipErr != nil {
			return skipErr
		}
		if skip {
			log.Printf(locMsg("skipping_file", cfg.Language), path)
			return nil
		}

		isLogger := isPathTheLogger(path, cfg)
		if isLogger {
			log.Printf(locMsg("skipping_file", cfg.Language), path)
			return nil
		}

		if mkErr := os.MkdirAll(filepath.Dir(targetPath), 0755); mkErr != nil {
			return fmt.Errorf("failed to create target directory for %q: %w", targetPath, mkErr)
		}

		if moveErr := moveFile(path, targetPath, info); moveErr != nil {
			log.Printf(locMsg("move_error", cfg.Language), path, targetPath, moveErr)
			return moveErr
		}

		log.Printf(locMsg("moved_file", cfg.Language), path, targetPath)
		return nil
	})
}

func isPathTheLogger(path string, config MovementConfiguration) bool {
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
func moveFile(src, dst string, info os.FileInfo) error {
	uniqueDst, err := ensureUniquePath(dst)
	if err != nil {
		return fmt.Errorf("error ensuring unique path: %w", err)
	}

	err = os.Rename(src, uniqueDst)
	if err == nil {
		// Rename succeeded
		return nil
	}

	log.Printf("Rename failed, falling back to copy: %s => %s (err=%v)", src, uniqueDst, err)

	// Copy fallback
	if copyErr := copyFilePreserve(src, uniqueDst, info); copyErr != nil {
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
