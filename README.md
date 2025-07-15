# SCG Config

SCG Config is a configuration library for Go that wraps spf13/viper and exposes a Laravel‑like dot notation API. Its goal is to keep configuration simple, predictable, and idiomatic while embracing Go’s conventions.

[![CI](https://github.com/next-trace/scg-config/actions/workflows/ci.yml/badge.svg)](https://github.com/next-trace/scg-config/actions/workflows/ci.yml)

## Features

SCG Config offers a concise, type‑safe API for working with configuration:

- Dot notation API – Access nested configuration values using dot syntax (e.g. `app.name` or `database.host`). Arrays can be traversed by index (e.g. `auth.roles.0`).
- Single `Get` method – Retrieve values via one method by specifying the expected type via `contract.KeyType` (e.g. `contract.String`, `contract.Int`, `contract.Bool`). The method returns the value as `any` and an error if the key is missing or cannot be converted. Use `Has` to check for existence.
- Multiple sources – Load configuration from YAML or JSON files (supported extensions: `.yaml`, `.yml`, `.json`) from a single file or an entire directory. Environment variables can also be loaded with an optional prefix. Values loaded later override earlier ones.
- Case‑insensitive keys and nested structures – Keys are normalized to lower‑case dot notation, and you can navigate arbitrarily deep maps and arrays.
- Runtime overrides – Override values at runtime by writing to the underlying provider (`cfg.Provider().Set(key, value)`) and calling `cfg.Reload()` to refresh the getter snapshot.
- Hot reloading – Watch configuration files for changes and execute a callback when a file is modified.
- Struct loading with validation – Decode the current configuration snapshot into your own struct using `mapstructure` tags and validate it using `validate` tags powered by `go-playground/validator`.
- Viper integration – Uses a Viper‑backed provider under the hood; you can also construct and pass your own `viper`‑based provider wrapper.

## Installation

```bash
go get github.com/next-trace/scg-config
```

## Usage

The central type is `*config.Config`, created via `config.New()`. After loading configuration (from files and/or environment), call `Reload()` to refresh the internal getter snapshot with the latest data.

### Loading from files and environment

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/next-trace/scg-config/config"
	"github.com/next-trace/scg-config/contract"
)

func main() {
	cfg := config.New()

	// Load all .yaml/.yml/.json files from a directory. Each file’s basename
	// becomes the top-level namespace.
	if err := cfg.FileLoader().LoadFromDirectory("./config"); err != nil {
		log.Fatalf("failed to load directory: %v", err)
	}

	// Load environment variables with prefix APP_. The prefix is stripped and
	// the rest is normalized to dot notation (e.g. APP_APP_NAME -> app.name).
	// Environment values override those from files.
	_ = os.Setenv("APP_APP_NAME", "EnvName")
	if err := cfg.EnvLoader().LoadFromEnv("APP"); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	// Refresh the getter after loading.
	if err := cfg.Reload(); err != nil {
		log.Fatalf("failed to reload config: %v", err)
	}

	// Retrieve values using contract.KeyType and cast to the appropriate Go type.
	nameAny, err := cfg.Get("app.name", contract.String)
	if err != nil {
		log.Fatalf("app.name error: %v", err)
	}
	fmt.Println("Application Name:", nameAny.(string))

	portAny, err := cfg.Get("server.port", contract.Int)
	if err != nil {
		log.Fatalf("server.port error: %v", err)
	}
	fmt.Println("Server Port:", portAny.(int))
}
```

### Programmatic overrides

To set or override configuration at runtime, write to the provider and call `Reload()`:

```go
// Override the log level
cfg.Provider().Set("app.loglevel", "debug")
_ = cfg.Reload()
val, _ := cfg.Get("app.loglevel", contract.String)
fmt.Println("New log level:", val.(string))
```

### Checking for a key

```go
if cfg.Has("feature.newFlag") {
	// enable the new feature
}
```

### Watching for changes

To react to configuration changes at runtime, register a file and a watch callback. Note: calling `Watcher().Watch(cb)` overrides callbacks for tracked files; your callback should re‑read the file and refresh the getter.

```go
// Watch a specific config file
if err := cfg.StartWatching("config/app.yaml"); err != nil {
	log.Fatal(err)
}

cfg.Watcher().Watch(func() {
	// Read updated config from the file and refresh the getter
	if err := cfg.Provider().ReadInConfig(); err != nil {
		log.Println("read error:", err)
	}
	if err := cfg.Reload(); err != nil {
		log.Println("reload error:", err)
		return
	}
	if val, err := cfg.Get("app.name", contract.String); err == nil {
		fmt.Println("Updated app.name:", val.(string))
	}
})
```

### Loading into structs with validation

Use `Config.Load(out any)` to decode the current configuration snapshot into your struct and validate fields using `validate` tags.

```go
package main

import (
	"log"

	"github.com/next-trace/scg-config/config"
)

type AppConfig struct {
	App struct {
		Name string `mapstructure:"name" validate:"required,min=3"`
	} `mapstructure:"app"`
	Server struct {
		Port int `mapstructure:"port" validate:"required,min=1,max=65535"`
	} `mapstructure:"server"`
}

func main() {
	cfg := config.New()
	// ... load from files/env and cfg.Reload()

	var out AppConfig
	if err := cfg.Load(&out); err != nil {
		log.Fatalf("config validation failed: %v", err)
	}
	// out now contains strongly-typed, validated settings
}
```

### Using the Viper provider directly

`config.WithProvider` expects a `contract.Provider`, not a raw `*viper.Viper`. Use the provided wrapper `provider/viper.ConfigProvider`:

```go
import (
	"github.com/next-trace/scg-config/config"
	vprovider "github.com/next-trace/scg-config/provider/viper"
)

p := vprovider.NewConfigProvider()
p.SetConfigFile("./config/app.yaml")
_ = p.ReadInConfig()

cfg := config.New(config.WithProvider(p))
_ = cfg.Reload()
```

## License

MIT
