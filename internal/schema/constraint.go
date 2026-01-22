package schema

// Constraint represents a database constraint.
type Constraint struct {
	ConstraintName string   `json:"constraint_name"`
	ConstraintType string   `json:"constraint_type"`
	Columns        []string `json:"columns,omitempty"`
	Expression     *string  `json:"expression,omitempty"`
	Definition     *string  `json:"definition,omitempty"`
}

// ConstraintsOutput represents the table.constraints.json output file.
type ConstraintsOutput struct {
	PrimaryKey          *Constraint  `json:"primary_key,omitempty"`
	UniqueConstraints   []Constraint `json:"unique_constraints,omitempty"`
	CheckConstraints    []Constraint `json:"check_constraints,omitempty"`
	ExclusionConstraints []Constraint `json:"exclusion_constraints,omitempty"`
	NotNullConstraints  []Constraint `json:"not_null_constraints,omitempty"`
}
