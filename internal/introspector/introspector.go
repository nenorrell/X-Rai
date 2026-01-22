package introspector

import (
	"context"

	"github.com/nenorrell/xrai/internal/config"
	"github.com/nenorrell/xrai/internal/schema"
)

// Introspector defines the interface for database introspection.
type Introspector interface {
	// Introspect performs full database introspection.
	Introspect(ctx context.Context, cfg *config.Config) (*schema.Database, error)

	// DatabaseName returns the name of the connected database.
	DatabaseName() string

	// Version returns the database version.
	Version() string

	// Close closes the database connection.
	Close() error
}
