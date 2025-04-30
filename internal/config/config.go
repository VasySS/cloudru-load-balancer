// Package config parses and provides configuration for the application.
package config

import (
	"log/slog"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// RateLimiterType is a type of rate limiter.
type RateLimiterType string

// BalancerType is a type of load balancer.
type BalancerType string

// A list of available balancers and rate limiters algorithms.
const (
	LeastConnectionsType BalancerType    = "least-connections"
	RandomType           BalancerType    = "random"
	RoundRobinType       BalancerType    = "round-robin"
	TokenBucketType      RateLimiterType = "token-bucket"
	LeakyBucketType      RateLimiterType = "leaky-bucket"
)

// Postgres contains Postgres connection credentials.
type Postgres struct {
	User     string `env:"PG_USER" env-required:"true"`
	Password string `env:"PG_PASS" env-required:"true"`
	Host     string `env:"PG_HOST" env-required:"true"`
	Database string `env:"PG_DB"   env-required:"true"`
}

// Balancer contains configuration for balancers.
type Balancer struct {
	Type BalancerType `env-default:"least-connections" yaml:"type"`
}

// RateLimit contains configuration for rate limiters.
type RateLimit struct {
	Type     RateLimiterType `env-default:"token-bucket" yaml:"type"`
	Capacity int             `env-default:"100"          yaml:"capacity"`
}

// configYAML contains values from /config/config.yaml.
type configYAML struct {
	Backends  []string  `env-required:"true" yaml:"backends"`
	Balancer  Balancer  `yaml:"balancer"`
	RateLimit RateLimit `yaml:"rateLimit"`
}

// configENV contains values from .env.
type configENV struct {
	Port     string `env:"APP_PORT" env-default:"8080"`
	Postgres Postgres
}

// Config contains application configuration.
type Config struct {
	YAML configYAML
	ENV  configENV
}

// MustInit reads .yaml config, then environment variables and returns a new global config.
func MustInit() Config {
	var cfg Config

	if err := cleanenv.ReadConfig("./config/config.yaml", &cfg.YAML); err != nil {
		slog.Info("failed to read ./config/config.yaml", slog.Any("error", err))
	}

	if err := cleanenv.ReadConfig(".env", &cfg.ENV); err != nil {
		slog.Info("failed to read .env", slog.Any("error", err))
	}

	if err := cleanenv.ReadEnv(&cfg.ENV); err != nil {
		slog.Error("failed to read environment variables", slog.Any("error", err))
		os.Exit(1)
	}

	return cfg
}
