package cmd

import "github.com/spf13/cobra"

// --- gcloud iap (#345) ---
//
// Note: gcloud-go already ships an internal iap package under internal/iap for
// SSH tunneling. This surface registration exposes the gcloud-python iap
// subcommand set as stubs so callers can discover them.

var iapCmd = &cobra.Command{Use: "iap", Short: "Manage IAP"}

func init() {
	registerStubGroup(iapCmd, "oauth-brands", "(DEPRECATED) Manage OAuth brands", "create", "describe", "list")
	registerStubGroup(iapCmd, "oauth-clients", "(DEPRECATED) Manage OAuth clients", "create", "delete", "describe", "list", "reset-secret")
	registerStubGroup(iapCmd, "settings", "Manage IAP settings", "get", "set")
	tcp := &cobra.Command{Use: "tcp", Short: "Manage IAP TCP resources"}
	registerStubGroup(tcp, "dest-groups", "Manage TCP destination groups", "create", "delete", "describe", "list", "update")
	iapCmd.AddCommand(tcp)
	web := &cobra.Command{Use: "web", Short: "Manage IAP web policies"}
	for _, n := range []string{"enable", "disable", "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding"} {
		registerStubCommand(web, n, "Not yet implemented")
	}
	iapCmd.AddCommand(web)
	rootCmd.AddCommand(iapCmd)
}
