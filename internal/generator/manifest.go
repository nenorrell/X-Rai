package generator

import (
	"path/filepath"

	"github.com/nenorrell/xrai/internal/schema"
)

func (g *Generator) generateManifest(db *schema.Database) error {
	manifest := schema.Manifest{
		GenerationTimestamp: timestamp(),
		DatabaseEngine:      db.Engine,
		DatabaseVersion:     db.Version,
		DatabaseName:        db.Name,
		IncludedSchemas:     db.Schemas,
		IncludedTablesCount: len(db.Tables),
		EnabledArtifacts: schema.EnabledArtifacts{
			Tables:    true,
			Views:     g.cfg.IncludeViews && len(db.Views) > 0,
			Routines:  g.cfg.IncludeRoutines && len(db.Routines) > 0,
			Enums:     len(db.Enums) > 0,
			Sequences: len(db.Sequences) > 0,
			Types:     len(db.Types) > 0,
			Stats:     g.cfg.IncludeStats,
		},
		StatsEnabled: g.cfg.IncludeStats,
		UsageEnabled: false, // Usage heuristics not implemented yet
	}

	return g.writeTOON(filepath.Join(g.outputDir, "xrai.manifest.toon"), manifest)
}
