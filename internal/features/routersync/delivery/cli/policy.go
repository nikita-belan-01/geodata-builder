package cli

import (
	"os"

	corecli "github.com/nikita-belan-01/netcraze_route_builder/internal/core/delivery/cli"
	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
)

const (
	PolicyModeAsk     = "ask"
	PolicyModeDefault = "default"
)

func Selector(policies []domain.Policy) (string, error) {
	options := make([]corecli.Option, 0, len(policies)+1)
	options = append(options, corecli.Option{
		Label: "default — global dns-proxy section (all clients without a policy)",
		Value: domain.DefaultPolicy,
	})

	for _, policy := range policies {
		options = append(options, corecli.Option{
			Label: policy.Label(),
			Value: policy.Name,
		})
	}

	return corecli.SelectOption(os.Stdout, os.Stdin, "Access policies on the router:", options)
}

func Resolve(flagValue string, nonInteractive bool) (string, func([]domain.Policy) (string, error)) {
	switch flagValue {
	case PolicyModeAsk:
		return domain.DefaultPolicy, Selector
	case PolicyModeDefault:
		return domain.DefaultPolicy, nil
	case "":
		if nonInteractive || !corecli.IsTerminal() {
			return domain.DefaultPolicy, nil
		}

		return domain.DefaultPolicy, Selector
	default:
		return flagValue, nil
	}
}
