package heuristics

import (
	"strings"

	"github.com/nenorrell/xrai/internal/schema"
)

// GroupIntoDomains groups tables into logical domains based on naming patterns
// and relationships.
func GroupIntoDomains(tables []*schema.Table) map[string][]string {
	domains := make(map[string][]string)
	assigned := make(map[string]bool)

	// First pass: assign by common prefixes
	prefixGroups := groupByPrefix(tables)
	for prefix, tableNames := range prefixGroups {
		if len(tableNames) >= 2 {
			domainName := formatDomainName(prefix)
			domains[domainName] = tableNames
			for _, tn := range tableNames {
				assigned[tn] = true
			}
		}
	}

	// Second pass: group remaining by relationship clusters
	unassigned := make([]*schema.Table, 0)
	for _, t := range tables {
		if !assigned[t.TableName] {
			unassigned = append(unassigned, t)
		}
	}

	// Group remaining tables into "other" or by detected patterns
	for _, t := range unassigned {
		domain := detectDomainByContent(t)
		domains[domain] = append(domains[domain], t.TableName)
	}

	// Remove empty domains
	for name, tables := range domains {
		if len(tables) == 0 {
			delete(domains, name)
		}
	}

	return domains
}

func groupByPrefix(tables []*schema.Table) map[string][]string {
	groups := make(map[string][]string)

	for _, t := range tables {
		name := strings.ToLower(t.TableName)

		// Look for common prefixes (e.g., "user_", "order_", "product_")
		if idx := strings.Index(name, "_"); idx > 0 && idx < len(name)-1 {
			prefix := name[:idx]
			if len(prefix) >= 3 { // Ignore very short prefixes
				groups[prefix] = append(groups[prefix], t.TableName)
			}
		}
	}

	return groups
}

func formatDomainName(prefix string) string {
	// Capitalize first letter
	if len(prefix) == 0 {
		return "Other"
	}
	return strings.ToUpper(prefix[:1]) + prefix[1:]
}

func detectDomainByContent(table *schema.Table) string {
	name := strings.ToLower(table.TableName)

	// Common domain patterns
	domainPatterns := map[string][]string{
		"Auth":      {"user", "auth", "login", "session", "permission", "role", "token"},
		"Billing":   {"payment", "invoice", "subscription", "billing", "charge", "refund"},
		"Commerce":  {"order", "cart", "product", "catalog", "inventory", "price"},
		"Content":   {"post", "article", "comment", "media", "attachment", "content"},
		"Messaging": {"message", "notification", "email", "sms", "chat"},
		"Analytics": {"metric", "analytics", "stat", "tracking", "event"},
	}

	for domain, patterns := range domainPatterns {
		for _, pattern := range patterns {
			if strings.Contains(name, pattern) {
				return domain
			}
		}
	}

	// Check if it's a system/metadata table
	systemPatterns := []string{"schema", "migration", "meta", "sys_", "pg_"}
	for _, pattern := range systemPatterns {
		if strings.Contains(name, pattern) {
			return "System"
		}
	}

	return "Other"
}
