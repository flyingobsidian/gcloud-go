package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "dev"

// Values of the global flags, for command groups not yet migrated to
// invocation (#2070). See globalFlags.setLegacyGlobals.
var (
	flagProject           string
	flagZone              string
	flagAccount           string
	flagQuiet             bool
	flagAccessTokenFile   string
	flagBillingProject    string
	flagConfiguration     string
	flagFlagsFile         string
	flagFilter            string
	flagFlatten           []string
	flagFormat            string
	flagImpersonateSA     string
	flagLogHTTP           bool
	flagTraceToken        string
	flagUserOutputEnabled = true
	flagVerbosity         = "warning"
)

// IsInteractive returns true when stdin is a terminal (used for interactive prompts).
func IsInteractive() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

// rootCmd collects the command groups that have not been migrated to
// invocation yet (#2070): their init functions add them here and
// newRootCmd adopts them into the CLI's root. It is never executed itself.
var rootCmd = &cobra.Command{Use: "gcloud"}

// newRootCmd returns the root command for inv, with the global flags bound
// to inv.flags. With legacy set it also adopts the unmigrated command groups
// registered on rootCmd and copies the global flags into the package
// variables they read; that is only safe when the process runs a single
// invocation, as the CLI does.
func newRootCmd(inv *invocation, legacy bool) *cobra.Command {
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

	if legacy {
		root.PersistentPreRun = func(*cobra.Command, []string) { f.setLegacyGlobals() }
		for _, c := range rootCmd.Commands() {
			root.AddCommand(c)
		}
	}
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
	root := newRootCmd(inv, true)
	root.SetArgs(args)
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

// Execute runs gcloud-go with the process's arguments, standard streams and
// environment, and returns the resolved subcommand along with any error.
// Callers use the returned *cobra.Command to render the dotted-path error
// prefix ("gcloud.secrets.delete") that Python gcloud uses. Like Main it must
// be called at most once per process; prefer Main.
func Execute() (*cobra.Command, error) {
	return newRootCmd(newOSInvocation(), true).ExecuteContextC(context.Background())
}
