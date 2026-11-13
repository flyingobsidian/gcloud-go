package cmd

import "github.com/spf13/cobra"

// --- gcloud alpha (#294) ---
//
// Like `beta` and `preview`, `alpha` in gcloud-python is a release-stage alias
// tree. gcloud-go does not track release stages separately.

func (s *alphaCmds) initAlpha() {

}

// alphaCmds holds the commands and flag values of one invocation's "alpha"
// command group, so invocations don't share state (#2070).
type alphaCmds struct {
	inv *invocation

	alphaCmd *cobra.Command
}

// newAlphaCmd returns the "alpha" command group for inv.
func newAlphaCmd(inv *invocation) *cobra.Command {
	return newAlphaCmds(inv).alphaCmd
}

// newAlphaCmds builds the "alpha" command group for inv and returns its state.
func newAlphaCmds(inv *invocation) *alphaCmds {
	s := &alphaCmds{inv: inv}
	s.alphaCmd = &cobra.Command{
		Use:   "alpha",
		Short: "ALPHA aliases for the top-level command tree",
		Long:  "gcloud-go does not track release stages separately. Use the corresponding top-level command instead.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	s.initAlpha()
	return s
}
