// Package config loads configuration for user-svc from environment variables using viper.
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all configuration for user-svc.
type Config struct {
	Port string `mapstructure:"PORT"`

	// PostgreSQL
	DatabaseURL string `mapstructure:"DATABASE_URL"`

	// NATS
	NATSUrl string `mapstructure:"NATS_URL"`

	// OpenTelemetry
	OTLPEndpoint string `mapstructure:"OTLP_ENDPOINT"`

	// App
	Env string `mapstructure:"ENV"`
}

func mustBindEnv(keys ...string) {
	for _, key := range keys {
		if err := viper.BindEnv(key); err != nil {
			panic(fmt.Sprintf("viper.BindEnv(%q): %v", key, err))
		}
	}
}

// Load loads configuration from environment variables into a Config struct.
func Load() (*Config, error) {
	viper.AutomaticEnv()

	mustBindEnv(
		"PORT",
		"DATABASE_URL",
		"NATS_URL",
		"OTLP_ENDPOINT",
		"ENV",
	)

	viper.SetDefault("PORT", "8082")
	viper.SetDefault("ENV", "production")
	viper.SetDefault("OTLP_ENDPOINT", "http://jaeger:4318")

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}
