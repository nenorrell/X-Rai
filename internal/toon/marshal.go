package toon

import (
	"encoding/json"
	"fmt"
)

// Marshal encodes a Go value to TOON format.
// It first converts to a generic map/slice structure, then encodes to TOON.
func Marshal(v interface{}) ([]byte, error) {
	// Use JSON as intermediate to handle struct tags and complex types
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	// Parse JSON into generic structure
	var generic interface{}
	if err := json.Unmarshal(jsonBytes, &generic); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Encode to TOON
	toon, err := Encode(generic)
	if err != nil {
		return nil, fmt.Errorf("failed to encode TOON: %w", err)
	}

	return []byte(toon), nil
}

// MarshalIndent is an alias for Marshal (TOON is always indented).
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return Marshal(v)
}
