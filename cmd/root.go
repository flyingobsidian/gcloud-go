package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "dev"

var (
	flagProject               string
	flagZone                  string
	flagAccount               string
	flagQuiet                 bool
	flagAccessTokenFile       string
	flagBillingProject        string
	flagConfiguration         string
	flagFlagsFile             string
	flagFlatten               []string
	flagFormat                string
	flagImpersonateSA         string
	flagLogHTTP               bool
	flagTraceToken            string
	flagUserOutputEnabled     bool
	flagVerbosity             string
)

// IsInteractive returns true when stdin is a terminal (used for interactive prompts).
func IsInteractive() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

var rootCmd = &cobra.Command{
	Use:           "gcloud",
	Short:         "Lightweight gcloud CLI replacement",
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringVar(&flagProject, "project", "", "Google Cloud project ID")
	pf.StringVar(&flagZone, "zone", "", "Compute Engine zone")
	pf.StringVar(&flagAccount, "account", "", "Account to use for authentication")
	pf.BoolVarP(&flagQuiet, "quiet", "q", false, "Suppress all confirmation prompts")

	// Global flags mirroring gcloud-python (#530). Registered here so parsers
	// accept them uniformly; individual flags are wired to behavior where the
	// corresponding subsystem is implemented, and are otherwise accepted and
	// silently forwarded to subcommands that consult them.
	pf.StringVar(&flagAccessTokenFile, "access-token-file", "",
		"File containing an OAuth access token to use for this invocation")
	pf.StringVar(&flagBillingProject, "billing-project", "",
		"Cloud project to charge for quota/billing (overrides billing/quota_project)")
	pf.StringVar(&flagConfiguration, "configuration", "",
		"Named gcloud configuration to use for this invocation")
	pf.StringVar(&flagFlagsFile, "flags-file", "",
		"YAML/JSON file that supplies additional --flag=value pairs")
	pf.StringSliceVar(&flagFlatten, "flatten", nil,
		"Flatten name[] output slices in KEY into separate records")
	pf.StringVar(&flagFormat, "format", "",
		"Output format (config, csv, json, table, text, value, yaml, get, …)")
	pf.StringVar(&flagImpersonateSA, "impersonate-service-account", "",
		"Service account email(s) to impersonate for this invocation")
	pf.BoolVar(&flagLogHTTP, "log-http", false,
		"Log HTTP requests and responses")
	pf.StringVar(&flagTraceToken, "trace-token", "",
		"Trace token forwarded to Google APIs for this invocation")
	pf.BoolVar(&flagUserOutputEnabled, "user-output-enabled", true,
		"Print user-intended output; use --no-user-output-enabled to suppress")
	pf.StringVar(&flagVerbosity, "verbosity", "warning",
		"Log verbosity (debug, info, warning, error, critical, none)")

	rootCmd.SetVersionTemplate(versionTemplate())
}

func versionTemplate() string {
	return fmt.Sprintf("gcloud-go %s (Golang)\nGo %s %s/%s\n", Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

// Execute runs the root command and returns the resolved subcommand along
// with any error. Callers use the returned *cobra.Command to render the
// dotted-path error prefix ("gcloud.secrets.delete") that Python gcloud uses.
func Execute() (*cobra.Command, error) {
	return rootCmd.ExecuteC()
}
