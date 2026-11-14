package netcraze

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
	"github.com/nikita-belan-01/netcraze_route_builder/pkg/logger"
)

var promptRegexp = regexp.MustCompile(`\(config[^)]*\)>[ \t]*`)

type SessionConfig struct {
	Host           string
	Port           int
	User           string
	Password       string
	ConnectTimeout time.Duration
	CommandTimeout time.Duration
}

type Session struct {
	client  *ssh.Client
	session *ssh.Session
	stdin   writeCloser
	console *console
	timeout time.Duration
	log     *logger.Logger
}

type writeCloser interface {
	Write(p []byte) (int, error)
}

func Dial(cfg SessionConfig, log *logger.Logger) (*Session, error) {
	if log == nil {
		log = logger.Discard()
	}

	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))

	log.Debug("dialing router", "addr", addr, "user", cfg.User)

	clientConfig := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.KeyboardInteractive(func(_, _ string, questions []string, _ []bool) ([]string, error) {
				answers := make([]string, len(questions))
				for i := range answers {
					answers[i] = cfg.Password
				}

				return answers, nil
			}),
			ssh.Password(cfg.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         cfg.ConnectTimeout,
	}

	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return nil, domain.NewUnavailableError(fmt.Sprintf("SSH: connect to %s", addr), err)
	}

	s := &Session{
		client:  client,
		timeout: cfg.CommandTimeout,
		log:     log,
	}

	session, err := client.NewSession()
	if err != nil {
		s.abort()

		return nil, domain.NewUnavailableError("SSH: open session", err)
	}

	s.session = session

	stdout, err := session.StdoutPipe()
	if err != nil {
		s.abort()

		return nil, domain.NewUnavailableError("SSH: attach stdout", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		s.abort()

		return nil, domain.NewUnavailableError("SSH: attach stdin", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", 50, 1000, modes); err != nil {
		s.abort()

		return nil, domain.NewUnavailableError("SSH: request pty", err)
	}

	if err := session.Shell(); err != nil {
		s.abort()

		return nil, domain.NewUnavailableError("SSH: open shell", err)
	}

	s.stdin = stdin
	s.console = newConsole(stdout)

	if _, err := s.console.expect(promptRegexp, cfg.CommandTimeout); err != nil {
		s.abort()

		return nil, translateTransport(err,
			"CLI: timed out waiting for the (config)> prompt",
			"CLI: connection closed before the prompt appeared")
	}

	log.Debug("router CLI ready")

	return s, nil
}

func (s *Session) Send(command string) (string, error) {
	response, err := s.SendRaw(command)
	if err != nil {
		return response, err
	}

	if err := TranslateResponse(command, response); err != nil {
		return response, err
	}

	return response, nil
}

func (s *Session) SendRaw(command string) (string, error) {
	s.log.Debug("router command", "command", command)

	if _, err := s.stdin.Write([]byte(command + "\r")); err != nil {
		return "", domain.NewUnavailableError("CLI: write command", err)
	}

	response, err := s.console.expect(promptRegexp, s.timeout)
	if err != nil {
		return response, translateTransport(err,
			fmt.Sprintf("CLI: timed out running %q", command),
			fmt.Sprintf("CLI: connection closed running %q", command))
	}

	return response, nil
}

func (s *Session) Quit() error {
	if _, err := s.stdin.Write([]byte("exit\r")); err != nil {
		return domain.NewUnavailableError("SSH: send exit", err)
	}

	if err := s.console.waitClosed(s.timeout); err != nil {
		if isExpectTimeout(err) {
			return domain.NewUnavailableError("CLI commands completed, but the session did not close cleanly", err)
		}

		if !isEOF(err) {
			return domain.NewUnavailableError("SSH: session ended unexpectedly", err)
		}
	}

	return nil
}

func (s *Session) Close() error {
	var sessionErr, clientErr error

	if s.session != nil {
		sessionErr = s.session.Close()
	}

	if s.client != nil {
		clientErr = s.client.Close()
	}

	if isEOF(sessionErr) {
		sessionErr = nil
	}

	if sessionErr != nil {
		sessionErr = fmt.Errorf("close ssh session: %w", sessionErr)
	}

	if clientErr != nil {
		clientErr = fmt.Errorf("close ssh client: %w", clientErr)
	}

	return errors.Join(sessionErr, clientErr)
}

func (s *Session) abort() {
	if err := s.Close(); err != nil {
		s.log.Debug("close ssh after dial failure", "err", err)
	}
}
