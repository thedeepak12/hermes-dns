package dns

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/thedeepak12/hermes-dns/internal/blocklist"
	"github.com/thedeepak12/hermes-dns/internal/cache"
	"github.com/thedeepak12/hermes-dns/internal/logger"
)

type Handler struct {
	resolver  *Resolver
	cache     *cache.Cache
	blocklist *blocklist.Blocklist
	logger    *logger.Logger
}

func NewHandler(resolver *Resolver, c *cache.Cache, log *logger.Logger) *Handler {
	return &Handler{
		resolver:  resolver,
		cache:     c,
		blocklist: nil,
		logger:    log,
	}
}

func (h *Handler) SetBlocklist(bl *blocklist.Blocklist) {
	h.blocklist = bl
}

func (h *Handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	startTime := time.Now()
	clientIP, _, _ := net.SplitHostPort(w.RemoteAddr().String())

	var domain string
	var result string
	var responseCode int
	var queryType string

	if len(r.Question) > 0 {
		question := r.Question[0]
		log.Printf("DNS Query: %s %s", question.Name, dns.TypeToString[question.Qtype])

		domain = strings.ToLower(strings.TrimSuffix(question.Name, "."))

		if h.blocklist != nil && h.blocklist.IsBlocked(domain) {
			result = "blocked"
			responseCode = dns.RcodeNameError
			log.Printf("Blocked: %s", domain)

			h.logger.Log(&logger.QueryLog{
				Timestamp:    time.Now(),
				ClientIP:     clientIP,
				Domain:       domain,
				QueryType:    queryType,
				Result:       result,
				ResponseCode: responseCode,
				LatencyMs:    float64(time.Since(startTime).Microseconds()) / 1000,
			})

			msg := new(dns.Msg)
			msg.SetReply(r)
			msg.SetRcode(r, dns.RcodeNameError)
			w.WriteMsg(msg)
			return
		}

		cacheKey := fmt.Sprintf("%s-%s", question.Name, dns.TypeToString[question.Qtype])

		if cached, found := h.cache.Get(cacheKey); found {
			result = "cached"
			responseCode = cached.Rcode
			log.Printf("Cache hit: %s", cacheKey)

			if h.logger != nil {
				h.logger.Log(&logger.QueryLog{
					Timestamp:    time.Now(),
					ClientIP:     clientIP,
					Domain:       domain,
					QueryType:    queryType,
					Result:       result,
					ResponseCode: responseCode,
					LatencyMs:    float64(time.Since(startTime).Microseconds()) / 1000,
				})
			}

			cached.SetReply(r)
			w.WriteMsg(cached)
			return
		}

		log.Printf("Cache miss: %s", cacheKey)
	}

	ctx := context.Background()
	resp, err := h.resolver.Resolve(ctx, r)

	result = "resolved"
	responseCode = dns.RcodeSuccess
	if err != nil {
		result = "error"
		responseCode = dns.RcodeServerFailure
		log.Printf("Resolution failed: %v", err)
	}

	if h.logger != nil {
		h.logger.Log(&logger.QueryLog{
			Timestamp:    time.Now(),
			ClientIP:     clientIP,
			Domain:       domain,
			QueryType:    queryType,
			Result:       result,
			ResponseCode: responseCode,
			LatencyMs:    float64(time.Since(startTime).Microseconds()) / 1000,
		})
	}

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
