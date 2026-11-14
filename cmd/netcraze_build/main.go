package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/nikita-belan-01/netcraze_route_builder/config"
	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
	amnezia_file "github.com/nikita-belan-01/netcraze_route_builder/internal/features/amnezia/delivery/file"
	amnezia_dns "github.com/nikita-belan-01/netcraze_route_builder/internal/features/amnezia/repository/dns"
	amnezia_service "github.com/nikita-belan-01/netcraze_route_builder/internal/features/amnezia/service"
	domainlist_fs "github.com/nikita-belan-01/netcraze_route_builder/internal/features/domainlist/repository/fs"
	domainlist_v2fly "github.com/nikita-belan-01/netcraze_route_builder/internal/features/domainlist/repository/v2fly"
	domainlist_service "github.com/nikita-belan-01/netcraze_route_builder/internal/features/domainlist/service"
	"github.com/nikita-belan-01/netcraze_route_builder/pkg/logger"
)

const amneziaFileName = "amnezia_sites.json"

type args struct {
	configPath   string
	list         string
	linesPerFile int
}

func main() {
	if err := run(); err != nil {
		log.Printf("fatal: %v", err)
		os.Exit(1)
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

	if a.linesPerFile > 0 {
		conf.Lists.LinesPerFile = a.linesPerFile
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

	selected, err := selectLists(a.list)
	if err != nil {
		return err
	}

	appLogger.Debug("initializing features", "feature", "domainlist")

	source := domainlist_v2fly.NewDomainListRepository(conf.Lists.SourceURL, appLogger)
	storage := domainlist_fs.NewListFilesRepository(appLogger)
	domainLists := domainlist_service.NewDomainListService(source, storage, appLogger)

	appLogger.Debug("initializing features", "feature", "amnezia")

	resolver := amnezia_dns.NewResolverRepository(appLogger)
	amneziaExport := amnezia_service.NewAmneziaService(resolver, appLogger)

	for _, list := range selected {
		if err := buildList(ctx, conf, domainLists, amneziaExport, storage, list); err != nil {
			return fmt.Errorf("build %s: %w", list.Name, err)
		}
	}

	if err := storage.Archive(conf.Lists.OutputDir, conf.Lists.ArchivePath); err != nil {
		return err
	}

	return nil
}

func parseArgs() (*args, error) {
	a := &args{}

	flag.StringVar(&a.configPath, "config", ".env", "config file path")
	flag.StringVar(&a.configPath, "c", ".env", "config file path")
	flag.StringVar(&a.list, "list", "",
		"build only this list: "+fmt.Sprint(domain.ListNames())+" (default: all of them)")
	flag.IntVar(&a.linesPerFile, "lines-per-file", 0,
		fmt.Sprintf("lines per part (default %d)", domain.DefaultLinesPerFile))

	flag.Parse()

	return a, nil
}

func selectLists(name string) ([]domain.List, error) {
	if name == "" {
		return domain.Lists(), nil
	}

	list, err := domain.LookupList(name)
	if err != nil {
		return nil, err
	}

	return []domain.List{list}, nil
}

func buildList(
	ctx context.Context,
	conf *config.Config,
	domainLists *domainlist_service.DomainListService,
	amneziaExport *amnezia_service.AmneziaService,
	storage *domainlist_fs.ListFilesRepository,
	list domain.List,
) error {
	domains, err := domainLists.BuildDomains(ctx, list)
	if err != nil {
		return err
	}

	versionDir := filepath.Join(conf.Lists.OutputDir, list.DirName())

	chunks := domainLists.SplitDomains(domains, conf.Lists.LinesPerFile)
	if err := domainLists.SaveChunks(versionDir, chunks); err != nil {
		return err
	}

	sites, resolved := amneziaExport.Sites(ctx, domains)

	data, err := amnezia_file.Encode(sites)
	if err != nil {
		return err
	}

	if err := storage.WriteFile(filepath.Join(versionDir, amneziaFileName), data); err != nil {
		return err
	}

	fmt.Printf("%s: %d domains, %d parts of %d lines; amnezia_sites.json resolved %d/%d\n",
		list.Name, len(domains), len(chunks), conf.Lists.LinesPerFile, resolved, len(domains))

	return nil
}
