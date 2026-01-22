package generator

import (
	"path/filepath"
	"sort"

	"github.com/nenorrell/X-Rai/internal/schema"
)

func (g *Generator) generateDatabaseIndex(db *schema.Database) error {
	index := schema.DatabaseIndex{
		Tables: make([]schema.TableIndexEntry, 0, len(db.Tables)),
	}

	for _, table := range db.Tables {
		entry := schema.TableIndexEntry{
			TableName:          table.TableName,
			SchemaName:         table.SchemaName,
			RowCountEstimate:   table.RowCountEstimate,
			ForeignKeyOutCount: len(table.OutgoingForeignKeys),
			ForeignKeyInCount:  len(table.IncomingForeignKeys),
			Tags:               table.Tags,
		}

		// Add short description from comment if not redacted
		if !g.cfg.RedactComments && table.Comment != "" {
			entry.ShortDescription = truncateDescription(table.Comment, 200)
		}

		// Add primary key columns
		entry.PrimaryKeyColumns = getPrimaryKeyColumns(table)

		index.Tables = append(index.Tables, entry)
	}

	// Sort tables by name for deterministic output
	sort.Slice(index.Tables, func(i, j int) bool {
		if index.Tables[i].SchemaName != index.Tables[j].SchemaName {
			return index.Tables[i].SchemaName < index.Tables[j].SchemaName
		}
		return index.Tables[i].TableName < index.Tables[j].TableName
	})

	// Determine recommended start tables
	index.RecommendedStartTables = findRecommendedStartTables(db.Tables)

	return g.writeTOON(filepath.Join(g.outputDir, "db.index.toon"), index)
}

func getPrimaryKeyColumns(table *schema.Table) []string {
	for _, con := range table.Constraints {
		if con.ConstraintType == "PRIMARY KEY" {
			return con.Columns
		}
	}
	return nil
}

func truncateDescription(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// findRecommendedStartTables identifies good entry points for understanding the schema.
// Prioritizes tables that have many incoming foreign keys (central entities)
// and are tagged as "core".
func findRecommendedStartTables(tables []*schema.Table) []string {
	type tableScore struct {
		name  string
		score int
	}

	var scores []tableScore
	for _, t := range tables {
		score := 0

		// Incoming FK count is a strong signal of importance
		score += len(t.IncomingForeignKeys) * 2

		// Core tables are important
		for _, tag := range t.Tags {
			if tag == "core" {
				score += 5
			}
		}

		// Junction tables are less interesting as start points
		if t.IsJunction {
			score -= 3
		}

		// Lookup tables are also less interesting
		for _, tag := range t.Tags {
			if tag == "lookup" {
				score -= 2
			}
		}

		scores = append(scores, tableScore{name: t.TableName, score: score})
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Return top 5 with positive scores
	var result []string
	for i := 0; i < len(scores) && i < 5; i++ {
		if scores[i].score > 0 {
			result = append(result, scores[i].name)
		}
	}

	return result
}
