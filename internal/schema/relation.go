package schema

// ForeignKey represents a foreign key relationship.
type ForeignKey struct {
	ConstraintName string   `json:"constraint_name"`
	FromSchema     string   `json:"from_schema,omitempty"`
	FromTable      string   `json:"from_table"`
	FromColumns    []string `json:"from_columns"`
	ToSchema       string   `json:"to_schema,omitempty"`
	ToTable        string   `json:"to_table"`
	ToColumns      []string `json:"to_columns"`
	OnUpdate       string   `json:"on_update,omitempty"`
	OnDelete       string   `json:"on_delete,omitempty"`

	// Derived
	Nullable    bool   `json:"-"`
	Cardinality string `json:"-"`
}
