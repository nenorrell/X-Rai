package postgres

import (
	"context"
	"fmt"

	"github.com/nenorrell/X-Rai/internal/schema"
)

func (i *Introspector) introspectEnums(ctx context.Context, schemas []string) ([]*schema.Enum, error) {
	query := `
		SELECT
			n.nspname as schema_name,
			t.typname as enum_name,
			ARRAY(
				SELECT e.enumlabel
				FROM pg_enum e
				WHERE e.enumtypid = t.oid
				ORDER BY e.enumsortorder
			) as enum_values,
			COALESCE(obj_description(t.oid, 'pg_type'), '') as comment
		FROM pg_type t
		JOIN pg_namespace n ON n.oid = t.typnamespace
		WHERE t.typtype = 'e'
		  AND n.nspname = ANY($1)
		ORDER BY n.nspname, t.typname
	`

	rows, err := i.pool.Query(ctx, query, schemas)
	if err != nil {
		return nil, fmt.Errorf("failed to query enums: %w", err)
	}
	defer rows.Close()

	var enums []*schema.Enum
	for rows.Next() {
		var schemaName, enumName, comment string
		var values []string

		if err := rows.Scan(&schemaName, &enumName, &values, &comment); err != nil {
			return nil, fmt.Errorf("failed to scan enum row: %w", err)
		}

		enums = append(enums, &schema.Enum{
			EnumName:   enumName,
			SchemaName: schemaName,
			Values:     values,
			Comment:    comment,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating enum rows: %w", err)
	}

	return enums, nil
}
