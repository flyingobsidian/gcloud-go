package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// TokenSource returns an oauth2.TokenSource for the active or specified account.
// Resolution order:
//  1. GOOGLE_APPLICATION_CREDENTIALS env var
//  2. Stored credential for the specified account (flag)
//  3. Stored credential for the active account (from gcloud config)
//
// Supports both service_account (JWT) and authorized_user (refresh token) types.
func TokenSource(ctx context.Context, account string, scopes ...string) (oauth2.TokenSource, error) {
	// Try GOOGLE_APPLICATION_CREDENTIALS first.
	if f := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); f != "" {
		return tokenSourceFromFile(ctx, f, scopes)
	}

	// Resolve account from flag or gcloud config.
	if account == "" {
		props, err := config.Load()
		if err != nil {
			return nil, fmt.Errorf("loading config: %w", err)
		}
		account = props.Core.Account
	}
	if account == "" {
		return nil, fmt.Errorf("no active account; run 'gcloud auth login' or set GOOGLE_APPLICATION_CREDENTIALS")
	}

	store, err := NewStore()
	if err != nil {
		return nil, err
	}
	data, err := store.Load(account)
	if err != nil {
		return nil, fmt.Errorf("loading credentials for %s: %w", account, err)
	}

	// google.CredentialsFromJSON handles both service_account and authorized_user types.
	creds, err := google.CredentialsFromJSON(ctx, data, scopes...)
	if err != nil {
		return nil, fmt.Errorf("creating credentials for %s: %w", account, err)
	}
	return creds.TokenSource, nil
}

func tokenSourceFromFile(ctx context.Context, path string, scopes []string) (oauth2.TokenSource, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading credentials file: %w", err)
	}
	creds, err := google.CredentialsFromJSON(ctx, data, scopes...)
	if err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}
	return creds.TokenSource, nil
}
