package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nenorrell/X-Rai/internal/config"
	"github.com/nenorrell/X-Rai/internal/generator"
	"github.com/nenorrell/X-Rai/internal/introspector/postgres"
	"github.com/spf13/cobra"
)

var (
	dsn               string
	outputDir         string
	schemas           string
	includeViews      bool
	includeRoutines   bool
	includeStats      bool
	redactComments    bool
	redactDefinitions bool
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate schema snapshot from a database",
	Long: `Introspect a PostgreSQL database and generate a complete JSON schema
snapshot optimized for LLM consumption.

Example:
  xrai generate --dsn "postgres://user:pass@localhost:5432/mydb" --output ./schema`,
	RunE: runGenerate,
}

func init() {
	generateCmd.Flags().StringVar(&dsn, "dsn", "", "PostgreSQL connection string (required)")
	generateCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory path (required)")
	generateCmd.Flags().StringVar(&schemas, "schemas", "public", "Comma-separated list of schemas to include")
	generateCmd.Flags().BoolVar(&includeViews, "include-views", false, "Include view artifacts")
	generateCmd.Flags().BoolVar(&includeRoutines, "include-routines", false, "Include function/procedure artifacts")
	generateCmd.Flags().BoolVar(&includeStats, "stats", false, "Enable statistics collection")
	generateCmd.Flags().BoolVar(&redactComments, "redact-comments", false, "Redact all comments from output")
	generateCmd.Flags().BoolVar(&redactDefinitions, "redact-definitions", false, "Redact view/routine SQL definitions")

	generateCmd.MarkFlagRequired("dsn")
	generateCmd.MarkFlagRequired("output")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cfg := config.NewConfig()
	cfg.DSN = dsn
	cfg.OutputDir = outputDir
	cfg.Schemas = parseSchemas(schemas)
	cfg.IncludeViews = includeViews
	cfg.IncludeRoutines = includeRoutines
	cfg.IncludeStats = includeStats
	cfg.RedactComments = redactComments
	cfg.RedactDefinitions = redactDefinitions

	fmt.Fprintf(os.Stderr, "Connecting to database...\n")

	introspector, err := postgres.New(ctx, cfg.DSN)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer introspector.Close()

	fmt.Fprintf(os.Stderr, "Connected to %s (PostgreSQL %s)\n", introspector.DatabaseName(), introspector.Version())
	fmt.Fprintf(os.Stderr, "Introspecting schemas: %v\n", cfg.Schemas)

	db, err := introspector.Introspect(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to introspect database: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Found %d tables\n", len(db.Tables))

	gen := generator.New(cfg)
	if err := gen.Generate(db); err != nil {
		return fmt.Errorf("failed to generate output: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Schema snapshot generated at: %s\n", filepath.Join(cfg.OutputDir, ".xrai"))
	return nil
}

func parseSchemas(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
