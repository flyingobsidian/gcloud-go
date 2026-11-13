package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (s *surveyCmds) initSurvey() {

}

func (s *surveyCmds) runSurvey(cmd *cobra.Command, args []string) error {
	fmt.Fprintln(s.inv.stdout, "gcloud-go does not currently run a survey.")
	fmt.Fprintln(s.inv.stdout, "Please share feedback via:")
	fmt.Fprintln(s.inv.stdout, "  "+feedbackURL)
	return nil
}

// surveyCmds holds the commands and flag values of one invocation's "survey"
// command group, so invocations don't share state (#2070).
type surveyCmds struct {
	inv *invocation

	// gcloud-go does not run a satisfaction survey of its own; the command is
	// registered for CLI parity with gcloud-python (#535) and points at the
	// feedback tracker where users can share input.
	surveyCmd *cobra.Command
}

// newSurveyCmd returns the "survey" command group for inv.
func newSurveyCmd(inv *invocation) *cobra.Command {
	return newSurveyCmds(inv).surveyCmd
}

// newSurveyCmds builds the "survey" command group for inv and returns its state.
func newSurveyCmds(inv *invocation) *surveyCmds {
	s := &surveyCmds{inv: inv}
	s.surveyCmd = &cobra.Command{
		Use:   "survey",
		Short: "Invoke the customer satisfaction survey",
		Args:  cobra.NoArgs,
		RunE:  s.runSurvey,
	}
	s.initSurvey()
	return s
}
