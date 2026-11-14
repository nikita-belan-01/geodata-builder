package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/nikita-belan-01/netcraze_route_builder/config"
	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/repository/netcraze"
	credentials_keychain "github.com/nikita-belan-01/netcraze_route_builder/internal/features/credentials/repository/keychain"
	credentials_service "github.com/nikita-belan-01/netcraze_route_builder/internal/features/credentials/service"
	domainlist_fs "github.com/nikita-belan-01/netcraze_route_builder/internal/features/domainlist/repository/fs"
	domainlist_v2fly "github.com/nikita-belan-01/netcraze_route_builder/internal/features/domainlist/repository/v2fly"
	domainlist_service "github.com/nikita-belan-01/netcraze_route_builder/internal/features/domainlist/service"
	routersync_cli "github.com/nikita-belan-01/netcraze_route_builder/internal/features/routersync/delivery/cli"
	routersync_netcrazecli "github.com/nikita-belan-01/netcraze_route_builder/internal/features/routersync/repository/netcrazecli"
	routersync_service "github.com/nikita-belan-01/netcraze_route_builder/internal/features/routersync/service"
	"github.com/nikita-belan-01/netcraze_route_builder/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}

		log.Printf("fatal: %v", err)
		os.Exit(exitCode(err))
	}
}

