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

	"github.com/VasySS/cloudru-load-balancer/internal/backend"
	"github.com/VasySS/cloudru-load-balancer/internal/balancer"
	"github.com/VasySS/cloudru-load-balancer/internal/config"
	"github.com/VasySS/cloudru-load-balancer/internal/http/proxy"
	"github.com/VasySS/cloudru-load-balancer/internal/infrastructure/repository/postgres"
	"github.com/VasySS/cloudru-load-balancer/internal/ratelimit"
	"github.com/VasySS/cloudru-load-balancer/internal/ratelimit/leakybucket"
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

	rateLimiter := newRateLimiter(cfg, pgRepo, closer)

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
	backends, err := backend.NewBackendServers(cfg.YAML.Backends, cfg.YAML.Balancer.BackendsCheckInterval)
	if err != nil {
		return nil, fmt.Errorf("error creating backends array: %w", err)
	}

	// convert to balancer interface to pass the array
	balancerBackends := make([]balancer.BackendServer, 0, len(backends))
	for _, backend := range backends {
		balancerBackends = append(balancerBackends, backend)
	}

	var loadBalancer balancer.Balancer

	switch cfg.YAML.Balancer.Type {
	case config.LeastConnectionsType:
		slog.Info("using least connections algorithm for load balancing")

		loadBalancer = balancer.NewLeastConnections(balancerBackends)
	case config.RandomType:
		slog.Info("using random select for load balancing")

		loadBalancer = balancer.NewRandom(balancerBackends)
	case config.RoundRobinType:
		slog.Info("using round robin algorithm for load balancing")

		loadBalancer = balancer.NewRoundRobin(balancerBackends)
	}

	return loadBalancer, nil
}

//nolint:ireturn
func newRateLimiter(cfg config.Config, pgRepo *postgres.Repository, closer *Closer) ratelimit.Limiter {
	var rateLimiter ratelimit.Limiter

	switch cfg.YAML.RateLimit.Type {
	case config.TokenBucketType:
		slog.Info("using token bucket algorithm for rate limiting")

		tokenBucket := tokenbucket.NewUserBucket(pgRepo,
			cfg.YAML.RateLimit.Capacity,
			cfg.YAML.RateLimit.TokenRate,
			cfg.YAML.RateLimit.RefillInterval,
		)
		closer.Add(tokenBucket.Stop)

		rateLimiter = tokenBucket
	case config.LeakyBucketType:
		slog.Info("using leaky bucket algorithm for rate limiting")

		rateLimiter = leakybucket.NewUserBucket(pgRepo,
			cfg.YAML.RateLimit.Capacity,
			cfg.YAML.RateLimit.TokenRate,
		)
	}

	return rateLimiter
}
