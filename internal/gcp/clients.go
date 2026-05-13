package gcp

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-golang-cli/internal/auth"
	cloudscheduler "google.golang.org/api/cloudscheduler/v1"
	dataflow "google.golang.org/api/dataflow/v1b3"
	dataplex "google.golang.org/api/dataplex/v1"
	monitoring "google.golang.org/api/monitoring/v3"
	ondemandscanning "google.golang.org/api/ondemandscanning/v1"
	"google.golang.org/api/option"
	redis "google.golang.org/api/redis/v1"
	storage "google.golang.org/api/storage/v1"
)

const cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

func SchedulerService(ctx context.Context, account string) (*cloudscheduler.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return cloudscheduler.NewService(ctx, option.WithTokenSource(ts))
}

func DataflowService(ctx context.Context, account string) (*dataflow.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return dataflow.NewService(ctx, option.WithTokenSource(ts))
}

func StorageService(ctx context.Context, account string) (*storage.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return storage.NewService(ctx, option.WithTokenSource(ts))
}

func MonitoringService(ctx context.Context, account string) (*monitoring.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return monitoring.NewService(ctx, option.WithTokenSource(ts))
}

func RedisService(ctx context.Context, account string) (*redis.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return redis.NewService(ctx, option.WithTokenSource(ts))
}

func DataplexService(ctx context.Context, account string) (*dataplex.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return dataplex.NewService(ctx, option.WithTokenSource(ts))
}

func OnDemandScanningService(ctx context.Context, account string) (*ondemandscanning.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return ondemandscanning.NewService(ctx, option.WithTokenSource(ts))
}
