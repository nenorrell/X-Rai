package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nenorrell/X-Rai/internal/schema"
)

func (g *Generator) generateLLMsTxt(db *schema.Database) error {
	var sb strings.Builder

	// Header - brief and action-oriented
	sb.WriteString("# ")
	sb.WriteString(db.Name)
	sb.WriteString(" Schema Reference\n\n")

	sb.WriteString("This directory contains complete schema documentation for ")
	sb.WriteString(db.Name)
	sb.WriteString(" (")
	sb.WriteString(db.Engine)
	sb.WriteString(" ")
	sb.WriteString(db.Version)
	sb.WriteString(").\n\n")

	// Discovery guide - the key section
	sb.WriteString("## Finding What You Need\n\n")
	sb.WriteString("| Task | File to Read |\n")
	sb.WriteString("|------|-------------|\n")
	sb.WriteString("| List all tables | `db.index.toon` |\n")
	sb.WriteString("| See how tables connect | `db.relationships.toon` |\n")
	sb.WriteString("| Find tables by domain/feature | `db.domains.toon` |\n")
	sb.WriteString("| Get columns for a table | `tables/<name>/table.columns.toon` |\n")
	sb.WriteString("| Get primary/foreign keys | `tables/<name>/table.relations.toon` |\n")
	sb.WriteString("| Check indexes/constraints | `tables/<name>/table.indexes.toon` |\n")
	if len(db.Enums) > 0 {
		sb.WriteString("| Look up enum values | `enums/<name>.toon` |\n")
	}
	if len(db.Views) > 0 {
		sb.WriteString("| View definitions | `views/<name>/view.definition.toon` |\n")
	}
	if len(db.Routines) > 0 {
		sb.WriteString("| Function/procedure code | `routines/functions/<name>.toon` |\n")
	}
	sb.WriteString("\n")

	// Quick stats
	sb.WriteString("## At a Glance\n\n")
	sb.WriteString(fmt.Sprintf("- **%d tables** across %s\n", len(db.Tables), formatSchemaList(db.Schemas)))
	if len(db.Views) > 0 {
		sb.WriteString(fmt.Sprintf("- **%d views**\n", len(db.Views)))
	}
	if len(db.Enums) > 0 {
		sb.WriteString(fmt.Sprintf("- **%d enums**\n", len(db.Enums)))
	}
	if len(db.Routines) > 0 {
		sb.WriteString(fmt.Sprintf("- **%d routines**\n", len(db.Routines)))
	}
	sb.WriteString("\n")

	// Table index - organized by importance
	sb.WriteString("## Table Index\n\n")
	writeTableIndex(&sb, db.Tables)

	// Relationship hints - where to start exploring
	if hasRelationships(db.Tables) {
		sb.WriteString("## Key Entry Points\n\n")
		sb.WriteString("Start exploring from these highly-connected tables:\n\n")
		writeEntryPoints(&sb, db.Tables)
	}

	// Enums quick reference
	if len(db.Enums) > 0 {
		sb.WriteString("## Enum Quick Reference\n\n")
		for _, e := range db.Enums {
			sb.WriteString(fmt.Sprintf("- **%s**: `%s`\n", e.EnumName, strings.Join(e.Values, "`, `")))
		}
		sb.WriteString("\n")
	}

	return g.writeFile(filepath.Join(g.outputDir, "llms.txt"), sb.String())
}

func formatSchemaList(schemas []string) string {
	if len(schemas) == 1 {
		return fmt.Sprintf("`%s` schema", schemas[0])
	}
	return fmt.Sprintf("%d schemas (%s)", len(schemas), strings.Join(schemas, ", "))
}

func writeTableIndex(sb *strings.Builder, tables []*schema.Table) {
	// Group tables
	core := make([]*schema.Table, 0)
	junction := make([]*schema.Table, 0)
	lookup := make([]*schema.Table, 0)
	other := make([]*schema.Table, 0)

	for _, t := range tables {
		switch {
		case hasTag(t, "core"):
			core = append(core, t)
		case hasTag(t, "junction"):
			junction = append(junction, t)
		case hasTag(t, "lookup"):
			lookup = append(lookup, t)
		default:
			other = append(other, t)
		}
	}

	// Core tables - most important
	if len(core) > 0 {
		sb.WriteString("**Core** (central entities): ")
		sb.WriteString(tableNames(core))
		sb.WriteString("\n\n")
	}

	// Other tables
	if len(other) > 0 {
		sb.WriteString("**Tables**: ")
		sb.WriteString(tableNames(other))
		sb.WriteString("\n\n")
	}

	// Junction tables
	if len(junction) > 0 {
		sb.WriteString("**Junction** (many-to-many): ")
		sb.WriteString(tableNames(junction))
		sb.WriteString("\n\n")
	}

	// Lookup tables
	if len(lookup) > 0 {
		sb.WriteString("**Lookup**: ")
		sb.WriteString(tableNames(lookup))
		sb.WriteString("\n\n")
	}
}

func hasTag(t *schema.Table, tag string) bool {
	for _, tg := range t.Tags {
		if tg == tag {
			return true
		}
	}
	return false
}

func tableNames(tables []*schema.Table) string {
	names := make([]string, 0, len(tables))
	for _, t := range tables {
		names = append(names, fmt.Sprintf("`%s`", t.TableName))
	}
	return strings.Join(names, ", ")
}

func hasRelationships(tables []*schema.Table) bool {
	for _, t := range tables {
		if len(t.IncomingForeignKeys) > 0 || len(t.OutgoingForeignKeys) > 0 {
			return true
		}
	}
	return false
}

func writeEntryPoints(sb *strings.Builder, tables []*schema.Table) {
	type entry struct {
		name     string
		incoming int
		refs     []string
	}

	entries := make([]entry, 0)
	for _, t := range tables {
		if len(t.IncomingForeignKeys) >= 2 {
			refs := make([]string, 0)
			for _, fk := range t.IncomingForeignKeys {
				refs = append(refs, fk.FromTable)
			}
			entries = append(entries, entry{
				name:     t.TableName,
				incoming: len(t.IncomingForeignKeys),
				refs:     refs,
			})
		}
	}

	// Sort by incoming count
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].incoming > entries[j].incoming
	})

	// Show top 5
	limit := 5
	if len(entries) < limit {
		limit = len(entries)
	}

	for i := 0; i < limit; i++ {
		e := entries[i]
		sb.WriteString(fmt.Sprintf("- `%s` ← referenced by %d tables\n", e.name, e.incoming))
	}
	sb.WriteString("\n")
}

func (g *Generator) writeFile(path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return os.WriteFile(path, []byte(content), 0644)
}
