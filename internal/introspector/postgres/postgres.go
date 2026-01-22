package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nenorrell/X-Rai/internal/config"
	"github.com/nenorrell/X-Rai/internal/schema"
)

// Introspector implements database introspection for PostgreSQL.
type Introspector struct {
	pool         *pgxpool.Pool
	databaseName string
	version      string
}

// New creates a new PostgreSQL introspector.
func New(ctx context.Context, dsn string) (*Introspector, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	i := &Introspector{pool: pool}

	if err := i.fetchMetadata(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return i, nil
}

func (i *Introspector) fetchMetadata(ctx context.Context) error {
	row := i.pool.QueryRow(ctx, "SELECT current_database(), version()")
	var fullVersion string
	if err := row.Scan(&i.databaseName, &fullVersion); err != nil {
		return fmt.Errorf("failed to fetch database metadata: %w", err)
	}

	// Extract version number from full version string
	row = i.pool.QueryRow(ctx, "SHOW server_version")
	if err := row.Scan(&i.version); err != nil {
		// Fallback to empty if we can't get it
		i.version = ""
	}

	return nil
}

// DatabaseName returns the name of the connected database.
func (i *Introspector) DatabaseName() string {
	return i.databaseName
}

// Version returns the PostgreSQL version.
func (i *Introspector) Version() string {
	return i.version
}

// Close closes the database connection pool.
func (i *Introspector) Close() error {
	i.pool.Close()
	return nil
}

// Introspect performs full database introspection.
func (i *Introspector) Introspect(ctx context.Context, cfg *config.Config) (*schema.Database, error) {
	db := &schema.Database{
		Name:    i.databaseName,
		Engine:  "postgresql",
		Version: i.version,
		Schemas: cfg.Schemas,
	}

	// Introspect tables
	tables, err := i.introspectTables(ctx, cfg.Schemas)
	if err != nil {
		return nil, fmt.Errorf("failed to introspect tables: %w", err)
	}
	db.Tables = tables

	// Introspect columns for each table
	for _, table := range db.Tables {
		columns, err := i.introspectColumns(ctx, table.SchemaName, table.TableName)
		if err != nil {
			return nil, fmt.Errorf("failed to introspect columns for %s.%s: %w", table.SchemaName, table.TableName, err)
		}
		table.Columns = columns

		// Introspect indexes
		indexes, err := i.introspectIndexes(ctx, table.SchemaName, table.TableName)
		if err != nil {
			return nil, fmt.Errorf("failed to introspect indexes for %s.%s: %w", table.SchemaName, table.TableName, err)
		}
		table.Indexes = indexes

		// Introspect constraints
		constraints, err := i.introspectConstraints(ctx, table.SchemaName, table.TableName)
		if err != nil {
			return nil, fmt.Errorf("failed to introspect constraints for %s.%s: %w", table.SchemaName, table.TableName, err)
		}
		table.Constraints = constraints

		// Introspect foreign keys (outgoing)
		fks, err := i.introspectForeignKeys(ctx, table.SchemaName, table.TableName)
		if err != nil {
			return nil, fmt.Errorf("failed to introspect foreign keys for %s.%s: %w", table.SchemaName, table.TableName, err)
		}
		table.OutgoingForeignKeys = fks

		// Introspect triggers
		triggers, err := i.introspectTriggers(ctx, table.SchemaName, table.TableName)
		if err != nil {
			return nil, fmt.Errorf("failed to introspect triggers for %s.%s: %w", table.SchemaName, table.TableName, err)
		}
		table.Triggers = triggers

		// Introspect row count estimate
		rowCount, err := i.introspectRowCount(ctx, table.SchemaName, table.TableName)
		if err == nil && rowCount != nil {
			table.RowCountEstimate = rowCount
		}

		// Introspect column comments
		if !cfg.RedactComments {
			if err := i.introspectColumnComments(ctx, table); err != nil {
				// Non-fatal, just log
			}
		}

		// Introspect stats if enabled
		if cfg.IncludeStats {
			stats, err := i.introspectStats(ctx, table.SchemaName, table.TableName)
			if err == nil {
				table.Stats = stats
			}
		}
	}

	// Build incoming foreign key references
	i.buildIncomingForeignKeys(db.Tables)

	// Introspect views if enabled
	if cfg.IncludeViews {
		views, err := i.introspectViews(ctx, cfg.Schemas, cfg.RedactDefinitions, cfg.RedactComments)
		if err != nil {
			return nil, fmt.Errorf("failed to introspect views: %w", err)
		}
		db.Views = views
	}

	// Introspect routines if enabled
	if cfg.IncludeRoutines {
		routines, err := i.introspectRoutines(ctx, cfg.Schemas, cfg.RedactDefinitions)
		if err != nil {
			return nil, fmt.Errorf("failed to introspect routines: %w", err)
		}
		db.Routines = routines
	}

	// Introspect enums
	enums, err := i.introspectEnums(ctx, cfg.Schemas)
	if err != nil {
		return nil, fmt.Errorf("failed to introspect enums: %w", err)
	}
	db.Enums = enums

	// Introspect sequences
	sequences, err := i.introspectSequences(ctx, cfg.Schemas)
	if err != nil {
		return nil, fmt.Errorf("failed to introspect sequences: %w", err)
	}
	db.Sequences = sequences

	// Introspect custom types
	types, err := i.introspectTypes(ctx, cfg.Schemas)
	if err != nil {
		return nil, fmt.Errorf("failed to introspect types: %w", err)
	}
	db.Types = types

	return db, nil
}

// buildIncomingForeignKeys populates IncomingForeignKeys for each table
func (i *Introspector) buildIncomingForeignKeys(tables []*schema.Table) {
	tableMap := make(map[string]*schema.Table)
	for _, t := range tables {
		key := t.SchemaName + "." + t.TableName
		tableMap[key] = t
	}

	for _, t := range tables {
		for _, fk := range t.OutgoingForeignKeys {
			targetKey := fk.ToSchema + "." + fk.ToTable
			if targetTable, ok := tableMap[targetKey]; ok {
				incoming := &schema.ForeignKey{
					ConstraintName: fk.ConstraintName,
					FromSchema:     t.SchemaName,
					FromTable:      t.TableName,
					FromColumns:    fk.FromColumns,
					ToSchema:       fk.ToSchema,
					ToTable:        fk.ToTable,
					ToColumns:      fk.ToColumns,
					OnUpdate:       fk.OnUpdate,
					OnDelete:       fk.OnDelete,
					Nullable:       fk.Nullable,
					Cardinality:    fk.Cardinality,
				}
				targetTable.IncomingForeignKeys = append(targetTable.IncomingForeignKeys, incoming)
			}
		}
	}
}

// Helper to scan nullable strings
func scanNullString(rows pgx.Rows, dest *string) error {
	var ns *string
	if err := rows.Scan(&ns); err != nil {
		return err
	}
	if ns != nil {
		*dest = *ns
	}
	return nil
}
