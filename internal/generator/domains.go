package generator

import (
	"path/filepath"
	"sort"

	"github.com/nenorrell/X-Rai/internal/heuristics"
	"github.com/nenorrell/X-Rai/internal/schema"
)

func (g *Generator) generateDomains(db *schema.Database) error {
	domains := heuristics.GroupIntoDomains(db.Tables)

	grouping := schema.DomainGrouping{
		Domains: make([]schema.Domain, 0, len(domains)),
	}

	for name, tables := range domains {
		d := schema.Domain{
			DomainName: name,
			Tables:     tables,
		}
		grouping.Domains = append(grouping.Domains, d)
	}

	// Sort domains for deterministic output
	sort.Slice(grouping.Domains, func(i, j int) bool {
		return grouping.Domains[i].DomainName < grouping.Domains[j].DomainName
	})

	// Sort tables within each domain
	for i := range grouping.Domains {
		sort.Strings(grouping.Domains[i].Tables)
	}

	return g.writeTOON(filepath.Join(g.outputDir, "db.domains.toon"), grouping)
}
