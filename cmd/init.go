package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

func (s *initCmds) initInit() {
	s.initCmd.Flags().BoolVar(&s.flagInitNoBrowser, "no-browser", false,
		"Do not launch a browser for authorization")
	s.initCmd.Flags().BoolVar(&s.flagInitConsoleOnly, "console-only", false,
		"Alias for --no-launch-browser")
	s.initCmd.Flags().BoolVar(&s.flagInitConsoleOnly, "no-launch-browser", false,
		"Do not launch a browser for authorization")
	s.initCmd.Flags().BoolVar(&s.flagInitSkipDiagnostics, "skip-diagnostics", false,
		"Skip diagnostic checks (currently a no-op in gcloud-go)")

}

func (s *initCmds) runInit(cmd *cobra.Command, args []string) error {
	if s.inv.flags.quiet {
		return fmt.Errorf("gcloud init is interactive; run 'gcloud auth login' and 'gcloud config set project PROJECT_ID' non-interactively instead")
	}

	fmt.Fprintln(s.inv.stderr, "Welcome! This command will take you through the configuration of gcloud-go.")
	fmt.Fprintln(s.inv.stderr)
	fmt.Fprintln(s.inv.stderr, "Step 1 of 2: Authorize an account.")

	// Delegate to the existing auth login flow so behavior stays consistent
	// with `gcloud auth login`. runAuthLogin honours --cred-file when set,
	// or opens the browser otherwise.
	if err := newAuthCmds(s.inv).runAuthLogin(cmd, nil); err != nil {
		return err
	}

	fmt.Fprintln(s.inv.stderr)
	fmt.Fprintln(s.inv.stderr, "Step 2 of 2: Choose a default project.")
	projectID, err := promptForProject(s.inv.stdin, s.inv.stderr)
	if err != nil {
		return err
	}
	if projectID != "" {
		props, err := s.inv.properties()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		props.Core.Project = projectID
		if err := s.inv.env.Save(props); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Fprintf(s.inv.stderr, "Updated property [core/project] to [%s].\n", projectID)
	}

	fmt.Fprintln(s.inv.stderr)
	fmt.Fprintln(s.inv.stderr, "gcloud has now been configured.")
	return nil
}

// promptForProject asks the user to type a project ID. An empty reply is
// permitted and returns "" (the caller then leaves [core/project] untouched).
func promptForProject(in io.Reader, err io.Writer) (string, error) {
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

// initCmds holds the commands and flag values of one invocation's "init"
// command group, so invocations don't share state (#2070).
type initCmds struct {
	inv *invocation

	// initCmd implements `gcloud init` — an interactive workflow that authorises
	// a user account and sets [core/project]. The full gcloud-python workflow
	// also picks a named configuration and default compute region/zone; those
	// steps are deferred (see #532 follow-up TODO comments below), but the core
	// account+project flow is wired so scripts using --quiet-friendly answers
	// work end-to-end. See #532.
	initCmd                 *cobra.Command
	flagInitNoBrowser       bool
	flagInitConsoleOnly     bool
	flagInitSkipDiagnostics bool
}

// newInitCmd returns the "init" command group for inv.
func newInitCmd(inv *invocation) *cobra.Command {
	return newInitCmds(inv).initCmd
}

// newInitCmds builds the "init" command group for inv and returns its state.
func newInitCmds(inv *invocation) *initCmds {
	s := &initCmds{inv: inv}
	s.initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize or reinitialize gcloud",
		Args:  cobra.NoArgs,
		RunE:  s.runInit,
	}
	s.initInit()
	return s
}
