package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func main() {
	// The user specified this path earlier.
	// We will make this configurable with flags later.
	cmd := exec.Command("go", "env", "GOBIN")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Error getting GOBIN path: %v", err)
	}
	toolsPath := strings.TrimSpace(string(output))
	if toolsPath == "" {
		log.Fatalf("Error: GOBIN path is empty. Please check your Go environment setup.")
	}

	fmt.Printf("Scanning for executables in: %s\n", toolsPath)

	executables, err := ScanDirectory(toolsPath)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if len(executables) == 0 {
		fmt.Println("No executables found.")
		return
	}

	fmt.Println("Checking for updates...")
	var wg sync.WaitGroup
	updatableToolsChan := make(chan *Tool)
	var updatableTools []*Tool

	for _, exe := range executables {
		wg.Add(1)
		go func(exe string) {
			defer wg.Done()
			toolInfo := GetToolInfo(exe)
			if toolInfo == nil {
				return
			}

			err := CheckForUpdate(toolInfo)
			if err != nil {
				log.Printf("Could not check for update for %s: %v", toolInfo.Name, err)
				return
			}

			if toolInfo.IsUpdatable {
				fmt.Printf("[UPDATE AVAILABLE] %s: %s -> %s\n", toolInfo.Name, toolInfo.CurrentVersion, toolInfo.LatestVersion)
				updatableToolsChan <- toolInfo
			} else {
				fmt.Printf("[OK] %s: %s (latest)\n", toolInfo.Name, toolInfo.CurrentVersion)
			}
		}(exe)
	}

	go func() {
		wg.Wait()
		close(updatableToolsChan)
	}()

	for tool := range updatableToolsChan {
		updatableTools = append(updatableTools, tool)
	}

	fmt.Printf("\nFound %d tools that can be updated.\n", len(updatableTools))

	if len(updatableTools) > 0 {
		fmt.Print("Do you want to update them all? (y/n): ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(input)) == "y" {
			var updateWg sync.WaitGroup
			for _, tool := range updatableTools {
				updateWg.Add(1)
				go func(t *Tool) {
					defer updateWg.Done()
					err := UpdateTool(t)
					if err != nil {
						log.Printf("Error updating %s: %v\n", t.Name, err)
					}
				}(tool)
			}
			updateWg.Wait()
			fmt.Println("\nAll updates completed.")
		} else {
			fmt.Println("Update cancelled.")
		}
	}
}
