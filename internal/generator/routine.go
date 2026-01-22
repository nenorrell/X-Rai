package generator

import (
	"path/filepath"

	"github.com/nenorrell/xrai/internal/schema"
)

func (g *Generator) generateRoutines(db *schema.Database) error {
	routinesDir := filepath.Join(g.outputDir, "routines")
	functionsDir := filepath.Join(routinesDir, "functions")
	proceduresDir := filepath.Join(routinesDir, "procedures")

	for _, routine := range db.Routines {
		var targetDir string
		if routine.RoutineType == "procedure" {
			targetDir = proceduresDir
		} else {
			targetDir = functionsDir
		}

		filename := sanitizeName(routine.RoutineName) + ".toon"

		output := *routine
		if g.cfg.RedactDefinitions {
			output.Definition = nil
		}

		if err := g.writeTOON(filepath.Join(targetDir, filename), output); err != nil {
			return err
		}
	}

	return nil
}
