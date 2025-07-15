// Package config provides the core configuration service and high-level API for
// loading, accessing and watching configuration data.
package config

import (
	"fmt"
	"sync"

	"github.com/next-trace/scg-config/contract"
	"github.com/next-trace/scg-config/loader/env"
	"github.com/next-trace/scg-config/loader/file"
	"github.com/next-trace/scg-config/provider/viper"
	"github.com/next-trace/scg-config/watcher"
)

// Config is the core config service, exposing only ValueAccessor API.
type Config struct {
	provider     contract.Provider
	getter       *Getter
	watcher      contract.Watcher
	fileLoader   contract.FileLoader
	envLoader    contract.EnvLoader
	watchedFiles map[string]bool
	done         chan struct{}
	mu           sync.RWMutex
}

// Option is a functional option for configuring the Config instance.
type Option func(*Config)

// WithProvider sets a custom Provider implementation on the Config.
func WithProvider(p contract.Provider) Option { return func(c *Config) { c.provider = p } }

// WithWatcher sets a custom Watcher implementation on the Config.
func WithWatcher(w contract.Watcher) Option { return func(c *Config) { c.watcher = w } }

// WithFileLoader sets a custom FileLoader implementation on the Config.
func WithFileLoader(fl contract.FileLoader) Option { return func(c *Config) { c.fileLoader = fl } }

// WithEnvLoader sets a custom EnvLoader implementation on the Config.
func WithEnvLoader(el contract.EnvLoader) Option { return func(c *Config) { c.envLoader = el } }

// New constructs a Config with optional components. Missing components default
// to Viper-based provider, file/env loaders, and a file watcher.
func New(opts ...Option) *Config {
	cfg := &Config{
		provider:     nil,
		getter:       nil,
		watcher:      nil,
		fileLoader:   nil,
		envLoader:    nil,
		watchedFiles: make(map[string]bool),
		done:         make(chan struct{}),
		mu:           sync.RWMutex{},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.provider == nil {
		cfg.provider = viper.NewConfigProvider()
	}

	if cfg.fileLoader == nil {
		cfg.fileLoader = file.NewFileLoader(cfg.provider)
	}

	if cfg.envLoader == nil {
		cfg.envLoader = env.NewEnvLoader(cfg.provider)
	}

	if cfg.watcher == nil {
		cfg.watcher = watcher.NewWatcher(nil)
	}
	// Snapshot config map for the getter
	cfg.getter = NewGetter(cfg.provider.AllSettings())

	// Set the config reference in the watcher after the config is fully constructed
	if w, ok := cfg.watcher.(*watcher.Watcher); ok {
		w.SetConfig(cfg)
	}

	return cfg
}

// Get returns the value associated with key converted to the provided KeyType.
// It supports both flat lookups and dot-notation for nested structures.
func (c *Config) Get(key string, typ contract.KeyType) (any, error) {
	return c.getter.Get(key, typ)
}

// Has reports whether the given key exists in the configuration.
func (c *Config) Has(key string) bool {
	return c.getter.HasKey(key)
}

// ReadInConfig asks the Provider to read configuration from its sources.
func (c *Config) ReadInConfig() error {
	err := c.provider.ReadInConfig()
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	return nil
}

// WatchFile registers a file path to be tracked by the watcher.
func (c *Config) WatchFile(filePath string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.watchedFiles[filePath] = true
}

// UnwatchFile unregisters a file path from being tracked by the watcher.
func (c *Config) UnwatchFile(filePath string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.watchedFiles, filePath)
}

// WatchedFiles returns the list of file paths currently being watched.
func (c *Config) WatchedFiles() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	files := make([]string, 0, len(c.watchedFiles))
	for f := range c.watchedFiles {
		files = append(files, f)
	}

	return files
}

// StartWatching registers the file with the watcher and begins watching.
func (c *Config) StartWatching(filePath string) error {
	err := c.watcher.AddFile(filePath, func() {
		_ = c.provider.ReadInConfig()
	})
	if err != nil {
		return fmt.Errorf("error starting watcher for file %s: %w", filePath, err)
	}

	return nil
}

// Close stops the watcher and releases resources held by the Config.
func (c *Config) Close() error {
	close(c.done)

	if c.watcher != nil {
		err := c.watcher.Close()
		if err != nil {
			return fmt.Errorf("error closing watcher: %w", err)
		}
	}

	return nil
}

// Provider returns the underlying provider.
//
//nolint:ireturn // returning an interface is required by the contract API
func (c *Config) Provider() contract.Provider {
	return c.provider
}

// EnvLoader returns the underlying environment loader.
//
//nolint:ireturn // returning an interface is required by the contract API
func (c *Config) EnvLoader() contract.EnvLoader {
	return c.envLoader
}

// FileLoader returns the underlying file loader.
//
//nolint:ireturn // returning an interface is required by the contract API
func (c *Config) FileLoader() contract.FileLoader {
	return c.fileLoader
}

// Watcher returns the underlying watcher instance.
//
//nolint:ireturn // returning an interface is required by the contract API
func (c *Config) Watcher() contract.Watcher {
	return c.watcher
}

// Reload reloads the configuration from the provider and updates the getter.
func (c *Config) Reload() error {
	err := c.provider.ReadInConfig()
	if err != nil {
		return fmt.Errorf("error reloading config: %w", err)
	}

	c.getter = NewGetter(c.provider.AllSettings())

	return nil
}

// --- Interface assertion: only ValueAccessor, not ValueReader! ---.
var _ contract.Config = (*Config)(nil)
