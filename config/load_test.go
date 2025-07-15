package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/next-trace/scg-config/config"
	"github.com/next-trace/scg-config/provider/viper"
)

type appConfig struct {
	App struct {
		Name string `mapstructure:"name" validate:"required,min=3"`
	} `mapstructure:"app"`
	Server struct {
		Port int `mapstructure:"port" validate:"required,min=1,max=65535"`
	} `mapstructure:"server"`
}

func TestConfig_Load_ValidationSuccess(t *testing.T) {
	t.Parallel()

	prov := viper.NewConfigProvider()
	prov.Set("app.name", "ValidApp")
	prov.Set("server.port", 8080)

	cfg := config.New(config.WithProvider(prov))

	var out appConfig
	require.NoError(t, cfg.Load(&out))
	require.Equal(t, "ValidApp", out.App.Name)
	require.Equal(t, 8080, out.Server.Port)
}

func TestConfig_Load_ValidationFailure(t *testing.T) {
	t.Parallel()

	prov := viper.NewConfigProvider()
	// Missing app.name; invalid port
	prov.Set("server.port", 0)

	cfg := config.New(config.WithProvider(prov))

	var out appConfig
	err := cfg.Load(&out)
	require.Error(t, err)
}

func TestConfig_Load_Nested_Decode_And_Validate(t *testing.T) {
	t.Parallel()

	prov := viper.NewConfigProvider()
	prov.Set("app.name", "NestedOK")

	cfg := config.New(config.WithProvider(prov))

	var out struct {
		App struct {
			Name string `mapstructure:"name" validate:"required"`
		} `mapstructure:"app"`
	}
	require.NoError(t, cfg.Load(&out))
	require.Equal(t, "NestedOK", out.App.Name)
}

// --- Merged from load_more_test.go ---
func TestConfig_Load_NilOut_ReturnsError(t *testing.T) {
	t.Parallel()

	cfg := config.New()
	require.Error(t, cfg.Load(nil))
}

func TestConfig_Load_NonPointer_ReturnsError(t *testing.T) {
	t.Parallel()

	prov := viper.NewConfigProvider()
	prov.Set("app.name", "x")
	cfg := config.New(config.WithProvider(prov))

	// Pass a non-pointer value; mapstructure requires a pointer target
	var out struct {
		App struct {
			Name string `mapstructure:"name"`
		} `mapstructure:"app"`
	}
	err := cfg.Load(out) // intentionally not &out
	require.Error(t, err)
}
