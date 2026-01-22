package generator

import (
	"path/filepath"

	"github.com/nenorrell/xrai/internal/schema"
)

func (g *Generator) generateEnums(db *schema.Database) error {
	enumsDir := filepath.Join(g.outputDir, "enums")

	for _, enum := range db.Enums {
		filename := sanitizeName(enum.EnumName) + ".toon"

		if err := g.writeTOON(filepath.Join(enumsDir, filename), enum); err != nil {
			return err
		}
	}

	return nil
}
