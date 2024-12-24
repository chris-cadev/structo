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
