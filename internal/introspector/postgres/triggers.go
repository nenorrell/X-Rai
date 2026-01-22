package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/nenorrell/X-Rai/internal/schema"
)

func (i *Introspector) introspectTriggers(ctx context.Context, schemaName, tableName string) ([]*schema.Trigger, error) {
	query := `
		SELECT
			t.tgname as trigger_name,
			CASE
				WHEN t.tgtype & 2 = 2 THEN 'BEFORE'
				WHEN t.tgtype & 64 = 64 THEN 'INSTEAD OF'
				ELSE 'AFTER'
			END as timing,
			ARRAY_REMOVE(ARRAY[
				CASE WHEN t.tgtype & 4 = 4 THEN 'INSERT' END,
				CASE WHEN t.tgtype & 8 = 8 THEN 'DELETE' END,
				CASE WHEN t.tgtype & 16 = 16 THEN 'UPDATE' END,
				CASE WHEN t.tgtype & 32 = 32 THEN 'TRUNCATE' END
			], NULL) as events,
			p.proname as function_name,
			np.nspname as function_schema,
			pg_get_triggerdef(t.oid) as definition
		FROM pg_trigger t
		JOIN pg_class c ON c.oid = t.tgrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		JOIN pg_proc p ON p.oid = t.tgfoid
		JOIN pg_namespace np ON np.oid = p.pronamespace
		WHERE n.nspname = $1 AND c.relname = $2
		  AND NOT t.tgisinternal
		ORDER BY t.tgname
	`

	rows, err := i.pool.Query(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query triggers: %w", err)
	}
	defer rows.Close()

	var triggers []*schema.Trigger
	for rows.Next() {
		var (
			triggerName, timing, funcName, funcSchema, definition string
			events                                                []string
		)

		err := rows.Scan(
			&triggerName, &timing, &events, &funcName, &funcSchema, &definition,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trigger row: %w", err)
		}

		funcRef := funcName
		if funcSchema != "" && funcSchema != "public" {
			funcRef = funcSchema + "." + funcName
		}

		trigger := &schema.Trigger{
			TriggerName:         triggerName,
			Timing:              strings.ToLower(timing),
			Events:              lowercaseStrings(events),
			FunctionOrProcedure: funcRef,
			Definition:          &definition,
		}

		triggers = append(triggers, trigger)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trigger rows: %w", err)
	}

	return triggers, nil
}

func lowercaseStrings(ss []string) []string {
	result := make([]string, len(ss))
	for i, s := range ss {
		result[i] = strings.ToLower(s)
	}
	return result
}
