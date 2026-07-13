package cmd

import "github.com/spf13/cobra"

// --- gcloud kms (#349) ---

var kmsCmd = &cobra.Command{Use: "kms", Short: "Manage Cloud KMS"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(kmsCmd, "autokey-config", "Manage AutokeyConfig", "describe", "update")
	registerStubGroup(kmsCmd, "ekm-config", "Manage EkmConfig", "describe", "update")
	registerStubGroup(kmsCmd, "ekm-connections", "Manage EKM connections", append(crud, "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding", "test-iam-permissions", "verify-connectivity")...)
	registerStubGroup(kmsCmd, "import-jobs", "Manage import jobs", append(crud, "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding")...)
	registerStubGroup(kmsCmd, "inventory", "KMS Inventory & Key Tracking", "list", "list-protected-resources")
	registerStubGroup(kmsCmd, "key-handles", "Manage KeyHandle resources", "create", "delete", "describe", "list")
	registerStubGroup(kmsCmd, "keyrings", "Manage keyrings", append(crud, "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding")...)
	registerStubGroup(kmsCmd, "keys", "Manage keys", append(crud, "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding", "versions", "add-rotation-schedule", "remove-rotation-schedule", "set-primary-version", "restore")...)
	registerStubGroup(kmsCmd, "locations", "View locations", "describe", "list")
	registerStubGroup(kmsCmd, "operations", "Manage operations", "describe", "list", "cancel")
	registerStubGroup(kmsCmd, "retired-resources", "Manage retired resources", "list")
	registerStubGroup(kmsCmd, "single-tenant-hsm", "Manage single-tenant HSM", crud...)
	for _, name := range []string{
		"asymmetric-decrypt", "asymmetric-sign", "decapsulate", "decrypt", "encrypt",
		"mac-sign", "mac-verify", "raw-decrypt", "raw-encrypt",
	} {
		registerStubCommand(kmsCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(kmsCmd)
}
