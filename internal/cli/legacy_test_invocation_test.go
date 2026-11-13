package cli

import (
	"context"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/auth"
	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

// legacyTestInvocation returns an invocation that behaves like the process
// state that tests written before a group's migration (#2070) set up: it
// uses whatever os.Stdin, os.Stdout and os.Stderr are when it reads or
// writes (so captureStdout works), the process environment, and the REST
// test hooks SetRestTokenSourceForTest and SetRestUserProjectForTest.
//
// New tests should use newTestInvocation or newFakeAPIInvocation instead.
func legacyTestInvocation() *invocation {
	env := config.OSEnv()
	inv := &invocation{
		stdin:  osStdin{},
		stdout: osStdout{},
		stderr: osStderr{},
		env:    env,
		tokens: auth.NewTokenProvider(env),
		flags:  globalFlags{userOutputEnabled: true, verbosity: "warning"},
		restTokenSource: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
			return restTokenSource(ctx, scopes...)
		},
	}
	inv.restUserProject = func() string {
		if restUserProject != nil {
			return restUserProject()
		}
		return inv.billingProject()
	}
	return inv
}

// restTokenSource and restUserProject supply the credentials and billing
// project of legacyTestInvocation's REST clients; a nil restUserProject means
// the invocation's own billing project. Tests override them with
// SetRestTokenSourceForTest and SetRestUserProjectForTest, so those tests
// must not run in parallel.
var (
	restTokenSource = func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
		return auth.NewTokenProvider(config.OSEnv()).TokenSource(ctx, "", scopes...)
	}
	restUserProject func() string
)

// SetRestTokenSourceForTest makes legacyTestInvocation's REST clients use
// ts. The returned function restores the previous token source.
func SetRestTokenSourceForTest(ts oauth2.TokenSource) (restore func()) {
	prev := restTokenSource
	restTokenSource = func(context.Context, ...string) (oauth2.TokenSource, error) {
		return ts, nil
	}
	return func() { restTokenSource = prev }
}

// SetRestUserProjectForTest makes legacyTestInvocation's REST clients send
// project as the billing project. The returned function restores the
// previous one.
func SetRestUserProjectForTest(project string) (restore func()) {
	prev := restUserProject
	restUserProject = func() string { return project }
	return func() { restUserProject = prev }
}

type osStdin struct{}

func (osStdin) Read(p []byte) (int, error) { return os.Stdin.Read(p) }

type osStdout struct{}

func (osStdout) Write(p []byte) (int, error) { return os.Stdout.Write(p) }

type osStderr struct{}

func (osStderr) Write(p []byte) (int, error) { return os.Stderr.Write(p) }

// setContexts gives c and its subcommands a background context, for tests
// that call run functions directly instead of executing the command.
func setContexts(c *cobra.Command) {
	c.SetContext(context.Background())
	for _, sub := range c.Commands() {
		setContexts(sub)
	}
}
