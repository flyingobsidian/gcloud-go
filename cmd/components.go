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

func (s *componentsCmds) initComponents() {
	s.componentsReposCmd.AddCommand(s.componentsReposAddCmd, s.componentsReposListCmd, s.componentsReposRemoveCmd)
	s.componentsCmd.AddCommand(s.componentsReposCmd)
	for _, name := range []string{"install", "list", "reinstall", "remove", "update"} {
		registerStubCommand(s.componentsCmd, name, "Not applicable to gcloud-go (single-binary distribution)")
	}

}

func (s *componentsCmds) runComponentsReposAdd(cmd *cobra.Command, args []string) error {
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
		fmt.Fprintf(s.inv.stdout, "Added repository: [%s]\n", url)
	}
	for _, url := range existing {
		fmt.Fprintf(s.inv.stdout, "Repository already added, skipping: [%s]\n", url)
	}
	return nil
}

func (s *componentsCmds) runComponentsReposList(cmd *cobra.Command, args []string) error {
	repos, err := config.LoadAdditionalRepositories()
	if err != nil {
		return err
	}
	if len(repos) == 0 {
		fmt.Fprintln(s.inv.stdout, "No additional component repositories are configured.")
		return nil
	}
	for _, r := range repos {
		fmt.Fprintln(s.inv.stdout, r)
	}
	return nil
}

func (s *componentsCmds) runComponentsReposRemove(cmd *cobra.Command, args []string) error {
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
		fmt.Fprintf(s.inv.stdout, "Removed repository: [%s]\n", url)
	}
	return nil
}

// componentsCmds holds the commands and flag values of one invocation's "components"
// command group, so invocations don't share state (#2070).
type componentsCmds struct {
	inv *invocation

	componentsCmd            *cobra.Command
	componentsReposCmd       *cobra.Command
	componentsReposAddCmd    *cobra.Command
	componentsReposListCmd   *cobra.Command
	componentsReposRemoveCmd *cobra.Command
}

// newComponentsCmd returns the "components" command group for inv.
func newComponentsCmd(inv *invocation) *cobra.Command {
	return newComponentsCmds(inv).componentsCmd
}

// newComponentsCmds builds the "components" command group for inv and returns its state.
func newComponentsCmds(inv *invocation) *componentsCmds {
	s := &componentsCmds{inv: inv}
	s.componentsCmd = &cobra.Command{Use: "components", Short: "Manage Google Cloud CLI components"}
	s.componentsReposCmd = &cobra.Command{
		Use:   "repositories",
		Short: "Manage additional component repositories",
	}
	s.componentsReposAddCmd = &cobra.Command{
		Use:   "add URL [URL ...]",
		Short: "Add one or more Trusted Tester component repositories",
		Args:  cobra.MinimumNArgs(1),
		RunE:  s.runComponentsReposAdd,
	}
	s.componentsReposListCmd = &cobra.Command{
		Use:   "list",
		Short: "List the configured additional component repositories",
		Args:  cobra.NoArgs,
		RunE:  s.runComponentsReposList,
	}
	s.componentsReposRemoveCmd = &cobra.Command{
		Use:   "remove URL [URL ...]",
		Short: "Remove one or more previously-added component repositories",
		Args:  cobra.MinimumNArgs(1),
		RunE:  s.runComponentsReposRemove,
	}
	s.initComponents()
	return s
}
