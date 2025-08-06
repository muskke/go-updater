package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Tool represents information about a Go tool.
type Tool struct {
	Name           string // e.g., "dlv.exe"
	Path           string // Full path to the executable
	PackagePath    string // e.g., "github.com/go-delve/delve/cmd/dlv"
	ModulePath     string // e.g., "github.com/go-delve/delve"
	CurrentVersion string // e.g., "v1.22.1"
	LatestVersion  string // Will be filled later
	IsUpdatable    bool   // Will be set later
}

// GetToolInfo runs `go version -m` on the given executable path and
// returns a Tool struct with the parsed information.
// If the executable is not a Go program or doesn't have module info,
// it returns nil.
func GetToolInfo(execPath string) *Tool {
	cmd := exec.Command("go", "version", "-m", execPath)
	var out bytes.Buffer
	cmd.Stdout = &out

	// We don't care about stderr for now, as non-Go programs will write to it.
	if err := cmd.Run(); err != nil {
		return nil
	}

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// The first line is: <exec_path>: <go_version>
	// We are interested in the "path" line for the package path,
	// and the "mod" line for the module path and version.
	// e.g., path    github.com/go-delve/delve/cmd/dlv
	// e.g., mod     github.com/go-delve/delve       v1.22.1 h1:...
	var pkgPath, modPath, modVersion string
	for _, line := range lines[1:] { // Start from the second line
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			switch parts[0] {
			case "path":
				pkgPath = parts[1]
			case "mod":
				modPath = parts[1]
				if len(parts) >= 3 {
					modVersion = parts[2]
				}
			}
		}
	}

	// We need both the package path and the module path to proceed.
	if pkgPath == "" || modPath == "" {
		return nil
	}

	return &Tool{
		Name:           filepath.Base(execPath),
		Path:           execPath,
		PackagePath:    pkgPath,
		ModulePath:     modPath,
		CurrentVersion: modVersion,
	}
}

// CheckForUpdate checks for a newer version of the tool by calling `go list`.
// It updates the Tool struct with the latest version if an update is available.
func CheckForUpdate(tool *Tool) error {
	if tool == nil || tool.ModulePath == "" {
		return nil // Nothing to check
	}

	// This logic attempts to find the latest version, including across major versions.
	// It first checks if the *next* major version exists (e.g., if on v3, it checks for v4).
	// If it exists, it uses that to find the latest version.
	// If not, it falls back to finding the latest within the current major version.
	pathToQuery := ""
	re := regexp.MustCompile(`^(.*)/v([2-9]\d*)$`)
	matches := re.FindStringSubmatch(tool.ModulePath)

	if len(matches) == 3 { // Path has a /vN suffix, e.g., .../v3
		basePath := matches[1]
		currentVersionNum, _ := strconv.Atoi(matches[2])
		nextMajorPath := fmt.Sprintf("%s/v%d", basePath, currentVersionNum+1)

		// Use `go list` to check for the existence of the next major version.
		if exec.Command("go", "list", "-m", nextMajorPath+"@latest").Run() == nil {
			pathToQuery = nextMajorPath // It exists, so we'll query it.
		}
	} else { // Path is v0/v1
		nextMajorPath := tool.ModulePath + "/v2"
		if exec.Command("go", "list", "-m", nextMajorPath+"@latest").Run() == nil {
			pathToQuery = nextMajorPath // v2 exists, so we'll query it.
		}
	}

	// If we didn't find a newer major version path, fall back to the current one.
	if pathToQuery == "" {
		pathToQuery = tool.ModulePath
	} else {
		// Replace the module path part of the package path with the new one to query.
		// For example, if PackagePath is "A/B" and ModulePath is "A", and we found a new
		// major version at "A/v2", we want to query "A/v2/B" to get the tool's latest version.
		// However, we only want to do this if the package path actually contains the module path.
		if strings.HasPrefix(tool.PackagePath, tool.ModulePath) {
			tool.PackagePath = strings.Replace(tool.PackagePath, tool.ModulePath, pathToQuery, 1)
		}
	}

	cmd := exec.Command("go", "list", "-m", "-u", "-json", pathToQuery+"@latest")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to check for update for %s: %w", tool.Name, err)
	}

	// Define a struct to parse the JSON output from `go list`
	var modInfo struct {
		Path      string
		Version   string
		Query     string
		Time      time.Time
		GoMod     string
		GoVersion string
	}

	if err := json.Unmarshal(out.Bytes(), &modInfo); err != nil {
		return fmt.Errorf("failed to parse json for %s: %w", tool.Name, err)
	}

	// If the version from `go list` is different from the binary's version, an update is available.
	if modInfo.Version != "" && tool.CurrentVersion != modInfo.Version {
		tool.LatestVersion = modInfo.Version
		tool.IsUpdatable = true
	}

	return nil
}
