package cmd

import "github.com/spf13/cobra"

// --- gcloud vmware (#395) ---

var vmwareCmd = &cobra.Command{Use: "vmware", Short: "Manage VMware Engine"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(vmwareCmd, "announcements", "Manage announcements", "describe", "list")
	registerStubGroup(vmwareCmd, "datastores", "Manage datastores", crud...)
	registerStubGroup(vmwareCmd, "dns-bind-permission", "Manage DNS binding permission", "describe", "grant", "revoke")
	registerStubGroup(vmwareCmd, "locations", "List locations", "list", "describe")
	registerStubGroup(vmwareCmd, "network-peerings", "Manage VPC peerings", crud...)
	registerStubGroup(vmwareCmd, "network-policies", "Manage network policies", append(crud, "external-access-rules")...)
	registerStubGroup(vmwareCmd, "networks", "Manage networks", crud...)
	registerStubGroup(vmwareCmd, "node-types", "Show node types", "list", "describe")
	registerStubGroup(vmwareCmd, "operations", "Manage operations", "describe", "list")
	registerStubGroup(vmwareCmd, "private-clouds", "Manage private clouds", append(crud, "clusters", "external-addresses", "hcx-activation-keys", "logging-servers", "management-dns-zone-bindings", "subnets", "nsx", "vcenter", "dns-forwarding", "get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding", "reset-nsx-credentials", "reset-vcenter-credentials", "upgrades")...)
	registerStubGroup(vmwareCmd, "private-connections", "Manage private connections", crud...)
	rootCmd.AddCommand(vmwareCmd)
}
