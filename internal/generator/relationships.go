package generator

import (
	"path/filepath"
	"sort"

	"github.com/nenorrell/xrai/internal/schema"
)

func (g *Generator) generateRelationships(db *schema.Database) error {
	graph := schema.RelationshipGraph{
		Nodes: make([]schema.RelationshipNode, 0, len(db.Tables)),
		Edges: make([]schema.RelationshipEdge, 0),
	}

	// Add nodes
	for _, table := range db.Tables {
		graph.Nodes = append(graph.Nodes, schema.RelationshipNode{
			TableName:  table.TableName,
			SchemaName: table.SchemaName,
		})
	}

	// Sort nodes for deterministic output
	sort.Slice(graph.Nodes, func(i, j int) bool {
		if graph.Nodes[i].SchemaName != graph.Nodes[j].SchemaName {
			return graph.Nodes[i].SchemaName < graph.Nodes[j].SchemaName
		}
		return graph.Nodes[i].TableName < graph.Nodes[j].TableName
	})

	// Add edges from foreign keys
	for _, table := range db.Tables {
		for _, fk := range table.OutgoingForeignKeys {
			graph.Edges = append(graph.Edges, schema.RelationshipEdge{
				FromTable:      table.TableName,
				FromSchema:     table.SchemaName,
				FromColumns:    fk.FromColumns,
				ToTable:        fk.ToTable,
				ToSchema:       fk.ToSchema,
				ToColumns:      fk.ToColumns,
				ConstraintName: fk.ConstraintName,
			})
		}
	}

	// Sort edges for deterministic output
	sort.Slice(graph.Edges, func(i, j int) bool {
		return graph.Edges[i].ConstraintName < graph.Edges[j].ConstraintName
	})

	// Identify junction table candidates
	for _, table := range db.Tables {
		if table.IsJunction {
			graph.JunctionTableCandidates = append(graph.JunctionTableCandidates, table.TableName)
		}
	}
	sort.Strings(graph.JunctionTableCandidates)

	return g.writeTOON(filepath.Join(g.outputDir, "db.relationships.toon"), graph)
}
