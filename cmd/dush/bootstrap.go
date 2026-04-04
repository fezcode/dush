package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"dush/internal/app"
	"dush/internal/config"
)

// Bootstrap initializes the application.
func Bootstrap() {
	// Initialize the App singleton early
	_ = app.GetApp()

	// Initialize config (runtime aliases, no file needed)
	config.InitConfig()

	// Ensure ~/.dush/ directory exists
	usr, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not get user home directory: %v\n", err)
		return
	}

	dushConfigDir := filepath.Join(usr.HomeDir, ".dush")
	if err := os.MkdirAll(dushConfigDir, 0755); err != nil {
		DebugPrint("Error creating .dush config directory: %v", err)
	}
}
