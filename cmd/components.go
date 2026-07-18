package cmd

import (
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/spf13/cobra"
)

// --- gcloud components (#318) ---
//
// gcloud-python's `components` group installs and updates CLI components.
// gcloud-go is distributed as a single binary and has no component model, so
// install/list/reinstall/remove/update are registered as stubs. The
// `repositories` subgroup (#1501) manages the additional_repositories property
// used by callers to point the (Python) SDK at Trusted Tester repositories;
// gcloud-go persists the same list so switching between the two binaries
// stays consistent.

var componentsCmd = &cobra.Command{Use: "components", Short: "Manage Google Cloud CLI components"}

var componentsReposCmd = &cobra.Command{
	Use:   "repositories",
	Short: "Manage additional component repositories",
}

var componentsReposAddCmd = &cobra.Command{
	Use:   "add URL [URL ...]",
	Short: "Add one or more Trusted Tester component repositories",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runComponentsReposAdd,
}

var componentsReposListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the configured additional component repositories",
	Args:  cobra.NoArgs,
	RunE:  runComponentsReposList,
}

var componentsReposRemoveCmd = &cobra.Command{
	Use:   "remove URL [URL ...]",
	Short: "Remove one or more previously-added component repositories",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runComponentsReposRemove,
}

func init() {
	componentsReposCmd.AddCommand(componentsReposAddCmd, componentsReposListCmd, componentsReposRemoveCmd)
	componentsCmd.AddCommand(componentsReposCmd)
	for _, name := range []string{"install", "list", "reinstall", "remove", "update"} {
		registerStubCommand(componentsCmd, name, "Not applicable to gcloud-go (single-binary distribution)")
	}
	rootCmd.AddCommand(componentsCmd)
}

func runComponentsReposAdd(cmd *cobra.Command, args []string) error {
	repos, err := config.LoadAdditionalRepositories()
	if err != nil {
		return err
	}
	seen := make(map[string]bool, len(repos))
	for _, r := range repos {
		seen[r] = true
	}
	var added, existing []string
	for _, url := range args {
		if seen[url] {
			existing = append(existing, url)
			continue
		}
		seen[url] = true
		repos = append(repos, url)
		added = append(added, url)
	}
	if err := config.SaveAdditionalRepositories(repos); err != nil {
		return err
	}
	for _, url := range added {
		fmt.Printf("Added repository: [%s]\n", url)
	}
	for _, url := range existing {
		fmt.Printf("Repository already added, skipping: [%s]\n", url)
	}
	return nil
}

func runComponentsReposList(cmd *cobra.Command, args []string) error {
	repos, err := config.LoadAdditionalRepositories()
	if err != nil {
		return err
	}
	if len(repos) == 0 {
		fmt.Println("No additional component repositories are configured.")
		return nil
	}
	for _, r := range repos {
		fmt.Println(r)
	}
	return nil
}

func runComponentsReposRemove(cmd *cobra.Command, args []string) error {
	repos, err := config.LoadAdditionalRepositories()
	if err != nil {
		return err
	}
	toRemove := make(map[string]bool, len(args))
	for _, a := range args {
		toRemove[a] = true
	}
	kept := repos[:0]
	var removed, missing []string
	seenRemove := make(map[string]bool, len(args))
	for _, r := range repos {
		if toRemove[r] {
			if !seenRemove[r] {
				removed = append(removed, r)
				seenRemove[r] = true
			}
			continue
		}
		kept = append(kept, r)
	}
	for _, url := range args {
		if !seenRemove[url] {
			missing = append(missing, url)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("repository not present: %v", missing)
	}
	if err := config.SaveAdditionalRepositories(kept); err != nil {
		return err
	}
	for _, url := range removed {
		fmt.Printf("Removed repository: [%s]\n", url)
	}
	return nil
}
