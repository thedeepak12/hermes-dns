package dns

import (
	"context"
	"log"

	"github.com/miekg/dns"
)

type Handler struct {
	resolver *Resolver
}

func NewHandler(resolver *Resolver) *Handler {
	return &Handler{
		resolver: resolver,
	}
}

func (h *Handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) > 0 {
		question := r.Question[0]
		log.Printf("DNS Query: %s %s", question.Name, dns.TypeToString[question.Qtype])
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

	if err := w.WriteMsg(resp); err != nil {
		log.Printf("Failed to send response: %v", err)
	}
}
