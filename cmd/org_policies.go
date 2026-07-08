package cmd

import "github.com/spf13/cobra"

// --- gcloud org-policies (#368) ---
//
// Note: gcloud-go already exposes org-policies commands under
// `gcloud resource-manager org-policies` (see #281). This registers the
// standalone `gcloud org-policies` top-level command called out in the
// gcloud-python surface; the underlying operations are stubs pending
// consolidation with the resource-manager implementations.

var orgPoliciesTopCmd = &cobra.Command{Use: "org-policies", Short: "Manage Organization Policies (stubbed)"}

func init() {
	for _, name := range []string{
		"delete", "delete-custom-constraint", "describe", "describe-custom-constraint",
		"list", "list-custom-constraints", "reset", "set-custom-constraint", "set-policy",
	} {
		registerStubCommand(orgPoliciesTopCmd, name, "Not yet implemented (see also: resource-manager org-policies)")
	}
	rootCmd.AddCommand(orgPoliciesTopCmd)
}
