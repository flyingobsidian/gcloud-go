package cli

import "github.com/spf13/cobra"

// --- gcloud org-policies (#368) ---
//
// Note: gcloud-go already exposes org-policies commands under
// `gcloud resource-manager org-policies` (see #281). This registers the
// standalone `gcloud org-policies` top-level command called out in the
// gcloud-python surface; the underlying operations are stubs pending
// consolidation with the resource-manager implementations.

func (s *orgPoliciesTopCmds) initOrgPolicies() {
	for _, name := range []string{
		"delete", "delete-custom-constraint", "describe", "describe-custom-constraint",
		"list", "list-custom-constraints", "reset", "set-custom-constraint", "set-policy",
	} {
		registerStubCommand(s.orgPoliciesTopCmd, name, "Not yet implemented (see also: resource-manager org-policies)")
	}

}

// orgPoliciesTopCmds holds the commands and flag values of one invocation's "orgPoliciesTop"
// command group, so invocations don't share state (#2070).
type orgPoliciesTopCmds struct {
	inv *invocation

	orgPoliciesTopCmd *cobra.Command
}

// newOrgPoliciesTopCmd returns the "orgPoliciesTop" command group for inv.
func newOrgPoliciesTopCmd(inv *invocation) *cobra.Command {
	return newOrgPoliciesTopCmds(inv).orgPoliciesTopCmd
}

// newOrgPoliciesTopCmds builds the "orgPoliciesTop" command group for inv and returns its state.
func newOrgPoliciesTopCmds(inv *invocation) *orgPoliciesTopCmds {
	s := &orgPoliciesTopCmds{inv: inv}
	s.orgPoliciesTopCmd = &cobra.Command{Use: "org-policies", Short: "Manage Organization Policies"}
	s.initOrgPolicies()
	return s
}
