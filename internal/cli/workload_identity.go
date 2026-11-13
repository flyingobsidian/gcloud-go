package cli

import "github.com/spf13/cobra"

// --- gcloud workload-identity (#398) ---
//
// Note: gcloud-go already ships workload identity federation commands under
// `gcloud iam workload-identity-pools` (see #201 / #202). This file adds the
// standalone `gcloud workload-identity` surface called out in gcloud-python
// as a stub pending consolidation with the iam implementations.

func (s *workloadIdentityCmds) initWorkloadIdentity() {
	registerStubGroup(s.workloadIdentityCmd, "service-agents", "Manage Workload Identity service agents", "list", "describe")

}

// workloadIdentityCmds holds the commands and flag values of one invocation's "workloadIdentity"
// command group, so invocations don't share state (#2070).
type workloadIdentityCmds struct {
	inv *invocation

	workloadIdentityCmd *cobra.Command
}

// newWorkloadIdentityCmd returns the "workloadIdentity" command group for inv.
func newWorkloadIdentityCmd(inv *invocation) *cobra.Command {
	return newWorkloadIdentityCmds(inv).workloadIdentityCmd
}

// newWorkloadIdentityCmds builds the "workloadIdentity" command group for inv and returns its state.
func newWorkloadIdentityCmds(inv *invocation) *workloadIdentityCmds {
	s := &workloadIdentityCmds{inv: inv}
	s.workloadIdentityCmd = &cobra.Command{Use: "workload-identity", Short: "Manage Workload Identity"}
	s.initWorkloadIdentity()
	return s
}
