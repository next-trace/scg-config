// Package utils provides helper functions for normalizing env keys, file checks,
// and safe type conversions used by the configuration system.
//
//revive:disable:var-naming // 'utils' is a conventional and accepted package name in this repository
package utils

import (
	"fmt"
	"math"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/next-trace/scg-config/configerrors"
	"github.com/next-trace/scg-config/contract"
)

const (
	splitEnvParts = 2
)

// NormalizeEnvKey converts an environment variable key (e.g. APP_NAME) to dot notation (e.g. app.name).
func NormalizeEnvKey(key string) string {
	return strings.ToLower(strings.ReplaceAll(key, "_", "."))
}

// NormalizePrefix prepares the prefix for env matching.
func NormalizePrefix(prefix string) string {
	prefix = strings.ToUpper(prefix)
	if prefix != "" {
		return prefix + "_"
	}

	return ""
}

// ShouldProcessEnv checks if env matches the prefix (or no prefix).
func ShouldProcessEnv(env, prefix string) bool {
	return prefix == "" || strings.HasPrefix(env, prefix)
}

// SplitEnv splits "KEY=VALUE" into key, value.
func SplitEnv(env string) (string, string) {
	parts := strings.SplitN(env, "=", splitEnvParts)
	if len(parts) != splitEnvParts {
		return env, ""
	}

	return parts[0], parts[1]
}

// StripPrefix strips prefix from key, if present.
func StripPrefix(key, prefix string) string {
	if prefix != "" && strings.HasPrefix(key, prefix) {
		return key[len(prefix):]
	}

	return key
}

// IsSupportedConfigFile returns true if the file has a supported config extension.
func IsSupportedConfigFile(filename string) bool {
	switch filepath.Ext(filename) {
	case contract.ExtYAML, contract.ExtYML, contract.ExtJSON:
		return true
	default:
		return false
	}
}

// --- Type conversion helpers with overflow checks and static errors ---

// ToInt converts val to int with range checking.
func ToInt(val any) (int, error) {
	switch value := val.(type) {
	case int:
		return value, nil
	case int64:
		if value > int64(math.MaxInt) || value < int64(math.MinInt) {
			return 0, fmt.Errorf("%w: int64 out of int range", configerrors.ErrNotInt)
		}

		converted := int(value)

		return converted, nil
	case float64:
		if value > float64(math.MaxInt) || value < float64(math.MinInt) {
			return 0, fmt.Errorf("%w: float64 out of int range", configerrors.ErrNotInt)
		}

		converted := int(value)

		return converted, nil
	case string:
		i, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("%w: %w", configerrors.ErrNotInt, err)
		}

		converted := i

		return converted, nil
	default:
		return 0, configerrors.ErrNotInt
	}
}

// ToInt32 converts val to int32 with range checking.
func ToInt32(val any) (int32, error) {
	switch value := val.(type) {
	case int32:
		return value, nil
	case int:
		if value > math.MaxInt32 || value < math.MinInt32 {
			return 0, fmt.Errorf("%w: int out of int32 range", configerrors.ErrNotInt32)
		}

		return int32(value), nil
	case string:
		// Use ParseInt with explicit bit size to avoid potential overflow converting from int
		int64Value, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("%w: %w", configerrors.ErrNotInt32, err)
		}
		// At this point the parsed value is guaranteed to fit into 32‑bits.
		return int32(int64Value), nil
	default:
		return 0, configerrors.ErrNotInt32
	}
}

// ToInt64 converts val to int64.
func ToInt64(val any) (int64, error) {
	switch value := val.(type) {
	case int64:
		return value, nil
	case int:
		if value > math.MaxInt {
			return 0, fmt.Errorf("%w: int out of int64 range", configerrors.ErrNotInt64)
		}

		return int64(value), nil
	case float64:
		return int64(value), nil
	case string:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%w: %w", configerrors.ErrNotInt64, err)
		}

		return intValue, nil
	default:
		return 0, configerrors.ErrNotInt64
	}
}

// ToUint converts val to uint with range checking and non-negative enforcement.
func ToUint(val any) (uint, error) {
	switch value := val.(type) {
	case uint:
		return value, nil
	case int:
		if value < 0 {
			return 0, fmt.Errorf("%w: int is negative", configerrors.ErrNotUint)
		}

		return uint(value), nil
	case string:
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("%w: %w", configerrors.ErrNotUint, err)
		}

		if intVal < 0 {
			return 0, fmt.Errorf("%w: string is negative", configerrors.ErrNotUint)
		}

		return uint(intVal), nil
	default:
		return 0, configerrors.ErrNotUint
	}
}

