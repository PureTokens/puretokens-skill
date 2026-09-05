package main

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

// sanitizeResponseJSON removes the active credential from decoded response
// strings and object keys, so JSON escaping cannot bypass redaction. It is only
// for JSON API responses, never outgoing requests or downloaded media bytes.
// Invalid JSON and key collisions fail closed as an unreadable response.
func sanitizeResponseJSON(body []byte, token string) []byte {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil
	}
	value, ok := sanitizeResponseValue(value, token)
	if !ok {
		return nil
	}
	sanitized, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return sanitized
}

func sanitizeResponseValue(value any, token string) (any, bool) {
	redact := func(text string) string {
		if token == "" {
			return text
		}
		return strings.ReplaceAll(text, token, "[redacted credential]")
	}
	switch current := value.(type) {
	case string:
		return redact(current), true
	case []any:
		for index, item := range current {
			sanitized, ok := sanitizeResponseValue(item, token)
			if !ok {
				return nil, false
			}
			current[index] = sanitized
		}
		return current, true
	case map[string]any:
		result := make(map[string]any, len(current))
		for key, item := range current {
			key = redact(key)
			if _, exists := result[key]; exists {
				return nil, false
			}
			sanitized, ok := sanitizeResponseValue(item, token)
			if !ok {
				return nil, false
			}
			result[key] = sanitized
		}
		return result, true
	default:
		return current, true
	}
}
