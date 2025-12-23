// Package env provides environment variable loading and normalization utilities for scg-config.
package env

import (
	"os"

	"github.com/next-trace/scg-config/configerrors"
	"github.com/next-trace/scg-config/contract"
	"github.com/next-trace/scg-config/utils"
)

// Loader loads configuration from environment variables into the provider provider.
type Loader struct {
	provider contract.Provider
}

// NewEnvLoader creates a new Loader for the given provider provider.
func NewEnvLoader(p contract.Provider) *Loader {
	return &Loader{provider: p}
}

// LoadFromEnv loads environment variables with the given prefix into the provider.
// Prefix is stripped and keys are normalized to dot notation (e.g. APP_NAME -> app.name).
func (el *Loader) LoadFromEnv(prefix string) error {
	provider := el.provider
	if provider == nil {
		return configerrors.ErrBackendProviderNotSet
	}

	prefix = utils.NormalizePrefix(prefix)

	for _, envString := range os.Environ() {
		if !utils.ShouldProcessEnv(envString, prefix) {
			continue
		}

		key, value := utils.SplitEnv(envString)
		key = utils.StripPrefix(key, prefix)
		key = utils.NormalizeEnvKey(key)

		provider.Set(key, value)
	}

	return nil
}

// GetProvider returns the Provider associated with the Loader.
//
//nolint:ireturn // returning an interface is required by the contract API
func (el *Loader) GetProvider() contract.Provider {
	return el.provider
}
