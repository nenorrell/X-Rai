package postgres

import (
	"context"
	"fmt"

	"github.com/nenorrell/X-Rai/internal/schema"
)

func (i *Introspector) introspectConstraints(ctx context.Context, schemaName, tableName string) ([]*schema.Constraint, error) {
	query := `
		SELECT
			con.conname as constraint_name,
			con.contype::text as constraint_type,
			ARRAY(
				SELECT a.attname
				FROM unnest(con.conkey) WITH ORDINALITY AS k(attnum, ord)
				JOIN pg_attribute a ON a.attrelid = con.conrelid AND a.attnum = k.attnum
				ORDER BY k.ord
			) as columns,
			pg_get_constraintdef(con.oid) as definition
		FROM pg_constraint con
		JOIN pg_class c ON c.oid = con.conrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = $1 AND c.relname = $2
		  AND con.contype IN ('p', 'u', 'c', 'x')
		ORDER BY
			CASE con.contype
				WHEN 'p' THEN 1
				WHEN 'u' THEN 2
				WHEN 'c' THEN 3
				WHEN 'x' THEN 4
			END,
			con.conname
	`

	rows, err := i.pool.Query(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query constraints: %w", err)
	}
	defer rows.Close()

	var constraints []*schema.Constraint
	for rows.Next() {
		var (
			constraintName, constraintType, definition string
			columns                                    []string
		)

		err := rows.Scan(&constraintName, &constraintType, &columns, &definition)
		if err != nil {
			return nil, fmt.Errorf("failed to scan constraint row: %w", err)
		}

		conType := mapConstraintType(constraintType)

		con := &schema.Constraint{
			ConstraintName: constraintName,
			ConstraintType: conType,
			Columns:        columns,
			Definition:     &definition,
		}

		// Extract check expression for check constraints
		if constraintType == "c" {
			con.Expression = &definition
		}

		constraints = append(constraints, con)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating constraint rows: %w", err)
	}

	return constraints, nil
}

func mapConstraintType(pgType string) string {
	switch pgType {
	case "p":
		return "PRIMARY KEY"
	case "u":
		return "UNIQUE"
	case "c":
		return "CHECK"
	case "x":
		return "EXCLUSION"
	case "f":
		return "FOREIGN KEY"
	default:
		return pgType
	}
}
