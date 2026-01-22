package generator

import (
	"path/filepath"

	"github.com/nenorrell/X-Rai/internal/schema"
)

func (g *Generator) generateSequences(db *schema.Database) error {
	sequencesDir := filepath.Join(g.outputDir, "sequences")

	for _, seq := range db.Sequences {
		filename := sanitizeName(seq.SequenceName) + ".toon"

		if err := g.writeTOON(filepath.Join(sequencesDir, filename), seq); err != nil {
			return err
		}
	}

	return nil
}
