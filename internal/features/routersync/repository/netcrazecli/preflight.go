package netcrazecli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
)

type preflightInfo struct {
	existingGroups     map[string]bool
	describedGroups    map[string]string
	routeSuffixByGroup map[string]string
	policies           []domain.Policy
}

func parsePreflight(preflight, prefix, policy string) preflightInfo {
	info := preflightInfo{
		existingGroups:     map[string]bool{},
		describedGroups:    map[string]string{},
		routeSuffixByGroup: map[string]string{},
	}

	currentGroup := ""
	currentPolicy := domain.DefaultPolicy

	for line := range strings.SplitSeq(preflight, "\n") {
		clean := strings.TrimSpace(line)
		words := strings.Fields(clean)
		wordCount := len(words)

		if wordCount >= 3 && words[0] == "ip" && words[1] == "policy" {
			currentPolicy = words[2]
			currentGroup = ""

			info.policies = append(info.policies, domain.Policy{Name: currentPolicy})

			continue
		}

		if currentPolicy != domain.DefaultPolicy && wordCount >= 2 &&
			words[0] == "description" && len(info.policies) > 0 {
			description := decodeEscapes(strings.Trim(strings.Join(words[1:], " "), `"`))
			info.policies[len(info.policies)-1].Description = description

			continue
		}

		for position := 0; position+2 < wordCount; position++ {
			first, second, name := words[position], words[position+1], words[position+2]

			if first == "object-group" && second == "fqdn" {
				info.existingGroups[name] = true
				currentGroup = name
			}

			if first == "route" && second == "object-group" && currentPolicy == policy {
				info.routeSuffixByGroup[name] = strings.Join(words[position+3:], " ")
			}
		}

		if currentGroup != "" && wordCount >= 2 && words[0] == "description" {
			description := strings.Join(words[1:], " ")
			if strings.HasPrefix(description, prefix) {
				info.describedGroups[description] = currentGroup
			}
		}

		if clean == "!" {
			currentGroup = ""
			currentPolicy = domain.DefaultPolicy
		}
	}

	return info
}

func decodeEscapes(s string) string {
	if !strings.Contains(s, `\x`) {
		return s
	}

	var out []byte

	for i := 0; i < len(s); {
		if strings.HasPrefix(s[i:], `\x`) && i+4 <= len(s) {
			if b, err := strconv.ParseUint(s[i+2:i+4], 16, 8); err == nil {
				out = append(out, byte(b))
				i += 4

				continue
			}
		}

		out = append(out, s[i])
		i++
	}

	return string(out)
}

func highestManagedIndex(info preflightInfo, prefix string) int {
	highest := 0

	check := func(name string) {
		rest, ok := strings.CutPrefix(name, prefix)
		if !ok || rest == "" || strings.Trim(rest, "0123456789") != "" {
			return
		}

		if index, err := strconv.Atoi(rest); err == nil && index > highest {
			highest = index
		}
	}

	for name := range info.existingGroups {
		check(name)
	}

	for displayName := range info.describedGroups {
		check(displayName)
	}

	return highest
}

func clearCommands(prefix string, from, to int) []string {
	var commands []string

	for index := from + 1; index <= to; index++ {
		commands = append(commands,
			fmt.Sprintf("object-group fqdn %s%d", prefix, index),
			"no include",
			"exit")
	}

	return commands
}

type managedGroups struct {
	displayToActual map[string]string
	missingGroups   int
	missingRoutes   []string
	routeSuffix     string
}

func resolveManagedGroups(
	preflightLen int,
	info preflightInfo,
	prefix string,
	partCount int,
	routeGateway, routeInterface string,
) (managedGroups, error) {
	mg := managedGroups{displayToActual: map[string]string{}}

	for index := 1; index <= partCount; index++ {
		displayName := fmt.Sprintf("%s%d", prefix, index)

		actualName := displayName

		switch {
		case info.describedGroups[displayName] != "":
			actualName = info.describedGroups[displayName]
		case info.existingGroups[displayName]:
			actualName = displayName
		default:
			mg.missingGroups++
		}

		mg.displayToActual[displayName] = actualName

		if suffix, ok := info.routeSuffixByGroup[actualName]; ok {
			if mg.routeSuffix == "" {
				mg.routeSuffix = suffix
			}

			continue
		}

		mg.missingRoutes = append(mg.missingRoutes, actualName)
	}

	if len(mg.missingRoutes) > 0 && mg.routeSuffix == "" {
		mg.routeSuffix = strings.TrimSpace(fmt.Sprintf("%s %s auto reject", routeGateway, routeInterface))

		if routeGateway == "" && routeInterface == "" {
			return mg, domain.NewInvalidArgumentError(fmt.Sprintf(
				"no route template found. For the initial setup, provide --route-interface and, if needed, "+
					"--route-gateway; the router was not changed. Diagnostics: bytes=%d, FQDN groups=%d, "+
					"managed descriptions=%d, routes=%d",
				preflightLen, len(info.existingGroups), len(info.describedGroups), len(info.routeSuffixByGroup)),
				domain.ErrNoRouteTemplate)
		}
	}

	return mg, nil
}

func rewriteObjectGroupCommand(command string, displayToActual map[string]string) string {
	words := strings.Fields(command)
	if len(words) != 3 || words[0] != "object-group" || words[1] != "fqdn" {
		return command
	}

	if actual, ok := displayToActual[words[2]]; ok {
		return "object-group fqdn " + actual
	}

	return command
}
