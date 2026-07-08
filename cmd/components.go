package cmd

import "github.com/spf13/cobra"

// --- gcloud components (#318) ---
//
// The gcloud Python `components` command manages the CLI's own installed
// components. gcloud-go is distributed as a single binary and does not have
// a component model; these commands are registered as stubs so callers can
// discover the surface, but they do not perform installation.

var componentsCmd = &cobra.Command{Use: "components", Short: "Manage Google Cloud CLI components (stubbed)"}

func init() {
	registerStubGroup(componentsCmd, "repositories", "Manage additional component repositories", "add", "list", "remove")
	for _, name := range []string{"install", "list", "reinstall", "remove", "update"} {
		registerStubCommand(componentsCmd, name, "Not applicable to gcloud-go (single-binary distribution)")
	}
	rootCmd.AddCommand(componentsCmd)
}
