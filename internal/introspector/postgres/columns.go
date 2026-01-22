package postgres

import (
	"context"
	"fmt"

	"github.com/nenorrell/X-Rai/internal/schema"
)

func (i *Introspector) introspectColumns(ctx context.Context, schemaName, tableName string) ([]*schema.Column, error) {
	query := `
		SELECT
			column_name,
			data_type,
			udt_name,
			is_nullable,
			column_default,
			character_maximum_length,
			numeric_precision,
			numeric_scale,
			COALESCE(is_identity, 'NO') as is_identity,
			identity_generation,
			COALESCE(is_generated, 'NEVER') as is_generated,
			generation_expression,
			collation_name,
			ordinal_position
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position
	`

	rows, err := i.pool.Query(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []*schema.Column
	for rows.Next() {
		var (
			columnName, dataType, udtName, isNullable string
			isIdentity, isGenerated                   string
			ordinalPosition                           int
			columnDefault, identityGeneration         *string
			generationExpression, collation           *string
			charMaxLength, numPrecision, numScale     *int
		)

		err := rows.Scan(
			&columnName, &dataType, &udtName, &isNullable,
			&columnDefault, &charMaxLength, &numPrecision, &numScale,
			&isIdentity, &identityGeneration, &isGenerated, &generationExpression,
			&collation, &ordinalPosition,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column row: %w", err)
		}

		col := &schema.Column{
			ColumnName:           columnName,
			DataType:             dataType,
			UDTName:              udtName,
			Nullable:             isNullable == "YES",
			DefaultValue:         columnDefault,
			Generated:            isGenerated != "NEVER",
			GenerationExpression: generationExpression,
			IsIdentity:           isIdentity == "YES",
			Collation:            collation,
			CharacterMaxLength:   charMaxLength,
			NumericPrecision:     numPrecision,
			NumericScale:         numScale,
			OrdinalPosition:      ordinalPosition,
		}

		if identityGeneration != nil {
			col.IdentityGeneration = *identityGeneration
		}

		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating column rows: %w", err)
	}

	return columns, nil
}

func (i *Introspector) introspectColumnComments(ctx context.Context, table *schema.Table) error {
	query := `
		SELECT
			a.attname,
			COALESCE(col_description(c.oid, a.attnum), '')
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		JOIN pg_attribute a ON a.attrelid = c.oid
		WHERE n.nspname = $1
		  AND c.relname = $2
		  AND a.attnum > 0
		  AND NOT a.attisdropped
		ORDER BY a.attnum
	`

	rows, err := i.pool.Query(ctx, query, table.SchemaName, table.TableName)
	if err != nil {
		return fmt.Errorf("failed to query column comments: %w", err)
	}
	defer rows.Close()

	commentMap := make(map[string]string)
	for rows.Next() {
		var colName, comment string
		if err := rows.Scan(&colName, &comment); err != nil {
			return fmt.Errorf("failed to scan column comment row: %w", err)
		}
		if comment != "" {
			commentMap[colName] = comment
		}
	}

	// Assign comments to columns
	for _, col := range table.Columns {
		if comment, ok := commentMap[col.ColumnName]; ok {
			col.Comment = comment
		}
	}

	return rows.Err()
}
