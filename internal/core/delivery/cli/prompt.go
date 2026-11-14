package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"

	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
)

type Option struct {
	Label string
	Value string
}

func SelectOption(out io.Writer, in io.Reader, title string, options []Option) (string, error) {
	var menu strings.Builder

	fmt.Fprintf(&menu, "\n%s\n", title)

	for index, option := range options {
		fmt.Fprintf(&menu, "  %d) %s\n", index, option.Label)
	}

	fmt.Fprintf(&menu, "Choose [0-%d]: ", len(options)-1)

	if _, err := io.WriteString(out, menu.String()); err != nil {
		return "", fmt.Errorf("print the menu: %w", err)
	}

	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read the choice: %w", err)
	}

	answer := strings.TrimSpace(line)

	choice, err := strconv.Atoi(answer)
	if err != nil || choice < 0 || choice >= len(options) {
		return "", domain.NewInvalidArgumentError(
			fmt.Sprintf("invalid choice %q: enter a number between 0 and %d", answer, len(options)-1),
			domain.ErrInvalidChoice)
	}

	return options[choice].Value, nil
}

func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func ReadSecret(prompt string) (string, error) {
	if !IsTerminal() {
		return "", domain.NewInvalidArgumentError("read secret", domain.ErrNoTerminal)
	}

	fmt.Print(prompt)

	secret, err := term.ReadPassword(int(os.Stdin.Fd()))

	fmt.Println()

	if err != nil {
		return "", fmt.Errorf("read secret: %w", err)
	}

	return strings.TrimSpace(string(secret)), nil
}
