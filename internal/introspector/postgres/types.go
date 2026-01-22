package postgres

import (
	"context"
	"fmt"

	"github.com/nenorrell/xrai/internal/schema"
)

func (i *Introspector) introspectTypes(ctx context.Context, schemas []string) ([]*schema.Type, error) {
	// Query composite types
	query := `
		SELECT
			n.nspname as schema_name,
			t.typname as type_name,
			CASE t.typtype
				WHEN 'c' THEN 'composite'
				WHEN 'd' THEN 'domain'
				WHEN 'r' THEN 'range'
				ELSE 'other'
			END as type_kind,
			COALESCE(obj_description(t.oid, 'pg_type'), '') as comment
		FROM pg_type t
		JOIN pg_namespace n ON n.oid = t.typnamespace
		WHERE t.typtype IN ('c', 'd', 'r')
		  AND n.nspname = ANY($1)
		  AND NOT EXISTS (
			-- Exclude types auto-created for tables
			SELECT 1 FROM pg_class c WHERE c.reltype = t.oid AND c.relkind IN ('r', 'v', 'm', 'p')
		  )
		ORDER BY n.nspname, t.typname
	`

	rows, err := i.pool.Query(ctx, query, schemas)
	if err != nil {
		return nil, fmt.Errorf("failed to query types: %w", err)
	}
	defer rows.Close()

	var types []*schema.Type
	for rows.Next() {
		var schemaName, typeName, typeKind, comment string

		if err := rows.Scan(&schemaName, &typeName, &typeKind, &comment); err != nil {
			return nil, fmt.Errorf("failed to scan type row: %w", err)
		}

		t := &schema.Type{
			TypeName:   typeName,
			SchemaName: schemaName,
			TypeKind:   typeKind,
			Comment:    comment,
		}

		types = append(types, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating type rows: %w", err)
	}

	// Fetch attributes for composite types
	for _, t := range types {
		if t.TypeKind == "composite" {
			attrs, err := i.introspectTypeAttributes(ctx, t.SchemaName, t.TypeName)
			if err != nil {
				// Non-fatal
				continue
			}
			t.Attributes = attrs
		} else if t.TypeKind == "domain" {
			base, constraint, err := i.introspectDomainInfo(ctx, t.SchemaName, t.TypeName)
			if err == nil {
				t.BaseType = base
				t.Constraint = constraint
			}
		}
	}

	return types, nil
}

func (i *Introspector) introspectTypeAttributes(ctx context.Context, schemaName, typeName string) ([]schema.TypeAttribute, error) {
	query := `
		SELECT
			a.attname as attr_name,
			format_type(a.atttypid, a.atttypmod) as data_type
		FROM pg_attribute a
		JOIN pg_type t ON t.typrelid = a.attrelid
		JOIN pg_namespace n ON n.oid = t.typnamespace
		WHERE n.nspname = $1
		  AND t.typname = $2
		  AND a.attnum > 0
		  AND NOT a.attisdropped
		ORDER BY a.attnum
	`

	rows, err := i.pool.Query(ctx, query, schemaName, typeName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attrs []schema.TypeAttribute
	for rows.Next() {
		var name, dataType string
		if err := rows.Scan(&name, &dataType); err != nil {
			return nil, err
		}
		attrs = append(attrs, schema.TypeAttribute{
			Name:     name,
			DataType: dataType,
		})
	}

	return attrs, rows.Err()
}

func (i *Introspector) introspectDomainInfo(ctx context.Context, schemaName, typeName string) (string, *string, error) {
	query := `
		SELECT
			format_type(t.typbasetype, t.typtypmod) as base_type,
			pg_get_constraintdef(con.oid) as constraint_def
		FROM pg_type t
		JOIN pg_namespace n ON n.oid = t.typnamespace
		LEFT JOIN pg_constraint con ON con.contypid = t.oid
		WHERE n.nspname = $1
		  AND t.typname = $2
		  AND t.typtype = 'd'
	`

	var baseType string
	var constraint *string

	err := i.pool.QueryRow(ctx, query, schemaName, typeName).Scan(&baseType, &constraint)
	if err != nil {
		return "", nil, err
	}

	return baseType, constraint, nil
}
