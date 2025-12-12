package main

import (
	_ "embed" // New import for go:embed
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"dush/cmd/dush/buildinfo" // Import the new buildinfo package
	"dush/internal/config"
)

//go:embed config.piml
var embeddedDefaultPIMLConfig string

// Bootstrap initializes the application, including loading the configuration.
func Bootstrap() {
	var configPath string

	if buildinfo.IsTestBuild() { // Use buildinfo.IsTestBuild()
		DebugPrint("Running in test mode. Using cmd/dush/config.piml")
		configPath = "cmd/dush/config.piml"
	} else {
		// Try to find ~/.dush/config.piml
		usr, err := user.Current()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not get user home directory: %v. Falling back to current directory config path.\n", err)
			configPath = "config.piml" // Fallback to current directory
		} else {
			dushConfigDir := filepath.Join(usr.HomeDir, ".dush")
			configPath = filepath.Join(dushConfigDir, "config.piml")
			DebugPrint("Attempting to load config from: %s", configPath)

			// Check if config file exists
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				DebugPrint("Config file not found at %s. Creating default config.", configPath)

				// Create ~/.dush directory if it doesn't exist
				if err := os.MkdirAll(dushConfigDir, 0755); err != nil {
					DebugPrint("Error creating .dush config directory: %v", err) // Use DebugPrint, don't panic
					// Proceeding with default config path, might panic later if dir is essential
				} else {
					// Write default config
					if err := os.WriteFile(configPath, []byte(embeddedDefaultPIMLConfig), 0644); err != nil {
						DebugPrint("Error writing default config file to %s: %v", configPath, err) // Use DebugPrint, don't panic
						// Proceeding, but config load will likely fail
					} else {
						DebugPrint("Default config file created at %s", configPath)
					}
				}
			}
		}
	}

	cfg := config.GetConfig(configPath)
	DebugPrint("Configuration loaded: UserName=%s", cfg.UserName)
	// Additional bootstrap logic can be added here
}
