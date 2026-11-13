package cmd

import "github.com/spf13/cobra"

// --- gcloud lustre (#352) ---

func (s *lustreCmds) initLustre() {
	registerStubGroup(s.lustreCmd, "instances", "Manage Lustre instances", "create", "delete", "describe", "list", "update", "import-data", "export-data")
	registerStubGroup(s.lustreCmd, "operations", "Manage operations", "cancel", "describe", "list")

}

// lustreCmds holds the commands and flag values of one invocation's "lustre"
// command group, so invocations don't share state (#2070).
type lustreCmds struct {
	inv *invocation

	lustreCmd *cobra.Command
}

// newLustreCmd returns the "lustre" command group for inv.
func newLustreCmd(inv *invocation) *cobra.Command {
	return newLustreCmds(inv).lustreCmd
}

// newLustreCmds builds the "lustre" command group for inv and returns its state.
func newLustreCmds(inv *invocation) *lustreCmds {
	s := &lustreCmds{inv: inv}
	s.lustreCmd = &cobra.Command{Use: "lustre", Short: "Manage Lustre"}
	s.initLustre()
	return s
}
