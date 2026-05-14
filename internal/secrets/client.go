package secrets

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/auth"
	secretmanager "google.golang.org/api/secretmanager/v1"
	"google.golang.org/api/option"
)

const cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

// NewService creates an authenticated Secret Manager API client.
func NewService(ctx context.Context, account string) (*secretmanager.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	svc, err := secretmanager.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("creating secret manager service: %w", err)
	}
	return svc, nil
}

// SecretParent returns the parent resource for secrets in a project.
func SecretParent(project, location string) string {
	if location != "" {
		return fmt.Sprintf("projects/%s/locations/%s", project, location)
	}
	return fmt.Sprintf("projects/%s", project)
}

// SecretName returns the full resource name for a secret.
func SecretName(project, secretID, location string) string {
	if location != "" {
		return fmt.Sprintf("projects/%s/locations/%s/secrets/%s", project, location, secretID)
	}
	return fmt.Sprintf("projects/%s/secrets/%s", project, secretID)
}

// VersionName returns the full resource name for a secret version.
func VersionName(project, secretID, version, location string) string {
	if location != "" {
		return fmt.Sprintf("projects/%s/locations/%s/secrets/%s/versions/%s", project, location, secretID, version)
	}
	return fmt.Sprintf("projects/%s/secrets/%s/versions/%s", project, secretID, version)
}
