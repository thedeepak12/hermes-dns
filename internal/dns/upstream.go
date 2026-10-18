package dns

import (
	"context"
	"fmt"
	"log"

	"github.com/miekg/dns"
	"github.com/thedeepak12/hermes-dns/internal/config"
)

type Resolver struct {
	config *config.Config
	client *dns.Client
}

func NewResolver(cfg *config.Config) *Resolver {
	return &Resolver{
		config: cfg,
		client: &dns.Client{
			Timeout: cfg.Upstream.Timeout,
		},
	}
}

func (r *Resolver) Resolve(ctx context.Context, req *dns.Msg) (*dns.Msg, error) {
	if len(r.config.Upstream.Providers) == 0 {
		return nil, fmt.Errorf("No upstream providers configured")
	}

	upstreamReq := req.Copy()
	var lastErr error

	for _, upstreamAddr := range r.config.Upstream.Providers {
		resp, _, err := r.client.ExchangeContext(ctx, upstreamReq, upstreamAddr)
		if err != nil {
			log.Printf("Upstream %s failed: %v", upstreamAddr, err)
			lastErr = err
			continue
		}

		if resp == nil {
			log.Printf("Upstream %s returned nil response", upstreamAddr)
			lastErr = fmt.Errorf("Upstream returned nil response")
			continue
		}

		log.Printf("Resolved via: %s", upstreamAddr)
		return resp, nil
	}

	return nil, fmt.Errorf("all upstream providers failed: %w", lastErr)
}
