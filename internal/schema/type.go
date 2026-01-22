package schema

// Type represents a custom database type (composite, domain, range, etc.).
type Type struct {
	TypeName   string       `json:"type_name"`
	SchemaName string       `json:"schema_name,omitempty"`
	TypeKind   string       `json:"type_kind"`
	Attributes []TypeAttribute `json:"attributes,omitempty"`
	BaseType   string       `json:"base_type,omitempty"`
	Constraint *string      `json:"constraint,omitempty"`
	Comment    string       `json:"comment,omitempty"`
}

// TypeAttribute represents an attribute of a composite type.
type TypeAttribute struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
}
