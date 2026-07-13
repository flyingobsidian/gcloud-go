package cmd

import "github.com/spf13/cobra"

// --- gcloud app (#299) ---

var appCmd = &cobra.Command{Use: "app", Short: "Manage App Engine"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(appCmd, "domain-mappings", "Manage domain mappings", crud...)
	registerStubGroup(appCmd, "firewall-rules", "Manage firewall rules", append(crud, "list-ingress-rules", "test-ip")...)
	registerStubGroup(appCmd, "instances", "Manage instances", "delete", "describe", "disable-debug", "enable-debug", "list", "scp", "ssh")
	registerStubGroup(appCmd, "logs", "Manage logs", "read", "tail")
	registerStubGroup(appCmd, "operations", "Manage operations", "describe", "list", "wait")
	registerStubGroup(appCmd, "regions", "View regional availability", "list", "describe")
	registerStubGroup(appCmd, "runtimes", "List runtimes", "list")
	registerStubGroup(appCmd, "services", "Manage services", "delete", "describe", "list", "set-traffic", "browse")
	registerStubGroup(appCmd, "ssl-certificates", "Manage SSL certificates", crud...)
	registerStubGroup(appCmd, "versions", "Manage versions", "browse", "delete", "describe", "list", "migrate", "start", "stop", "update")
	for _, name := range []string{"browse", "create", "deploy", "describe", "open-console", "update"} {
		registerStubCommand(appCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(appCmd)
}
