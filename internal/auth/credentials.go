package auth

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/flyingobsidian/gcloud-go/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// TokenProvider resolves credentials and hands out token sources. Sources are
// cached per credential and scope set, so a token is fetched once and then
// reused by every service client until it expires. A CLI process creates one
// provider; a long-lived embedder keeps one so cached tokens also outlive a
// single invocation. It is safe for concurrent use.
//
// Cached credentials are not reloaded: after `gcloud auth login` replaces a
// stored credential, a long-lived embedder needs a new provider to see it.
type TokenProvider struct {
	env config.Env

	mu    sync.Mutex
	cache map[tokenKey]oauth2.TokenSource
}

// tokenKey identifies a cached token source: where the credential came from
// (a credentials file or a stored account) and the requested scopes.
type tokenKey struct {
	source string
	scopes string
}

// NewTokenProvider returns a TokenProvider that reads credentials and gcloud
// configuration from env.
func NewTokenProvider(env config.Env) *TokenProvider {
	return &TokenProvider{env: env, cache: map[tokenKey]oauth2.TokenSource{}}
}

// TokenSource returns an oauth2.TokenSource for the active or specified account.
// Resolution order:
//  1. GOOGLE_APPLICATION_CREDENTIALS env var
//  2. Stored credential for the specified account (flag)
//  3. Stored credential for the active account (from gcloud config)
//
// Supports both service_account (JWT) and authorized_user (refresh token) types.
//
// The returned source outlives ctx: tokens are refreshed with ctx's values but
// not its cancellation, because the source is cached and shared.
func (p *TokenProvider) TokenSource(ctx context.Context, account string, scopes ...string) (oauth2.TokenSource, error) {
	key, load, err := p.resolve(account)
	if err != nil {
		return nil, err
	}
	key.scopes = scopeKey(scopes)

	p.mu.Lock()
	defer p.mu.Unlock()
	if ts, ok := p.cache[key]; ok {
		return ts, nil
	}
	data, err := load()
	if err != nil {
		return nil, err
	}
	creds, err := google.CredentialsFromJSON(context.WithoutCancel(ctx), data, scopes...)
	if err != nil {
		if strings.HasPrefix(key.source, "account:") {
			return nil, fmt.Errorf("creating credentials for %s: %w", strings.TrimPrefix(key.source, "account:"), err)
		}
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}
	ts := oauth2.ReuseTokenSource(nil, creds.TokenSource)
	p.cache[key] = ts
	return ts, nil
}

// resolve works out which credential to use, without reading it, and returns
// its cache key and a function that loads its JSON.
func (p *TokenProvider) resolve(account string) (tokenKey, func() ([]byte, error), error) {
	if f := p.env.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); f != "" {
		return tokenKey{source: "file:" + f}, func() ([]byte, error) {
			data, err := os.ReadFile(f)
			if err != nil {
				return nil, fmt.Errorf("reading credentials file: %w", err)
			}
			return data, nil
		}, nil
	}

	// Resolve account from flag or gcloud config.
	if account == "" {
		props, err := p.env.Load()
		if err != nil {
			return tokenKey{}, nil, fmt.Errorf("loading config: %w", err)
		}
		account = props.Core.Account
	}
	if account == "" {
		return tokenKey{}, nil, fmt.Errorf("no active account; run 'gcloud auth login' or set GOOGLE_APPLICATION_CREDENTIALS")
	}
	return tokenKey{source: "account:" + account}, func() ([]byte, error) {
		store, err := NewStoreFor(p.env)
		if err != nil {
			return nil, err
		}
		data, err := store.Load(account)
		if err != nil {
			return nil, fmt.Errorf("loading credentials for %s: %w", account, err)
		}
		return data, nil
	}, nil
}

// scopeKey returns a canonical form of scopes, independent of order.
func scopeKey(scopes []string) string {
	s := slices.Clone(scopes)
	slices.Sort(s)
	return strings.Join(slices.Compact(s), " ")
}