// ToUint32 converts val to uint32 with range checking.
func ToUint32(val any) (uint32, error) {
	switch value := val.(type) {
	case uint32:
		return value, nil
	case int:
		// int may be negative or exceed the maximum for uint32 on this platform
		if value < 0 || value > math.MaxUint32 {
			return 0, fmt.Errorf("%w: int out of uint32 range", configerrors.ErrNotUint32)
		}

		return uint32(value), nil
	case string:
		i, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("%w: %w", configerrors.ErrNotUint32, err)
		}

		return uint32(i), nil
	default:
		return 0, configerrors.ErrNotUint32
	}
}

// ToUint64 converts val to uint64.
func ToUint64(val any) (uint64, error) {
	switch value := val.(type) {
	case uint64:
		// A uint64 value is always within [0, math.MaxUint64], so no additional checks are needed.
		return value, nil
	case int:
		if value < 0 {
			return 0, fmt.Errorf("%w: int is negative", configerrors.ErrNotUint64)
		}

		return uint64(value), nil
	case string:
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%w: %w", configerrors.ErrNotUint64, err)
		}

		return i, nil
	default:
		return 0, configerrors.ErrNotUint64
	}
}

// ToFloat32 converts val to float32 with range checking.
func ToFloat32(val any) (float32, error) {
	switch value := val.(type) {
	case float32:
		return value, nil
	case float64:
		if value > math.MaxFloat32 || value < -math.MaxFloat32 {
			return 0, fmt.Errorf("%w: float64 out of float32 range", configerrors.ErrNotFloat32)
		}

		return float32(value), nil
	case string:
		floatVal, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return 0, fmt.Errorf("%w: %w", configerrors.ErrNotFloat32, err)
		}

		if floatVal > math.MaxFloat32 || floatVal < -math.MaxFloat32 {
			return 0, fmt.Errorf("%w: parsed float out of float32 range", configerrors.ErrNotFloat32)
		}

		return float32(floatVal), nil
	default:
		return 0, configerrors.ErrNotFloat32
	}
}

// ToFloat64 converts val to float64.
func ToFloat64(val any) (float64, error) {
	switch value := val.(type) {
	case float64:
		return value, nil
	case float32:
		return float64(value), nil
	case string:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, fmt.Errorf("%w: %w", configerrors.ErrNotFloat64, err)
		}

		return f, nil
	default:
		return 0, configerrors.ErrNotFloat64
	}
}

// ToString converts val to string.
func ToString(val any) (string, error) {
	if s, ok := val.(string); ok {
		return s, nil
	}

	return "", configerrors.ErrNotString
}

// ToBool converts val to bool.
func ToBool(val any) (bool, error) {
	switch value := val.(type) {
	case bool:
		return value, nil
	case string:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return false, fmt.Errorf("%w: %w", configerrors.ErrNotBool, err)
		}

		return b, nil
	default:
		return false, configerrors.ErrNotBool
	}
}

// ToStringSlice converts val to a slice of strings, validating each element.
func ToStringSlice(val any) ([]string, error) {
	switch value := val.(type) {
	case []string:
		return value, nil
	case []any:
		result := make([]string, len(value))

		for idx, elem := range value {
			s, ok := elem.(string)
			if !ok {
				return nil, configerrors.ErrNotStringInSlice
			}

			result[idx] = s
		}

		return result, nil
	default:
		return nil, configerrors.ErrNotStringSlice
	}
}

// ToMap converts val to map[string]any.
func ToMap(val any) (map[string]any, error) {
	if m, ok := val.(map[string]any); ok {
		return m, nil
	}

	return nil, configerrors.ErrNotMap
}

// ToTime converts val to time.Time.
func ToTime(val any) (time.Time, error) {
	if t, ok := val.(time.Time); ok {
		return t, nil
	}

	return time.Time{}, configerrors.ErrNotTime
}

// ToDuration converts val to time.Duration.
func ToDuration(val any) (time.Duration, error) {
	if d, ok := val.(time.Duration); ok {
		return d, nil
	}

	return 0, configerrors.ErrNotDuration
}

// ToBytes converts val to a byte slice.
func ToBytes(val any) ([]byte, error) {
	switch value := val.(type) {
	case []byte:
		return value, nil
	case string:
		return []byte(value), nil
	default:
		return nil, configerrors.ErrNotBytes
	}
}

// ToUUID converts val to a uuid.UUID.
func ToUUID(val any) (uuid.UUID, error) {
	switch value := val.(type) {
	case uuid.UUID:
		return value, nil
	case string:
		u, err := uuid.Parse(value)
		if err != nil {
			return uuid.Nil, fmt.Errorf("%w: %w", configerrors.ErrNotUUID, err)
		}

		return u, nil
	default:
		return uuid.Nil, configerrors.ErrNotUUID
	}
}

// ToURL converts val to a parsed *url.URL.
func ToURL(val any) (*url.URL, error) {
	switch value := val.(type) {
	case *url.URL:
		return value, nil
	case string:
		u, err := url.Parse(value)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", configerrors.ErrNotURL, err)
		}

		return u, nil
	default:
		return nil, configerrors.ErrNotURL
	}
}
