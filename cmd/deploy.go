package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
)

// --- gcloud deploy (#327) ---

var deployCmd = &cobra.Command{Use: "deploy", Short: "Manage Cloud Deploy"}

func init() {
	for _, name := range []string{"apply", "delete", "get-config"} {
		registerStubCommand(deployCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(deployCmd)
}

func deployLocationParent(project, region string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, region)
}

func deployChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func deployIamMemberFlags(c *cobra.Command, member, role, condExpr, condTitle, condDesc *string) {
	c.Flags().StringVar(member, "member", "", "IAM member (required)")
	c.Flags().StringVar(role, "role", "", "IAM role to bind (required)")
	c.Flags().StringVar(condExpr, "condition-expression", "", "CEL expression for a conditional binding")
	c.Flags().StringVar(condTitle, "condition-title", "", "Title for a conditional binding")
	c.Flags().StringVar(condDesc, "condition-description", "", "Description for a conditional binding")
	_ = c.MarkFlagRequired("member")
	_ = c.MarkFlagRequired("role")
}

func deployBuildCondition(expr, title, desc string) *clouddeploy.Expr {
	if expr == "" && title == "" && desc == "" {
		return nil
	}
	return &clouddeploy.Expr{Expression: expr, Title: title, Description: desc}
}

func deployCondsEqual(a, b *clouddeploy.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

func deployAddBinding(policy *clouddeploy.Policy, role, member string, cond *clouddeploy.Expr) {
	for _, b := range policy.Bindings {
		if b.Role != role || !deployCondsEqual(b.Condition, cond) {
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
	policy.Bindings = append(policy.Bindings, &clouddeploy.Binding{
		Role: role, Members: []string{member}, Condition: cond,
	})
}

func deployRemoveBinding(policy *clouddeploy.Policy, role, member string, cond *clouddeploy.Expr, allConds bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConds || deployCondsEqual(b.Condition, cond))
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

func deployUpdatedIam(who string) {
	fmt.Fprintf(os.Stderr, "Updated IAM policy for %s.\n", who)
}
