package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/nenorrell/X-Rai/internal/config"
	"github.com/nenorrell/X-Rai/internal/heuristics"
	"github.com/nenorrell/X-Rai/internal/schema"
	"github.com/nenorrell/X-Rai/internal/toon"
)

// Generator handles JSON output generation.
type Generator struct {
	cfg       *config.Config
	outputDir string
}

// New creates a new Generator.
func New(cfg *config.Config) *Generator {
	return &Generator{
		cfg:       cfg,
		outputDir: filepath.Join(cfg.OutputDir, ".xrai"),
	}
}

// Generate produces all output files for the database schema.
func (g *Generator) Generate(db *schema.Database) error {
	// Create output directory
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Apply heuristics
	heuristics.ApplyAllHeuristics(db)

	// Generate llms.txt (LLM entry point)
	if err := g.generateLLMsTxt(db); err != nil {
		return fmt.Errorf("failed to generate llms.txt: %w", err)
	}

	// Generate manifest
	if err := g.generateManifest(db); err != nil {
		return fmt.Errorf("failed to generate manifest: %w", err)
	}

	// Generate database index
	if err := g.generateDatabaseIndex(db); err != nil {
		return fmt.Errorf("failed to generate database index: %w", err)
	}

	// Generate relationships graph
	if err := g.generateRelationships(db); err != nil {
		return fmt.Errorf("failed to generate relationships: %w", err)
	}

	// Generate domain groupings
	if err := g.generateDomains(db); err != nil {
		return fmt.Errorf("failed to generate domains: %w", err)
	}

	// Generate table artifacts
	if err := g.generateTables(db); err != nil {
		return fmt.Errorf("failed to generate tables: %w", err)
	}

	// Generate view artifacts (if enabled)
	if g.cfg.IncludeViews && len(db.Views) > 0 {
		if err := g.generateViews(db); err != nil {
			return fmt.Errorf("failed to generate views: %w", err)
		}
	}

	// Generate routine artifacts (if enabled)
	if g.cfg.IncludeRoutines && len(db.Routines) > 0 {
		if err := g.generateRoutines(db); err != nil {
			return fmt.Errorf("failed to generate routines: %w", err)
		}
	}

	// Generate enum files
	if len(db.Enums) > 0 {
		if err := g.generateEnums(db); err != nil {
			return fmt.Errorf("failed to generate enums: %w", err)
		}
	}

	// Generate sequence files
	if len(db.Sequences) > 0 {
		if err := g.generateSequences(db); err != nil {
			return fmt.Errorf("failed to generate sequences: %w", err)
		}
	}

	// Generate type files
	if len(db.Types) > 0 {
		if err := g.generateTypes(db); err != nil {
			return fmt.Errorf("failed to generate types: %w", err)
		}
	}

	return nil
}

// writeTOON writes a value as TOON format to a file.
func (g *Generator) writeTOON(path string, v interface{}) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	data, err := toon.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal TOON: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// sanitizeName converts a database object name to a filesystem-safe name.
func sanitizeName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")

	// Replace any non-alphanumeric characters (except underscore) with underscore
	re := regexp.MustCompile(`[^a-z0-9_]`)
	name = re.ReplaceAllString(name, "_")

	// Collapse multiple underscores
	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}

	// Trim leading/trailing underscores
	name = strings.Trim(name, "_")

	if name == "" {
		name = "unnamed"
	}

	return name
}

// timestamp returns the current UTC time in RFC3339 format.
func timestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
