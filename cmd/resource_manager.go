package cmd

import (
	"github.com/spf13/cobra"
	crm "google.golang.org/api/cloudresourcemanager/v3"
)

func newResourceManagerCmd(inv *invocation) *cobra.Command {

	resourceManagerCmd := &cobra.Command{
		Use:   "resource-manager",
		Short: "Manage Cloud Resource Manager resources",
	}

	resourceManagerCmd.AddCommand(newCapabilitiesCmd(inv))
	resourceManagerCmd.AddCommand(newFoldersCmd(inv))
	resourceManagerCmd.AddCommand(newOrgPoliciesCmd(inv))
	resourceManagerCmd.AddCommand(newTagsCmd(inv))
	return resourceManagerCmd
}

// rmBuildCondition returns a *crm.Expr from the given condition components,
// or nil if none are set. Shared across resource-manager IAM binding commands.
func rmBuildCondition(expression, title, description string) *crm.Expr {
	if expression == "" && title == "" && description == "" {
		return nil
	}
	return &crm.Expr{
		Expression:  expression,
		Title:       title,
		Description: description,
	}
}
