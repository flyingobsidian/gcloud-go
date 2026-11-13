package cmd

import "github.com/spf13/cobra"

// --- gcloud compliance-manager (#317) ---

func (s *complianceManagerCmds) initComplianceManager() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(s.complianceManagerCmd, "cloud-control-deployments", "Manage cloud control deployments", crud...)
	registerStubGroup(s.complianceManagerCmd, "cloud-controls", "Manage cloud controls", crud...)
	registerStubGroup(s.complianceManagerCmd, "framework-deployments", "Manage framework deployments", crud...)
	registerStubGroup(s.complianceManagerCmd, "frameworks", "Manage frameworks", crud...)
	registerStubGroup(s.complianceManagerCmd, "operations", "Manage operations", "describe", "list")

}

// complianceManagerCmds holds the commands and flag values of one invocation's "complianceManager"
// command group, so invocations don't share state (#2070).
type complianceManagerCmds struct {
	inv *invocation

	complianceManagerCmd *cobra.Command
}

// newComplianceManagerCmd returns the "complianceManager" command group for inv.
func newComplianceManagerCmd(inv *invocation) *cobra.Command {
	return newComplianceManagerCmds(inv).complianceManagerCmd
}

// newComplianceManagerCmds builds the "complianceManager" command group for inv and returns its state.
func newComplianceManagerCmds(inv *invocation) *complianceManagerCmds {
	s := &complianceManagerCmds{inv: inv}
	s.complianceManagerCmd = &cobra.Command{Use: "compliance-manager", Short: "Manage Compliance Manager"}
	s.initComplianceManager()
	return s
}
