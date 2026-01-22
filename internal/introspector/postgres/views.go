package postgres

import (
	"context"
	"fmt"

	"github.com/nenorrell/X-Rai/internal/schema"
)

func (i *Introspector) introspectViews(ctx context.Context, schemas []string, redactDef, redactComments bool) ([]*schema.View, error) {
	query := `
		SELECT
			v.table_schema,
			v.table_name,
			v.view_definition,
			COALESCE(obj_description((v.table_schema || '.' || v.table_name)::regclass), '') as view_comment
		FROM information_schema.views v
		WHERE v.table_schema = ANY($1)
		ORDER BY v.table_schema, v.table_name
	`

	rows, err := i.pool.Query(ctx, query, schemas)
	if err != nil {
		return nil, fmt.Errorf("failed to query views: %w", err)
	}
	defer rows.Close()

	var views []*schema.View
	for rows.Next() {
		var schemaName, viewName, comment string
		var definition *string

		if err := rows.Scan(&schemaName, &viewName, &definition, &comment); err != nil {
			return nil, fmt.Errorf("failed to scan view row: %w", err)
		}

		view := &schema.View{
			ViewName:   viewName,
			SchemaName: schemaName,
		}

		if !redactDef && definition != nil {
			view.Definition = definition
		}

		if !redactComments {
			view.Comment = comment
		}

		views = append(views, view)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating view rows: %w", err)
	}

	// Introspect columns and dependencies for each view
	for _, view := range views {
		columns, err := i.introspectViewColumns(ctx, view.SchemaName, view.ViewName)
		if err != nil {
			// Non-fatal, continue
			continue
		}
		view.Columns = columns

		deps, err := i.introspectViewDependencies(ctx, view.SchemaName, view.ViewName)
		if err != nil {
			// Non-fatal, continue
			continue
		}
		view.DependsOnTables = deps.Tables
		view.DependsOnViews = deps.Views
	}

	return views, nil
}

func (i *Introspector) introspectViewColumns(ctx context.Context, schemaName, viewName string) ([]*schema.Column, error) {
	// Reuse column introspection logic
	return i.introspectColumns(ctx, schemaName, viewName)
}

type viewDeps struct {
	Tables []string
	Views  []string
}

func (i *Introspector) introspectViewDependencies(ctx context.Context, schemaName, viewName string) (*viewDeps, error) {
	query := `
		SELECT DISTINCT
			d.refclassid::regclass::text,
			dn.nspname as dep_schema,
			dc.relname as dep_name,
			dc.relkind
		FROM pg_depend d
		JOIN pg_rewrite r ON r.oid = d.objid
		JOIN pg_class c ON c.oid = r.ev_class
		JOIN pg_namespace n ON n.oid = c.relnamespace
		JOIN pg_class dc ON dc.oid = d.refobjid
		JOIN pg_namespace dn ON dn.oid = dc.relnamespace
		WHERE n.nspname = $1
		  AND c.relname = $2
		  AND d.classid = 'pg_rewrite'::regclass
		  AND d.refclassid = 'pg_class'::regclass
		  AND dc.relkind IN ('r', 'v')
		  AND NOT (dn.nspname = $1 AND dc.relname = $2)
	`

	rows, err := i.pool.Query(ctx, query, schemaName, viewName)
	if err != nil {
		return nil, fmt.Errorf("failed to query view dependencies: %w", err)
	}
	defer rows.Close()

	deps := &viewDeps{}
	for rows.Next() {
		var refClass, depSchema, depName, relKind string
		if err := rows.Scan(&refClass, &depSchema, &depName, &relKind); err != nil {
			return nil, fmt.Errorf("failed to scan view dependency row: %w", err)
		}

		fullName := depName
		if depSchema != "public" {
			fullName = depSchema + "." + depName
		}

		if relKind == "r" {
			deps.Tables = append(deps.Tables, fullName)
		} else if relKind == "v" {
			deps.Views = append(deps.Views, fullName)
		}
	}

	return deps, rows.Err()
}
