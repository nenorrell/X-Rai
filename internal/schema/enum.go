package schema

// Enum represents a PostgreSQL enum type.
type Enum struct {
	EnumName   string   `json:"enum_name"`
	SchemaName string   `json:"schema_name,omitempty"`
	Values     []string `json:"values"`
	Comment    string   `json:"comment,omitempty"`
}
