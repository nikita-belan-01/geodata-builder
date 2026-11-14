package config

import (
	"time"

	"github.com/nikita-belan-01/netcraze_route_builder/pkg/default_env"
)

const (
	defaultRouterHost     = "192.168.1.1"
	defaultRouterPort     = 22
	defaultRouterUser     = "admin"
	defaultKeychainSvc    = "netcraze-full-updater"
	defaultConnectTimeout = 10 * time.Second
	defaultCommandTimeout = 30 * time.Second
)

type Router struct {
	Host            string
	Port            int
	User            string
	KeychainService string
	ConnectTimeout  time.Duration
	CommandTimeout  time.Duration
	RouteGateway    string
	RouteInterface  string
}

func newRouter() (Router, error) {
	port, err := default_env.GetNumber("NETCRAZE_PORT", defaultRouterPort)
	if err != nil {
		return Router{}, err
	}

	if err := validatePort(port); err != nil {
		return Router{}, err
	}

	connectTimeout, err := default_env.GetDuration("NETCRAZE_CONNECT_TIMEOUT", defaultConnectTimeout)
	if err != nil {
		return Router{}, err
	}

	if err := validateDuration(connectTimeout); err != nil {
		return Router{}, err
	}

	commandTimeout, err := default_env.GetDuration("NETCRAZE_COMMAND_TIMEOUT", defaultCommandTimeout)
	if err != nil {
		return Router{}, err
	}

	if err := validateDuration(commandTimeout); err != nil {
		return Router{}, err
	}

	host := default_env.GetString("NETCRAZE_ROUTER", defaultRouterHost)
	if err := validateNotEmpty(host); err != nil {
		return Router{}, err
	}

	return Router{
		Host:            host,
		Port:            port,
		User:            default_env.GetString("NETCRAZE_USER", defaultRouterUser),
		KeychainService: default_env.GetString("NETCRAZE_KEYCHAIN_SERVICE", defaultKeychainSvc),
		ConnectTimeout:  connectTimeout,
		CommandTimeout:  commandTimeout,
		RouteGateway:    default_env.GetString("NETCRAZE_ROUTE_GATEWAY", ""),
		RouteInterface:  default_env.GetString("NETCRAZE_ROUTE_INTERFACE", ""),
	}, nil
}
