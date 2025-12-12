package config

import (
	"fmt"
	"os"
	"sync" // Import the sync package

	"github.com/fezcode/go-piml"
)

// Global config instance and once object for singleton pattern
var (
	_cfg  *Config
	_once sync.Once
	_err  error // To store error from config loading
)

// Config holds the application's configuration.
type Config struct {
	UserName     string `piml:"user_name"`
	PromptPrefix string `piml:"prompt_prefix"`
	PromptSuffix string `piml:"prompt_suffix"`
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

	return cfg, nil
}

// GetConfig returns the singleton Config instance.
// It ensures that the configuration is loaded only once and panics if an error occurs.
func GetConfig(configPath ...string) *Config {
	_once.Do(func() {
		cp := ""
		if configPath != nil || len(configPath) > 0 {
			cp = configPath[0]
		}
		_cfg, _err = loadConfig(cp)
		if _err != nil {
			panic(fmt.Sprintf("Failed to load configuration: %v", _err))
		}
	})
	return _cfg
}
