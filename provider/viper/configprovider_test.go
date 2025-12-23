package viper_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/next-trace/scg-config/provider/viper"
)

const (
	testKey   = "foo"
	testValue = "bar"
	testNum   = 42
)

func TestConfigProvider_Basic(t *testing.T) {
	t.Parallel()

	provider := viper.NewConfigProvider()
	provider.Set(testKey, testValue)

	if v := provider.GetKey(testKey); v != testValue {
		t.Errorf("Get/Set mismatch, got %v", v)
	}

	if !provider.IsSet(testKey) {
		t.Errorf("IsSet false for existing key")
	}

	all := provider.AllSettings()
	if all[testKey] != testValue {
		t.Errorf("AllSettings missing value, got %v", all)
	}
}

func TestConfigProvider_ConfigFile(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.yaml")
	yaml := []byte("foo: bar\nnum: 42")

	if err := os.WriteFile(path, yaml, 0o600); err != nil { // fixed permission to 0o600 for gosec
		t.Fatalf("write: %v", err)
	}

	provider := viper.NewConfigProvider()
	provider.SetConfigFile(path)

	if err := provider.ReadInConfig(); err != nil {
		t.Fatalf("ReadInConfig: %v", err)
	}

	if provider.GetKey(testKey) != testValue || provider.GetKey("num") != testNum {
		t.Errorf("unexpected config: foo=%v num=%v", provider.GetKey(testKey), provider.GetKey("num"))
	}
}

// --- Consolidated from configprovider_more_test.go ---
func TestConfigProvider_ReadInConfig_Error(t *testing.T) {
	t.Parallel()
	p := viper.NewConfigProvider()
	p.SetConfigFile(filepath.Join(t.TempDir(), "nope.yaml"))
	err := p.ReadInConfig()
	require.Error(t, err)
}

func TestConfigProvider_MergeConfigMap(t *testing.T) {
	t.Parallel()
	// create a base config file
	dir := t.TempDir()
	path := filepath.Join(dir, "base.yaml")
	require.NoError(t, os.WriteFile(path, []byte("app:\n  name: base\n  port: 80"), 0o600))

	p := viper.NewConfigProvider()
	p.SetConfigFile(path)
	require.NoError(t, p.ReadInConfig())

	// merge new values
	m := map[string]any{"app": map[string]any{"name": "merged", "debug": true}}
	require.NoError(t, p.MergeConfigMap(m))

	all := p.AllSettings()
	app := all["app"].(map[string]any)
	require.Equal(t, "merged", app["name"]) // overridden
	require.Equal(t, true, app["debug"])    // added
}

// --- Consolidated from provider_method_test.go ---
func TestConfigProvider_Provider_ReturnsViper(t *testing.T) {
	t.Parallel()
	p := viper.NewConfigProvider()
	require.NotNil(t, p.Provider())
}

// --- ENV-first configuration tests ---

// TestConfigProvider_EnvOnly_NoConfigFile verifies that the provider works with
// environment variables only, without any config file set (ENV-first, no panic).
func TestConfigProvider_EnvOnly_NoConfigFile(t *testing.T) {
	// Note: Cannot use t.Parallel() with t.Setenv()

	// Set environment variables
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_PORT", "8080")
	t.Setenv("APP_DEBUG", "true")

	// Create provider without setting any config file
	p := viper.NewConfigProvider()

	// ReadInConfig should not panic and should return nil (env-only mode)
	err := p.ReadInConfig()
	require.NoError(t, err, "ReadInConfig should not error in env-only mode")

	// Environment variables should be accessible via Get
	// Note: viper's AutomaticEnv requires keys to match env var names
	require.Equal(t, "TestApp", p.GetKey("app.name"))
	require.Equal(t, "8080", p.GetKey("app.port"))
	require.Equal(t, "true", p.GetKey("app.debug"))
}

// TestConfigProvider_EnvOnly_WithDotAndDashKeys verifies env key replacer
// handles both dots and dashes in config keys.
func TestConfigProvider_EnvOnly_WithDotAndDashKeys(t *testing.T) {
	// Note: Cannot use t.Parallel() with t.Setenv()

	// Set environment variables with underscores
	t.Setenv("DATABASE_HOST", "localhost")
	t.Setenv("DATABASE_MAX_CONNECTIONS", "100")

	p := viper.NewConfigProvider()

	// Keys with dots should map to env vars with underscores
	require.Equal(t, "localhost", p.GetKey("database.host"))
	require.Equal(t, "100", p.GetKey("database.max.connections"))

	// Keys with dashes should also map to env vars with underscores
	require.Equal(t, "100", p.GetKey("database-max-connections"))
}

// TestConfigProvider_MissingFile_ReturnsError verifies that when a config file
// is explicitly set but missing, ReadInConfig returns an error (not panic).
func TestConfigProvider_MissingFile_ReturnsError(t *testing.T) {
	t.Parallel()

	p := viper.NewConfigProvider()

	// Set a config file path that doesn't exist
	missingFile := filepath.Join(t.TempDir(), "missing.yaml")
	p.SetConfigFile(missingFile)

	// ReadInConfig should return an error, not panic
	err := p.ReadInConfig()
	require.Error(t, err, "ReadInConfig should error when config file is missing")
	require.Contains(t, err.Error(), "provider: failed to read config")
}

// TestConfigProvider_NoFile_ThenSetFile verifies transition from env-only to file mode.
func TestConfigProvider_NoFile_ThenSetFile(t *testing.T) {
	// Note: Cannot use t.Parallel() with t.Setenv()

	// Create a test config file
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.yaml")
	yaml := []byte("server:\n  host: filehost\n  port: 9000")
	require.NoError(t, os.WriteFile(path, yaml, 0o600))

	// Set environment variable
	t.Setenv("SERVER_HOST", "envhost")

	p := viper.NewConfigProvider()

	// Initially in env-only mode
	err := p.ReadInConfig()
	require.NoError(t, err)
	require.Equal(t, "envhost", p.GetKey("server.host"))

	// Now set config file and read it
	p.SetConfigFile(path)
	err = p.ReadInConfig()
	require.NoError(t, err)

	// File value should override (viper precedence: file > env when file is explicitly set)
	// Actually, with AutomaticEnv, env vars take precedence. Let's test a key only in file.
	require.Equal(t, 9000, p.GetKey("server.port"))
}

// TestConfigProvider_EnvOverridesFile verifies ENV variables override file values.
func TestConfigProvider_EnvOverridesFile(t *testing.T) {
	// Note: Cannot use t.Parallel() with t.Setenv()

	// Create a test config file
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.yaml")
	yaml := []byte("app:\n  name: FromFile\n  version: \"1.0\"")
	require.NoError(t, os.WriteFile(path, yaml, 0o600))

	// Set environment variable for the same key
	t.Setenv("APP_NAME", "FromEnv")

	p := viper.NewConfigProvider()
	p.SetConfigFile(path)
	err := p.ReadInConfig()
	require.NoError(t, err)

	// ENV should override file (12-factor principle)
	require.Equal(t, "FromEnv", p.GetKey("app.name"))
	// Key only in file should still work
	require.Equal(t, "1.0", p.GetKey("app.version"))
}
