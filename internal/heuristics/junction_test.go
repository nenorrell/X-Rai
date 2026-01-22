package heuristics

import (
	"testing"

	"github.com/nenorrell/X-Rai/internal/schema"
)

func TestDetectJunctionTable_TwoFKsCompositePK(t *testing.T) {
	// Classic M:N junction: PK is (user_id, role_id) which are both FKs
	table := &schema.Table{
		TableName: "user_roles",
		Columns: []*schema.Column{
			{ColumnName: "user_id"},
			{ColumnName: "role_id"},
		},
		OutgoingForeignKeys: []*schema.ForeignKey{
			{FromColumns: []string{"user_id"}, ToTable: "users"},
			{FromColumns: []string{"role_id"}, ToTable: "roles"},
		},
		Constraints: []*schema.Constraint{
			{ConstraintType: "PRIMARY KEY", Columns: []string{"user_id", "role_id"}},
		},
	}

	detectJunctionTable(table)

	if !table.IsJunction {
		t.Errorf("expected table to be detected as junction")
	}
	if table.JunctionReasoning == "" {
		t.Errorf("expected junction reasoning to be set")
	}
}

func TestDetectJunctionTable_TwoFKsFewColumns(t *testing.T) {
	// Junction with separate PK but minimal columns
	table := &schema.Table{
		TableName: "project_members",
		Columns: []*schema.Column{
			{ColumnName: "id"},
			{ColumnName: "project_id"},
			{ColumnName: "member_id"},
			{ColumnName: "created_at"},
		},
		OutgoingForeignKeys: []*schema.ForeignKey{
			{FromColumns: []string{"project_id"}, ToTable: "projects"},
			{FromColumns: []string{"member_id"}, ToTable: "users"},
		},
		Constraints: []*schema.Constraint{
			{ConstraintType: "PRIMARY KEY", Columns: []string{"id"}},
		},
	}

	detectJunctionTable(table)

	if !table.IsJunction {
		t.Errorf("expected table to be detected as junction (few columns heuristic)")
	}
}

func TestDetectJunctionTable_NamePattern(t *testing.T) {
	patterns := []string{"user_to_group", "post_has_tag", "item_x_category", "order_link_product"}

	for _, name := range patterns {
		t.Run(name, func(t *testing.T) {
			table := &schema.Table{
				TableName: name,
				Columns: []*schema.Column{
					{ColumnName: "id"},
					{ColumnName: "left_id"},
					{ColumnName: "right_id"},
					{ColumnName: "extra_data"},
					{ColumnName: "more_data"},
				},
				OutgoingForeignKeys: []*schema.ForeignKey{
					{FromColumns: []string{"left_id"}, ToTable: "left_table"},
					{FromColumns: []string{"right_id"}, ToTable: "right_table"},
				},
			}

			detectJunctionTable(table)

			if !table.IsJunction {
				t.Errorf("expected table %q to be detected as junction by name pattern", name)
			}
		})
	}
}

func TestDetectJunctionTable_NotJunction(t *testing.T) {
	tests := []struct {
		name  string
		table *schema.Table
	}{
		{
			name: "single FK",
			table: &schema.Table{
				TableName: "orders",
				Columns: []*schema.Column{
					{ColumnName: "id"},
					{ColumnName: "user_id"},
					{ColumnName: "total"},
				},
				OutgoingForeignKeys: []*schema.ForeignKey{
					{FromColumns: []string{"user_id"}, ToTable: "users"},
				},
			},
		},
		{
			name: "no FKs",
			table: &schema.Table{
				TableName: "settings",
				Columns: []*schema.Column{
					{ColumnName: "key"},
					{ColumnName: "value"},
				},
			},
		},
		{
			name: "many non-FK columns",
			table: &schema.Table{
				TableName: "transactions",
				Columns: []*schema.Column{
					{ColumnName: "id"},
					{ColumnName: "from_account_id"},
					{ColumnName: "to_account_id"},
					{ColumnName: "amount"},
					{ColumnName: "currency"},
					{ColumnName: "description"},
					{ColumnName: "status"},
					{ColumnName: "processed_at"},
					{ColumnName: "reference_number"},
					{ColumnName: "notes"},
				},
				OutgoingForeignKeys: []*schema.ForeignKey{
					{FromColumns: []string{"from_account_id"}, ToTable: "accounts"},
					{FromColumns: []string{"to_account_id"}, ToTable: "accounts"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detectJunctionTable(tt.table)

			if tt.table.IsJunction {
				t.Errorf("expected table %q to NOT be detected as junction", tt.table.TableName)
			}
		})
	}
}
