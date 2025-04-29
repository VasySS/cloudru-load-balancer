// Package app contains logic for creating services and starting the application.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/VasySS/cloudru-load-balancer/internal/balancer"
	"github.com/VasySS/cloudru-load-balancer/internal/config"
	"github.com/VasySS/cloudru-load-balancer/internal/http/proxy"
	"github.com/VasySS/cloudru-load-balancer/internal/infrastructure/repository/postgres"
	"github.com/VasySS/cloudru-load-balancer/internal/ratelimit"
	"github.com/VasySS/cloudru-load-balancer/internal/ratelimit/tokenbucket"
)

// Run starts the application.
func Run(ctx context.Context, cfg config.Config) error {
	closer := NewCloser()

	pgConnectionURL := fmt.Sprintf(
		"postgres://%s:%s@%s/%s",
		cfg.ENV.Postgres.User,
		cfg.ENV.Postgres.Password,
		cfg.ENV.Postgres.Host,
		cfg.ENV.Postgres.Database,
	)

	pgRepo, err := newPostgresRepo(ctx, closer, pgConnectionURL)
	if err != nil {
		return err
	}

	rateLimiter := newRateLimiter(cfg, pgRepo)

	loadBalancer, err := newLoadBalancer(cfg)
	if err != nil {
		return err
	}

	r := proxy.New(rateLimiter, loadBalancer)

	go startHTTP(closer, r, cfg.ENV.Port)

	<-ctx.Done()
	slog.Info("gracefully shutting down...")

	closeCtx, stop := context.WithCancel(context.Background())
	defer stop()

	//nolint:contextcheck
	if err := closer.Close(closeCtx); err != nil {
		return err
	}

	return nil
}

func startHTTP(closer *Closer, r http.Handler, port string) {
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
		IdleTimeout:  time.Second * 10,
	}

	slog.Info("starting http server", slog.String("addr", srv.Addr))
	closer.AddWithCtx(srv.Shutdown)

	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start http server")
		os.Exit(1)
	}
}

func newPostgresRepo(ctx context.Context, closer *Closer, connectionURL string) (*postgres.Repository, error) {
	pool, err := pgxpool.New(ctx, connectionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	slog.Info("postgres connected")
	closer.Add(pool.Close)

	txManager := postgres.NewTxManager(pool)
	pgRepo := postgres.New(txManager)

	return pgRepo, nil
}

//nolint:ireturn
func newLoadBalancer(cfg config.Config) (balancer.Balancer, error) {
	backends, err := balancer.BackendServersFromArray(cfg.YAML.Backends)
	if err != nil {
		return nil, fmt.Errorf("error creating backends array: %w", err)
	}

	var loadBalancer balancer.Balancer

	switch cfg.YAML.Balancer.Type {
	case config.LeastConnectionsType:
		loadBalancer = balancer.NewLeastConnections(backends)
	case config.RandomType:
		loadBalancer = balancer.NewRandom(backends)
	case config.RoundRobinType:
		loadBalancer = balancer.NewRoundRobin(backends)
	}

	return loadBalancer, nil
}

//nolint:ireturn
func newRateLimiter(cfg config.Config, pgRepo *postgres.Repository) ratelimit.Limiter {
	var rateLimiter ratelimit.Limiter

	switch cfg.YAML.RateLimit.Type {
	case config.TokenBucketType:
		rateLimiter = tokenbucket.New(pgRepo)
	case config.LeakyBucketType:
		rateLimiter = tokenbucket.New(pgRepo)
	}

	return rateLimiter
}
