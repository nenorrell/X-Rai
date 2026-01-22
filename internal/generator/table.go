package generator

import (
	"path/filepath"

	"github.com/nenorrell/xrai/internal/schema"
)

func (g *Generator) generateTables(db *schema.Database) error {
	tablesDir := filepath.Join(g.outputDir, "tables")

	for _, table := range db.Tables {
		tableDir := filepath.Join(tablesDir, sanitizeName(table.TableName))

		// Generate table.structure.toon
		if err := g.generateTableStructure(tableDir, table); err != nil {
			return err
		}

		// Generate table.columns.toon
		if err := g.generateTableColumns(tableDir, table); err != nil {
			return err
		}

		// Generate table.indexes.toon
		if err := g.generateTableIndexes(tableDir, table); err != nil {
			return err
		}

		// Generate table.constraints.toon
		if err := g.generateTableConstraints(tableDir, table); err != nil {
			return err
		}

		// Generate table.relations.toon
		if err := g.generateTableRelations(tableDir, table); err != nil {
			return err
		}

		// Generate table.triggers.toon
		if err := g.generateTableTriggers(tableDir, table); err != nil {
			return err
		}

		// Generate table.comments.toon
		if err := g.generateTableComments(tableDir, table); err != nil {
			return err
		}

		// Generate table.stats.toon (if enabled)
		if g.cfg.IncludeStats && table.Stats != nil {
			if err := g.generateTableStats(tableDir, table); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Generator) generateTableStructure(tableDir string, table *schema.Table) error {
	structure := schema.TableStructure{
		TableName:        table.TableName,
		SchemaName:       table.SchemaName,
		TableType:        table.TableType,
		RowCountEstimate: table.RowCountEstimate,
	}

	// Find primary key
	for _, con := range table.Constraints {
		if con.ConstraintType == "PRIMARY KEY" {
			structure.PrimaryKey = &schema.PrimaryKey{
				Columns:        con.Columns,
				ConstraintName: con.ConstraintName,
			}
			break
		}
	}

	return g.writeTOON(filepath.Join(tableDir, "table.structure.toon"), structure)
}

func (g *Generator) generateTableColumns(tableDir string, table *schema.Table) error {
	columns := make([]schema.Column, 0, len(table.Columns))
	for _, col := range table.Columns {
		columns = append(columns, *col)
	}

	output := schema.ColumnsOutput{Columns: columns}
	return g.writeTOON(filepath.Join(tableDir, "table.columns.toon"), output)
}

func (g *Generator) generateTableIndexes(tableDir string, table *schema.Table) error {
	indexes := make([]schema.Index, 0, len(table.Indexes))
	for _, idx := range table.Indexes {
		// Create a copy without the Definition field (internal use only)
		indexes = append(indexes, schema.Index{
			IndexName:      idx.IndexName,
			Unique:         idx.Unique,
			Primary:        idx.Primary,
			IndexType:      idx.IndexType,
			Columns:        idx.Columns,
			IncludeColumns: idx.IncludeColumns,
			Predicate:      idx.Predicate,
			Expression:     idx.Expression,
		})
	}

	output := schema.IndexesOutput{Indexes: indexes}
	return g.writeTOON(filepath.Join(tableDir, "table.indexes.toon"), output)
}

func (g *Generator) generateTableConstraints(tableDir string, table *schema.Table) error {
	output := schema.ConstraintsOutput{
		UniqueConstraints:    make([]schema.Constraint, 0),
		CheckConstraints:     make([]schema.Constraint, 0),
		ExclusionConstraints: make([]schema.Constraint, 0),
		NotNullConstraints:   make([]schema.Constraint, 0),
	}

	for _, con := range table.Constraints {
		switch con.ConstraintType {
		case "PRIMARY KEY":
			output.PrimaryKey = con
		case "UNIQUE":
			output.UniqueConstraints = append(output.UniqueConstraints, *con)
		case "CHECK":
			output.CheckConstraints = append(output.CheckConstraints, *con)
		case "EXCLUSION":
			output.ExclusionConstraints = append(output.ExclusionConstraints, *con)
		}
	}

	// Add NOT NULL constraints from columns
	for _, col := range table.Columns {
		if !col.Nullable {
			output.NotNullConstraints = append(output.NotNullConstraints, schema.Constraint{
				ConstraintName: col.ColumnName + "_not_null",
				ConstraintType: "NOT NULL",
				Columns:        []string{col.ColumnName},
			})
		}
	}

	return g.writeTOON(filepath.Join(tableDir, "table.constraints.toon"), output)
}

func (g *Generator) generateTableRelations(tableDir string, table *schema.Table) error {
	output := schema.TableRelations{
		OutgoingForeignKeys: make([]schema.ForeignKeyOutput, 0, len(table.OutgoingForeignKeys)),
		IncomingForeignKeys: make([]schema.IncomingFKOutput, 0, len(table.IncomingForeignKeys)),
	}

	for _, fk := range table.OutgoingForeignKeys {
		output.OutgoingForeignKeys = append(output.OutgoingForeignKeys, schema.ForeignKeyOutput{
			ConstraintName: fk.ConstraintName,
			FromColumns:    fk.FromColumns,
			ToTable:        fk.ToTable,
			ToSchema:       fk.ToSchema,
			ToColumns:      fk.ToColumns,
			OnUpdate:       fk.OnUpdate,
			OnDelete:       fk.OnDelete,
			Nullable:       fk.Nullable,
			Cardinality:    fk.Cardinality,
		})
	}

	for _, fk := range table.IncomingForeignKeys {
		output.IncomingForeignKeys = append(output.IncomingForeignKeys, schema.IncomingFKOutput{
			ConstraintName: fk.ConstraintName,
			FromTable:      fk.FromTable,
			FromSchema:     fk.FromSchema,
			FromColumns:    fk.FromColumns,
			ToColumns:      fk.ToColumns,
			Cardinality:    fk.Cardinality,
		})
	}

	if table.IsJunction {
		output.JunctionTableDetection = &schema.JunctionDetection{
			IsJunction: true,
			Reasoning:  table.JunctionReasoning,
		}
	} else {
		output.JunctionTableDetection = &schema.JunctionDetection{
			IsJunction: false,
		}
	}

	return g.writeTOON(filepath.Join(tableDir, "table.relations.toon"), output)
}

func (g *Generator) generateTableTriggers(tableDir string, table *schema.Table) error {
	triggers := make([]schema.Trigger, 0, len(table.Triggers))

	for _, trig := range table.Triggers {
		t := schema.Trigger{
			TriggerName:         trig.TriggerName,
			Timing:              trig.Timing,
			Events:              trig.Events,
			FunctionOrProcedure: trig.FunctionOrProcedure,
		}

		// Include definition if not redacted
		if !g.cfg.RedactDefinitions {
			t.Definition = trig.Definition
		}

		triggers = append(triggers, t)
	}

	output := schema.TriggersOutput{Triggers: triggers}
	return g.writeTOON(filepath.Join(tableDir, "table.triggers.toon"), output)
}

func (g *Generator) generateTableComments(tableDir string, table *schema.Table) error {
	output := schema.TableComments{
		ColumnComments: make(map[string]string),
	}

	if !g.cfg.RedactComments {
		output.TableComment = table.Comment

		for _, col := range table.Columns {
			if col.Comment != "" {
				output.ColumnComments[col.ColumnName] = col.Comment
			}
		}
	}

	// Add inferred semantics
	output.InferredSemantics = inferColumnSemantics(table.Columns)

	return g.writeTOON(filepath.Join(tableDir, "table.comments.toon"), output)
}

func (g *Generator) generateTableStats(tableDir string, table *schema.Table) error {
	if table.Stats == nil {
		return nil
	}
	return g.writeTOON(filepath.Join(tableDir, "table.stats.toon"), table.Stats)
}

func inferColumnSemantics(columns []*schema.Column) *schema.InferredSemantics {
	semantics := &schema.InferredSemantics{}

	for _, col := range columns {
		name := col.ColumnName
		dtype := col.DataType

		// ID columns
		if isIDColumn(name) {
			semantics.IDColumns = append(semantics.IDColumns, name)
		}

		// Timestamp columns
		if isTimestampColumn(name, dtype) {
			semantics.TimestampColumns = append(semantics.TimestampColumns, name)
		}

		// Status columns
		if isStatusColumn(name) {
			semantics.StatusColumns = append(semantics.StatusColumns, name)
		}

		// Currency/amount columns
		if isCurrencyColumn(name, dtype) {
			semantics.CurrencyAmountColumns = append(semantics.CurrencyAmountColumns, name)
		}
	}

	// Only return if we found something
	if len(semantics.IDColumns) == 0 && len(semantics.TimestampColumns) == 0 &&
		len(semantics.StatusColumns) == 0 && len(semantics.CurrencyAmountColumns) == 0 {
		return nil
	}

	return semantics
}

func isIDColumn(name string) bool {
	return name == "id" ||
		hasSuffix(name, "_id") ||
		hasSuffix(name, "_uuid") ||
		name == "uuid"
}

func isTimestampColumn(name, dtype string) bool {
	if contains(dtype, "timestamp") || dtype == "date" {
		return true
	}
	return hasSuffix(name, "_at") ||
		hasSuffix(name, "_date") ||
		hasSuffix(name, "_time") ||
		name == "created" ||
		name == "updated" ||
		name == "deleted"
}

func isStatusColumn(name string) bool {
	return name == "status" ||
		hasSuffix(name, "_status") ||
		name == "state" ||
		hasSuffix(name, "_state")
}

func isCurrencyColumn(name, dtype string) bool {
	if dtype == "money" || dtype == "numeric" || contains(dtype, "decimal") {
		return contains(name, "amount") ||
			contains(name, "price") ||
			contains(name, "cost") ||
			contains(name, "total") ||
			contains(name, "balance")
	}
	return false
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
