package heuristics

import "github.com/nenorrell/xrai/internal/schema"

// ApplyAllHeuristics runs all heuristic analyses on the database schema.
func ApplyAllHeuristics(db *schema.Database) {
	for _, table := range db.Tables {
		// Detect junction tables
		detectJunctionTable(table)

		// Apply tags
		applyTableTags(table)

		// Determine FK cardinality
		for _, fk := range table.OutgoingForeignKeys {
			fk.Cardinality = determineCardinality(table, fk)
		}
	}
}
