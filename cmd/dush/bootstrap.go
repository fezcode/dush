package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"dush/cmd/dush/buildinfo" // Import the new buildinfo package
	"dush/internal/config"
)

// Bootstrap initializes the application, including loading the configuration.
func Bootstrap() {
	var configPath string

	if buildinfo.IsTestBuild() { // Use buildinfo.IsTestBuild()
		fmt.Println("Running in test mode. Using internal/config/config.piml")
		configPath = "internal/config/config.piml"
	} else {
		// Try to find ~/.dush/config.piml
		usr, err := user.Current()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not get user home directory: %v. Falling back to default config path.\n", err)
			configPath = "config.piml" // Or provide a sensible default
		} else {
			configPath = filepath.Join(usr.HomeDir, ".dush", "config.piml")
			fmt.Printf("Attempting to load config from: %s\n", configPath)
		}
	}

	cfg := config.GetConfig(configPath)
	fmt.Printf("Configuration loaded: UserName=%s\n", cfg.UserName)
	// Additional bootstrap logic can be added here
}
