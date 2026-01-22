package generator

import (
	"path/filepath"

	"github.com/nenorrell/X-Rai/internal/schema"
)

func (g *Generator) generateTypes(db *schema.Database) error {
	typesDir := filepath.Join(g.outputDir, "types")

	for _, t := range db.Types {
		filename := sanitizeName(t.TypeName) + ".toon"

		if err := g.writeTOON(filepath.Join(typesDir, filename), t); err != nil {
			return err
		}
	}

	return nil
}
