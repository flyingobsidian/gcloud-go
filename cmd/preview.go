package cmd

import "github.com/spf13/cobra"

// --- gcloud preview (#373) ---
//
// Like `beta`, `preview` in gcloud-python is a release-stage alias tree.
// gcloud-go does not track release stages separately.

func (s *previewCmds) initPreview() {

}

// previewCmds holds the commands and flag values of one invocation's "preview"
// command group, so invocations don't share state (#2070).
type previewCmds struct {
	inv *invocation

	previewCmd *cobra.Command
}

// newPreviewCmd returns the "preview" command group for inv.
func newPreviewCmd(inv *invocation) *cobra.Command {
	return newPreviewCmds(inv).previewCmd
}

// newPreviewCmds builds the "preview" command group for inv and returns its state.
func newPreviewCmds(inv *invocation) *previewCmds {
	s := &previewCmds{inv: inv}
	s.previewCmd = &cobra.Command{
		Use:   "preview",
		Short: "PREVIEW aliases for the top-level command tree",
		Long:  "gcloud-go does not track release stages separately. Use the corresponding top-level command instead.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	s.initPreview()
	return s
}
