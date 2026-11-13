package domain

import "strings"

type Rule struct {
	Name  string
	Attrs map[string]bool
}

func (r Rule) HasExcludedAttr() bool {
	for attr := range r.Attrs {
		if ExcludeAttrsOptimized[attr] {
			return true
		}
	}

	return false
}

func NormalizeRule(rule string) string {
	if strings.HasPrefix(rule, "include:") ||
		strings.HasPrefix(rule, "regexp:") ||
		strings.HasPrefix(rule, "keyword:") {
		return ""
	}

	switch {
	case strings.HasPrefix(rule, "domain:"):
		rule = rule[len("domain:"):]
	case strings.HasPrefix(rule, "full:"):
		rule = rule[len("full:"):]
	}

	rule = strings.ToLower(strings.TrimSpace(rule))
	rule = strings.TrimRight(rule, ".")

	if rule == "" {
		return ""
	}

	for _, ch := range []string{":", "/", "\\", "*", " "} {
		if strings.Contains(rule, ch) {
			return ""
		}
	}

	return rule
}
