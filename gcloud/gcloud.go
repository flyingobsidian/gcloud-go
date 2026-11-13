// Package gcloud runs gcloud-go commands inside the calling program.
//
// A Runner is created once and used for many commands, including at the
// same time from several goroutines. Each command gets its own arguments,
// standard streams and context; credentials are shared, so an access token
// is fetched once and reused until it expires.
//
//	r := gcloud.NewRunner(gcloud.Options{})
//	var out bytes.Buffer
//	err := r.Run(ctx, gcloud.Invocation{
//		Args:   []string{"projects", "list", "--format=json"},
//		Stdout: &out,
//	})
//
// Commands run in-process never prompt and never change project settings
// (such as enabling an API) as a side effect.
//
// Every gcloud-go command runs in-process, and runs concurrently with others
// (https://github.com/flyingobsidian/gcloud-go/issues/2070). Run returns
// ErrNotInProcess only for commands gcloud-go does not have; callers can run
// another gcloud for those instead.
package gcloud

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"runtime/debug"

	"github.com/flyingobsidian/gcloud-go/cmd"
	"github.com/flyingobsidian/gcloud-go/internal/auth"
	"github.com/flyingobsidian/gcloud-go/internal/config"
)

const modulePath = "github.com/flyingobsidian/gcloud-go"

// ErrNotInProcess is returned by Runner.Run for commands that gcloud-go does
// not have.
var ErrNotInProcess = cmd.ErrNotInProcess

// CommandError is returned by Runner.Run when a command ran and failed.
// Error returns the message the CLI would print, e.g.
// "ERROR: (gcloud.projects.describe) NOT_FOUND: …"; errors.As reaches the
// underlying cause, such as a *googleapi.Error.
type CommandError = cmd.CommandError

// PanicError is returned by Runner.Run when a command panicked. Run recovers
// the panic so it doesn't take down the caller; Stack holds the trace.
type PanicError = cmd.PanicError

// Options configures a Runner. The zero value uses the process environment.
type Options struct {
	// Getenv returns the environment variables commands see, such as
	// CLOUDSDK_CONFIG or CLOUDSDK_CORE_PROJECT. nil means os.Getenv.
	Getenv func(key string) string
	// HomeDir returns the home directory, used to find the gcloud
	// configuration when CLOUDSDK_CONFIG is unset. nil means os.UserHomeDir.
	HomeDir func() (string, error)
}

// Runner runs gcloud-go commands in-process. It is safe for concurrent use.
//
// A Runner caches credentials: after `gcloud auth login` replaces stored
// credentials, create a new Runner to use them.
type Runner struct {
	env    config.Env
	tokens *auth.TokenProvider
}

// NewRunner returns a Runner configured by opts.
func NewRunner(opts Options) *Runner {
	env := config.OSEnv()
	if opts.Getenv != nil {
		env.Getenv = opts.Getenv
	}
	if opts.HomeDir != nil {
		env.HomeDir = opts.HomeDir
	}
	return &Runner{env: env, tokens: auth.NewTokenProvider(env)}
}

// Invocation is one command for Runner.Run.
type Invocation struct {
	// Args are the command-line arguments without the program name, e.g.
	// {"projects", "list", "--format=json"}.
	Args []string
	// Stdin is the command's input; nil means none.
	Stdin io.Reader
	// Stdout and Stderr receive the command's output; nil discards it.
	Stdout, Stderr io.Writer
}

// Run runs one command and returns when it has finished or ctx is done.
//
// It returns nil on success, ErrNotInProcess for commands that gcloud-go
// does not have, a *CommandError when the command failed, a *PanicError
// when it panicked, or another error when the arguments could not be parsed.
func (r *Runner) Run(ctx context.Context, inv Invocation) error {
	return cmd.RunInProcess(ctx, cmd.InProcess{
		Args:   inv.Args,
		Stdin:  inv.Stdin,
		Stdout: inv.Stdout,
		Stderr: inv.Stderr,
		Env:    r.env,
		Tokens: r.tokens,
	})
}

// ExitCode returns the process exit status the gcloud-go CLI would use for
// err, as returned by Run: 0 for nil, the command's own status for commands
// that pass through a subprocess's status (such as ssh), 2 for a panic as
// with any Go program, and 1 otherwise.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var panicErr *PanicError
	if errors.As(err, &panicErr) {
		return 2
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 1
}

// Version describes the gcloud-go build: the version stamped at build time,
// else the module version recorded by the Go toolchain, else "dev".
func Version() string {
	if cmd.Version != "" && cmd.Version != "dev" {
		return cmd.Version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Path == modulePath && info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
		for _, dep := range info.Deps {
			if dep.Path != modulePath {
				continue
			}
			if dep.Replace != nil {
				return "local " + dep.Replace.Path
			}
			return dep.Version
		}
	}
	return "dev"
}
