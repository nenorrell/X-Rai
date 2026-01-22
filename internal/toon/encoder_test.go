package toon

import (
	"strings"
	"testing"
)

func TestEncode_Primitives(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil", nil, "null"},
		{"true", true, "true"},
		{"false", false, "false"},
		{"int", 42, "42"},
		{"int64", int64(123456789), "123456789"},
		{"float64", 3.14, "3.14"},
		{"float64 whole", float64(42), "42"},
		{"string simple", "hello", "hello"},
		{"string with space", "hello world", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Encode(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestEncode_StringQuoting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"reserved true", "true", `"true"`},
		{"reserved false", "false", `"false"`},
		{"reserved null", "null", `"null"`},
		{"numeric looking", "123", `"123"`},
		{"negative looking", "-5", `"-5"`},
		{"empty string", "", `""`},
		{"leading space", " hello", `" hello"`},
		{"trailing space", "hello ", `"hello "`},
		{"contains colon", "key:value", `"key:value"`},
		{"contains newline", "line1\nline2", `"line1\nline2"`},
		{"contains quote", `say "hi"`, `"say \"hi\""`},
		{"starts with dash", "-flag", `"-flag"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Encode(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestEncode_ArrayInline(t *testing.T) {
	// Primitive arrays should be inline
	input := []interface{}{"a", "b", "c"}
	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain count and be comma-separated
	if !strings.HasPrefix(result, "[3]:") {
		t.Errorf("expected prefix [3]:, got %q", result)
	}
	if !strings.Contains(result, "a,b,c") {
		t.Errorf("expected comma-separated values, got %q", result)
	}
}

func TestEncode_ArrayEmpty(t *testing.T) {
	input := []interface{}{}
	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "[0]:" {
		t.Errorf("got %q, want [0]:", result)
	}
}

func TestEncode_Object(t *testing.T) {
	input := map[string]interface{}{
		"name": "test",
	}
	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "name:") {
		t.Errorf("expected 'name:', got %q", result)
	}
	if !strings.Contains(result, "test") {
		t.Errorf("expected 'test', got %q", result)
	}
}

func TestFormatKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with_underscore", "with_underscore"},
		{"CamelCase", "CamelCase"},
		{"has space", "has space"},
		{"123numeric", `"123numeric"`},
		{"with:colon", `"with:colon"`},
		{" leading", `" leading"`},
		{"trailing ", `"trailing "`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatKey(tt.input)
			if result != tt.expected {
				t.Errorf("formatKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNeedsQuoting(t *testing.T) {
	tests := []struct {
		input    string
		delim    byte
		expected bool
	}{
		{"simple", ',', false},
		{"", ',', true},
		{"true", ',', true},
		{"false", ',', true},
		{"null", ',', true},
		{"123", ',', true},
		{" leading", ',', true},
		{"trailing ", ',', true},
		{"has,comma", ',', true},
		{"no comma", ';', false},
		{"has;semi", ';', true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := needsQuoting(tt.input, tt.delim)
			if result != tt.expected {
				t.Errorf("needsQuoting(%q, %q) = %v, want %v", tt.input, tt.delim, result, tt.expected)
			}
		})
	}
}

func TestAllPrimitives(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected bool
	}{
		{"all strings", []interface{}{"a", "b", "c"}, true},
		{"mixed primitives", []interface{}{"a", 1, true, nil}, true},
		{"contains map", []interface{}{"a", map[string]interface{}{}}, false},
		{"contains array", []interface{}{"a", []interface{}{}}, false},
		{"empty", []interface{}{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := allPrimitives(tt.input)
			if result != tt.expected {
				t.Errorf("allPrimitives() = %v, want %v", result, tt.expected)
			}
		})
	}
}
