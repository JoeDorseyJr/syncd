package defaults

import "fmt"

// CompareValue returns true if the raw defaults output matches the config value.
func CompareValue(rawOutput, configType string, configValue interface{}) bool {
	switch configType {
	case "bool":
		return compareBool(rawOutput, configValue)
	case "int", "float", "string":
		return rawOutput == fmt.Sprintf("%v", configValue)
	}
	return false
}

func compareBool(raw string, value interface{}) bool {
	var desired string
	switch v := value.(type) {
	case bool:
		if v {
			desired = "1"
		} else {
			desired = "0"
		}
	case string:
		if v == "true" {
			desired = "1"
		} else {
			desired = "0"
		}
	}
	return raw == desired
}
