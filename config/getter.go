package config

import (
	"time"

	"github.com/next-trace/scg-config/configerrors"
	"github.com/next-trace/scg-config/contract"
	"github.com/next-trace/scg-config/dotmap"
	"github.com/next-trace/scg-config/utils"
)

// Getter provides typed accessors to configuration values backed by a
// snapshot map captured from the Provider.
type Getter struct {
	config map[string]any
}

// NewGetter creates a Getter over the provided configuration map.
func NewGetter(config map[string]any) *Getter {
	return &Getter{config: config}
}

// Get returns the value associated with key, converted to the provided KeyType.
// It first attempts a flat key lookup, then falls back to dot-notation path resolution.
func (gt *Getter) Get(key string, typ contract.KeyType) (any, error) {
	if key == "" || gt.config == nil {
		return nil, configerrors.ErrKeyNotFound
	}

	if value, ok := gt.config[key]; ok {
		result, err := tryTypeCast(value, typ)
		if err != nil {
			return nil, err
		}

		return result, nil
	}

	value := dotmap.Resolve(gt.config, key)
	if value == nil {
		return nil, configerrors.ErrKeyNotFound
	}

	value, err := tryTypeCast(value, typ)
	if err != nil {
		return nil, configerrors.ErrWrongType
	}

	return value, nil
}

// GetKey returns the raw value for key as any, or nil when not found.
func (gt *Getter) GetKey(key string) any {
	value, _ := gt.Get(key, contract.String)

	return value
}

// GetString returns the string value for key, or "" if not found/convertible.
func (gt *Getter) GetString(key string) string {
	value, _ := gt.Get(key, contract.String)
	if stringValue, ok := value.(string); ok {
		return stringValue
	}

	return ""
}

// GetInt returns the int value for key, or 0 if not found/convertible.
func (gt *Getter) GetInt(key string) int {
	value, _ := gt.Get(key, contract.Int)
	if intValue, ok := value.(int); ok {
		return intValue
	}

	return 0
}

// GetBool returns the bool value for key, or false if not found/convertible.
func (gt *Getter) GetBool(key string) bool {
	value, _ := gt.Get(key, contract.Bool)
	if boolValue, ok := value.(bool); ok {
		return boolValue
	}

	return false
}

// GetFloat64 returns the float64 value for key, or 0 if not found/convertible.
func (gt *Getter) GetFloat64(key string) float64 {
	value, _ := gt.Get(key, contract.Float64)
	if floatValue, ok := value.(float64); ok {
		return floatValue
	}

	return 0
}

// GetDuration returns the time.Duration value for key, or 0 if not found/convertible.
func (gt *Getter) GetDuration(key string) time.Duration {
	value, _ := gt.Get(key, contract.Duration)
	if duration, ok := value.(time.Duration); ok {
		return duration
	}

	return 0
}

// GetInt64 returns the int64 value for key, or 0 if not found/convertible.
func (gt *Getter) GetInt64(key string) int64 {
	value, _ := gt.Get(key, contract.Int64)
	if int64Value, ok := value.(int64); ok {
		return int64Value
	}

	return 0
}

// GetStringSlice returns a []string for key, or nil if not found/convertible.
func (gt *Getter) GetStringSlice(key string) []string {
	value, _ := gt.Get(key, contract.StringSlice)
	if slice, ok := value.([]string); ok {
		return slice
	}

	return nil
}

// GetStringMap returns a map[string]interface{} for key, or nil if not found/convertible.
func (gt *Getter) GetStringMap(key string) map[string]interface{} {
	value, _ := gt.Get(key, contract.Map)
	if mapValue, ok := value.(map[string]interface{}); ok {
		return mapValue
	}

	return nil
}

// GetStringMapString returns a map[string]string for key, or nil if not found/convertible.
func (gt *Getter) GetStringMapString(key string) map[string]string {
	value, _ := gt.Get(key, contract.Map)
	if mapValue, ok := value.(map[string]string); ok {
		return mapValue
	}

	return nil
}

