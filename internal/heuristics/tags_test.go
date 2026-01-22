package heuristics

import (
	"testing"

	"github.com/nenorrell/X-Rai/internal/schema"
)

func TestIsLookupTable_NamePatterns(t *testing.T) {
	patterns := []string{
		"order_status",
		"user_types",
		"product_category",
		"payment_kinds",
		"ref_countries",
		"lookup_currencies",
	}

	for _, name := range patterns {
		t.Run(name, func(t *testing.T) {
			table := &schema.Table{TableName: name}
			if !isLookupTable(table) {
				t.Errorf("expected %q to be detected as lookup table", name)
			}
		})
	}
}

func TestIsLookupTable_StructureHeuristic(t *testing.T) {
	// Few columns, no outgoing FKs, has incoming FKs
	table := &schema.Table{
		TableName: "currencies",
		Columns: []*schema.Column{
			{ColumnName: "id"},
			{ColumnName: "code"},
			{ColumnName: "name"},
		},
		IncomingForeignKeys: []*schema.ForeignKey{
			{FromTable: "orders", FromColumns: []string{"currency_id"}},
		},
	}

	if !isLookupTable(table) {
		t.Errorf("expected table with few columns and incoming FKs to be lookup")
	}
}

func TestIsLookupTable_NotLookup(t *testing.T) {
	table := &schema.Table{
		TableName: "orders",
		Columns: []*schema.Column{
			{ColumnName: "id"},
			{ColumnName: "user_id"},
			{ColumnName: "total"},
			{ColumnName: "status"},
			{ColumnName: "created_at"},
			{ColumnName: "updated_at"},
		},
		OutgoingForeignKeys: []*schema.ForeignKey{
			{FromColumns: []string{"user_id"}, ToTable: "users"},
		},
	}

	if isLookupTable(table) {
		t.Errorf("expected orders table to NOT be detected as lookup")
	}
}

func TestIsAuditTable(t *testing.T) {
	tests := []struct {
		name     string
		table    *schema.Table
		expected bool
	}{
		{
			name:     "name pattern _audit",
			table:    &schema.Table{TableName: "user_audit"},
			expected: true,
		},
		{
			name:     "name pattern _history",
			table:    &schema.Table{TableName: "order_history"},
			expected: true,
		},
		{
			name: "audit columns",
			table: &schema.Table{
				TableName: "changes",
				Columns: []*schema.Column{
					{ColumnName: "id"},
					{ColumnName: "action"},
					{ColumnName: "old_value"},
					{ColumnName: "new_value"},
				},
			},
			expected: true,
		},
		{
			name:     "regular table",
			table:    &schema.Table{TableName: "users"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAuditTable(tt.table)
			if result != tt.expected {
				t.Errorf("isAuditTable(%q) = %v, want %v", tt.table.TableName, result, tt.expected)
			}
		})
	}
}

func TestIsConfigTable(t *testing.T) {
	tests := []struct {
		name     string
		table    *schema.Table
		expected bool
	}{
		{
			name:     "name contains config",
			table:    &schema.Table{TableName: "app_config"},
			expected: true,
		},
		{
			name:     "name contains setting",
			table:    &schema.Table{TableName: "user_settings"},
			expected: true,
		},
		{
			name: "key-value structure",
			table: &schema.Table{
				TableName: "properties",
				Columns: []*schema.Column{
					{ColumnName: "id"},
					{ColumnName: "key"},
					{ColumnName: "value"},
				},
			},
			expected: true,
		},
		{
			name:     "regular table",
			table:    &schema.Table{TableName: "users"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isConfigTable(tt.table)
			if result != tt.expected {
				t.Errorf("isConfigTable(%q) = %v, want %v", tt.table.TableName, result, tt.expected)
			}
		})
	}
}

func TestIsLogTable(t *testing.T) {
	tests := []struct {
		name     string
		table    *schema.Table
		expected bool
	}{
		{
			name:     "name pattern _log",
			table:    &schema.Table{TableName: "access_log"},
			expected: true,
		},
		{
			name:     "name pattern _events",
			table:    &schema.Table{TableName: "user_events"},
			expected: true,
		},
		{
			name: "log-like columns",
			table: &schema.Table{
				TableName: "activity",
				Columns: []*schema.Column{
					{ColumnName: "id"},
					{ColumnName: "event_type"},
					{ColumnName: "timestamp"},
					{ColumnName: "payload"},
				},
			},
			expected: true,
		},
		{
			name:     "regular table",
			table:    &schema.Table{TableName: "orders"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isLogTable(tt.table)
			if result != tt.expected {
				t.Errorf("isLogTable(%q) = %v, want %v", tt.table.TableName, result, tt.expected)
			}
		})
	}
}

func TestApplyTableTags_Core(t *testing.T) {
	// Core tables have 3+ incoming FKs
	table := &schema.Table{
		TableName: "users",
		IncomingForeignKeys: []*schema.ForeignKey{
			{FromTable: "orders"},
			{FromTable: "posts"},
			{FromTable: "comments"},
		},
	}

	applyTableTags(table)

	found := false
	for _, tag := range table.Tags {
		if tag == "core" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected 'core' tag, got %v", table.Tags)
	}
}
