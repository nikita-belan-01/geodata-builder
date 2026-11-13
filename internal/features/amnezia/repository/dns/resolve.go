package dns

import (
	"context"
	"net"
	"sort"
	"sync"

	"github.com/nikita-belan-01/netcraze_route_builder/internal/core/domain"
)

func (r *ResolverRepository) ResolveSites(ctx context.Context, domains []string) []domain.AmneziaSite {
	var (
		mu        sync.Mutex
		sites     []domain.AmneziaSite
		waitGroup sync.WaitGroup
		semaphore = make(chan struct{}, r.concurrency)
	)

	for _, name := range domains {
		waitGroup.Add(1)
		semaphore <- struct{}{}

		go func(name string) {
			defer waitGroup.Done()
			defer func() { <-semaphore }()

			for _, ip := range r.lookupIPv4(ctx, name) {
				mu.Lock()
				sites = append(sites, domain.AmneziaSite{Hostname: name, IP: ip})
				mu.Unlock()
			}
		}(name)
	}

	waitGroup.Wait()

	sort.Slice(sites, func(i, j int) bool {
		if sites[i].Hostname != sites[j].Hostname {
			return sites[i].Hostname < sites[j].Hostname
		}

		return sites[i].IP < sites[j].IP
	})

	r.log.Debug("domains resolved", "requested", len(domains), "entries", len(sites))

	return sites
}

func (r *ResolverRepository) lookupIPv4(ctx context.Context, name string) []string {
	lookupCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	addrs, err := net.DefaultResolver.LookupIP(lookupCtx, "ip4", name)
	if err != nil {
		return nil
	}

	ips := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		ips = append(ips, addr.String())
	}

	return ips
}
