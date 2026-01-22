package toon

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Encoder encodes data to TOON format.
type Encoder struct {
	sb         strings.Builder
	indent     int
	indentSize int
}

// NewEncoder creates a new TOON encoder.
func NewEncoder() *Encoder {
	return &Encoder{
		indentSize: 2,
	}
}

// Encode encodes a value to TOON format.
func Encode(v interface{}) (string, error) {
	e := NewEncoder()
	if err := e.encode(v, true); err != nil {
		return "", err
	}
	return e.sb.String(), nil
}

func (e *Encoder) encode(v interface{}, root bool) error {
	switch val := v.(type) {
	case nil:
		e.sb.WriteString("null")
	case bool:
		if val {
			e.sb.WriteString("true")
		} else {
			e.sb.WriteString("false")
		}
	case int:
		e.sb.WriteString(strconv.Itoa(val))
	case int64:
		e.sb.WriteString(strconv.FormatInt(val, 10))
	case float64:
		e.sb.WriteString(formatNumber(val))
	case string:
		e.sb.WriteString(quoteString(val, ','))
	case []interface{}:
		return e.encodeArray(val)
	case map[string]interface{}:
		return e.encodeObject(val, root)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
	return nil
}

func (e *Encoder) encodeObject(obj map[string]interface{}, root bool) error {
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}

	for i, k := range keys {
		if !root || i > 0 {
			e.writeIndent()
		}
		e.sb.WriteString(formatKey(k))
		e.sb.WriteString(":")

		v := obj[k]
		switch val := v.(type) {
		case map[string]interface{}:
			if len(val) == 0 {
				e.sb.WriteString(" {}")
			} else {
				e.sb.WriteString("\n")
				e.indent++
				if err := e.encodeObject(val, false); err != nil {
					return err
				}
				e.indent--
			}
		case []interface{}:
			if err := e.encodeArrayValue(val); err != nil {
				return err
			}
		default:
			e.sb.WriteString(" ")
			if err := e.encode(v, false); err != nil {
				return err
			}
		}

		if i < len(keys)-1 {
			e.sb.WriteString("\n")
		}
	}
	return nil
}

func (e *Encoder) encodeArrayValue(arr []interface{}) error {
	if len(arr) == 0 {
		e.sb.WriteString(fmt.Sprintf("[%d]:", len(arr)))
		return nil
	}

	// Check if all elements are primitives (inline array)
	if allPrimitives(arr) {
		e.sb.WriteString(fmt.Sprintf("[%d]: ", len(arr)))
		for i, v := range arr {
			if i > 0 {
				e.sb.WriteString(",")
			}
			if err := e.encode(v, false); err != nil {
				return err
			}
		}
		return nil
	}

	// Check if all elements are uniform objects (tabular array)
	if fields, ok := uniformObjectFields(arr); ok && len(fields) > 0 {
		e.sb.WriteString(fmt.Sprintf("[%d]{%s}:\n", len(arr), strings.Join(fields, ",")))
		e.indent++
		for i, v := range arr {
			e.writeIndent()
			obj := v.(map[string]interface{})
			for j, f := range fields {
				if j > 0 {
					e.sb.WriteString(",")
				}
				if err := e.encode(obj[f], false); err != nil {
					return err
				}
			}
			if i < len(arr)-1 {
				e.sb.WriteString("\n")
			}
		}
		e.indent--
		return nil
	}

	// Mixed array - use dash notation
	e.sb.WriteString(fmt.Sprintf("[%d]:\n", len(arr)))
	e.indent++
	for i, v := range arr {
		e.writeIndent()
		e.sb.WriteString("- ")
		switch val := v.(type) {
		case map[string]interface{}:
			if len(val) == 0 {
				e.sb.WriteString("{}")
			} else {
				e.sb.WriteString("\n")
				e.indent++
				if err := e.encodeObject(val, false); err != nil {
					return err
				}
				e.indent--
			}
		default:
			if err := e.encode(v, false); err != nil {
				return err
			}
		}
		if i < len(arr)-1 {
			e.sb.WriteString("\n")
		}
	}
	e.indent--
	return nil
}

func (e *Encoder) encodeArray(arr []interface{}) error {
	return e.encodeArrayValue(arr)
}

func (e *Encoder) writeIndent() {
	for i := 0; i < e.indent*e.indentSize; i++ {
		e.sb.WriteByte(' ')
	}
}

// formatNumber returns canonical decimal form per TOON spec.
func formatNumber(f float64) string {
	if f == 0 {
		return "0"
	}
	// Check if it's an integer
	if f == float64(int64(f)) {
		return strconv.FormatInt(int64(f), 10)
	}
	s := strconv.FormatFloat(f, 'f', -1, 64)
	return s
}

var (
	unquotedKeyRe   = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.]*$`)
	numericPatternRe = regexp.MustCompile(`^-?[0-9]`)
)

// formatKey formats an object key, quoting if necessary.
func formatKey(k string) string {
	if unquotedKeyRe.MatchString(k) {
		return k
	}
	return quoteString(k, ',')
}

// quoteString quotes a string if necessary per TOON spec.
func quoteString(s string, delim byte) string {
	if needsQuoting(s, delim) {
		return `"` + escapeString(s) + `"`
	}
	return s
}

func needsQuoting(s string, delim byte) bool {
	if s == "" {
		return true
	}
	if s == "true" || s == "false" || s == "null" {
		return true
	}
	if numericPatternRe.MatchString(s) {
		return true
	}
	if s[0] == ' ' || s[len(s)-1] == ' ' {
		return true
	}
	if s[0] == '-' {
		return true
	}
	for _, r := range s {
		switch r {
		case ':', '"', '\\', '[', ']', '{', '}', '\n', '\r', '\t':
			return true
		}
		if byte(r) == delim {
			return true
		}
		if r < 32 {
			return true
		}
	}
	return false
}

func escapeString(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			sb.WriteString(`\\`)
		case '"':
			sb.WriteString(`\"`)
		case '\n':
			sb.WriteString(`\n`)
		case '\r':
			sb.WriteString(`\r`)
		case '\t':
			sb.WriteString(`\t`)
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func allPrimitives(arr []interface{}) bool {
	for _, v := range arr {
		switch v.(type) {
		case nil, bool, int, int64, float64, string:
			continue
		default:
			return false
		}
	}
	return true
}

func uniformObjectFields(arr []interface{}) ([]string, bool) {
	if len(arr) == 0 {
		return nil, false
	}

	var fields []string
	for i, v := range arr {
		obj, ok := v.(map[string]interface{})
		if !ok {
			return nil, false
		}

		// Check all values are primitives
		for _, val := range obj {
			switch val.(type) {
			case nil, bool, int, int64, float64, string:
				continue
			default:
				return nil, false
			}
		}

		if i == 0 {
			fields = make([]string, 0, len(obj))
			for k := range obj {
				fields = append(fields, k)
			}
		} else {
			// Check same fields
			if len(obj) != len(fields) {
				return nil, false
			}
			for _, f := range fields {
				if _, ok := obj[f]; !ok {
					return nil, false
				}
			}
		}
	}
	return fields, true
}
