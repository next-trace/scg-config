package env_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/next-trace/scg-config/errors"
	"github.com/next-trace/scg-config/loader/env"
)

func TestEnvLoader_NilProvider_Error(t *testing.T) {
	t.Parallel()
	ldr := env.NewEnvLoader(nil)
	err := ldr.LoadFromEnv("APP")
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrBackendProviderNotSet)
}
