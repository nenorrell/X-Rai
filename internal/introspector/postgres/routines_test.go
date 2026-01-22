package postgres

import (
	"testing"

	"github.com/nenorrell/X-Rai/internal/schema"
)

func TestDerefString(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{"nil", nil, ""},
		{"empty string", ptr(""), ""},
		{"non-empty", ptr("hello"), "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := derefString(tt.input)
			if result != tt.expected {
				t.Errorf("derefString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestMapArgMode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"i", "IN"},
		{"I", "IN"},
		{"o", "OUT"},
		{"O", "OUT"},
		{"b", "INOUT"},
		{"B", "INOUT"},
		{"v", "VARIADIC"},
		{"V", "VARIADIC"},
		{"t", "TABLE"},
		{"T", "TABLE"},
		{"", "IN"},
		{"x", "IN"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapArgMode(tt.input)
			if result != tt.expected {
				t.Errorf("mapArgMode(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuildArguments(t *testing.T) {
	tests := []struct {
		name     string
		names    []string
		types    []string
		modes    []string
		expected []schema.Argument
	}{
		{
			name:     "empty",
			names:    nil,
			types:    nil,
			modes:    nil,
			expected: nil,
		},
		{
			name:  "single arg with all info",
			names: []string{"user_id"},
			types: []string{"integer"},
			modes: []string{"i"},
			expected: []schema.Argument{
				{Name: "user_id", DataType: "integer", Mode: "IN"},
			},
		},
		{
			name:  "multiple args",
			names: []string{"a", "b"},
			types: []string{"text", "integer"},
			modes: []string{"i", "o"},
			expected: []schema.Argument{
				{Name: "a", DataType: "text", Mode: "IN"},
				{Name: "b", DataType: "integer", Mode: "OUT"},
			},
		},
		{
			name:  "missing names",
			names: nil,
			types: []string{"text", "integer"},
			modes: []string{"i", "i"},
			expected: []schema.Argument{
				{Name: "", DataType: "text", Mode: "IN"},
				{Name: "", DataType: "integer", Mode: "IN"},
			},
		},
		{
			name:  "missing modes defaults to IN",
			names: []string{"x"},
			types: []string{"text"},
			modes: nil,
			expected: []schema.Argument{
				{Name: "x", DataType: "text", Mode: "IN"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildArguments(tt.names, tt.types, tt.modes)

			if len(result) != len(tt.expected) {
				t.Fatalf("got %d args, want %d", len(result), len(tt.expected))
			}

			for i, arg := range result {
				exp := tt.expected[i]
				if arg.Name != exp.Name || arg.DataType != exp.DataType || arg.Mode != exp.Mode {
					t.Errorf("arg[%d] = %+v, want %+v", i, arg, exp)
				}
			}
		})
	}
}

// ptr is a helper to create string pointers
func ptr(s string) *string {
	return &s
}
