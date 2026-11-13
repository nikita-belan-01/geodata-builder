package domain

import "errors"

var (
	ErrUnknownList         = errors.New("unknown list")
	ErrEmptyDomainList     = errors.New("domain list is empty")
	ErrInvalidDomain       = errors.New("invalid domain")
	ErrTooManyDomains      = errors.New("domains do not fit into the configured parts")
	ErrInvalidPartCount    = errors.New("part count must be greater than zero")
	ErrInvalidLinesPerFile = errors.New("lines per file must be greater than zero")
	ErrSourceUnavailable   = errors.New("domain list source is unavailable")
)
