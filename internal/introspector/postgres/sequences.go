package postgres

import (
	"context"
	"fmt"

	"github.com/nenorrell/X-Rai/internal/schema"
)

func (i *Introspector) introspectSequences(ctx context.Context, schemas []string) ([]*schema.Sequence, error) {
	query := `
		SELECT
			n.nspname as schema_name,
			c.relname as sequence_name,
			s.seqtypid::regtype::text as data_type,
			s.seqstart as start_value,
			s.seqmin as min_value,
			s.seqmax as max_value,
			s.seqincrement as increment,
			s.seqcycle as cycle,
			COALESCE(
				(
					SELECT a.attrelid::regclass::text || '.' || a.attname
					FROM pg_depend d
					JOIN pg_attribute a ON a.attrelid = d.refobjid AND a.attnum = d.refobjsubid
					WHERE d.objid = c.oid
					  AND d.classid = 'pg_class'::regclass
					  AND d.refclassid = 'pg_class'::regclass
					  AND d.deptype = 'a'
					LIMIT 1
				),
				''
			) as owned_by,
			COALESCE(obj_description(c.oid, 'pg_class'), '') as comment
		FROM pg_sequence s
		JOIN pg_class c ON c.oid = s.seqrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = ANY($1)
		ORDER BY n.nspname, c.relname
	`

	rows, err := i.pool.Query(ctx, query, schemas)
	if err != nil {
		return nil, fmt.Errorf("failed to query sequences: %w", err)
	}
	defer rows.Close()

	var sequences []*schema.Sequence
	for rows.Next() {
		var (
			schemaName, seqName, dataType, ownedBy, comment string
			startValue, minValue, maxValue, increment       int64
			cycle                                           bool
		)

		err := rows.Scan(
			&schemaName, &seqName, &dataType,
			&startValue, &minValue, &maxValue, &increment,
			&cycle, &ownedBy, &comment,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sequence row: %w", err)
		}

		seq := &schema.Sequence{
			SequenceName: seqName,
			SchemaName:   schemaName,
			DataType:     dataType,
			StartValue:   &startValue,
			MinValue:     &minValue,
			MaxValue:     &maxValue,
			Increment:    &increment,
			CycleOption:  cycle,
			OwnedBy:      ownedBy,
			Comment:      comment,
		}

		sequences = append(sequences, seq)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sequence rows: %w", err)
	}

	return sequences, nil
}
