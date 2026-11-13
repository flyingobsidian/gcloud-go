package cli

import "github.com/spf13/cobra"

// --- gcloud anthos (#295) ---

func (s *anthosCmds) initAnthos() {
	registerStubGroup(s.anthosCmd, "auth", "Authenticate clusters", "login", "token")
	registerStubGroup(s.anthosCmd, "config", "Anthos config", "controller")
	registerStubCommand(s.anthosCmd, "create-login-config", "Generate a login configuration file")

}

// anthosCmds holds the commands and flag values of one invocation's "anthos"
// command group, so invocations don't share state (#2070).
type anthosCmds struct {
	inv *invocation

	anthosCmd *cobra.Command
}

// newAnthosCmd returns the "anthos" command group for inv.
func newAnthosCmd(inv *invocation) *cobra.Command {
	return newAnthosCmds(inv).anthosCmd
}

// newAnthosCmds builds the "anthos" command group for inv and returns its state.
func newAnthosCmds(inv *invocation) *anthosCmds {
	s := &anthosCmds{inv: inv}
	s.anthosCmd = &cobra.Command{Use: "anthos", Short: "Anthos commands"}
	s.initAnthos()
	return s
}
