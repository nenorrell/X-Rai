package generator

import (
	"path/filepath"

	"github.com/nenorrell/xrai/internal/schema"
)

func (g *Generator) generateViews(db *schema.Database) error {
	viewsDir := filepath.Join(g.outputDir, "views")

	for _, view := range db.Views {
		viewDir := filepath.Join(viewsDir, sanitizeName(view.ViewName))

		// Generate view.definition.toon
		if err := g.generateViewDefinition(viewDir, view); err != nil {
			return err
		}

		// Generate view.columns.toon
		if err := g.generateViewColumns(viewDir, view); err != nil {
			return err
		}

		// Generate view.dependencies.toon
		if err := g.generateViewDependencies(viewDir, view); err != nil {
			return err
		}

		// Generate view.comments.toon
		if err := g.generateViewComments(viewDir, view); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generateViewDefinition(viewDir string, view *schema.View) error {
	def := schema.ViewDefinition{
		ViewName:   view.ViewName,
		SchemaName: view.SchemaName,
	}

	if !g.cfg.RedactDefinitions {
		def.Definition = view.Definition
	}

	return g.writeTOON(filepath.Join(viewDir, "view.definition.toon"), def)
}

func (g *Generator) generateViewColumns(viewDir string, view *schema.View) error {
	columns := make([]schema.Column, 0, len(view.Columns))
	for _, col := range view.Columns {
		columns = append(columns, *col)
	}

	output := schema.ColumnsOutput{Columns: columns}
	return g.writeTOON(filepath.Join(viewDir, "view.columns.toon"), output)
}

func (g *Generator) generateViewDependencies(viewDir string, view *schema.View) error {
	deps := schema.ViewDependencies{
		DependsOnTables: view.DependsOnTables,
		DependsOnViews:  view.DependsOnViews,
	}

	return g.writeTOON(filepath.Join(viewDir, "view.dependencies.toon"), deps)
}

func (g *Generator) generateViewComments(viewDir string, view *schema.View) error {
	output := schema.ViewComments{
		ColumnComments: make(map[string]string),
	}

	if !g.cfg.RedactComments {
		output.ViewComment = view.Comment

		for _, col := range view.Columns {
			if col.Comment != "" {
				output.ColumnComments[col.ColumnName] = col.Comment
			}
		}
	}

	return g.writeTOON(filepath.Join(viewDir, "view.comments.toon"), output)
}
