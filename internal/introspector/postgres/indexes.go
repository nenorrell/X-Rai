package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/nenorrell/xrai/internal/schema"
)

func (i *Introspector) introspectIndexes(ctx context.Context, schemaName, tableName string) ([]*schema.Index, error) {
	query := `
		SELECT
			i.relname as index_name,
			ix.indisunique as is_unique,
			ix.indisprimary as is_primary,
			am.amname as index_type,
			pg_get_indexdef(ix.indexrelid) as index_def,
			ARRAY(
				SELECT a.attname
				FROM unnest(ix.indkey) WITH ORDINALITY AS k(attnum, ord)
				JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = k.attnum
				ORDER BY k.ord
			) as columns,
			pg_get_expr(ix.indpred, ix.indrelid) as predicate,
			pg_get_expr(ix.indexprs, ix.indrelid) as expression
		FROM pg_index ix
		JOIN pg_class t ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_namespace n ON n.oid = t.relnamespace
		JOIN pg_am am ON am.oid = i.relam
		WHERE n.nspname = $1 AND t.relname = $2
		ORDER BY i.relname
	`

	rows, err := i.pool.Query(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes: %w", err)
	}
	defer rows.Close()

	var indexes []*schema.Index
	for rows.Next() {
		var (
			indexName, indexType, indexDef string
			isUnique, isPrimary            bool
			columns                        []string
			predicate, expression          *string
		)

		err := rows.Scan(
			&indexName, &isUnique, &isPrimary, &indexType,
			&indexDef, &columns, &predicate, &expression,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan index row: %w", err)
		}

		idx := &schema.Index{
			IndexName:  indexName,
			Unique:     isUnique,
			Primary:    isPrimary,
			IndexType:  indexType,
			Columns:    columns,
			Predicate:  predicate,
			Expression: expression,
			Definition: indexDef,
		}

		// Extract INCLUDE columns from definition if present
		idx.IncludeColumns = extractIncludeColumns(indexDef)

		indexes = append(indexes, idx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating index rows: %w", err)
	}

	return indexes, nil
}

// extractIncludeColumns parses INCLUDE columns from index definition
func extractIncludeColumns(def string) []string {
	// Look for INCLUDE (col1, col2, ...)
	upper := strings.ToUpper(def)
	idx := strings.Index(upper, " INCLUDE ")
	if idx == -1 {
		return nil
	}

	rest := def[idx+9:] // Skip " INCLUDE "
	start := strings.Index(rest, "(")
	end := strings.Index(rest, ")")
	if start == -1 || end == -1 || end <= start {
		return nil
	}

	colsPart := rest[start+1 : end]
	parts := strings.Split(colsPart, ",")
	var cols []string
	for _, p := range parts {
		col := strings.TrimSpace(p)
		if col != "" {
			cols = append(cols, col)
		}
	}
	return cols
}
