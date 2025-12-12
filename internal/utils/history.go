package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync" // Import sync package for mutex
)

const historyFileName = ".dush_history"
const maxHistorySize = 1000 // Limit the history to prevent excessively large files

var commandHistory []string
var historyMutex sync.Mutex // Mutex to protect commandHistory and file operations

// getHistoryFilePath returns the full path to the history file.
func getHistoryFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	dushDir := filepath.Join(homeDir, ".dush")
	// Ensure the .dush directory exists
	if _, err := os.Stat(dushDir); os.IsNotExist(err) {
		err = os.Mkdir(dushDir, 0700)
		if err != nil {
			return "", fmt.Errorf("failed to create .dush directory: %w", err)
		}
	}
	return filepath.Join(dushDir, historyFileName), nil
}

// readHistoryFile reads history from the file, returns a new slice.
func readHistoryFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil // File doesn't exist, return empty history
		}
		return nil, fmt.Errorf("error opening history file: %w", err)
	}
	defer file.Close()

	var historyFromFile []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		historyFromFile = append(historyFromFile, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading history file: %w", err)
	}
	return historyFromFile, nil
}

// LoadHistory loads command history from the history file into memory.
func LoadHistory() {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	filePath, err := getHistoryFilePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting history file path: %v\n", err)
		commandHistory = make([]string, 0)
		return
	}

	loadedHistory, err := readHistoryFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading history from file: %v\n", err)
		commandHistory = make([]string, 0)
		return
	}

	commandHistory = loadedHistory
	// Trim history if it exceeds maxHistorySize
	if len(commandHistory) > maxHistorySize {
		commandHistory = commandHistory[len(commandHistory)-maxHistorySize:]
	}
}

// AddCommand adds a command to the in-memory history.
func AddCommand(command string) {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	trimmedCommand := strings.TrimSpace(command)
	if trimmedCommand == "" {
		return // Don't add empty commands
	}

	commandHistory = append(commandHistory, trimmedCommand)
	if len(commandHistory) > maxHistorySize {
		commandHistory = commandHistory[1:] // Remove the oldest command
	}
}

// SaveHistory writes the in-memory history to the history file, merging with external changes.
// reload, merge, save strategy.
func SaveHistory() {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	filePath, err := getHistoryFilePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting history file path for saving: %v\n", err)
		return
	}

	// 1. Read existing file history
	fileHistory, err := readHistoryFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading history file for merging: %v\n", err)
		// If we can't read, we'll just write our current in-memory history (commandHistory)
		fileHistory = []string{} // Treat as empty to avoid nil issues
	}

	// 2. Combine and deduplicate
	var mergedHistory []string
	seen := make(map[string]bool)

	// Add fileHistory to mergedHistory, keeping unique and in order
	for _, cmd := range fileHistory {
		trimmedCmd := strings.TrimSpace(cmd)
		if trimmedCmd != "" && !seen[trimmedCmd] {
			mergedHistory = append(mergedHistory, trimmedCmd)
			seen[trimmedCmd] = true
		}
	}

	// Add current session's commandHistory to mergedHistory, keeping unique and in order
	for _, cmd := range commandHistory {
		trimmedCmd := strings.TrimSpace(cmd)
		if trimmedCmd != "" && !seen[trimmedCmd] {
			mergedHistory = append(mergedHistory, trimmedCmd)
			seen[trimmedCmd] = true
		}
	}

	// 3. Trim to maxHistorySize
	if len(mergedHistory) > maxHistorySize {
		mergedHistory = mergedHistory[len(mergedHistory)-maxHistorySize:]
	}

	// 4. Write mergedHistory to file
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening history file for writing: %v\n", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, cmd := range mergedHistory {
		_, err := writer.WriteString(cmd + "\n")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing command to history file: %v\n", err)
			return
		}
	}
	writer.Flush()
}

// GetHistory returns a copy of the current in-memory command history.
func GetHistory() []string {
	historyMutex.Lock()
	defer historyMutex.Unlock()
	// Return a copy to prevent external modification of the internal slice
	historyCopy := make([]string, len(commandHistory))
	copy(historyCopy, commandHistory)
	return historyCopy
}
