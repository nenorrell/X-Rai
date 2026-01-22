package schema

// Index represents a database index.
type Index struct {
	IndexName      string   `json:"index_name"`
	Unique         bool     `json:"unique"`
	Primary        bool     `json:"primary,omitempty"`
	IndexType      string   `json:"index_type,omitempty"`
	Columns        []string `json:"columns"`
	IncludeColumns []string `json:"include_columns,omitempty"`
	Predicate      *string  `json:"predicate,omitempty"`
	Expression     *string  `json:"expression,omitempty"`
	Definition     string   `json:"-"`
}

// IndexesOutput represents the table.indexes.json output file.
type IndexesOutput struct {
	Indexes []Index `json:"indexes"`
}
