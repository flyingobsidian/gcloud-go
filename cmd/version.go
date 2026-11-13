package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

func (s *versionCmds) initVersion() {

}

func (s *versionCmds) runVersion(cmd *cobra.Command, args []string) error {
	fmt.Fprintf(s.inv.stdout, "gcloud-go %s\n", Version)
	fmt.Fprintf(s.inv.stdout, "go %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	return nil
}

// versionCmds holds the commands and flag values of one invocation's "version"
// command group, so invocations don't share state (#2070).
type versionCmds struct {
	inv *invocation

	// versionCmd implements `gcloud version`, which prints version information
	// for the CLI itself. gcloud-python prints per-component versions
	// ("bq 2.1.33", "gsutil 5.37", …); gcloud-go is a single binary so there is
	// only one line, but the top-level command exists for CLI-parity with
	// gcloud-python and to match `--version` output. See #531.
	versionCmd *cobra.Command
}

// newVersionCmd returns the "version" command group for inv.
func newVersionCmd(inv *invocation) *cobra.Command {
	return newVersionCmds(inv).versionCmd
}

// newVersionCmds builds the "version" command group for inv and returns its state.
func newVersionCmds(inv *invocation) *versionCmds {
	s := &versionCmds{inv: inv}
	s.versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information for gcloud-go",
		Args:  cobra.NoArgs,
		RunE:  s.runVersion,
	}
	s.initVersion()
	return s
}
