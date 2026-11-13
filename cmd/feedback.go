package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// gcloud-go's issue tracker. gcloud-python points at its own; gcloud-go
// forwards feedback to the fork's repository.
const feedbackURL = "https://github.com/flyingobsidian/gcloud-go/issues/new"

func (s *feedbackCmds) initFeedback() {
	s.feedbackCmd.Flags().StringVar(&s.flagFeedbackLogFile, "log-file", "",
		"Log file to attach in the browser feedback form (accepted for gcloud-python parity; not currently transmitted)")

}

func (s *feedbackCmds) runFeedback(cmd *cobra.Command, args []string) error {
	fmt.Fprintln(s.inv.stdout, "File a gcloud-go issue at:")
	fmt.Fprintln(s.inv.stdout, "  "+feedbackURL)
	if s.flagFeedbackLogFile != "" {
		fmt.Fprintf(s.inv.stdout, "Log file to attach: %s\n", s.flagFeedbackLogFile)
	}
	// Best-effort browser open; ignore failure since not all environments
	// have a UI and gcloud-python also degrades gracefully.
	openBrowser(feedbackURL)
	return nil
}

// feedbackCmds holds the commands and flag values of one invocation's "feedback"
// command group, so invocations don't share state (#2070).
type feedbackCmds struct {
	inv *invocation

	feedbackCmd         *cobra.Command
	flagFeedbackLogFile string
}

// newFeedbackCmd returns the "feedback" command group for inv.
func newFeedbackCmd(inv *invocation) *cobra.Command {
	return newFeedbackCmds(inv).feedbackCmd
}

// newFeedbackCmds builds the "feedback" command group for inv and returns its state.
func newFeedbackCmds(inv *invocation) *feedbackCmds {
	s := &feedbackCmds{inv: inv}
	s.feedbackCmd = &cobra.Command{
		Use:   "feedback",
		Short: "Provide feedback to the gcloud-go team",
		Args:  cobra.NoArgs,
		RunE:  s.runFeedback,
	}
	s.initFeedback()
	return s
}
