package heuristics

import (
	"strings"

	"github.com/nenorrell/xrai/internal/schema"
)

// applyTableTags assigns heuristic tags to a table based on its structure.
func applyTableTags(table *schema.Table) {
	var tags []string

	// Core: Tables with many incoming FKs (referenced by others)
	if len(table.IncomingForeignKeys) >= 3 {
		tags = append(tags, "core")
	}

	// Junction: Already detected
	if table.IsJunction {
		tags = append(tags, "junction")
	}

	// Lookup: Small reference tables
	if isLookupTable(table) {
		tags = append(tags, "lookup")
	}

	// Audit: Tables that track changes/history
	if isAuditTable(table) {
		tags = append(tags, "audit")
	}

	// Config: Configuration/settings tables
	if isConfigTable(table) {
		tags = append(tags, "config")
	}

	// Log: Event/log tables
	if isLogTable(table) {
		tags = append(tags, "log")
	}

	table.Tags = tags
}

func isLookupTable(table *schema.Table) bool {
	name := strings.ToLower(table.TableName)

	// Common lookup table patterns
	lookupPatterns := []string{
		"_type", "_types", "_status", "_statuses",
		"_category", "_categories", "_kind", "_kinds",
		"_code", "_codes", "lookup_", "ref_",
	}

	for _, pattern := range lookupPatterns {
		if strings.Contains(name, pattern) || strings.HasSuffix(name, pattern) {
			return true
		}
	}

	// Heuristic: Few columns, no outgoing FKs, possibly has incoming FKs
	if len(table.Columns) <= 5 &&
		len(table.OutgoingForeignKeys) == 0 &&
		len(table.IncomingForeignKeys) > 0 {
		return true
	}

	// Tables with name, code, or value column patterns
	hasCodeOrName := false
	for _, col := range table.Columns {
		colName := strings.ToLower(col.ColumnName)
		if colName == "name" || colName == "code" || colName == "value" || colName == "label" {
			hasCodeOrName = true
			break
		}
	}

	if hasCodeOrName && len(table.Columns) <= 4 {
		return true
	}

	return false
}

func isAuditTable(table *schema.Table) bool {
	name := strings.ToLower(table.TableName)

	auditPatterns := []string{
		"_audit", "_history", "_log", "_changelog",
		"audit_", "history_",
	}

	for _, pattern := range auditPatterns {
		if strings.Contains(name, pattern) {
			return true
		}
	}

	// Check for audit-like columns
	auditCols := 0
	for _, col := range table.Columns {
		colName := strings.ToLower(col.ColumnName)
		if colName == "action" || colName == "old_value" || colName == "new_value" ||
			colName == "changed_by" || colName == "changed_at" || colName == "revision" {
			auditCols++
		}
	}

	return auditCols >= 2
}

func isConfigTable(table *schema.Table) bool {
	name := strings.ToLower(table.TableName)

	configPatterns := []string{
		"config", "setting", "preference", "option",
	}

	for _, pattern := range configPatterns {
		if strings.Contains(name, pattern) {
			return true
		}
	}

	// Key-value style tables
	if len(table.Columns) <= 4 {
		hasKey := false
		hasValue := false
		for _, col := range table.Columns {
			colName := strings.ToLower(col.ColumnName)
			if colName == "key" || colName == "name" || colName == "setting_key" {
				hasKey = true
			}
			if colName == "value" || colName == "setting_value" {
				hasValue = true
			}
		}
		if hasKey && hasValue {
			return true
		}
	}

	return false
}

func isLogTable(table *schema.Table) bool {
	name := strings.ToLower(table.TableName)

	logPatterns := []string{
		"_log", "_logs", "_event", "_events",
		"log_", "event_",
	}

	for _, pattern := range logPatterns {
		if strings.Contains(name, pattern) || strings.HasSuffix(name, pattern) {
			return true
		}
	}

	// Check for log-like columns
	hasTimestamp := false
	hasEventType := false
	for _, col := range table.Columns {
		colName := strings.ToLower(col.ColumnName)
		if strings.Contains(colName, "timestamp") || colName == "logged_at" || colName == "event_time" {
			hasTimestamp = true
		}
		if colName == "event_type" || colName == "log_type" || colName == "action" {
			hasEventType = true
		}
	}

	return hasTimestamp && hasEventType
}
