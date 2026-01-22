package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/nenorrell/xrai/internal/schema"
)

func (i *Introspector) introspectRoutines(ctx context.Context, schemas []string, redactDef bool) ([]*schema.Routine, error) {
	query := `
		SELECT
			n.nspname as schema_name,
			p.proname as routine_name,
			CASE p.prokind
				WHEN 'f' THEN 'function'
				WHEN 'p' THEN 'procedure'
				WHEN 'a' THEN 'aggregate'
				WHEN 'w' THEN 'window'
				ELSE 'function'
			END as routine_type,
			l.lanname as language,
			pg_get_function_result(p.oid) as return_type,
			pg_get_functiondef(p.oid) as definition,
			COALESCE(obj_description(p.oid, 'pg_proc'), '') as comment,
			p.proargnames as arg_names,
			p.proargtypes::regtype[]::text[] as arg_types,
			p.proargmodes::text[] as arg_modes,
			p.pronargdefaults as num_defaults
		FROM pg_proc p
		JOIN pg_namespace n ON n.oid = p.pronamespace
		JOIN pg_language l ON l.oid = p.prolang
		WHERE n.nspname = ANY($1)
		  AND p.prokind IN ('f', 'p')
		ORDER BY n.nspname, p.proname
	`

	rows, err := i.pool.Query(ctx, query, schemas)
	if err != nil {
		return nil, fmt.Errorf("failed to query routines: %w", err)
	}
	defer rows.Close()

	var routines []*schema.Routine
	for rows.Next() {
		var (
			schemaName, routineName, routineType string
			language, returnType, comment        string
			definition                           *string
			argNames                             []string
			argTypes                             []string
			argModes                             []string
			numDefaults                          int
		)

		err := rows.Scan(
			&schemaName, &routineName, &routineType, &language,
			&returnType, &definition, &comment,
			&argNames, &argTypes, &argModes, &numDefaults,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan routine row: %w", err)
		}

		routine := &schema.Routine{
			RoutineName: routineName,
			SchemaName:  schemaName,
			RoutineType: routineType,
			Language:    language,
			ReturnType:  returnType,
			Comment:     comment,
		}

		if !redactDef && definition != nil {
			routine.Definition = definition
		}

		// Build arguments list
		routine.Arguments = buildArguments(argNames, argTypes, argModes)

		routines = append(routines, routine)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating routine rows: %w", err)
	}

	return routines, nil
}

func buildArguments(names, types, modes []string) []schema.Argument {
	if len(types) == 0 {
		return nil
	}

	args := make([]schema.Argument, len(types))
	for i, t := range types {
		args[i] = schema.Argument{
			DataType: t,
		}

		if i < len(names) && names[i] != "" {
			args[i].Name = names[i]
		}

		if i < len(modes) {
			args[i].Mode = mapArgMode(modes[i])
		} else {
			args[i].Mode = "IN"
		}
	}

	return args
}

func mapArgMode(mode string) string {
	switch strings.ToLower(mode) {
	case "i":
		return "IN"
	case "o":
		return "OUT"
	case "b":
		return "INOUT"
	case "v":
		return "VARIADIC"
	case "t":
		return "TABLE"
	default:
		return "IN"
	}
}
