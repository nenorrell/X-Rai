package schema

// Table represents a database table and all its associated metadata.
type Table struct {
	TableName        string `json:"table_name"`
	SchemaName       string `json:"schema_name,omitempty"`
	TableType        string `json:"table_type,omitempty"`
	Comment          string `json:"-"`
	RowCountEstimate *int64 `json:"-"`

	Columns     []*Column     `json:"-"`
	Indexes     []*Index      `json:"-"`
	Constraints []*Constraint `json:"-"`
	Triggers    []*Trigger    `json:"-"`

	// Relations
	OutgoingForeignKeys []*ForeignKey `json:"-"`
	IncomingForeignKeys []*ForeignKey `json:"-"`

	// Heuristics
	IsJunction       bool     `json:"-"`
	JunctionReasoning string   `json:"-"`
	Tags             []string `json:"-"`

	// Optional
	Stats *Stats `json:"-"`
	Usage *Usage `json:"-"`
}

// TableStructure represents the table.structure.json output file.
type TableStructure struct {
	TableName        string      `json:"table_name"`
	SchemaName       string      `json:"schema_name,omitempty"`
	TableType        string      `json:"table_type,omitempty"`
	PrimaryKey       *PrimaryKey `json:"primary_key,omitempty"`
	RowCountEstimate *int64      `json:"row_count_estimate,omitempty"`
}

// PrimaryKey represents primary key information.
type PrimaryKey struct {
	Columns        []string `json:"columns"`
	ConstraintName string   `json:"constraint_name,omitempty"`
}

// TableRelations represents the table.relations.json output file.
type TableRelations struct {
	OutgoingForeignKeys    []ForeignKeyOutput   `json:"outgoing_foreign_keys"`
	IncomingForeignKeys    []IncomingFKOutput   `json:"incoming_foreign_keys"`
	JunctionTableDetection *JunctionDetection   `json:"junction_table_detection,omitempty"`
}

// ForeignKeyOutput represents an outgoing foreign key in JSON output.
type ForeignKeyOutput struct {
	ConstraintName string   `json:"constraint_name"`
	FromColumns    []string `json:"from_columns"`
	ToTable        string   `json:"to_table"`
	ToSchema       string   `json:"to_schema,omitempty"`
	ToColumns      []string `json:"to_columns"`
	OnUpdate       string   `json:"on_update,omitempty"`
	OnDelete       string   `json:"on_delete,omitempty"`
	Nullable       bool     `json:"nullable"`
	Cardinality    string   `json:"cardinality,omitempty"`
}

// IncomingFKOutput represents an incoming foreign key in JSON output.
type IncomingFKOutput struct {
	ConstraintName string   `json:"constraint_name"`
	FromTable      string   `json:"from_table"`
	FromSchema     string   `json:"from_schema,omitempty"`
	FromColumns    []string `json:"from_columns"`
	ToColumns      []string `json:"to_columns"`
	Cardinality    string   `json:"cardinality,omitempty"`
}

// JunctionDetection represents junction table detection results.
type JunctionDetection struct {
	IsJunction bool   `json:"is_junction"`
	Reasoning  string `json:"reasoning,omitempty"`
}

// TableComments represents the table.comments.json output file.
type TableComments struct {
	TableComment      string            `json:"table_comment,omitempty"`
	ColumnComments    map[string]string `json:"column_comments,omitempty"`
	InferredSemantics *InferredSemantics `json:"inferred_semantics,omitempty"`
}

// InferredSemantics represents heuristically detected column semantics.
type InferredSemantics struct {
	EnumLikeColumns       []string `json:"enum_like_columns,omitempty"`
	StatusColumns         []string `json:"status_columns,omitempty"`
	TimestampColumns      []string `json:"timestamp_columns,omitempty"`
	CurrencyAmountColumns []string `json:"currency_amount_columns,omitempty"`
	IDColumns             []string `json:"id_columns,omitempty"`
}
