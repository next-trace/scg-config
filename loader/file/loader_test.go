package file_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/next-trace/scg-config/config"
	"github.com/next-trace/scg-config/contract"
	"github.com/next-trace/scg-config/errors"
	"github.com/next-trace/scg-config/loader/file"
	"github.com/next-trace/scg-config/provider/viper"
)

func TestFileLoader_LoadFromFile_AllSupportedExtensions(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		ext     string
		content string
		key     string
		want    string
	}
	// Supported config file extensions and their syntax
	cases := []testCase{
		{
			name:    "yaml",
			ext:     ".yaml",
			content: "app:\n  name: scg",
			key:     "app.name",
			want:    "scg",
		},
		{
			name:    "yml",
			ext:     ".yml",
			content: "app:\n  name: scg",
			key:     "app.name",
			want:    "scg",
		},
		{
			name:    "json",
			ext:     ".json",
			content: `{"app": {"name": "scg"}}`,
			key:     "app.name",
			want:    "scg",
		},
	}

	for _, testCase := range cases {
		// capture
		t.Run("LoadFromFile_"+testCase.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()

			tmpFile := filepath.Join(tmpDir, "config"+testCase.ext)
			if err := os.WriteFile(tmpFile, []byte(testCase.content), 0o600); err != nil {
				t.Fatalf("failed to write temp config: %v", err)
			}

			provider := viper.NewConfigProvider()
			loader := file.NewFileLoader(provider)

			err := loader.LoadFromFile(tmpFile)
			if err != nil {
				t.Fatalf("LoadFromFile error: %v", err)
			}

			cfg := config.New(config.WithFileLoader(loader), config.WithProvider(provider))

			val, err := cfg.Get(testCase.key, contract.String)
			if err != nil {
				t.Fatalf("Get(%q, String) error: %v", testCase.key, err)
			}

			if val != testCase.want {
				t.Errorf("Get(%q, String) = %q, want %q", testCase.key, val, testCase.want)
			}
		})
	}
}

// --- Consolidated from loader_dir_test.go ---
func TestLoadFromDirectory_HappyPath_MergeMultipleFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// First file (alphabetically) app.yaml
	app := "app:\n  name: scg\n  log: info\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "app.yaml"), []byte(app), 0o600))
	// Second file database.json
	db := `{"database": {"host": "localhost", "port": 3306}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "database.json"), []byte(db), 0o600))

	provider := viper.NewConfigProvider()
	ldr := file.NewFileLoader(provider)
	require.NoError(t, ldr.LoadFromDirectory(dir))

	cfg := config.New(config.WithProvider(provider))
	// Ensure values from both files are present after merge
	name := provider.GetKey("app.name")
	require.Equal(t, "scg", name)
	portAny, err := cfg.Get("database.port", contract.Int)
	require.NoError(t, err)
	require.Equal(t, 3306, portAny)
}

func TestLoadFromDirectory_EmptyDirectory_NoError(t *testing.T) {
	t.Parallel()
	// Empty temp dir
	dir := t.TempDir()
	provider := viper.NewConfigProvider()
	ldr := file.NewFileLoader(provider)
	require.NoError(t, ldr.LoadFromDirectory(dir))
}

func TestLoadFromDirectory_NonExistentDirectory_Error(t *testing.T) {
	t.Parallel()
	provider := viper.NewConfigProvider()
	ldr := file.NewFileLoader(provider)
	err := ldr.LoadFromDirectory(filepath.Join(t.TempDir(), "does-not-exist"))
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrFailedReadDirectory)
}

// --- Fake provider to force MergeConfigMap error path

type fakeProvider struct{}

func (f *fakeProvider) ReadInConfig() error                         { return nil }
func (f *fakeProvider) AllSettings() map[string]interface{}         { return map[string]interface{}{} }
func (f *fakeProvider) GetKey(string) any                           { return nil }
func (f *fakeProvider) Set(string, any)                             {}
func (f *fakeProvider) IsSet(string) bool                           { return false }
func (f *fakeProvider) Provider() any                               { return nil }
func (f *fakeProvider) SetConfigFile(string)                        {}
func (f *fakeProvider) MergeConfigMap(map[string]interface{}) error { return assertErr }

type assertError string

func (e assertError) Error() string { return string(e) }

var assertErr = assertError("merge error")

func TestLoadFromDirectory_MergeError_Propagates(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// first file to be loaded normally
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.yaml"), []byte("a: 1"), 0o600))
	// second file triggers merge path; content is valid but our fake will fail merge
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.yaml"), []byte("b: 2"), 0o600))

	ldr := file.NewFileLoader(&fakeProvider{})
	err := ldr.LoadFromDirectory(dir)
	require.Error(t, err)
}

// --- Consolidated from loader_dir_error_syntax_test.go ---
func TestLoadFromDirectory_FirstFileInvalid_ReturnsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// invalid yaml content
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.yaml"), []byte(": bad"), 0o600))

	p := viper.NewConfigProvider()
	ldr := file.NewFileLoader(p)
	require.Error(t, ldr.LoadFromDirectory(dir))
}

func TestLoadFromDirectory_MergeFileInvalid_ReturnsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// valid first file
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.yaml"), []byte("foo: 1"), 0o600))
	// invalid second file (merge path)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.yaml"), []byte("::bad"), 0o600))

	p := viper.NewConfigProvider()
	ldr := file.NewFileLoader(p)
	require.Error(t, ldr.LoadFromDirectory(dir))
}

// --- Consolidated from loader_more_test.go ---
func TestFileLoader_LoadFromFile_NoProvider_Error(t *testing.T) {
	t.Parallel()
	ldr := file.NewFileLoader(nil)
	err := ldr.LoadFromFile(filepath.Join(t.TempDir(), "x.yaml"))
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrBackendProviderHasNoConfig)
}

func TestFileLoader_LoadFromDirectory_NoProvider_Error(t *testing.T) {
	t.Parallel()
	ldr := file.NewFileLoader(nil)
	err := ldr.LoadFromDirectory(t.TempDir())
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrBackendProviderHasNoConfig)
}
