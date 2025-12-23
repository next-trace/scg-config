// Package viper contains the Viper-based implementation of the config Provider.
package viper

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/next-trace/scg-config/contract"
)

// ConfigProvider implements contract.Provider using Viper.
type ConfigProvider struct {
	v             *viper.Viper
	configFileSet bool // tracks if a config file path was explicitly set
}

// NewConfigProvider returns a new ConfigProvider instance (satisfies contract.Provider).
// It configures Viper for ENV-first operation with automatic environment variable support.
func NewConfigProvider() *ConfigProvider {
	v := viper.New()

	// Enable automatic environment variable reading (ENV-first, 12-factor compliant)
	v.AutomaticEnv()

	// Replace '.' and '-' with '_' in env var names for consistent key mapping
	// e.g., "app.name" or "app-name" will match env var "APP_NAME"
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	return &ConfigProvider{
		v:             v,
		configFileSet: false,
	}
}

// AllSettings returns the entire config as a nested map.
func (cp *ConfigProvider) AllSettings() map[string]interface{} {
	return cp.v.AllSettings()
}

// GetKey returns the value for a key (flat lookup only, for bootstrapping and tests).
func (cp *ConfigProvider) GetKey(key string) any {
	return cp.v.Get(key)
}

// IsSet checks if a config key is present (flat lookup).
func (cp *ConfigProvider) IsSet(key string) bool {
	return cp.v.IsSet(key)
}

// Set sets a key in the Viper store (for tests or live editing).
func (cp *ConfigProvider) Set(key string, value any) {
	cp.v.Set(key, value)
}

// ReadInConfig reloads from file/env if supported by Viper.
// If no config file is set, this is a no-op (environment-only mode).
// Environment variables are always read automatically via AutomaticEnv().
func (cp *ConfigProvider) ReadInConfig() error {
	// Only try to read config file if one was explicitly set
	// This implements ENV-first: environment variables work without any config file
	if !cp.configFileSet {
		// No config file configured - skip file reading (env-only mode)
		// Environment variables are still available via AutomaticEnv()
		return nil
	}

	// Config file was explicitly set - attempt to read it
	if err := cp.v.ReadInConfig(); err != nil {
		return fmt.Errorf("provider: failed to read config: %w", err)
	}

	return nil
}

// SetConfigFile sets which file to read and marks file config as enabled.
func (cp *ConfigProvider) SetConfigFile(file string) {
	cp.v.SetConfigFile(file)
	cp.configFileSet = true
}

// MergeConfigMap merges another map into config.
func (cp *ConfigProvider) MergeConfigMap(configMap map[string]interface{}) error {
	if err := cp.v.MergeConfigMap(configMap); err != nil {
		return fmt.Errorf("provider: failed to merge config map: %w", err)
	}

	return nil
}

// Provider returns the underlying Viper object for advanced use.
func (cp *ConfigProvider) Provider() any {
	return cp.v
}

// Interface assertion: this struct implements contract.Provider.
var _ contract.Provider = (*ConfigProvider)(nil)
