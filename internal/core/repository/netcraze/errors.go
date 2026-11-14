package netcraze

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
)

var errorRegexp = regexp.MustCompile(
	`(?i)(syntax error|command not found|unknown command|invalid command|no such command|argument parse error|` +
		`ambiguous command|::\s*\w+\s*error\[|system failed|execute denied|access denied|ошибк|неизвестн|доступ.*запрещ)`,
)

func TranslateResponse(command, response string) error {
	match := errorRegexp.FindString(response)
	if match == "" {
		return nil
	}

	return domain.NewAppError(
		domain.KindInvalidArgument,
		fmt.Sprintf("command %q failed: %s", command, strings.TrimSpace(response)),
		domain.ErrRouterCLI,
	)
}

func translateTransport(err error, timeoutMessage, closedMessage string) error {
	if isExpectTimeout(err) {
		return domain.NewUnavailableError(timeoutMessage, err)
	}

	return domain.NewUnavailableError(closedMessage, err)
}

func isExpectTimeout(err error) bool {
	return errors.Is(err, errExpectTimeout)
}

func isEOF(err error) bool {
	return errors.Is(err, io.EOF)
}
