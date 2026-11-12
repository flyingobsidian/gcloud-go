package gcp

import (
	"context"
	"fmt"
	"io"

	"github.com/flyingobsidian/gcloud-go/internal/auth"
	"google.golang.org/api/option"
)

// Clients builds Google API service clients for one invocation: it fixes the
// account, where credentials come from and whether requests are logged, so
// that command code only names the service it wants. Clients holds no
// mutable state; concurrent invocations each use their own, typically sharing
// one TokenProvider.
type Clients struct {
	// Tokens supplies credentials. It must not be nil.
	Tokens *auth.TokenProvider
	// Account selects a stored credential; "" means the active account.
	Account string
	// DebugOut, if not nil, receives every outgoing request as a curl
	// command (see internal/httplog), as for --log-http.
	DebugOut io.Writer
}

// newService builds a generated API client with ctor, authenticated as c's
// account. extra options (such as a regional endpoint) are applied last.
func newService[T any](ctx context.Context, c *Clients, ctor func(context.Context, ...option.ClientOption) (T, error), extra ...option.ClientOption) (T, error) {
	var zero T
	ts, err := c.Tokens.TokenSource(ctx, c.Account, cloudPlatformScope)
	if err != nil {
		return zero, fmt.Errorf("obtaining credentials: %w", err)
	}
	opts, err := withRequestLogging(ctx, c.DebugOut, option.WithTokenSource(ts))
	if err != nil {
		return zero, err
	}
	return ctor(ctx, append(opts, extra...)...)
}
