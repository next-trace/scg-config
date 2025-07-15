// Package file provides file loading utilities for scg-config.
package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/next-trace/scg-config/contract"
	"github.com/next-trace/scg-config/errors"
	"github.com/next-trace/scg-config/utils"
)

// Loader loads configuration files into the provider provider.
type Loader struct {
	provider contract.Provider
}

// NewFileLoader creates a new Loader for the given provider provider.
func NewFileLoader(p contract.Provider) *Loader {
	return &Loader{provider: p}
}

// LoadFromFile loads a single configuration file into the provider.
func (l *Loader) LoadFromFile(configFile string) error {
	provider := l.provider
	if provider == nil {
		return errors.ErrBackendProviderHasNoConfig
	}

	provider.SetConfigFile(configFile)

	if err := provider.ReadInConfig(); err != nil {
		return fmt.Errorf("%w: %w", errors.ErrReadConfigFileFailed, err)
	}

	return nil
}

// LoadFromDirectory loads all supported config files from a directory.
// Files are processed in alphabetical order, with the first file loaded normally
// and subsequent files merged to preserve nested block structures.
func (l *Loader) LoadFromDirectory(dir string) error {
	provider := l.provider
	if provider == nil {
		return errors.ErrBackendProviderHasNoConfig
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("%w: %w", errors.ErrFailedReadDirectory, err)
	}

	// Filter and collect supported config files
	var configFiles []string

	for _, file := range files {
		if file.IsDir() || !utils.IsSupportedConfigFile(file.Name()) {
			continue
		}

		configFiles = append(configFiles, file.Name())
	}

	if len(configFiles) == 0 {
		return nil // No config files found, not an error
	}

	isFirst := true

	for _, fileName := range configFiles {
		path := filepath.Join(dir, fileName)

		if isFirst {
			// Load the first file normally to establish the base configuration
			provider.SetConfigFile(path)

			if err := provider.ReadInConfig(); err != nil {
				return fmt.Errorf("failed to load initial config file %s: %w", path, err)
			}

			isFirst = false
		} else {
			// For subsequent files, use a more robust merging approach
			if err := l.mergeConfigFile(path); err != nil {
				return fmt.Errorf("failed to merge config file %s: %w", path, err)
			}
		}
	}

	return nil
}

// mergeConfigFile merges a configuration file into the existing provider configuration.
// This method parses the file to a generic map and merges via the Provider interface,
// keeping this loader decoupled from any specific backend implementation.
func (l *Loader) mergeConfigFile(configFile string) error {
	// #nosec G304 -- configFile path originates from os.ReadDir(dir) and is joined via filepath.Join
	// with a whitelist of supported extensions. This read is limited to files within the specified
	// configuration directory and is considered safe in this context.
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file for merging: %w", err)
	}

	var m map[string]interface{}
	ext := strings.ToLower(filepath.Ext(configFile))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &m); err != nil {
			return fmt.Errorf("failed to parse YAML config for merging: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &m); err != nil {
			return fmt.Errorf("failed to parse JSON config for merging: %w", err)
		}
	default:
		return fmt.Errorf("unsupported config file extension %q", ext)
	}

	if err := l.provider.MergeConfigMap(m); err != nil {
		return fmt.Errorf("failed to merge configuration map: %w", err)
	}
	return nil
}

// GetProvider returns the Provider associated with the Loader.
//
//nolint:ireturn // returning an interface is required by the contract API
func (l *Loader) GetProvider() contract.Provider {
	return l.provider
}
