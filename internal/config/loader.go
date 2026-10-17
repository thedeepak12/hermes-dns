package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func Load(configPath string) (*Config, error) {
	v := viper.New()

	if configPath == "" {
		configPath = "./config/config.yaml"
	}

	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.listen", ":53")
	v.SetDefault("server.doh.enabled", false)
	v.SetDefault("server.doh.listen", ":443")
	v.SetDefault("server.dot.enabled", false)
	v.SetDefault("server.dot.listen", ":853")
	v.SetDefault("upstream.timeout", "5s")
	v.SetDefault("cache.enabled", true)
	v.SetDefault("cache.max_size", 10000)
	v.SetDefault("cache.default_ttl", 3600)
}

func validate(cfg *Config) error {
	if len(cfg.Upstream.Providers) == 0 {
		return fmt.Errorf("at least one upstream provider is required")
	}

	if cfg.Cache.MaxSize <= 0 {
		return fmt.Errorf("cache max_size must be positive")
	}

	return nil
}
