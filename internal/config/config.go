package config

import (
	"sync"
)

var (
	_cfg  *Config
	_once sync.Once
)

// Config holds runtime configuration.
// Prompt settings come from the evaluator environment (set in ~/.dush/env),
// falling back to defaults from defaults.go.
type Config struct {
	Aliases map[string]string
}

// InitConfig initializes the singleton Config instance.
func InitConfig() {
	_once.Do(func() {
		_cfg = &Config{
			Aliases: make(map[string]string),
		}
	})
}

// GetConfig returns the singleton Config instance.
func GetConfig() *Config {
	if _cfg == nil {
		panic("Configuration not initialized. Call InitConfig() first.")
	}
	return _cfg
}
