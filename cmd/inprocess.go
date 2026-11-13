package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"runtime/debug"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/auth"
	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ErrNotInProcess is returned by RunInProcess for commands in groups that
// still keep state in package globals (#2070). Run those with the gcloud-go
// binary instead.
var ErrNotInProcess = errors.New("command cannot run in-process yet")

// migratedGroups returns the command groups that keep all their state in
// inv, built for inv. Only these can run in-process.
func migratedGroups(inv *invocation) []*cobra.Command {
	return []*cobra.Command{
		newOrganizationsCmd(inv),
		newProjectsCmd(inv),
		newResourceManagerCmd(inv),
		newComputeCmd(inv),
		newLoggingCmd(inv),
		newSchedulerCmd(inv),
		newStorageCmd(inv),
		newWorkflowsCmd(inv),
		newTasksCmd(inv),
	}
}

// InProcess describes one command run by RunInProcess.
type InProcess struct {
	// Args are the command-line arguments, without the program name.
	Args []string
	// Stdin, Stdout and Stderr are the command's standard streams. nil
	// means no input and discarded output.
	Stdin          io.Reader
	Stdout, Stderr io.Writer
	// Env supplies environment variables and the home directory.
	Env config.Env
	// Tokens supplies credentials. Sharing one between runs reuses tokens.
	Tokens *auth.TokenProvider
}

// CommandError is the error from a command that RunInProcess ran and that
// failed. Error returns the text the CLI would print; Unwrap gives the
// underlying error (for example a *googleapi.Error).
type CommandError struct {
	// Command is the dotted path of the command, e.g. "gcloud.projects.list".
	Command string
	Err     error
	text    string
}

func (e *CommandError) Error() string { return e.text }
func (e *CommandError) Unwrap() error { return e.Err }

// PanicError is the error from a command that panicked while RunInProcess
// ran it. The panic is recovered so that it doesn't take down the caller
// (#2084). Panics in goroutines a command starts are not recovered.
type PanicError struct {
	// Value is the value passed to panic.
	Value any
	// Stack is the stack trace of the panicking goroutine.
	Stack []byte
}

func (e *PanicError) Error() string { return fmt.Sprintf("command panicked: %v", e.Value) }

// RunInProcess runs one gcloud-go command in the calling process. It is safe
// to call concurrently: each run has its own command tree, flags, streams
// and environment. It never prompts and never offers to enable APIs.
//
// It returns ErrNotInProcess for commands that cannot run in-process yet,
// a *CommandError if the command ran and failed, a *PanicError if the
// command panicked, or another error if the arguments could not be parsed.
func RunInProcess(ctx context.Context, p InProcess) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = &PanicError{Value: r, Stack: debug.Stack()}
		}
	}()
	inv := &invocation{
		stdin:  p.Stdin,
		stdout: p.Stdout,
		stderr: p.Stderr,
		env:    p.Env,
		tokens: p.Tokens,
	}
	if inv.stdin == nil {
		inv.stdin = strings.NewReader("")
	}
	if inv.stdout == nil {
		inv.stdout = io.Discard
	}
	if inv.stderr == nil {
		inv.stderr = io.Discard
	}

	root := newRootCmd(inv, false)
	args := p.Args
	if args == nil {
		args = []string{}
	}
	if found, rest, err := root.Find(args); found == root && (err != nil || hasPositional(root, rest)) {
		// No migrated group matched the command name.
		return ErrNotInProcess
	}
	root.SetArgs(args)
	inv.ctx = ctx
	executed, err := root.ExecuteContextC(ctx)
	if err != nil {
		return &CommandError{Command: commandDottedPath(executed), Err: err, text: inv.formatError(executed, err)}
	}
	return nil
}

// hasPositional reports whether args, parsed against c's local and
// persistent flags, contain a positional argument.
func hasPositional(c *cobra.Command, args []string) bool {
	flags := pflag.NewFlagSet(c.Name(), pflag.ContinueOnError)
	flags.AddFlagSet(c.LocalFlags())
	flags.AddFlagSet(c.PersistentFlags())
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--":
			return i+1 < len(args)
		case strings.HasPrefix(a, "--"):
			name := strings.TrimPrefix(a, "--")
			if strings.Contains(name, "=") {
				continue
			}
			if f := flags.Lookup(name); f != nil && f.NoOptDefVal == "" {
				i++ // the flag's value
			}
		case strings.HasPrefix(a, "-") && len(a) == 2:
			if f := flags.ShorthandLookup(a[1:]); f != nil && f.NoOptDefVal == "" {
				i++
			}
		case strings.HasPrefix(a, "-") && len(a) > 2:
			// -xVALUE or combined shorthands: no separate value.
		default:
			return true
		}
	}
	return false
}
