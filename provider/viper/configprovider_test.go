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
