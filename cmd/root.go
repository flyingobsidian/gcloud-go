package cmd

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "dev" //nolint:gochecknoglobals // set by the linker (-X), never at run time

// newRootCmd returns the root command for inv, with the global flags bound
// to inv.flags and the command groups (see migratedGroups).
func newRootCmd(inv *invocation) *cobra.Command {
	root := &cobra.Command{
		Use:           "gcloud",
		Short:         "Lightweight gcloud CLI replacement",
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetIn(inv.stdin)
	root.SetOut(inv.stdout)
	root.SetErr(inv.stderr)

	f := &inv.flags
	pf := root.PersistentFlags()
	pf.StringVar(&f.project, "project", "", "Google Cloud project ID")
	pf.StringVar(&f.zone, "zone", "", "Compute Engine zone")
	pf.StringVar(&f.account, "account", "", "Account to use for authentication")
	pf.BoolVarP(&f.quiet, "quiet", "q", false, "Suppress all confirmation prompts")

	// Global flags mirroring gcloud-python (#530). Registered here so parsers
	// accept them uniformly; individual flags are wired to behavior where the
	// corresponding subsystem is implemented, and are otherwise accepted and
	// silently forwarded to subcommands that consult them.
	pf.StringVar(&f.accessTokenFile, "access-token-file", "",
		"File containing an OAuth access token to use for this invocation")
	pf.StringVar(&f.billingProject, "billing-project", "",
		"Cloud project to charge for quota/billing (overrides billing/quota_project)")
	pf.StringVar(&f.configuration, "configuration", "",
		"Named gcloud configuration to use for this invocation")
	pf.StringVar(&f.flagsFile, "flags-file", "",
		"YAML/JSON file that supplies additional --flag=value pairs")
	pf.StringVar(&f.filter, "filter", "",
		"Client-side filter expression (see 'gcloud topic filters')")
	pf.StringSliceVar(&f.flatten, "flatten", nil,
		"Flatten name[] output slices in KEY into separate records")
	pf.StringVar(&f.format, "format", "",
		"Output format (config, csv, json, table, text, value, yaml, get, …)")
	pf.StringVar(&f.impersonateSA, "impersonate-service-account", "",
		"Service account email(s) to impersonate for this invocation")
	pf.BoolVar(&f.logHTTP, "log-http", false,
		"Log HTTP requests and responses")
	pf.StringVar(&f.traceToken, "trace-token", "",
		"Trace token forwarded to Google APIs for this invocation")
	pf.BoolVar(&f.userOutputEnabled, "user-output-enabled", true,
		"Print user-intended output; use --no-user-output-enabled to suppress")
	pf.StringVar(&f.verbosity, "verbosity", "warning",
		"Log verbosity (debug, info, warning, error, critical, none)")

	root.SetVersionTemplate(versionTemplate())

	root.AddCommand(migratedGroups(inv)...)
	return root
}

func versionTemplate() string {
	return fmt.Sprintf("gcloud-go %s (Golang)\nGo %s %s/%s\n", Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

// Main runs the gcloud-go CLI with args, using the process's standard
// streams and environment, and returns the exit status. It runs a single
// invocation and must be called at most once per process.
//
// If the command fails because an API is not enabled and the user agrees to
// enable it (see invocation.offerEnableAPI), the command is run once more.
func Main(ctx context.Context, args []string) int {
	inv := newOSInvocation()
	root := newRootCmd(inv)
	root.SetArgs(args)
	inv.ctx = ctx
	executed, err := root.ExecuteContextC(ctx)
	if err != nil {
		if inv.offerEnableAPI(ctx, err) {
			executed, err = root.ExecuteContextC(ctx)
		}
	}
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	fmt.Fprintln(inv.stderr, inv.formatError(executed, err))
	return 1
}
