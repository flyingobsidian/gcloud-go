package cmd

import "github.com/spf13/cobra"

// --- gcloud container (#320) ---
//
// Container/GKE is a very large surface. This stub registers the top-level
// subgroups and common commands so callers can discover them; individual
// commands should be wired up to google.golang.org/api/container/v1 in
// follow-up PRs.

var containerCmd = &cobra.Command{Use: "container", Short: "Manage Google Kubernetes Engine"}

func init() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(containerCmd, "ai", "Manage AI-related workloads", "profiles")
	registerStubGroup(containerCmd, "attached", "Manage attached clusters", append(crud, "generate-install-manifest", "import", "get-credentials")...)
	registerStubGroup(containerCmd, "aws", "(DEPRECATED) Manage AWS clusters", crud...)
	registerStubGroup(containerCmd, "azure", "(DEPRECATED) Manage Azure clusters", crud...)
	registerStubGroup(containerCmd, "binauthz", "Manage Binary Authorization attestations",
		"create", "delete", "describe", "list", "sign", "verify", "policy")
	registerStubGroup(containerCmd, "images", "Manage container images",
		"delete", "describe", "list", "list-tags", "add-tag", "remove-tag", "untag")
	registerStubGroup(containerCmd, "subnets", "Manage subnets", "list-usable")
	registerStubGroup(containerCmd, "workload", "Manage Workload Optimizer",
		"list-recommendations", "get-recommendation", "apply-recommendation")
	rootCmd.AddCommand(containerCmd)
}