// GetTime returns the time.Time value for key, or the zero time if not found/convertible.
func (gt *Getter) GetTime(key string) time.Time {
	value, _ := gt.Get(key, contract.Time)
	if timeValue, ok := value.(time.Time); ok {
		return timeValue
	}

	return time.Time{}
}

// HasKey checks if a key exists in the configuration.
// It first attempts a flat key lookup, then falls back to dot-notation path resolution.
func (gt *Getter) HasKey(key string) bool {
	if key == "" || gt.config == nil {
		return false
	}

	if _, ok := gt.config[key]; ok {
		return true
	}

	return dotmap.Resolve(gt.config, key) != nil
}

// TypeConverter defines a function that converts a value to a specific type.
type TypeConverter func(val any) (any, error)

// typeConverters maps a KeyType to its converter function and associated error type.
//
//nolint:gochecknoglobals // a global converter map simplifies lookups and avoids an overly long helper function
var typeConverters = map[contract.KeyType]struct {
	converter TypeConverter
	errorType error
}{
	contract.Int: {
		converter: func(val any) (any, error) {
			return utils.ToInt(val)
		},
		errorType: configerrors.ErrNotInt,
	},
	contract.Int32: {
		converter: func(val any) (any, error) {
			return utils.ToInt32(val)
		},
		errorType: configerrors.ErrNotInt32,
	},
	contract.Int64: {
		converter: func(val any) (any, error) {
			return utils.ToInt64(val)
		},
		errorType: configerrors.ErrNotInt64,
	},
	contract.Uint: {
		converter: func(val any) (any, error) {
			return utils.ToUint(val)
		},
		errorType: configerrors.ErrNotUint,
	},
	contract.Uint32: {
		converter: func(val any) (any, error) {
			return utils.ToUint32(val)
		},
		errorType: configerrors.ErrNotUint32,
	},
	contract.Uint64: {
		converter: func(val any) (any, error) {
			return utils.ToUint64(val)
		},
		errorType: configerrors.ErrNotUint64,
	},
	contract.Float32: {
		converter: func(val any) (any, error) {
			return utils.ToFloat32(val)
		},
		errorType: configerrors.ErrNotFloat32,
	},
	contract.Float64: {
		converter: func(val any) (any, error) {
			return utils.ToFloat64(val)
		},
		errorType: configerrors.ErrNotFloat64,
	},
	contract.String: {
		converter: func(val any) (any, error) {
			return utils.ToString(val)
		},
		errorType: configerrors.ErrNotString,
	},
	contract.Bool: {
		converter: func(val any) (any, error) {
			return utils.ToBool(val)
		},
		errorType: configerrors.ErrNotBool,
	},
	contract.StringSlice: {
		converter: func(val any) (any, error) {
			return utils.ToStringSlice(val)
		},
		errorType: configerrors.ErrNotStringSlice,
	},
	contract.Map: {
		converter: func(val any) (any, error) {
			return utils.ToMap(val)
		},
		errorType: configerrors.ErrNotMap,
	},
	contract.Time: {
		converter: func(val any) (any, error) {
			return utils.ToTime(val)
		},
		errorType: configerrors.ErrNotTime,
	},
	contract.Duration: {
		converter: func(val any) (any, error) {
			return utils.ToDuration(val)
		},
		errorType: configerrors.ErrNotDuration,
	},
	contract.Bytes: {
		converter: func(val any) (any, error) {
			return utils.ToBytes(val)
		},
		errorType: configerrors.ErrNotBytes,
	},
	contract.UUID: {
		converter: func(val any) (any, error) {
			return utils.ToUUID(val)
		},
		errorType: configerrors.ErrNotUUID,
	},
	contract.URL: {
		converter: func(val any) (any, error) {
			return utils.ToURL(val)
		},
		errorType: configerrors.ErrNotURL,
	},
}

// tryTypeCast converts a value to the specified type using a function map approach.
// This reduces cognitive complexity by eliminating the large switch statement.
func tryTypeCast(val any, typ contract.KeyType) (any, error) {
	converterInfo, exists := typeConverters[typ]
	if !exists {
		return nil, configerrors.ErrUnknownType
	}

	value, err := converterInfo.converter(val)
	if err != nil {
		return nil, converterInfo.errorType
	}

	return value, nil
}
