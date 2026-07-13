package cmd

import "github.com/spf13/cobra"

// --- gcloud identity (#346) ---

var identityCmd = &cobra.Command{Use: "identity", Short: "Manage Cloud Identity"}

func init() {
	groups := &cobra.Command{Use: "groups", Short: "Manage Cloud Identity Groups"}
	for _, n := range []string{"create", "delete", "describe", "list", "update", "search", "get-iam-policy", "set-iam-policy"} {
		registerStubCommand(groups, n, "Not yet implemented")
	}
	registerStubGroup(groups, "memberships", "Manage group memberships", "create", "delete", "describe", "list", "update", "search-transitive-memberships", "search-transitive-groups", "modify-membership-roles")
	registerStubGroup(groups, "preview", "Preview commands", "list")
	identityCmd.AddCommand(groups)
	rootCmd.AddCommand(identityCmd)
}
