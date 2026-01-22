package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/nenorrell/xrai/internal/schema"
)

func (i *Introspector) introspectTables(ctx context.Context, schemas []string) ([]*schema.Table, error) {
	query := `
		SELECT
			t.table_schema,
			t.table_name,
			COALESCE(obj_description((t.table_schema || '.' || t.table_name)::regclass), '') as table_comment
		FROM information_schema.tables t
		WHERE t.table_schema = ANY($1)
		  AND t.table_type = 'BASE TABLE'
		ORDER BY t.table_schema, t.table_name
	`

	rows, err := i.pool.Query(ctx, query, schemas)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []*schema.Table
	for rows.Next() {
		var schemaName, tableName, comment string
		if err := rows.Scan(&schemaName, &tableName, &comment); err != nil {
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}

		tables = append(tables, &schema.Table{
			TableName:  tableName,
			SchemaName: schemaName,
			TableType:  "BASE TABLE",
			Comment:    comment,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating table rows: %w", err)
	}

	return tables, nil
}

func (i *Introspector) introspectRowCount(ctx context.Context, schemaName, tableName string) (*int64, error) {
	// Use pg_class for estimated row count (much faster than COUNT(*))
	query := `
		SELECT reltuples::bigint
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = $1 AND c.relname = $2
	`

	var count int64
	err := i.pool.QueryRow(ctx, query, schemaName, tableName).Scan(&count)
	if err != nil {
		return nil, err
	}

	if count < 0 {
		return nil, nil
	}

	return &count, nil
}

// schemaPlaceholders generates $1, $2, ... for schema list
func schemaPlaceholders(schemas []string) string {
	placeholders := make([]string, len(schemas))
	for i := range schemas {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	return strings.Join(placeholders, ", ")
}
