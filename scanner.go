package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// ScanDirectory scans the given directory path and returns a slice of paths
// to all executable files found within it.
func ScanDirectory(dirPath string) ([]string, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	var executables []string
	for _, file := range files {
		// Skip directories
		if file.IsDir() {
			continue
		}

		// On Windows, we primarily care about .exe files.
		// On Unix-like systems, we would check for the executable permission bit.
		if runtime.GOOS == "windows" {
			if filepath.Ext(file.Name()) == ".exe" {
				fullPath := filepath.Join(dirPath, file.Name())
				executables = append(executables, fullPath)
			}
		} else {
			// For Linux and macOS, check the executable bit.
			info, err := file.Info()
			if err != nil {
				// Could log this error but continue for now
				continue
			}
			if info.Mode()&0111 != 0 {
				fullPath := filepath.Join(dirPath, file.Name())
				executables = append(executables, fullPath)
			}
		}
	}

	return executables, nil
}