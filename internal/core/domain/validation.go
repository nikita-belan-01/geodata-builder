package domain

import "fmt"

func ValidateDomain(domain string) error {
	if NormalizeRule(domain) != domain {
		return fmt.Errorf("%w: %q", ErrInvalidDomain, domain)
	}

	return nil
}
