package json

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/goccy/go-json"
)

// GetValuesFromJSONMap extracts values from a nested JSON using a list of specified paths.
// Ex:
// paths: ["settings.background.desktop.attachment", "settings.background.desktop.iconSpacing", "settings.background.desktop.color"]
// data: { settings { background { desktop {attachment : "123px", iconSpacing: "12px" }}}}
// result: ["123px", "12px", nil]
func GetValuesFromJSONMap(paths []string, data map[string]any) ([]any, error) {
	result := make([]any, 0, len(paths))
	for _, path := range paths {
		value, err := GetValueFromJSONMap(path, data)
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, nil
}

// GetValueFromJSONMap extracts value from a nested JSON using a specified path.
// Ex:
// paths: "settings.background.desktop.attachment"
// data: { settings { background { desktop {attachment : "123px", iconSpacing: "12px" }}}}
// result: "123px"
func GetValueFromJSONMap(path string, data map[string]any) (any, error) {
	keys := strings.Split(path, ".")

	// Traverse the JSON structure using the keys
	current := data
	for i, key := range keys {
		if val, ok := current[key]; ok {
			switch val := val.(type) {
			case map[string]interface{}:
				current = val
			default:
				if i == len(keys)-1 {
					return val, nil
				}
				return nil, fmt.Errorf("key not found: %s", path)
			}
		} else {
			return nil, fmt.Errorf("key not found: %s", path)
		}
	}
	return current, nil
}

// IsValid checks whether a string is a slice or a map
func IsValid(value string, t reflect.Kind) bool {
	switch t {
	case reflect.Slice: // [{}]
		var jsonArr []interface{}
		return json.Unmarshal([]byte(value), &jsonArr) == nil
	case reflect.Map: // {}
		var jsonStr map[string]interface{}
		return json.Unmarshal([]byte(value), &jsonStr) == nil
	default:
		return false
	}
}
