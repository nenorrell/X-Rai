package heuristics

import (
	"strings"

	"github.com/nenorrell/xrai/internal/schema"
)

// detectJunctionTable determines if a table is a junction/bridge table.
// Junction tables typically:
// - Have exactly 2 foreign keys (or sometimes more for multi-way junctions)
// - Primary key consists of the FK columns
// - Few or no additional columns beyond FKs
func detectJunctionTable(table *schema.Table) {
	fkCount := len(table.OutgoingForeignKeys)

	// Need at least 2 FKs
	if fkCount < 2 {
		table.IsJunction = false
		return
	}

	// Collect FK columns
	fkColumns := make(map[string]bool)
	for _, fk := range table.OutgoingForeignKeys {
		for _, col := range fk.FromColumns {
			fkColumns[col] = true
		}
	}

	// Find PK columns
	var pkColumns []string
	for _, con := range table.Constraints {
		if con.ConstraintType == "PRIMARY KEY" {
			pkColumns = con.Columns
			break
		}
	}

	// Check if PK is composed entirely of FK columns
	pkIsAllFKs := len(pkColumns) > 0
	for _, pkCol := range pkColumns {
		if !fkColumns[pkCol] {
			pkIsAllFKs = false
			break
		}
	}

	// Count non-FK columns (excluding common metadata columns)
	nonFKColCount := 0
	metadataCols := map[string]bool{
		"created_at": true,
		"updated_at": true,
		"created_by": true,
		"updated_by": true,
		"deleted_at": true,
		"id":         true,
		"uuid":       true,
	}

	for _, col := range table.Columns {
		if !fkColumns[col.ColumnName] && !metadataCols[strings.ToLower(col.ColumnName)] {
			nonFKColCount++
		}
	}

	// Junction table heuristics:
	// 1. PK is entirely FK columns
	// 2. Or: Has 2+ FKs and few extra columns
	if pkIsAllFKs {
		table.IsJunction = true
		table.JunctionReasoning = "Primary key consists entirely of foreign key columns"
		return
	}

	if fkCount >= 2 && nonFKColCount <= 2 && len(table.Columns) <= 8 {
		table.IsJunction = true
		table.JunctionReasoning = "Table has multiple foreign keys with minimal additional columns"
		return
	}

	// Check for common junction table naming patterns
	name := strings.ToLower(table.TableName)
	junctionPatterns := []string{
		"_to_", "_x_", "_has_", "_rel_", "_link_", "_map_", "_assoc_",
	}
	for _, pattern := range junctionPatterns {
		if strings.Contains(name, pattern) && fkCount >= 2 {
			table.IsJunction = true
			table.JunctionReasoning = "Table name suggests junction pattern and has multiple foreign keys"
			return
		}
	}

	table.IsJunction = false
}
