package cmd

import "github.com/spf13/cobra"

// --- gcloud api-gateway (#296) ---

var apiGatewayCmd = &cobra.Command{Use: "api-gateway", Short: "Manage API Gateway (stubbed)"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(apiGatewayCmd, "api-configs", "Manage API configs", append(crud, "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding")...)
	registerStubGroup(apiGatewayCmd, "apis", "Manage APIs", append(crud, "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding")...)
	registerStubGroup(apiGatewayCmd, "gateways", "Manage gateways", append(crud, "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding")...)
	registerStubGroup(apiGatewayCmd, "operations", "Manage operations", "cancel", "describe", "list", "wait")
	rootCmd.AddCommand(apiGatewayCmd)
}
