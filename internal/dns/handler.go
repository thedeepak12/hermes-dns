package dns

import (
	"context"
	"fmt"
	"log"

	"github.com/miekg/dns"
	"github.com/thedeepak12/hermes-dns/internal/cache"
)

type Handler struct {
	resolver *Resolver
	cache    *cache.Cache
}

func NewHandler(resolver *Resolver, c *cache.Cache) *Handler {
	return &Handler{
		resolver: resolver,
		cache:    c,
	}
}

func (h *Handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) > 0 {
		question := r.Question[0]
		log.Printf("DNS Query: %s %s", question.Name, dns.TypeToString[question.Qtype])

		cacheKey := fmt.Sprintf("%s-%s", question.Name, dns.TypeToString[question.Qtype])

		if cached, found := h.cache.Get(cacheKey); found {
			log.Printf("Cache hit: %s", cacheKey)
			cached.SetReply(r)
			w.WriteMsg(cached)
			return
		}

		log.Printf("Cache miss: %s", cacheKey)
	}

	ctx := context.Background()
	resp, err := h.resolver.Resolve(ctx, r)
	if err != nil {
		log.Printf("Resolution failed: %v", err)

		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.SetRcode(r, dns.RcodeServerFailure)
		w.WriteMsg(msg)
		return
	}

	if len(r.Question) > 0 {
		question := r.Question[0]
		cacheKey := fmt.Sprintf("%s-%s", question.Name, dns.TypeToString[question.Qtype])
		h.cache.Set(cacheKey, resp)
		log.Printf("Cached: %s", cacheKey)
	}

	if err := w.WriteMsg(resp); err != nil {
		log.Printf("Failed to send response: %v", err)
	}
}
