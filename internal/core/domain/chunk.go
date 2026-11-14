package domain

import (
	"sort"
	"strings"
)

func DomainsFromRules(rules []Rule, optimized bool) []string {
	domainSet := make(map[string]bool, len(rules))

	for _, rule := range rules {
		if optimized && rule.HasExcludedAttr() {
			continue
		}

		if domain := NormalizeRule(rule.Name); domain != "" {
			domainSet[domain] = true
		}
	}

	domains := make([]string, 0, len(domainSet))
	for domain := range domainSet {
		domains = append(domains, domain)
	}

	sort.Strings(domains)

	return domains
}

func CollapseCoveredSubdomains(domains []string) []string {
	domainSet := make(map[string]bool, len(domains))
	for _, domain := range domains {
		domainSet[domain] = true
	}

	result := make([]string, 0, len(domains))

	for _, domain := range domains {
		labels := strings.Split(domain, ".")

		covered := false
		for i := 1; i < len(labels)-1; i++ {
			if domainSet[strings.Join(labels[i:], ".")] {
				covered = true

				break
			}
		}

		if !covered {
			result = append(result, domain)
		}
	}

	return result
}

func SplitIntoParts(domains []string, linesPerFile int) [][]string {
	if linesPerFile <= 0 {
		linesPerFile = DefaultLinesPerFile
	}

	total := (len(domains) + linesPerFile - 1) / linesPerFile

	parts := make([][]string, 0, total)
	for i := range total {
		start := i * linesPerFile
		end := min(start+linesPerFile, len(domains))
		parts = append(parts, domains[start:end])
	}

	return parts
}

func SplitIntoRouterParts(domains []string, linesPerFile int) ([][]string, error) {
	if linesPerFile < 1 {
		return nil, NewInvalidArgumentError("split domains", ErrInvalidLinesPerFile)
	}

	if len(domains) == 0 {
		return nil, NewInvalidArgumentError("split domains", ErrEmptyDomainList)
	}

	return SplitIntoParts(domains, linesPerFile), nil
}

func CountDomains(chunks [][]string) (total, nonEmpty int) {
	for _, chunk := range chunks {
		total += len(chunk)

		if len(chunk) > 0 {
			nonEmpty++
		}
	}

	return total, nonEmpty
}
