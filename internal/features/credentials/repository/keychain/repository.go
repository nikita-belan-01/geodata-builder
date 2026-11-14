package keychain

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/nikita-belan-01/netcraze_route_builder/pkg/logger"
)

type KeychainRepository struct {
	log *logger.Logger
}

func NewKeychainRepository(log *logger.Logger) *KeychainRepository {
	if log == nil {
		log = logger.Discard()
	}

	return &KeychainRepository{log: log}
}

func (r *KeychainRepository) FindPassword(account, service string) (string, bool) {
	if runtime.GOOS != "darwin" || service == "" {
		return "", false
	}

	out, err := exec.Command("/usr/bin/security",
		"find-generic-password", "-a", account, "-s", service, "-w").Output()
	if err != nil {
		r.log.Debug("keychain lookup missed", "account", account, "service", service)

		return "", false
	}

	password := strings.TrimRight(string(out), "\n")
	if password == "" {
		return "", false
	}

	r.log.Debug("password taken from the keychain", "account", account, "service", service)

	return password, true
}
