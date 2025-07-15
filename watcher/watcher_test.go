package watcher_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/next-trace/scg-config/config"
	"github.com/next-trace/scg-config/watcher"
)

func TestWatcher(t *testing.T) {
	t.Parallel()

	t.Run("WatchWithCallback", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "test.yaml")

		require.NoError(t, os.WriteFile(configFile, []byte("test: value"), 0o600))

		cfg := config.New()
		watcher := cfg.Watcher()

		defer func() { _ = watcher.Close() }()

		callbackCalled := make(chan struct{}, 1)

		require.NoError(t, watcher.AddFile(configFile, func() {
			select {
			case callbackCalled <- struct{}{}:
			default:
			}
		}))

		time.Sleep(200 * time.Millisecond)

		require.NoError(t, os.WriteFile(configFile, []byte("test: modified"), 0o600))
		require.NoError(t, os.Chtimes(configFile, time.Now(), time.Now()))

		select {
		case <-callbackCalled:
			// Success
		case <-time.After(2 * time.Second):
			t.Fatal("Callback was not called within timeout")
		}
	})

	t.Run("CloseWatcher", func(t *testing.T) {
		t.Parallel()

		cfg := config.New()
		watcher := cfg.Watcher()
		require.NoError(t, watcher.Close())
	})

	t.Run("WatchNonExistentFile", func(t *testing.T) {
		t.Parallel()

		cfg := config.New()
		watcher := cfg.Watcher()

		defer func() { _ = watcher.Close() }()

		err := watcher.AddFile("/non/existent/file.yaml", func() {})
		require.Error(t, err)
	})
}

// --- Consolidated from watcher_more_test.go ---
func TestWatcher_CloseTwice_IsIdempotent(t *testing.T) {
	t.Parallel()
	cfg := config.New()
	w := cfg.Watcher()
	require.NoError(t, w.Close())
	// Closing again should be a no-op
	require.NoError(t, w.Close())
}

func TestWatcher_WatchOverridesCallbacks(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "x.yaml")
	require.NoError(t, os.WriteFile(path, []byte("a: 1"), 0o600))

	cfg := config.New()
	w := cfg.Watcher()
	defer func() { _ = w.Close() }()

	chA := make(chan struct{}, 1)
	chB := make(chan struct{}, 1)

	require.NoError(t, w.AddFile(path, func() { chA <- struct{}{} }))
	// Override all callbacks via Watch
	w.Watch(func() { chB <- struct{}{} })

	// Trigger change
	require.NoError(t, os.WriteFile(path, []byte("a: 2"), 0o600))
	require.NoError(t, os.Chtimes(path, time.Now(), time.Now()))

	select {
	case <-chB:
		// new callback fired
	case <-time.After(2 * time.Second):
		t.Fatal("new callback not called")
	}

	// Ensure old callback wasn't invoked after override
	select {
	case <-chA:
		t.Fatal("old callback should not be called after Watch override")
	default:
		// ok
	}
}

// --- Consolidated from setget_config_test.go ---
func TestWatcher_SetGetConfig(t *testing.T) {
	t.Parallel()
	w := watcher.NewWatcher(nil)
	require.Nil(t, w.GetConfig())
	cfg := config.New()
	w.SetConfig(cfg)
	require.Equal(t, cfg, w.GetConfig())
	_ = w.Close()
}
