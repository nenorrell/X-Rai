package schema

// View represents a database view.
type View struct {
	ViewName   string    `json:"view_name"`
	SchemaName string    `json:"schema_name,omitempty"`
	Definition *string   `json:"-"`
	Comment    string    `json:"-"`
	Columns    []*Column `json:"-"`

	// Dependencies
	DependsOnTables []string `json:"-"`
	DependsOnViews  []string `json:"-"`
}

// ViewDefinition represents the view.definition.json output file.
type ViewDefinition struct {
	ViewName   string  `json:"view_name"`
	SchemaName string  `json:"schema_name,omitempty"`
	Definition *string `json:"definition,omitempty"`
}

// ViewDependencies represents the view.dependencies.json output file.
type ViewDependencies struct {
	DependsOnTables []string `json:"depends_on_tables,omitempty"`
	DependsOnViews  []string `json:"depends_on_views,omitempty"`
}

// ViewComments represents the view.comments.json output file.
type ViewComments struct {
	ViewComment    string            `json:"view_comment,omitempty"`
	ColumnComments map[string]string `json:"column_comments,omitempty"`
}
