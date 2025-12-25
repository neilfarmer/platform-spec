package kubernetes

// JSON helper functions for navigating kubectl JSON output

// getNestedString navigates nested maps to extract a string value
func getNestedString(m map[string]interface{}, keys ...string) (string, bool) {
	current := m
	for i, key := range keys {
		if i == len(keys)-1 {
			if val, ok := current[key].(string); ok {
				return val, true
			}
			return "", false
		}
		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			return "", false
		}
	}
	return "", false
}

// getNestedSlice navigates nested maps to extract a slice value
func getNestedSlice(m map[string]interface{}, keys ...string) ([]interface{}, bool) {
	current := m
	for i, key := range keys {
		if i == len(keys)-1 {
			if val, ok := current[key].([]interface{}); ok {
				return val, true
			}
			return nil, false
		}
		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			return nil, false
		}
	}
	return nil, false
}

// getNestedMap navigates nested maps to extract a map[string]string value
func getNestedMap(m map[string]interface{}, keys ...string) (map[string]string, bool) {
	current := m
	for i, key := range keys {
		if i == len(keys)-1 {
			if val, ok := current[key].(map[string]interface{}); ok {
				result := make(map[string]string)
				for k, v := range val {
					if strVal, ok := v.(string); ok {
						result[k] = strVal
					}
				}
				return result, true
			}
			return nil, false
		}
		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			return nil, false
		}
	}
	return nil, false
}

// getNestedFloat64 navigates nested maps to extract a float64 value (for numbers)
func getNestedFloat64(m map[string]interface{}, keys ...string) (float64, bool) {
	current := m
	for i, key := range keys {
		if i == len(keys)-1 {
			if val, ok := current[key].(float64); ok {
				return val, true
			}
			return 0, false
		}
		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			return 0, false
		}
	}
	return 0, false
}
