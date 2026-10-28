package config

import "time"

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Upstream UpstreamConfig `mapstructure:"upstream"`
	Cache    CacheConfig    `mapstructure:"cache"`
	Logging  LoggingConfig  `mapstructure:"logging"`
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

type LoggingConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	LogFile string `mapstructure:"log_file"`
	Format  string `mapstructure:"format"`
}
