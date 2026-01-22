package config

// Config holds all configuration options for xrai generation.
type Config struct {
	// Connection
	DSN string

	// Output
	OutputDir string

	// Scope
	Schemas []string

	// Feature flags
	IncludeViews    bool
	IncludeRoutines bool
	IncludeStats    bool

	// Redaction
	RedactComments    bool
	RedactDefinitions bool
}

// NewConfig creates a Config with default values.
func NewConfig() *Config {
	return &Config{
		Schemas:           []string{"public"},
		IncludeViews:      false,
		IncludeRoutines:   false,
		IncludeStats:      false,
		RedactComments:    false,
		RedactDefinitions: false,
	}
}
