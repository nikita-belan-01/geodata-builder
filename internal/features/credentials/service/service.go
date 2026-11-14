package service

import (
	"fmt"
	"os"

	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/delivery/cli"
	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
	"github.com/nikita-belan-01/netcraze_route_builder/pkg/logger"
)

const PasswordEnvVar = "NETCRAZE_PASSWORD"

type SecretsRepository interface {
	FindPassword(account, service string) (string, bool)
}

type CredentialsService struct {
	secrets SecretsRepository
	log     *logger.Logger
}

func NewCredentialsService(secrets SecretsRepository, log *logger.Logger) *CredentialsService {
	if log == nil {
		log = logger.Discard()
	}

	return &CredentialsService{
		secrets: secrets,
		log:     log,
	}
}

func (s *CredentialsService) ResolvePassword(
	user, host, keychainService string,
	nonInteractive bool,
) (string, error) {
	if password := os.Getenv(PasswordEnvVar); password != "" {
		s.log.Debug("password taken from the environment", "var", PasswordEnvVar)

		return password, nil
	}

	if password, ok := s.secrets.FindPassword(user, keychainService); ok {
		return password, nil
	}

	if nonInteractive {
		return "", domain.NewInvalidArgumentError(fmt.Sprintf(
			"password not found: set %s or save it in the keychain with service=%q",
			PasswordEnvVar, keychainService), domain.ErrPasswordNotFound)
	}

	password, err := cli.ReadSecret(fmt.Sprintf("SSH password %s@%s: ", user, host))
	if err != nil {
		return "", err
	}

	return password, nil
}
