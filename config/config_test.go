package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/next-trace/scg-config/config"
	"github.com/next-trace/scg-config/contract"
	"github.com/next-trace/scg-config/provider/viper"
)

func TestConfig_Get(t *testing.T) {
	t.Parallel()

	prov := viper.NewConfigProvider()
	prov.Set("str.int", "420")
	prov.Set("my.int", 123)
	prov.Set("my.str", "abc")
	prov.Set("my.bool", true)

	cfg := config.New(config.WithProvider(prov))

	tests := []struct {
		name     string
		key      string
		keyType  contract.KeyType
		expected interface{}
		hasError bool
	}{
		{"str int can parsed to int", "str.int", contract.Int, 420, false},
		{"existing int", "my.int", contract.Int, 123, false},
		{"existing string", "my.str", contract.String, "abc", false},
		{"existing bool", "my.bool", contract.Bool, true, false},
		{"nonexistent returns error", "missing", contract.Int, nil, true},
		{"type mismatch returns error", "my.str", contract.Int, nil, true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := cfg.Get(testCase.key, testCase.keyType)
			if testCase.hasError {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expected, got)
			}
		})
	}
}

func TestConfig_Has(t *testing.T) {
	t.Parallel()

	prov := viper.NewConfigProvider()
	prov.Set("foo", "bar")
	prov.Set("baz", 42)
	cfg := config.New(config.WithProvider(prov))

	tests := []struct {
		name string
		key  string
		want bool
	}{
		{"existing key", "foo", true},
		{"existing key 2", "baz", true},
		{"nonexistent key", "nope", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, cfg.Has(tc.key))
		})
	}
}

// --- Merged from config_more_test.go ---

type fakeProvider struct {
	all   map[string]any
	readE error
}

func (f *fakeProvider) ReadInConfig() error                 { return f.readE }
func (f *fakeProvider) AllSettings() map[string]interface{} { return f.all }
func (f *fakeProvider) GetKey(key string) any               { return f.all[key] }
func (f *fakeProvider) Set(key string, value any)           { f.all[key] = value }
func (f *fakeProvider) IsSet(key string) bool               { _, ok := f.all[key]; return ok }
func (f *fakeProvider) Provider() any                       { return nil }
func (f *fakeProvider) SetConfigFile(string)                {}
func (f *fakeProvider) MergeConfigMap(cfg map[string]interface{}) error {
	for k, v := range cfg {
		f.all[k] = v
	}
	return nil
}

type fakeWatcher struct {
	addErr   error
	closed   bool
	files    []string
	callback func()
}

func (w *fakeWatcher) AddFile(path string, cb func()) error {
	w.files = append(w.files, path)
	w.callback = cb
	return w.addErr
}
func (w *fakeWatcher) Watch(cb func()) { cb() }
func (w *fakeWatcher) Close() error    { w.closed = true; return nil }

func TestConfig_WatchList_AddRemoveAndList(t *testing.T) {
	t.Parallel()
	prov := &fakeProvider{all: map[string]any{"k": "v"}}
	cfg := config.New(config.WithProvider(prov))

	cfg.WatchFile("a")
	cfg.WatchFile("b")
	require.ElementsMatch(t, []string{"a", "b"}, cfg.WatchedFiles())
	cfg.UnwatchFile("a")
	require.ElementsMatch(t, []string{"b"}, cfg.WatchedFiles())
}

func TestConfig_ReadInConfig_ErrorWrap(t *testing.T) {
	t.Parallel()
	prov := &fakeProvider{all: map[string]any{}, readE: errors.New("boom")}
	cfg := config.New(config.WithProvider(prov))
	err := cfg.ReadInConfig()
	require.Error(t, err)
}

func TestConfig_StartWatching_UsesWatcherAndClose(t *testing.T) {
	t.Parallel()
	w := &fakeWatcher{}
	prov := &fakeProvider{all: map[string]any{"server.port": 8080}}
	cfg := config.New(config.WithProvider(prov), config.WithWatcher(w))

	require.NoError(t, cfg.StartWatching("/tmp/file.yaml"))
	require.Equal(t, []string{"/tmp/file.yaml"}, w.files)

	// Close should close watcher without error
	require.NoError(t, cfg.Close())
	require.True(t, w.closed)
}

func TestConfig_Reload_UpdatesGetterSnapshot(t *testing.T) {
	t.Parallel()
	prov := &fakeProvider{all: map[string]any{"app": map[string]any{"name": "a"}}}
	cfg := config.New(config.WithProvider(prov))

	val1, err := cfg.Get("app.name", contract.String)
	require.NoError(t, err)
	require.Equal(t, "a", val1)

	// mutate provider and reload
	prov.all = map[string]any{"app": map[string]any{"name": "b"}}
	require.NoError(t, cfg.Reload())
	val2, err := cfg.Get("app.name", contract.String)
	require.NoError(t, err)
	require.Equal(t, "b", val2)
}

// --- Merged from close_nowatcher_test.go ---
func TestConfig_Close_NoWatcher(t *testing.T) {
	t.Parallel()
	cfg := config.New(config.WithWatcher(nil))
	require.NoError(t, cfg.Close())
}

// --- Merged from readinconfig_success_test.go ---
func TestConfig_ReadInConfig_Success(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "ok.yaml")
	require.NoError(t, os.WriteFile(path, []byte("foo: bar"), 0o600))

	cfg := config.New()
	cfg.Provider().SetConfigFile(path)
	require.NoError(t, cfg.ReadInConfig())
}
