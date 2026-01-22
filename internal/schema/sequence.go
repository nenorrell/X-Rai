package schema

// Sequence represents a database sequence.
type Sequence struct {
	SequenceName string `json:"sequence_name"`
	SchemaName   string `json:"schema_name,omitempty"`
	DataType     string `json:"data_type,omitempty"`
	StartValue   *int64 `json:"start_value,omitempty"`
	MinValue     *int64 `json:"min_value,omitempty"`
	MaxValue     *int64 `json:"max_value,omitempty"`
	Increment    *int64 `json:"increment,omitempty"`
	CycleOption  bool   `json:"cycle,omitempty"`
	OwnedBy      string `json:"owned_by,omitempty"`
	Comment      string `json:"comment,omitempty"`
}