func exitCode(err error) int {
	switch domain.KindOf(err) {
	case domain.KindInvalidArgument, domain.KindTooLarge:
		return 2
	case domain.KindUnavailable:
		return 3
	default:
		return 1
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	a, err := parseArgs()
	if err != nil {
		return err
	}

	conf, err := config.New(config.OptionalPath(a.configPath))
	if err != nil {
		return fmt.Errorf("init config: %w", err)
	}

	if err := a.applyTo(conf); err != nil {
		return err
	}

	list, err := domain.LookupList(conf.Lists.Name)
	if err != nil {
		return err
	}

	appLogger, err := logger.New(conf.Logger.Dir, conf.Logger.Level, conf.Logger.Format)
	if err != nil {
		return fmt.Errorf("init app logger: %w", err)
	}

	defer func() {
		if err := appLogger.Close(); err != nil {
			log.Printf("close app logger: %v", err)
		}
	}()

	prefix := a.prefix
	if prefix == "" {
		prefix = list.Prefix
	}

	sourceDir := a.sourceDir
	if sourceDir == "" {
		sourceDir = filepath.Join(conf.Lists.OutputDir, list.DirName())
	}

	appLogger.Debug("initializing features", "feature", "domainlist")

	storage := domainlist_fs.NewListFilesRepository(appLogger)
	domainLists := domainlist_service.NewDomainListService(
		domainlist_v2fly.NewDomainListRepository(conf.Lists.SourceURL, appLogger), storage, appLogger)

	chunks, source, err := loadChunks(ctx, conf, domainLists, list, sourceDir)
	if err != nil {
		return err
	}

	total, _ := domain.CountDomains(chunks)
	fmt.Printf("List: %s. Source: %s. Domains: %d; parts: %d.\n",
		list.Name, source, total, len(chunks))

	commands := routersync_service.BuildCommands(chunks, prefix)

	if a.printCommands {
		fmt.Println(strings.Join(commands, "\n"))
	}

	if a.dryRun {
		fmt.Println("Dry-run: files and the router were not changed.")

		return nil
	}

	if !conf.Lists.UseExisting {
		if err := domainLists.SaveChunks(sourceDir, chunks); err != nil {
			return err
		}

		fmt.Printf("Updated only the folder: %s\n", sourceDir)
	}

	appLogger.Debug("initializing features", "feature", "credentials")

	credentials := credentials_service.NewCredentialsService(
		credentials_keychain.NewKeychainRepository(appLogger), appLogger)

	password, err := credentials.ResolvePassword(
		conf.Router.User, conf.Router.Host, conf.Router.KeychainService, a.nonInteractive)
	if err != nil {
		return err
	}

	appLogger.Debug("initializing features", "feature", "routersync")

	router := routersync_netcrazecli.NewRouterRepository(netcraze.SessionConfig{
		Host:           conf.Router.Host,
		Port:           conf.Router.Port,
		User:           conf.Router.User,
		Password:       password,
		ConnectTimeout: conf.Router.ConnectTimeout,
		CommandTimeout: conf.Router.CommandTimeout,
	}, appLogger)

	policy, selectPolicy := routersync_cli.Resolve(a.policy, a.nonInteractive)

	result, err := routersync_service.NewRouterSyncService(router, appLogger).Sync(ctx, chunks,
		routersync_service.SyncParams{
			Prefix:         prefix,
			Policy:         policy,
			RouteGateway:   conf.Router.RouteGateway,
			RouteInterface: conf.Router.RouteInterface,
			SelectPolicy:   selectPolicy,
		})
	if err != nil {
		return err
	}

	fmt.Printf(
		"Done: synced %s1…%s%d in policy %s; missing groups/routes created (%d/%d), "+
			"stale parts cleared: %d; other lists were not touched.\n",
		prefix, prefix, result.Parts, domain.PolicyLabel(result.Policy),
		result.MissingGroups, result.MissingRoutes, result.ClearedParts,
	)

	return nil
}

func loadChunks(
	ctx context.Context,
	conf *config.Config,
	domainLists *domainlist_service.DomainListService,
	list domain.List,
	sourceDir string,
) (chunks [][]string, source string, err error) {
	if conf.Lists.UseExisting {
		chunks, err = domainLists.LoadChunks(sourceDir, conf.Lists.LinesPerFile)

		return chunks, sourceDir, err
	}

	chunks, err = domainLists.BuildChunks(ctx, list, conf.Lists.LinesPerFile)

	return chunks, "current v2fly/domain-list-community", err
}

type args struct {
	configPath     string
	router         string
	port           int
	user           string
	list           string
	linesPerFile   int
	prefix         string
	routeGateway   string
	routeInterface string
	useExisting    bool
	sourceDir      string
	dryRun         bool
	printCommands  bool
	nonInteractive bool
	keychainSvc    string
	timeout        int
	policy         string
}

func parseArgs() (*args, error) {
	a := &args{}
	fs := flag.NewFlagSet("netcraze-update", flag.ContinueOnError)

	fs.StringVar(&a.configPath, "config", ".env", "config file path")
	fs.StringVar(&a.configPath, "c", ".env", "config file path")
	fs.StringVar(&a.router, "router", "", "router IP/hostname")
	fs.IntVar(&a.port, "port", 0, "SSH port")
	fs.StringVar(&a.user, "user", "", "Netcraze user")
	fs.StringVar(&a.list, "list", "", "which list to update: "+strings.Join(domain.ListNames(), "/"))
	fs.IntVar(&a.linesPerFile, "lines-per-file", 0,
		fmt.Sprintf("lines per part (default %d)", domain.DefaultLinesPerFile))
	fs.StringVar(&a.prefix, "prefix", "", "group prefix; defaults to rus_<list>_part_")
	fs.StringVar(&a.routeGateway, "route-gateway", "",
		"gateway IP for creating routes, if the router has no rus_<list>_part_N route yet")
	fs.StringVar(&a.routeInterface, "route-interface", "",
		"CLI interface name (e.g. PPPoE0) for new routes")
	fs.BoolVar(&a.useExisting, "use-existing", false,
		"don't download v2fly, use the current files from the output directory")
	fs.StringVar(&a.sourceDir, "source-dir", "", "directory to read the existing parts from")
	fs.BoolVar(&a.dryRun, "dry-run", false, "don't change files or the router")
	fs.BoolVar(&a.printCommands, "print-commands", false, "print the CLI commands that would be sent")
	fs.BoolVar(&a.nonInteractive, "non-interactive", false,
		"don't prompt; fail if the password isn't in NETCRAZE_PASSWORD or the keychain")
	fs.StringVar(&a.keychainSvc, "keychain-service", "",
		"macOS keychain service name for the password; empty disables the keychain")
	fs.IntVar(&a.timeout, "timeout", 0, "timeout for a single CLI command, in seconds")
	fs.StringVar(&a.policy, "policy", "",
		"access policy to create the routes in: a policy name, \"default\" for the global dns-proxy "+
			"section, or \"ask\" to pick from a numbered list of the router's policies")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	return a, nil
}

func (a *args) applyTo(conf *config.Config) error {
	if a.router != "" {
		conf.Router.Host = a.router
	}

	if a.port != 0 {
		conf.Router.Port = a.port
	}

	if a.user != "" {
		conf.Router.User = a.user
	}

	if a.keychainSvc != "" {
		conf.Router.KeychainService = a.keychainSvc
	}

	if a.routeGateway != "" {
		conf.Router.RouteGateway = a.routeGateway
	}

	if a.routeInterface != "" {
		conf.Router.RouteInterface = a.routeInterface
	}

	if a.timeout != 0 {
		conf.Router.CommandTimeout = time.Duration(a.timeout) * time.Second
	}

	if a.list != "" {
		conf.Lists.Name = a.list
	}

	if a.linesPerFile != 0 {
		conf.Lists.LinesPerFile = a.linesPerFile
	}

	if a.useExisting {
		conf.Lists.UseExisting = true
	}

	if a.nonInteractive && a.policy == routersync_cli.PolicyModeAsk {
		return domain.NewInvalidArgumentError(
			"--policy ask needs a terminal; pass --policy <name> or --policy default instead",
			domain.ErrInvalidPolicy)
	}

	return nil
}
