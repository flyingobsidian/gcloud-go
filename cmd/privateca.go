package cmd

import "github.com/spf13/cobra"

// --- gcloud privateca (#374) ---

var privatecaCmd = &cobra.Command{Use: "privateca", Short: "Manage Private CA (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(privatecaCmd, "certificates", "Manage certificates", append(crud, "revoke", "export")...)
	registerStubGroup(privatecaCmd, "locations", "Manage locations", "list", "describe")
	registerStubGroup(privatecaCmd, "operations", "Manage operations", "cancel", "delete", "describe", "list")
	registerStubGroup(privatecaCmd, "pools", "Manage CA pools", append(crud, "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding")...)
	registerStubGroup(privatecaCmd, "roots", "Manage root CAs", append(crud, "activate", "disable", "enable", "undelete", "publish-crl", "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding")...)
	registerStubGroup(privatecaCmd, "subordinates", "Manage subordinate CAs", append(crud, "activate", "disable", "enable", "undelete", "publish-crl", "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding")...)
	registerStubGroup(privatecaCmd, "templates", "Manage certificate templates", append(crud, "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding")...)
	rootCmd.AddCommand(privatecaCmd)
}
