package main

import (
	_ "embed" // New import for go:embed
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"dush/cmd/dush/buildinfo"
	"dush/internal/app"
	"dush/internal/config"
)

//go:embed config.piml
var embeddedDefaultPIMLConfig string

// Bootstrap initializes the application, including loading the configuration.
func Bootstrap() {
	// Initialize the App singleton early
	_ = app.GetApp()

	var configPath string
	var aliasConfigPath string

	if buildinfo.IsTestBuild() { // Use buildinfo.IsTestBuild()
		DebugPrint("Running in test mode. Using cmd/dush/config.piml")
		configPath = "cmd/dush/config.piml"
		aliasConfigPath = "cmd/dush/alias.piml" // Assuming a test alias config path
		// For simplicity in test mode, we might not create this file if it doesn't exist.
		// If test mode requires alias functionality, we would need a default alias.piml here.
		// For now, it will attempt to load, and if not found, it will just use an empty map.
	} else {
		usr, err := user.Current()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not get user home directory: %v. Falling back to current directory config path.\n", err)
			configPath = "config.piml"     // Fallback to current directory
			aliasConfigPath = "alias.piml" // No specific alias path for fallback
		} else {
			dushConfigDir := filepath.Join(usr.HomeDir, ".dush")
			configPath = filepath.Join(dushConfigDir, "config.piml")
			aliasConfigPath = filepath.Join(dushConfigDir, "alias.piml")
			DebugPrint("Attempting to load config from: %s", configPath)
			DebugPrint("Attempting to load alias config from: %s", aliasConfigPath)

			// Create ~/.dush directory if it doesn't exist
			if err := os.MkdirAll(dushConfigDir, 0755); err != nil {
				DebugPrint("Error creating .dush config directory: %v", err)
			} else {
				// Check and create config.piml if it doesn't exist
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					DebugPrint("Config file not found at %s. Creating default config.", configPath)
					if err := os.WriteFile(configPath, []byte(embeddedDefaultPIMLConfig), 0644); err != nil {
						DebugPrint("Error writing default config file to %s: %v", configPath, err)
					} else {
						DebugPrint("Default config file created at %s", configPath)
					}
				}

				// Check and create alias.piml if it doesn't exist
				if _, err := os.Stat(aliasConfigPath); os.IsNotExist(err) {
					DebugPrint("Alias config file not found at %s. Creating empty alias config.", aliasConfigPath)
					if err := os.WriteFile(aliasConfigPath, []byte(""), 0644); err != nil {
						DebugPrint("Error writing empty alias config file to %s: %v", aliasConfigPath, err)
					} else {
						DebugPrint("Empty alias config file created at %s", aliasConfigPath)
					}
				}
			}
		}
	}

	config.InitConfig(configPath, aliasConfigPath)
	DebugPrint("Configuration initialized.")
	// Additional bootstrap logic can be added here
}
