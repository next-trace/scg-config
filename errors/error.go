// Package errors provides shared error values used by the configuration system.
package errors

import "errors"

// Generic getter/config errors.
var (
	// ErrKeyNotFound indicates that a requested configuration key does not exist.
	ErrKeyNotFound = errors.New("config: key not found")
	// ErrWrongType indicates that a key exists but its value cannot be converted to the requested type.
	ErrWrongType = errors.New("config: wrong type for key")
	// ErrUnknownType indicates that an unsupported target KeyType was requested.
	ErrUnknownType = errors.New("config: unknown type for key")
)

// Loader and provider related errors.
var (
	// ErrBackendProviderNotSet is returned when an EnvLoader has no backing provider configured.
	ErrBackendProviderNotSet = errors.New("no provider provider set for environment loader")
	// ErrBackendProviderHasNoConfig is returned when a provider has no config source configured.
	ErrBackendProviderHasNoConfig = errors.New("provider provider has no config provider set")
	// ErrReadConfigFileFailed indicates that reading a configuration file failed.
	ErrReadConfigFileFailed = errors.New("failed to read configuration file")
	// ErrFailedReadDirectory indicates that reading a configuration directory failed.
	ErrFailedReadDirectory = errors.New("failed to read directory")
)

// Type assertion / conversion errors for getter helpers.
var (
	ErrNotInt           = errors.New("not an int")
	ErrNotInt32         = errors.New("not an int32")
	ErrNotInt64         = errors.New("not an int64")
	ErrNotUint          = errors.New("not a uint")
	ErrNotUint32        = errors.New("not a uint32")
	ErrNotUint64        = errors.New("not a uint64")
	ErrNotFloat32       = errors.New("not a float32")
	ErrNotFloat64       = errors.New("not a float64")
	ErrNotString        = errors.New("not a string")
	ErrNotBool          = errors.New("not a bool")
	ErrNotStringInSlice = errors.New("not a string in slice")
	ErrNotStringSlice   = errors.New("not a string slice")
	ErrNotMap           = errors.New("not a map")
	ErrNotTime          = errors.New("not a time.Time")
	ErrNotDuration      = errors.New("not a duration")
	ErrNotBytes         = errors.New("not bytes")
	ErrNotUUID          = errors.New("not a uuid")
	ErrNotURL           = errors.New("not a URL")
)
