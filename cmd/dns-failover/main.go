package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yangs1202/dns-failover/internal/config"
	"github.com/yangs1202/dns-failover/internal/health"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.LoadFromEnv()
	if err != nil {
		logger.Error("load config failed", "error", err)
		os.Exit(1)
	}

	checker := health.NewHTTPChecker(cfg.HealthTimeout)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info(
		"agent started",
		"region", cfg.RegionID,
		"check_interval", cfg.CheckInterval.String(),
		"health_timeout", cfg.HealthTimeout.String(),
		"etcd_endpoints", len(cfg.Etcd.Endpoints),
		"etcd_key_prefix", cfg.Etcd.KeyPrefix,
	)

	runHealthCheckCycle(ctx, logger, checker, cfg)

	ticker := time.NewTicker(cfg.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("agent stopped", "region", cfg.RegionID)
			return
		case <-ticker.C:
			runHealthCheckCycle(ctx, logger, checker, cfg)
		}
	}
}

func runHealthCheckCycle(ctx context.Context, logger *slog.Logger, checker health.HTTPChecker, cfg config.Config) {
	cycleCtx, cancel := context.WithTimeout(ctx, cfg.HealthTimeout*time.Duration(len(cfg.Endpoints)))
	defer cancel()

	for _, endpoint := range cfg.Endpoints {
		result := checker.Check(cycleCtx, endpoint)
		logger.Info(
			"health observation",
			"observer_region", cfg.RegionID,
			"target_region", endpoint.RegionID,
			"healthy", result.Healthy,
			"status_code", result.StatusCode,
			"latency_ms", result.Latency.Milliseconds(),
			"error", errorString(result.Err),
		)
	}

	logger.Info("health check cycle completed", "observer_region", cfg.RegionID)
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
