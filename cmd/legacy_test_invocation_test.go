package cmd

import (
	"context"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/auth"
	"github.com/flyingobsidian/gcloud-go/internal/config"
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
	return &invocation{
		stdin:  osStdin{},
		stdout: osStdout{},
		stderr: osStderr{},
		env:    env,
		tokens: auth.NewTokenProvider(env),
		flags:  globalFlags{userOutputEnabled: true, verbosity: "warning"},
		restTokenSource: func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
			return restTokenSource(ctx, scopes...)
		},
		restUserProject: func() string { return restUserProject() },
	}
}

type osStdin struct{}

func (osStdin) Read(p []byte) (int, error) { return os.Stdin.Read(p) }

type osStdout struct{}

func (osStdout) Write(p []byte) (int, error) { return os.Stdout.Write(p) }

type osStderr struct{}

func (osStderr) Write(p []byte) (int, error) { return os.Stderr.Write(p) }
