package config

import "fmt"

var allowedTypes = map[string]bool{
	"string": true,
	"int":    true,
	"float":  true,
	"bool":   true,
}

// ValidateDefaults validates all entries in the defaults section.
func ValidateDefaults(entries []DefaultEntry) error {
	for i, e := range entries {
		if e.Domain == "" {
			return fmt.Errorf("defaults[%d]: domain is required", i)
		}
		if e.Key == "" {
			return fmt.Errorf("defaults[%d]: key is required", i)
		}
		if !allowedTypes[e.Type] {
			return fmt.Errorf("defaults[%d]: type must be string, int, float, or bool", i)
		}
		if e.Value == nil {
			return fmt.Errorf("defaults[%d]: value is required", i)
		}
		if err := validateTypeMatch(e.Type, e.Value); err != nil {
			return fmt.Errorf("defaults[%d]: %w", i, err)
		}
	}
	return nil
}

func validateTypeMatch(typ string, value interface{}) error {
	// Reject non-scalar values (arrays, maps)
	switch value.(type) {
	case []interface{}:
		return fmt.Errorf("value must be a scalar, got array")
	case map[string]interface{}:
		return fmt.Errorf("value must be a scalar, got map")
	}

	switch typ {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("value %v is not a valid string", value)
		}
	case "int":
		switch value.(type) {
		case int, int64, float64:
			// YAML parses integers as int; accept numeric types
			if f, ok := value.(float64); ok {
				if f != float64(int(f)) {
					return fmt.Errorf("value %v is not a valid int", value)
				}
			}
		default:
			return fmt.Errorf("value %v is not a valid int", value)
		}
	case "float":
		switch value.(type) {
		case float64, int, int64:
			// accept numeric types
		default:
			return fmt.Errorf("value %v is not a valid float", value)
		}
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("value %v is not a valid bool", value)
		}
	}
	return nil
}
