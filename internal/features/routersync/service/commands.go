package service

import "fmt"

func BuildCommands(chunks [][]string, prefix string) []string {
	commands := make([]string, 0, len(chunks)*3+1)

	for index, domains := range chunks {
		commands = append(commands,
			fmt.Sprintf("object-group fqdn %s%d", prefix, index+1),
			"no include")

		for _, name := range domains {
			commands = append(commands, "include "+name)
		}

		commands = append(commands, "exit")
	}

	return append(commands, "system configuration save")
}
