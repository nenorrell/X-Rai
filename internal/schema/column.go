package schema

// Column represents a database table column.
type Column struct {
	ColumnName           string  `json:"column_name"`
	DataType             string  `json:"data_type"`
	UDTName              string  `json:"udt_name,omitempty"`
	Nullable             bool    `json:"nullable"`
	DefaultValue         *string `json:"default_value,omitempty"`
	Generated            bool    `json:"generated,omitempty"`
	GenerationExpression *string `json:"generation_expression,omitempty"`
	IsIdentity           bool    `json:"is_identity,omitempty"`
	IdentityGeneration   string  `json:"identity_generation,omitempty"`
	Collation            *string `json:"collation,omitempty"`

	// Size/precision information
	CharacterMaxLength *int `json:"character_max_length,omitempty"`
	NumericPrecision   *int `json:"numeric_precision,omitempty"`
	NumericScale       *int `json:"numeric_scale,omitempty"`

	// Position in table
	OrdinalPosition int `json:"-"`

	// Comment
	Comment string `json:"-"`
}

// ColumnsOutput represents the table.columns.json output file.
type ColumnsOutput struct {
	Columns []Column `json:"columns"`
}
