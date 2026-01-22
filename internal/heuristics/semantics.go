package heuristics

import (
	"github.com/nenorrell/xrai/internal/schema"
)

// determineCardinality estimates the cardinality of a foreign key relationship.
// Returns: "one-to-one", "one-to-many", "many-to-one", "many-to-many"
func determineCardinality(table *schema.Table, fk *schema.ForeignKey) string {
	// Check if FK columns have a unique constraint
	fkColsAreUnique := hasFKUniqueConstraint(table, fk.FromColumns)

	// Junction tables suggest many-to-many
	if table.IsJunction {
		return "many-to-many"
	}

	// If FK columns are unique, it's one-to-one
	if fkColsAreUnique {
		return "one-to-one"
	}

	// If FK is nullable, relationship is optional many-to-one
	if fk.Nullable {
		return "many-to-one"
	}

	// Default: many-to-one (most common case)
	return "many-to-one"
}

// hasFKUniqueConstraint checks if the FK columns have a unique or PK constraint.
func hasFKUniqueConstraint(table *schema.Table, fkCols []string) bool {
	if len(fkCols) == 0 {
		return false
	}

	for _, con := range table.Constraints {
		if con.ConstraintType == "PRIMARY KEY" || con.ConstraintType == "UNIQUE" {
			if columnsMatch(con.Columns, fkCols) {
				return true
			}
		}
	}

	// Also check indexes
	for _, idx := range table.Indexes {
		if idx.Unique && columnsMatch(idx.Columns, fkCols) {
			return true
		}
	}

	return false
}

// columnsMatch checks if two column lists contain the same columns (order-insensitive).
func columnsMatch(cols1, cols2 []string) bool {
	if len(cols1) != len(cols2) {
		return false
	}

	set := make(map[string]bool)
	for _, c := range cols1 {
		set[c] = true
	}

	for _, c := range cols2 {
		if !set[c] {
			return false
		}
	}

	return true
}
