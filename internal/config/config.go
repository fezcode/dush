package config

import (
	"fmt"
	"os"
	"strings" // New import
	"sync"

	"github.com/fezcode/go-piml"
)

// Global config instance and once object for singleton pattern
var (
	_cfg             *Config
	_once            sync.Once
	_err             error  // To store error from config loading
	_aliasConfigPath string // To store the path to the alias config file
)

// Config holds the application's configuration.
type Config struct {
	UserName     string `piml:"user_name"`
	PromptPrefix string `piml:"prompt_prefix"`
	PromptSuffix string `piml:"prompt_suffix"`
	Aliases      map[string]string
}

// loadConfig reads configuration from the specified PIML file.
// This function is now unexported and called only once by GetConfig.
func loadConfig(configPath string) (*Config, error) {
	cfg := &Config{}

	// Read the PIML file content
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file at %s: %w", configPath, err)
	}

	// Unmarshal the PIML content into the Config struct
	err = piml.Unmarshal(content, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config from %s: %w", configPath, err)
	}

	// Initialize Aliases map
	cfg.Aliases = make(map[string]string)

	return cfg, nil
}

// loadAliasConfig reads aliases from the specified PIML file.
// It returns a map of aliases or an error.
func loadAliasConfig(aliasConfigPath string) (map[string]string, error) {
	aliases := make(map[string]string)

	// Read the PIML file content
	content, err := os.ReadFile(aliasConfigPath)
	if err != nil {
		// If the file doesn't exist, return an empty map and no error
		if os.IsNotExist(err) {
			return aliases, nil
		}
		return nil, fmt.Errorf("failed to read alias config file at %s: %w", aliasConfigPath, err)
	}

	// Unmarshal the PIML content directly into the map
	err = piml.Unmarshal(content, &aliases)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal aliases from %s: %w", aliasConfigPath, err)
	}

	return aliases, nil
}

// InitConfig initializes the singleton Config instance.
// It ensures that the configuration is loaded only once.
// This function should be called early in the application lifecycle.
func InitConfig(configPath string, aliasConfigPath string) {
	_once.Do(func() {
		// Store aliasConfigPath for later use by SaveAliases
		_aliasConfigPath = aliasConfigPath

		// Load main config
		_cfg, _err = loadConfig(configPath)
		if _err != nil {
			panic(fmt.Sprintf("Failed to load main configuration: %v", _err))
		}

		// Load aliases
		aliasMap, aliasErr := loadAliasConfig(aliasConfigPath)
		if aliasErr != nil {
			// If alias config fails to load, log a warning but don't panic
			fmt.Fprintf(os.Stderr, "Warning: Failed to load alias configuration from %s: %v\n", aliasConfigPath, aliasErr)
		} else {
			// Merge loaded aliases into the main config
			for k, v := range aliasMap {
				_cfg.Aliases[k] = v
			}
		}
	})
}

// GetConfig returns the singleton Config instance.
// It panics if the configuration has not been initialized.
func GetConfig() *Config {
	if _cfg == nil {
		panic("Configuration not initialized. Call InitConfig() first.")
	}
	return _cfg
}

// SaveAliases writes the current aliases to the alias config file.
func SaveAliases() error {
	if _cfg == nil {
		return fmt.Errorf("configuration not loaded, cannot save aliases")
	}
	if _aliasConfigPath == "" {
		return fmt.Errorf("alias config path not set, cannot save aliases")
	}

	// Create a temporary map to hold aliases, quoting values if they contain spaces.
	// This is necessary because go-piml's Unmarshal expects quoted strings for values with spaces
	// if it is to unmarshal them correctly into a single map entry.
	quotedAliasesForMarshal := make(map[string]string)
	for k, v := range _cfg.Aliases {
		// Check if the value contains spaces or special characters that PIML might misinterpret
		// and wrap it in single quotes.
		if strings.ContainsAny(v, " \t") {
			quotedAliasesForMarshal[k] = fmt.Sprintf("'%s'", v)
		} else {
			quotedAliasesForMarshal[k] = v
		}
	}

	// Marshal the (potentially) quoted aliases map to PIML content
	content, err := piml.Marshal(quotedAliasesForMarshal)
	if err != nil {
		return fmt.Errorf("failed to marshal aliases: %w", err)
	}

	// Write the content to the alias config file
	err = os.WriteFile(_aliasConfigPath, content, 0644)
	if err != nil {
		return fmt.Errorf("failed to write alias config file to %s: %w", _aliasConfigPath, err)
	}

	return nil
}
