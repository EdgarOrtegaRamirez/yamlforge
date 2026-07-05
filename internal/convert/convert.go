// Package convert provides YAML conversion to/from other formats.
package convert

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Engine converts between YAML and other formats.
type Engine struct{}

// NewEngine creates a new convert engine.
func NewEngine() *Engine {
	return &Engine{}
}

// ToJSON converts YAML to JSON.
func (e *Engine) ToJSON(data []byte) ([]byte, error) {
	var interfaceData interface{}
	if err := yaml.Unmarshal(data, &interfaceData); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	jsonData, err := json.MarshalIndent(interfaceData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling JSON: %w", err)
	}

	return jsonData, nil
}

// ToJSONCompact converts YAML to compact JSON.
func (e *Engine) ToJSONCompact(data []byte) ([]byte, error) {
	var interfaceData interface{}
	if err := yaml.Unmarshal(data, &interfaceData); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	jsonData, err := json.Marshal(interfaceData)
	if err != nil {
		return nil, fmt.Errorf("marshaling JSON: %w", err)
	}

	return jsonData, nil
}

// FromJSON converts JSON to YAML.
func (e *Engine) FromJSON(data []byte) ([]byte, error) {
	var interfaceData interface{}
	if err := json.Unmarshal(data, &interfaceData); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	yamlData, err := yaml.Marshal(interfaceData)
	if err != nil {
		return nil, fmt.Errorf("marshaling YAML: %w", err)
	}

	return yamlData, nil
}

// ToTOML converts YAML to TOML (simplified).
func (e *Engine) ToTOML(data []byte) ([]byte, error) {
	var interfaceData interface{}
	if err := yaml.Unmarshal(data, &interfaceData); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	tomlStr, err := interfaceToTOML(interfaceData, "")
	if err != nil {
		return nil, err
	}

	return []byte(tomlStr), nil
}

// interfaceToTOML recursively converts an interface to TOML string.
func interfaceToTOML(data interface{}, prefix string) (string, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		var sb strings.Builder
		for key, val := range v {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + "." + key
			}

			switch val.(type) {
			case map[string]interface{}:
				sb.WriteString(fmt.Sprintf("\n[%s]\n", fullKey))
				sub, err := interfaceToTOML(val, fullKey)
				if err != nil {
					return "", err
				}
				sb.WriteString(sub)
			case []interface{}:
				sb.WriteString(fmt.Sprintf("%s = [", key))
				for i, item := range val.([]interface{}) {
					if i > 0 {
						sb.WriteString(", ")
					}
					sb.WriteString(fmt.Sprintf("%v", item))
				}
				sb.WriteString("]\n")
			default:
				sb.WriteString(fmt.Sprintf("%s = %v\n", key, formatTOMLValue(val)))
			}
		}
		return sb.String(), nil

	case []interface{}:
		var items []string
		for _, item := range v {
			items = append(items, fmt.Sprintf("%v", item))
		}
		return strings.Join(items, ", "), nil

	default:
		return fmt.Sprintf("%v", data), nil
	}
}

// formatTOMLValue formats a value for TOML output.
func formatTOMLValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}

// ToCSV converts a YAML sequence of mappings to CSV.
func (e *Engine) ToCSV(data []byte) ([]byte, error) {
	var interfaceData interface{}
	if err := yaml.Unmarshal(data, &interfaceData); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	arr, ok := interfaceData.([]interface{})
	if !ok {
		return nil, fmt.Errorf("YAML root is not a sequence")
	}

	if len(arr) == 0 {
		return []byte(""), nil
	}

	// Get headers from first item
	first, ok := arr[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("sequence items are not mappings")
	}

	var headers []string
	for key := range first {
		headers = append(headers, key)
	}

	var sb strings.Builder
	sb.WriteString(strings.Join(headers, ",") + "\n")

	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		var values []string
		for _, header := range headers {
			val := m[header]
			values = append(values, fmt.Sprintf("%v", val))
		}
		sb.WriteString(strings.Join(values, ",") + "\n")
	}

	return []byte(sb.String()), nil
}

// FileToJSON converts a YAML file to JSON.
func (e *Engine) FileToJSON(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return e.ToJSON(data)
}

// FileFromJSON converts a JSON file to YAML.
func (e *Engine) FileFromJSON(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return e.FromJSON(data)
}
