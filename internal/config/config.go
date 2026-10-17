package config

import "time"

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Upstream UpstreamConfig `mapstructure:"upstream"`
	Cache    CacheConfig    `mapstructure:"cache"`
}

type ServerConfig struct {
	Listen string    `mapstructure:"listen"`
	DoH    DoHConfig `mapstructure:"doh"`
	DoT    DoTConfig `mapstructure:"dot"`
}

type DoHConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Listen  string `mapstructure:"listen"`
}

type DoTConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Listen  string `mapstructure:"listen"`
}

type UpstreamConfig struct {
	Providers []string      `mapstructure:"providers"`
	Timeout   time.Duration `mapstructure:"timeout"`
}

type CacheConfig struct {
	Enabled    bool `mapstructure:"enabled"`
	MaxSize    int  `mapstructure:"max_size"`
	DefaultTTL int  `mapstructure:"default_ttl"`
}
