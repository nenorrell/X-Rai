package postgres

import (
	"context"
	"fmt"

	"github.com/nenorrell/X-Rai/internal/schema"
)

func (i *Introspector) introspectForeignKeys(ctx context.Context, schemaName, tableName string) ([]*schema.ForeignKey, error) {
	query := `
		SELECT
			con.conname as constraint_name,
			ARRAY(
				SELECT a.attname
				FROM unnest(con.conkey) WITH ORDINALITY AS k(attnum, ord)
				JOIN pg_attribute a ON a.attrelid = con.conrelid AND a.attnum = k.attnum
				ORDER BY k.ord
			) as from_columns,
			nf.nspname as to_schema,
			cf.relname as to_table,
			ARRAY(
				SELECT a.attname
				FROM unnest(con.confkey) WITH ORDINALITY AS k(attnum, ord)
				JOIN pg_attribute a ON a.attrelid = con.confrelid AND a.attnum = k.attnum
				ORDER BY k.ord
			) as to_columns,
			con.confupdtype::text as on_update,
			con.confdeltype::text as on_delete
		FROM pg_constraint con
		JOIN pg_class c ON c.oid = con.conrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		JOIN pg_class cf ON cf.oid = con.confrelid
		JOIN pg_namespace nf ON nf.oid = cf.relnamespace
		WHERE n.nspname = $1 AND c.relname = $2
		  AND con.contype = 'f'
		ORDER BY con.conname
	`

	rows, err := i.pool.Query(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query foreign keys: %w", err)
	}
	defer rows.Close()

	var fks []*schema.ForeignKey
	for rows.Next() {
		var (
			constraintName, toSchema, toTable string
			fromColumns, toColumns            []string
			onUpdateChar, onDeleteChar        string
		)

		err := rows.Scan(
			&constraintName, &fromColumns, &toSchema, &toTable, &toColumns,
			&onUpdateChar, &onDeleteChar,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan foreign key row: %w", err)
		}

		fk := &schema.ForeignKey{
			ConstraintName: constraintName,
			FromSchema:     schemaName,
			FromTable:      tableName,
			FromColumns:    fromColumns,
			ToSchema:       toSchema,
			ToTable:        toTable,
			ToColumns:      toColumns,
			OnUpdate:       mapFKAction(onUpdateChar),
			OnDelete:       mapFKAction(onDeleteChar),
		}

		fks = append(fks, fk)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating foreign key rows: %w", err)
	}

	// Determine nullability for each FK based on column nullability
	if len(fks) > 0 {
		if err := i.determineFKNullability(ctx, schemaName, tableName, fks); err != nil {
			// Non-fatal
		}
	}

	return fks, nil
}

func (i *Introspector) determineFKNullability(ctx context.Context, schemaName, tableName string, fks []*schema.ForeignKey) error {
	query := `
		SELECT column_name, is_nullable
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
	`

	rows, err := i.pool.Query(ctx, query, schemaName, tableName)
	if err != nil {
		return err
	}
	defer rows.Close()

	nullability := make(map[string]bool)
	for rows.Next() {
		var colName, isNullable string
		if err := rows.Scan(&colName, &isNullable); err != nil {
			return err
		}
		nullability[colName] = isNullable == "YES"
	}

	for _, fk := range fks {
		// FK is nullable if any of its columns are nullable
		nullable := false
		for _, col := range fk.FromColumns {
			if nullability[col] {
				nullable = true
				break
			}
		}
		fk.Nullable = nullable
	}

	return rows.Err()
}

func mapFKAction(char string) string {
	switch char {
	case "a":
		return "NO ACTION"
	case "r":
		return "RESTRICT"
	case "c":
		return "CASCADE"
	case "n":
		return "SET NULL"
	case "d":
		return "SET DEFAULT"
	default:
		return ""
	}
}
