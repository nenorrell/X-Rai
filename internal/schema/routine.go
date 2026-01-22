package schema

// Routine represents a database function or procedure.
type Routine struct {
	RoutineName   string   `json:"routine_name"`
	SchemaName    string   `json:"schema_name,omitempty"`
	RoutineType   string   `json:"routine_type"`
	Language      string   `json:"language,omitempty"`
	Arguments     []Argument `json:"arguments,omitempty"`
	ReturnType    string   `json:"return_type,omitempty"`
	Definition    *string  `json:"definition,omitempty"`
	Comment       string   `json:"-"`

	// Dependencies
	ReferencedTables []string `json:"referenced_tables,omitempty"`
}

// Argument represents a function/procedure argument.
type Argument struct {
	Name      string `json:"name,omitempty"`
	DataType  string `json:"data_type"`
	Mode      string `json:"mode,omitempty"`
	Default   *string `json:"default,omitempty"`
}
