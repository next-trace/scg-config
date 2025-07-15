package config

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

// Load populates the provided struct pointer with values from the current
// configuration snapshot and validates it using struct tags.
//
// The decoding respects `mapstructure` tags on the target struct. After
// decoding, fields are validated using github.com/go-playground/validator
// according to any `validate` tags present. If validation fails, a detailed
// error describing invalid fields is returned.
func (c *Config) Load(out any) error { //nolint:ireturn // returning error (an interface) is idiomatic Go
	if out == nil {
		return fmt.Errorf("config: output target is nil")
	}

	// Decode the provider settings map into the target struct.
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "mapstructure",
		Result:           out,
		WeaklyTypedInput: true,
	})
	if err != nil {
		return fmt.Errorf("config: failed to create decoder: %w", err)
	}
	if err := decoder.Decode(c.provider.AllSettings()); err != nil {
		return fmt.Errorf("config: failed to unmarshal config into struct: %w", err)
	}

	// Validate the populated struct using `validate` tags.
	v := validator.New(validator.WithRequiredStructEnabled())
	if err := v.Struct(out); err != nil {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			// Build a human-friendly error message enumerating field errors.
			msg := "config validation failed:"
			for _, fe := range verrs {
				// fe.Namespace() gives full path; fe.Field() gives field name.
				msg += fmt.Sprintf(" field '%s' failed '%s'", fe.Namespace(), fe.Tag())
				if fe.Param() != "" {
					msg += fmt.Sprintf("='%s'", fe.Param())
				}
				msg += ";"
			}
			return errors.New(msg)
		}

		// Non-typed validation error; wrap and return for debugging.
		return fmt.Errorf("config validation error: %w", err)
	}

	return nil
}
