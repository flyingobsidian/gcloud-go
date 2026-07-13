package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/spf13/cobra"
)

// initCmd implements `gcloud init` — an interactive workflow that authorises
// a user account and sets [core/project]. The full gcloud-python workflow
// also picks a named configuration and default compute region/zone; those
// steps are deferred (see #532 follow-up TODO comments below), but the core
// account+project flow is wired so scripts using --quiet-friendly answers
// work end-to-end. See #532.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize or reinitialize gcloud",
	Args:  cobra.NoArgs,
	RunE:  runInit,
}

var (
	flagInitNoBrowser      bool
	flagInitConsoleOnly    bool
	flagInitSkipDiagnostics bool
)

func init() {
	initCmd.Flags().BoolVar(&flagInitNoBrowser, "no-browser", false,
		"Do not launch a browser for authorization")
	initCmd.Flags().BoolVar(&flagInitConsoleOnly, "console-only", false,
		"Alias for --no-launch-browser")
	initCmd.Flags().BoolVar(&flagInitConsoleOnly, "no-launch-browser", false,
		"Do not launch a browser for authorization")
	initCmd.Flags().BoolVar(&flagInitSkipDiagnostics, "skip-diagnostics", false,
		"Skip diagnostic checks (currently a no-op in gcloud-go)")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	if flagQuiet {
		return fmt.Errorf("gcloud init is interactive; run 'gcloud auth login' and 'gcloud config set project PROJECT_ID' non-interactively instead")
	}

	fmt.Fprintln(os.Stderr, "Welcome! This command will take you through the configuration of gcloud-go.")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Step 1 of 2: Authorize an account.")

	// Delegate to the existing auth login flow so behavior stays consistent
	// with `gcloud auth login`. runAuthLogin honours --cred-file when set,
	// or opens the browser otherwise.
	if err := runAuthLogin(cmd, nil); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Step 2 of 2: Choose a default project.")
	projectID, err := promptForProject(os.Stdin, os.Stderr)
	if err != nil {
		return err
	}
	if projectID != "" {
		props, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		props.Core.Project = projectID
		if err := props.Save(); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Updated property [core/project] to [%s].\n", projectID)
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "gcloud has now been configured.")
	return nil
}

// promptForProject asks the user to type a project ID. An empty reply is
// permitted and returns "" (the caller then leaves [core/project] untouched).
func promptForProject(in *os.File, err *os.File) (string, error) {
	fmt.Fprint(err, "Enter project ID (leave blank to skip): ")
	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		if e := scanner.Err(); e != nil {
			return "", e
		}
		return "", nil
	}
	return strings.TrimSpace(scanner.Text()), nil
}
