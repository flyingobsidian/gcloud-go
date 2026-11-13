package cmd

import "github.com/spf13/cobra"

// --- gcloud service-health (#384) ---

func (s *serviceHealthCmds) initServiceHealth() {
	registerStubGroup(s.serviceHealthCmd, "events", "Manage events", "describe", "list")
	registerStubGroup(s.serviceHealthCmd, "organization-events", "Manage organization events", "describe", "list")
	registerStubGroup(s.serviceHealthCmd, "organization-impacts", "Manage organization impacts", "describe", "list")

}

// serviceHealthCmds holds the commands and flag values of one invocation's "serviceHealth"
// command group, so invocations don't share state (#2070).
type serviceHealthCmds struct {
	inv *invocation

	serviceHealthCmd *cobra.Command
}

// newServiceHealthCmd returns the "serviceHealth" command group for inv.
func newServiceHealthCmd(inv *invocation) *cobra.Command {
	return newServiceHealthCmds(inv).serviceHealthCmd
}

// newServiceHealthCmds builds the "serviceHealth" command group for inv and returns its state.
func newServiceHealthCmds(inv *invocation) *serviceHealthCmds {
	s := &serviceHealthCmds{inv: inv}
	s.serviceHealthCmd = &cobra.Command{Use: "service-health", Short: "Manage Service Health"}
	s.initServiceHealth()
	return s
}
