package cli

import (
	"github.com/spf13/cobra"
)

// --- gcloud beta (#305) ---
//
// In gcloud-python, `gcloud beta X` mirrors `gcloud X` but exposes the BETA
// tracks of each surface. gcloud-go does not track release stages
// separately: every implemented command is exposed at its GA path only.
// This stub registers the `beta` group so `gcloud beta ...` invocations are
// recognized and redirected to the GA tree.

func (s *betaCmds) initBeta() {

}

// betaCmds holds the commands and flag values of one invocation's "beta"
// command group, so invocations don't share state (#2070).
type betaCmds struct {
	inv *invocation

	betaCmd *cobra.Command
}

// newBetaCmd returns the "beta" command group for inv.
func newBetaCmd(inv *invocation) *cobra.Command {
	return newBetaCmds(inv).betaCmd
}

// newBetaCmds builds the "beta" command group for inv and returns its state.
func newBetaCmds(inv *invocation) *betaCmds {
	s := &betaCmds{inv: inv}
	s.betaCmd = &cobra.Command{
		Use:   "beta",
		Short: "BETA aliases for the top-level command tree",
		Long: `gcloud-go does not currently split commands by release stage. Use the
corresponding top-level command instead (for example, ` + "`gcloud-go compute ...`" +
			` in place of ` + "`gcloud-go beta compute ...`" + `).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	s.initBeta()
	return s
}
