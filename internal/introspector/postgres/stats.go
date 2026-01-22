package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/nenorrell/X-Rai/internal/schema"
)

func (i *Introspector) introspectStats(ctx context.Context, schemaName, tableName string) (*schema.Stats, error) {
	stats := &schema.Stats{
		ComputedAt:       time.Now().UTC(),
		SamplingStrategy: "estimate",
		Note:             "Values are estimates from PostgreSQL statistics. Run ANALYZE for more accurate statistics.",
	}

	// Get row count estimate
	rowCount, err := i.introspectRowCount(ctx, schemaName, tableName)
	if err == nil {
		stats.RowCountEstimate = rowCount
	}

	// Get column statistics
	colStats, err := i.introspectColumnStats(ctx, schemaName, tableName)
	if err == nil {
		stats.ColumnStats = colStats
	}

	return stats, nil
}

func (i *Introspector) introspectColumnStats(ctx context.Context, schemaName, tableName string) (map[string]schema.ColStats, error) {
	query := `
		SELECT
			a.attname as column_name,
			s.null_frac as null_fraction,
			s.n_distinct as n_distinct,
			s.most_common_vals::text as most_common_vals
		FROM pg_stats s
		JOIN pg_attribute a ON a.attname = s.attname
		JOIN pg_class c ON c.oid = a.attrelid AND c.relname = s.tablename
		JOIN pg_namespace n ON n.oid = c.relnamespace AND n.nspname = s.schemaname
		WHERE s.schemaname = $1 AND s.tablename = $2
	`

	rows, err := i.pool.Query(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query column stats: %w", err)
	}
	defer rows.Close()

	result := make(map[string]schema.ColStats)
	for rows.Next() {
		var (
			colName        string
			nullFrac       float64
			nDistinct      float64
			mostCommonVals *string
		)

		if err := rows.Scan(&colName, &nullFrac, &nDistinct, &mostCommonVals); err != nil {
			return nil, fmt.Errorf("failed to scan column stats row: %w", err)
		}

		cs := schema.ColStats{
			NullFraction: &nullFrac,
		}

		// n_distinct interpretation:
		// > 0: estimated number of distinct values
		// < 0: negative fraction of rows (e.g., -0.5 means 50% of rows are distinct)
		if nDistinct > 0 {
			distinctCount := int64(nDistinct)
			cs.DistinctCountEstimate = &distinctCount
		}

		result[colName] = cs
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating column stats rows: %w", err)
	}

	return result, nil
}
