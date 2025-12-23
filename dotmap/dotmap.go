// Package dotmap provides utilities for traversing nested maps with dot notation.
package dotmap

import (
	"strconv"
	"strings"
)

// Resolve navigates a nested map using dot notation (e.g., "foo.bar.0.baz").
// Returns nil if any part of the path does not exist (case-insensitive fallback).
func Resolve(settings map[string]interface{}, path string) interface{} {
	if settings == nil || path == "" {
		return nil
	}

	parts := strings.Split(path, ".")

	// First try case-sensitive resolution
	result := resolvePath(settings, parts, false)
	if result != nil {
		return result
	}

	// If case-sensitive failed, try case-insensitive from the root
	return resolvePath(settings, parts, true)
}

// resolvePath traverses the path with the specified case sensitivity.
func resolvePath(current interface{}, parts []string, caseInsensitive bool) interface{} {
	for _, part := range parts {
		next, found := resolveStep(current, part, caseInsensitive)
		if !found {
			return nil
		}

		current = next
	}

	return current
}

// resolveStep attempts to resolve one step in the path.
func resolveStep(current interface{}, part string, caseInsensitive bool) (interface{}, bool) {
	// Handle array indexing first (same for both case modes)
	if index, err := strconv.Atoi(part); err == nil {
		return resolveArrayIndex(current, index)
	}

	// Handle map types
	switch typedMap := current.(type) {
	case map[string]interface{}:
		return resolveStringMap(typedMap, part, caseInsensitive)
	case map[string]string:
		return resolveStringStringMap(typedMap, part, caseInsensitive)
	case map[interface{}]interface{}:
		return resolveInterfaceMap(typedMap, part, caseInsensitive)
	}

	return nil, false
}

// resolveArrayIndex handles array/slice indexing for both []interface{} and []string.
func resolveArrayIndex(current interface{}, index int) (interface{}, bool) {
	switch typedSlice := current.(type) {
	case []interface{}:
		if index >= 0 && index < len(typedSlice) {
			return typedSlice[index], true
		}
	case []string:
		if index >= 0 && index < len(typedSlice) {
			return typedSlice[index], true
		}
	}

	return nil, false
}

// resolveStringMap handles map[string]interface{} with optional case-insensitive lookup.
func resolveStringMap(configData map[string]interface{}, key string, caseInsensitive bool) (interface{}, bool) {
	if !caseInsensitive {
		value, exists := configData[key]

		return value, exists
	}

	for mapKey, mapValue := range configData {
		if strings.EqualFold(mapKey, key) {
			return mapValue, true
		}
	}

	return nil, false
}

// resolveStringStringMap handles map[string]string with optional case-insensitive lookup.
func resolveStringStringMap(configData map[string]string, key string, caseInsensitive bool) (interface{}, bool) {
	if !caseInsensitive {
		value, exists := configData[key]

		return value, exists
	}

	for mapKey, mapValue := range configData {
		if strings.EqualFold(mapKey, key) {
			return mapValue, true
		}
	}

	return nil, false
}

// resolveInterfaceMap handles map[interface{}]interface{} with optional case-insensitive lookup.
func resolveInterfaceMap(configData map[interface{}]interface{}, key string, caseInsensitive bool) (interface{}, bool) {
	if !caseInsensitive {
		value, exists := configData[key]

		return value, exists
	}

	for mapKey, mapValue := range configData {
		if keyString, ok := mapKey.(string); ok && strings.EqualFold(keyString, key) {
			return mapValue, true
		}
	}

	return nil, false
}
