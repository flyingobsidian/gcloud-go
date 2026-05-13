package compute

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-golang-cli/internal/auth"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

const (
	ComputeScope    = "https://www.googleapis.com/auth/compute"
	ComputeROScope  = "https://www.googleapis.com/auth/compute.readonly"
	CloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"
)

// NewService creates an authenticated Compute Engine API client.
func NewService(ctx context.Context, account string) (*compute.Service, error) {
	ts, err := auth.TokenSource(ctx, account, ComputeScope, CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	svc, err := compute.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("creating compute service: %w", err)
	}
	return svc, nil
}
