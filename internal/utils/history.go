package utils

import (
	"bufio"
	"dush/internal/config"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const maxHistorySize = 1000

var commandHistory []string
var historyMutex sync.Mutex

// getHistoryFilePath returns the resolved history file path.
func getHistoryFilePath() (string, error) {
	path := config.ShellPaths.History
	if path == "" {
		return "", fmt.Errorf("history path not resolved")
	}
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return path, nil
}

func readHistoryFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
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
		return
	}

	commandHistory = append(commandHistory, trimmedCommand)
	if len(commandHistory) > maxHistorySize {
		commandHistory = commandHistory[1:]
	}
}

// SaveHistory writes the in-memory history to the history file, merging with external changes.
func SaveHistory() {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	filePath, err := getHistoryFilePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting history file path for saving: %v\n", err)
		return
	}

	fileHistory, err := readHistoryFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading history file for merging: %v\n", err)
		fileHistory = []string{}
	}

	var mergedHistory []string
	seen := make(map[string]bool)

	for _, cmd := range fileHistory {
		trimmedCmd := strings.TrimSpace(cmd)
		if trimmedCmd != "" && !seen[trimmedCmd] {
			mergedHistory = append(mergedHistory, trimmedCmd)
			seen[trimmedCmd] = true
		}
	}

	for _, cmd := range commandHistory {
		trimmedCmd := strings.TrimSpace(cmd)
		if trimmedCmd != "" && !seen[trimmedCmd] {
			mergedHistory = append(mergedHistory, trimmedCmd)
			seen[trimmedCmd] = true
		}
	}

	if len(mergedHistory) > maxHistorySize {
		mergedHistory = mergedHistory[len(mergedHistory)-maxHistorySize:]
	}

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
	historyCopy := make([]string, len(commandHistory))
	copy(historyCopy, commandHistory)
	return historyCopy
}
