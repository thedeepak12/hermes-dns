package dns

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/miekg/dns"
	"github.com/thedeepak12/hermes-dns/internal/cache"
	"github.com/thedeepak12/hermes-dns/internal/config"
)

type Server struct {
	server   *dns.Server
	config   *config.Config
	handler  *Handler
	resolver *Resolver
	cache    *cache.Cache
}

func NewServer(cfg *config.Config) *Server {
	resolver := NewResolver(cfg)
	cache, _ := cache.New(&cfg.Cache)
	handler := NewHandler(resolver, cache)

	return &Server{
		config:   cfg,
		resolver: resolver,
		handler:  handler,
		cache:    cache,
	}
}

func (s *Server) Start() error {
	s.server = &dns.Server{
		Addr:    s.config.Server.Listen,
		Net:     "udp",
		Handler: s.handler,
	}

	log.Printf("Starting DNS server on %s", s.config.Server.Listen)

	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			log.Fatalf("DNS server failed: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down DNS server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := s.server.ShutdownContext(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
		return err
	}

	defer s.cache.Close()

	log.Println("DNS server stopped")
	return nil
}
