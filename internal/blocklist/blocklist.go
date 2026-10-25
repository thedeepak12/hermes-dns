package blocklist

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

type Blocklist struct {
	domains map[string]bool
	mu      sync.RWMutex
}

func New(sources map[string][]string) (*Blocklist, error) {
	bl := &Blocklist{
		domains: make(map[string]bool),
	}

	for category, files := range sources {
		for _, file := range files {
			if err := bl.loadFile(file); err != nil {
				return nil, fmt.Errorf("failed to load %s (%s): %w", file, category, err)
			}
		}
	}

	return bl, nil
}

func (b *Blocklist) loadFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	b.mu.Lock()
	defer b.mu.Unlock()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		var domain string
		if len(parts) == 1 {
			domain = parts[0]
		} else if len(parts) >= 2 {
			if strings.Contains(parts[0], ".") {
				domain = parts[1]
			} else {
				domain = parts[0]
			}
		}

		domain = strings.ToLower(strings.TrimSuffix(domain, "."))
		if domain != "" {
			b.domains[domain] = true
		}
	}

	return scanner.Err()
}

func (b *Blocklist) IsBlocked(domain string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	domain = strings.ToLower(strings.TrimSuffix(domain, "."))

	if b.domains[domain] {
		return true
	}

	parts := strings.Split(domain, ".")
	for i := 1; i < len(parts); i++ {
		parent := strings.Join(parts[i:], ".")
		if b.domains[parent] {
			return true
		}
	}

	return false
}

func (b *Blocklist) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.domains)
}
