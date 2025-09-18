package tools

import (
	"encoding/json"
	"fmt"
)

func StructToMap(input interface{}) (map[string]interface{}, error) {
	if input == nil {
		return nil, fmt.Errorf("input is nil")
	}

	bytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshal struct: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, fmt.Errorf("unmarshal to map: %w", err)
	}

	return result, nil
}

func MapToStruct(input map[string]interface{}, output interface{}) error {
	if input == nil {
		return fmt.Errorf("input map is nil")
	}
	if output == nil {
		return fmt.Errorf("output struct is nil")
	}

	bytes, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("marshal map: %w", err)
	}

	if err := json.Unmarshal(bytes, output); err != nil {
		return fmt.Errorf("unmarshal to struct: %w", err)
	}

	return nil
}
