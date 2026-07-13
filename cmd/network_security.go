package cmd

import "github.com/spf13/cobra"

// --- gcloud network-security (#363) ---

var networkSecurityCmd = &cobra.Command{Use: "network-security", Short: "Manage Network Security"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(networkSecurityCmd, "address-groups", "Manage address groups", append(crud, "add-items", "remove-items", "clone-items", "list-references")...)
	registerStubGroup(networkSecurityCmd, "authorization-policies", "Manage authorization policies", crud...)
	registerStubGroup(networkSecurityCmd, "authz-policies", "Manage authorization policies (v2)", crud...)
	registerStubGroup(networkSecurityCmd, "client-tls-policies", "Manage client TLS policies", crud...)
	registerStubGroup(networkSecurityCmd, "firewall-endpoint-associations", "Manage firewall endpoint associations", crud...)
	registerStubGroup(networkSecurityCmd, "firewall-endpoints", "Manage firewall endpoints", crud...)
	registerStubGroup(networkSecurityCmd, "gateway-security-policies", "Manage gateway security policies", append(crud, "rules")...)
	registerStubGroup(networkSecurityCmd, "intercept-deployment-groups", "Manage intercept deployment groups", crud...)
	registerStubGroup(networkSecurityCmd, "intercept-deployments", "Manage intercept deployments", crud...)
	registerStubGroup(networkSecurityCmd, "intercept-endpoint-groups", "Manage intercept endpoint groups", crud...)
	registerStubGroup(networkSecurityCmd, "mirroring-deployment-groups", "Manage mirroring deployment groups", crud...)
	registerStubGroup(networkSecurityCmd, "mirroring-deployments", "Manage mirroring deployments", crud...)
	registerStubGroup(networkSecurityCmd, "mirroring-endpoint-groups", "Manage mirroring endpoint groups", crud...)
	registerStubGroup(networkSecurityCmd, "mirroring-endpoints", "Manage mirroring endpoints", crud...)
	registerStubGroup(networkSecurityCmd, "operations", "Manage operations", "cancel", "delete", "describe", "list")
	registerStubGroup(networkSecurityCmd, "org-address-groups", "Manage org address groups", append(crud, "add-items", "remove-items")...)
	registerStubGroup(networkSecurityCmd, "secure-access-connect", "Manage Secure Access Connect", crud...)
	registerStubGroup(networkSecurityCmd, "security-profile-groups", "Manage security profile groups", crud...)
	registerStubGroup(networkSecurityCmd, "security-profiles", "Manage security profiles", crud...)
	registerStubGroup(networkSecurityCmd, "server-tls-policies", "Manage server TLS policies", crud...)
	registerStubGroup(networkSecurityCmd, "tls-inspection-policies", "Manage TLS inspection policies", crud...)
	registerStubGroup(networkSecurityCmd, "ull-mirroring-collectors", "Manage ULL mirroring collectors", crud...)
	registerStubGroup(networkSecurityCmd, "ull-mirroring-engines", "Manage ULL mirroring engines", crud...)
	registerStubGroup(networkSecurityCmd, "url-lists", "Manage URL lists", crud...)
		registerStubGroup(networkSecurityCmd, "backend-authentication-configs", "Manage backend-authentication-configs", "list", "describe")
	registerStubGroup(networkSecurityCmd, "dns-threat-detectors", "Manage dns-threat-detectors", "list", "describe")
	registerStubGroup(networkSecurityCmd, "intercept-endpoint-group-associations", "Manage intercept-endpoint-group-associations", "list", "describe")
	registerStubGroup(networkSecurityCmd, "mirroring-endpoint-group-associations", "Manage mirroring-endpoint-group-associations", "list", "describe")
	rootCmd.AddCommand(networkSecurityCmd)
}
