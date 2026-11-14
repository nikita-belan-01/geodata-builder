package domain

import "fmt"

const DefaultPolicy = ""

type Policy struct {
	Name        string
	Description string
}

func (p Policy) Label() string {
	if p.Description == "" {
		return p.Name
	}

	return fmt.Sprintf("%s — %s", p.Name, p.Description)
}

func PolicyLabel(policy string) string {
	if policy == DefaultPolicy {
		return "default (global dns-proxy section)"
	}

	return policy
}

type SyncRequest struct {
	Prefix         string
	PartCount      int
	Commands       []string
	Policy         string
	RouteGateway   string
	RouteInterface string

	SelectPolicy func(policies []Policy) (string, error)
}

type SyncResult struct {
	Parts         int
	ClearedParts  int
	MissingGroups int
	MissingRoutes int
	Policy        string
}
