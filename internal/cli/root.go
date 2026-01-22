package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "xrai",
	Short: "Xrai - LLM-optimized database schema introspection",
	Long: `Xrai statically introspects a relational database and generates a
JSON-first, LLM-optimized schema snapshot.

The output is designed for machine consumption, providing structured
metadata about tables, columns, relationships, and more.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
