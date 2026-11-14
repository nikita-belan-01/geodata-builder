package domain

import "errors"

var (
	ErrUnknownList         = errors.New("unknown list")
	ErrEmptyDomainList     = errors.New("domain list is empty")
	ErrInvalidDomain       = errors.New("invalid domain")
	ErrTooManyDomains      = errors.New("too many domains in a part")
	ErrInvalidLinesPerFile = errors.New("lines per file must be greater than zero")
	ErrNoRouteTemplate     = errors.New("no route template found")
	ErrPasswordNotFound    = errors.New("password not found")
	ErrNoTerminal          = errors.New("no terminal available")
	ErrInvalidPolicy       = errors.New("invalid policy")
	ErrInvalidChoice       = errors.New("invalid choice")
	ErrRouterCLI           = errors.New("router CLI rejected a command")
	ErrSourceUnavailable   = errors.New("domain list source is unavailable")
)
