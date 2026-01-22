package schema

import "time"

// Database represents the complete introspected database schema.
type Database struct {
	Name    string `json:"database_name"`
	Engine  string `json:"database_engine"`
	Version string `json:"database_version,omitempty"`

	Schemas   []string    `json:"-"`
	Tables    []*Table    `json:"-"`
	Views     []*View     `json:"-"`
	Routines  []*Routine  `json:"-"`
	Enums     []*Enum     `json:"-"`
	Sequences []*Sequence `json:"-"`
	Types     []*Type     `json:"-"`
}

// Manifest represents the xrai.manifest.json output file.
type Manifest struct {
	GenerationTimestamp string           `json:"generation_timestamp"`
	DatabaseEngine      string           `json:"database_engine"`
	DatabaseVersion     string           `json:"database_version,omitempty"`
	DatabaseName        string           `json:"database_name"`
	IncludedSchemas     []string         `json:"included_schemas"`
	IncludedTablesCount int              `json:"included_tables_count"`
	EnabledArtifacts    EnabledArtifacts `json:"enabled_artifacts"`
	StatsEnabled        bool             `json:"stats_enabled"`
	UsageEnabled        bool             `json:"usage_enabled"`
}

// EnabledArtifacts tracks which artifact types were generated.
type EnabledArtifacts struct {
	Tables    bool `json:"tables"`
	Views     bool `json:"views"`
	Routines  bool `json:"routines"`
	Enums     bool `json:"enums"`
	Sequences bool `json:"sequences"`
	Types     bool `json:"types"`
	Stats     bool `json:"stats"`
}

// DatabaseIndex represents the db.index.json output file.
type DatabaseIndex struct {
	Tables                 []TableIndexEntry `json:"tables"`
	RecommendedStartTables []string          `json:"recommended_start_tables,omitempty"`
}

// TableIndexEntry is a single table entry in the database index.
type TableIndexEntry struct {
	TableName          string   `json:"table_name"`
	SchemaName         string   `json:"schema_name,omitempty"`
	ShortDescription   string   `json:"short_description,omitempty"`
	RowCountEstimate   *int64   `json:"row_count_estimate,omitempty"`
	PrimaryKeyColumns  []string `json:"primary_key_columns,omitempty"`
	ForeignKeyOutCount int      `json:"foreign_key_out_count"`
	ForeignKeyInCount  int      `json:"foreign_key_in_count"`
	Tags               []string `json:"tags,omitempty"`
}

// RelationshipGraph represents the db.relationships.json output file.
type RelationshipGraph struct {
	Nodes                       []RelationshipNode `json:"nodes"`
	Edges                       []RelationshipEdge `json:"edges"`
	JunctionTableCandidates     []string           `json:"junction_table_candidates,omitempty"`
	StronglyConnectedComponents [][]string         `json:"strongly_connected_components,omitempty"`
}

// RelationshipNode is a node (table) in the relationship graph.
type RelationshipNode struct {
	TableName  string `json:"table_name"`
	SchemaName string `json:"schema_name,omitempty"`
}

// RelationshipEdge is an edge (foreign key) in the relationship graph.
type RelationshipEdge struct {
	FromTable      string   `json:"from_table"`
	FromSchema     string   `json:"from_schema,omitempty"`
	FromColumns    []string `json:"from_columns"`
	ToTable        string   `json:"to_table"`
	ToSchema       string   `json:"to_schema,omitempty"`
	ToColumns      []string `json:"to_columns"`
	ConstraintName string   `json:"constraint_name"`
}

// DomainGrouping represents the db.domains.json output file.
type DomainGrouping struct {
	Domains []Domain `json:"domains"`
}

// Domain is a heuristic grouping of related tables.
type Domain struct {
	DomainName string   `json:"domain_name"`
	Tables     []string `json:"tables"`
	Keywords   []string `json:"keywords,omitempty"`
	Rationale  string   `json:"rationale,omitempty"`
}

// Stats represents optional table statistics.
type Stats struct {
	ComputedAt       time.Time           `json:"computed_at"`
	SamplingStrategy string              `json:"sampling_strategy"`
	RowCountEstimate *int64              `json:"row_count_estimate,omitempty"`
	ColumnStats      map[string]ColStats `json:"per_column_stats,omitempty"`
	Note             string              `json:"note,omitempty"`
}

// ColStats represents statistics for a single column.
type ColStats struct {
	NullFraction          *float64      `json:"null_fraction,omitempty"`
	DistinctCountEstimate *int64        `json:"distinct_count_estimate,omitempty"`
	Min                   interface{}   `json:"min,omitempty"`
	Max                   interface{}   `json:"max,omitempty"`
	TopValues             []interface{} `json:"top_values,omitempty"`
}

// Usage represents optional usage heuristics.
type Usage struct {
	CommonJoins         []string `json:"common_joins,omitempty"`
	FrequentlyFilteredBy []string `json:"frequently_filtered_by,omitempty"`
	FrequentlyGroupedBy []string `json:"frequently_grouped_by,omitempty"`
	DerivationNotes     string   `json:"derivation_notes,omitempty"`
}
