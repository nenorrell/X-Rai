package generator

import (
	"strings"
	"testing"

	"github.com/nenorrell/X-Rai/internal/schema"
)

func TestFormatSchemaList(t *testing.T) {
	tests := []struct {
		name     string
		schemas  []string
		expected string
	}{
		{"single schema", []string{"public"}, "`public` schema"},
		{"two schemas", []string{"public", "app"}, "2 schemas (public, app)"},
		{"three schemas", []string{"a", "b", "c"}, "3 schemas (a, b, c)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSchemaList(tt.schemas)
			if result != tt.expected {
				t.Errorf("formatSchemaList() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestHasTag(t *testing.T) {
	table := &schema.Table{
		TableName: "users",
		Tags:      []string{"core", "audit"},
	}

	tests := []struct {
		tag      string
		expected bool
	}{
		{"core", true},
		{"audit", true},
		{"junction", false},
		{"lookup", false},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			result := hasTag(table, tt.tag)
			if result != tt.expected {
				t.Errorf("hasTag(%q) = %v, want %v", tt.tag, result, tt.expected)
			}
		})
	}
}

func TestTableNames(t *testing.T) {
	tables := []*schema.Table{
		{TableName: "users"},
		{TableName: "orders"},
		{TableName: "products"},
	}

	result := tableNames(tables)
	expected := "`users`, `orders`, `products`"

	if result != expected {
		t.Errorf("tableNames() = %q, want %q", result, expected)
	}
}

func TestHasRelationships(t *testing.T) {
	tests := []struct {
		name     string
		tables   []*schema.Table
		expected bool
	}{
		{
			name: "with incoming FK",
			tables: []*schema.Table{
				{
					TableName:           "users",
					IncomingForeignKeys: []*schema.ForeignKey{{FromTable: "orders"}},
				},
			},
			expected: true,
		},
		{
			name: "with outgoing FK",
			tables: []*schema.Table{
				{
					TableName:           "orders",
					OutgoingForeignKeys: []*schema.ForeignKey{{ToTable: "users"}},
				},
			},
			expected: true,
		},
		{
			name: "no relationships",
			tables: []*schema.Table{
				{TableName: "settings"},
			},
			expected: false,
		},
		{
			name:     "empty tables",
			tables:   []*schema.Table{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasRelationships(tt.tables)
			if result != tt.expected {
				t.Errorf("hasRelationships() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsIDColumn(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"id", true},
		{"user_id", true},
		{"session_uuid", true},
		{"uuid", true},
		{"name", false},
		{"identifier", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isIDColumn(tt.name)
			if result != tt.expected {
				t.Errorf("isIDColumn(%q) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestIsTimestampColumn(t *testing.T) {
	tests := []struct {
		name     string
		dtype    string
		expected bool
	}{
		{"created_at", "timestamp with time zone", true},
		{"updated_at", "text", true},
		{"birth_date", "text", true},
		{"start_time", "text", true},
		{"created", "text", true},
		{"updated", "text", true},
		{"name", "text", false},
		{"any_col", "timestamp", true},
		{"any_col", "date", true},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_"+tt.dtype, func(t *testing.T) {
			result := isTimestampColumn(tt.name, tt.dtype)
			if result != tt.expected {
				t.Errorf("isTimestampColumn(%q, %q) = %v, want %v", tt.name, tt.dtype, result, tt.expected)
			}
		})
	}
}

func TestIsStatusColumn(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"status", true},
		{"order_status", true},
		{"state", true},
		{"payment_state", true},
		{"name", false},
		{"status_count", false}, // contains but doesn't end with
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isStatusColumn(tt.name)
			if result != tt.expected {
				t.Errorf("isStatusColumn(%q) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestIsCurrencyColumn(t *testing.T) {
	tests := []struct {
		name     string
		dtype    string
		expected bool
	}{
		{"amount", "numeric", true},
		{"total_price", "decimal(10,2)", true},
		{"order_cost", "money", true},
		{"balance", "numeric", true},
		{"total", "numeric", true},
		{"amount", "text", false},
		{"name", "numeric", false},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_"+tt.dtype, func(t *testing.T) {
			result := isCurrencyColumn(tt.name, tt.dtype)
			if result != tt.expected {
				t.Errorf("isCurrencyColumn(%q, %q) = %v, want %v", tt.name, tt.dtype, result, tt.expected)
			}
		})
	}
}

func TestInferColumnSemantics(t *testing.T) {
	columns := []*schema.Column{
		{ColumnName: "id", DataType: "integer"},
		{ColumnName: "user_id", DataType: "integer"},
		{ColumnName: "created_at", DataType: "timestamp"},
		{ColumnName: "status", DataType: "text"},
		{ColumnName: "amount", DataType: "numeric"},
		{ColumnName: "name", DataType: "text"},
	}

	result := inferColumnSemantics(columns)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Check ID columns
	if len(result.IDColumns) != 2 {
		t.Errorf("expected 2 ID columns, got %d: %v", len(result.IDColumns), result.IDColumns)
	}

	// Check timestamp columns
	if len(result.TimestampColumns) != 1 {
		t.Errorf("expected 1 timestamp column, got %d: %v", len(result.TimestampColumns), result.TimestampColumns)
	}

	// Check status columns
	if len(result.StatusColumns) != 1 {
		t.Errorf("expected 1 status column, got %d: %v", len(result.StatusColumns), result.StatusColumns)
	}

	// Check currency columns
	if len(result.CurrencyAmountColumns) != 1 {
		t.Errorf("expected 1 currency column, got %d: %v", len(result.CurrencyAmountColumns), result.CurrencyAmountColumns)
	}
}

func TestInferColumnSemantics_Empty(t *testing.T) {
	columns := []*schema.Column{
		{ColumnName: "foo", DataType: "text"},
		{ColumnName: "bar", DataType: "text"},
	}

	result := inferColumnSemantics(columns)

	if result != nil {
		t.Errorf("expected nil for columns with no semantics, got %+v", result)
	}
}

func TestWriteTableIndex(t *testing.T) {
	tables := []*schema.Table{
		{TableName: "users", Tags: []string{"core"}},
		{TableName: "user_roles", Tags: []string{"junction"}},
		{TableName: "countries", Tags: []string{"lookup"}},
		{TableName: "orders", Tags: []string{}},
	}

	var sb strings.Builder
	writeTableIndex(&sb, tables)
	result := sb.String()

	// Check each category is present
	if !strings.Contains(result, "**Core**") {
		t.Error("expected Core section")
	}
	if !strings.Contains(result, "`users`") {
		t.Error("expected users table in output")
	}
	if !strings.Contains(result, "**Junction**") {
		t.Error("expected Junction section")
	}
	if !strings.Contains(result, "**Lookup**") {
		t.Error("expected Lookup section")
	}
	if !strings.Contains(result, "**Tables**") {
		t.Error("expected Tables section for uncategorized")
	}
}
