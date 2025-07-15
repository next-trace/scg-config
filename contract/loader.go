// Package contract defines the public interfaces and shared types used across the
// configuration system.
package contract

// EnvLoader describes loading configuration from environment variables.
type EnvLoader interface {
	LoadFromEnv(prefix string) error
	GetProvider() Provider
}

// FileLoader describes loading configuration from files and directories.
type FileLoader interface {
	LoadFromFile(configFile string) error
	LoadFromDirectory(dir string) error
	GetProvider() Provider
}
