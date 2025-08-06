package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// UpdateTool takes a Tool struct and runs `go install` to update it to the latest version.
func UpdateTool(tool *Tool) error {
	if tool == nil || !tool.IsUpdatable {
		return nil // Nothing to update
	}

	fmt.Printf("Updating %s...\n", tool.Name)

	// With the refactoring of `checker.go`, `tool.PackagePath` now correctly holds the full
	// package path (e.g., "github.com/go-delve/delve/cmd/dlv"), which is what
	// `go install` needs.
	cmd := exec.Command("go", "install", tool.PackagePath+"@latest")

	// We can capture and show output for more detailed feedback
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update %s: %w\nOutput: %s", tool.Name, err, string(output))
	}

	// After `go install`, the new executable is in place. In the past, we might have
	// needed to remove an old version if the path or name changed, but `go install`
	// handles overwriting the binary at the correct destination.
	// The logic below is a safeguard. It compares the base name of the package path
	// with the executable name to decide if a separate removal is needed. This check
	// is made cross-platform compatible by only trimming the ".exe" suffix on Windows.
	packageBase := filepath.Base(tool.PackagePath)
	toolBase := filepath.Base(tool.Path)
	if runtime.GOOS == "windows" {
		toolBase = strings.TrimSuffix(toolBase, ".exe")
	}

	// This condition is based on the user's request to compare the command names.
	// If the names differ, it implies a potential mismatch that might require cleanup.
	if packageBase != toolBase {
		fmt.Printf("Removing old version of %s at %s\n", tool.Name, tool.Path)
		if err := os.Remove(tool.Path); err != nil {
			// If the file doesn't exist, we can ignore the error.
			if !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove old tool %s: %w", tool.Name, err)
			}
		}
	}

	fmt.Printf("Successfully updated %s to the latest version: %s.\n", tool.Name, tool.LatestVersion)
	return nil
}
