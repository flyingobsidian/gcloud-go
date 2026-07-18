package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	managedidentities "google.golang.org/api/managedidentities/v1"
)

// --- gcloud active-directory (#289) ---

var activeDirectoryCmd = &cobra.Command{Use: "active-directory", Short: "Manage Managed Microsoft AD"}

func init() {
	rootCmd.AddCommand(activeDirectoryCmd)
}

// adIamFlags registers the common IAM member/role/condition flags for
// add/remove-iam-policy-binding commands on Managed Microsoft AD resources.
func adIamFlags(c *cobra.Command, member, role, condExpr, condTitle, condDesc *string) {
	c.Flags().StringVar(member, "member", "", "IAM member (required)")
	c.Flags().StringVar(role, "role", "", "IAM role to bind (required)")
	c.Flags().StringVar(condExpr, "condition-expression", "", "CEL expression for a conditional binding")
	c.Flags().StringVar(condTitle, "condition-title", "", "Title for a conditional binding")
	c.Flags().StringVar(condDesc, "condition-description", "", "Description for a conditional binding")
	_ = c.MarkFlagRequired("member")
	_ = c.MarkFlagRequired("role")
}

func adBuildCondition(expr, title, desc string) *managedidentities.Expr {
	if expr == "" && title == "" && desc == "" {
		return nil
	}
	return &managedidentities.Expr{Expression: expr, Title: title, Description: desc}
}

func adCondsEqual(a, b *managedidentities.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

func adAddBinding(policy *managedidentities.Policy, role, member string, cond *managedidentities.Expr) {
	for _, b := range policy.Bindings {
		if b.Role != role || !adCondsEqual(b.Condition, cond) {
			continue
		}
		for _, m := range b.Members {
			if m == member {
				return
			}
		}
		b.Members = append(b.Members, member)
		return
	}
	policy.Bindings = append(policy.Bindings, &managedidentities.Binding{
		Role: role, Members: []string{member}, Condition: cond,
	})
}

func adRemoveBinding(policy *managedidentities.Policy, role, member string, cond *managedidentities.Expr, allConds bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConds || adCondsEqual(b.Condition, cond))
		if !match {
			kept = append(kept, b)
			continue
		}
		newMembers := b.Members[:0]
		for _, m := range b.Members {
			if m == member {
				continue
			}
			newMembers = append(newMembers, m)
		}
		if len(newMembers) != len(b.Members) {
			changed = true
		}
		b.Members = newMembers
		if len(b.Members) > 0 {
			kept = append(kept, b)
		} else {
			changed = true
		}
	}
	policy.Bindings = kept
	return changed
}

func adUpdatedIam(who string) {
	fmt.Fprintf(os.Stderr, "Updated IAM policy for %s.\n", who)
}
