package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	dataproc "google.golang.org/api/dataproc/v1"
)

// --- gcloud dataproc (#324) ---

var dataprocCmd = &cobra.Command{Use: "dataproc", Short: "Manage Dataproc"}

func init() {
	rootCmd.AddCommand(dataprocCmd)
}

func dpRegionParent(project, region string) string {
	return fmt.Sprintf("projects/%s/regions/%s", project, region)
}

func dpLocationParent(project, region string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, region)
}

func dpChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func dpIamMemberFlags(c *cobra.Command, member, role, condExpr, condTitle, condDesc *string) {
	c.Flags().StringVar(member, "member", "", "IAM member (e.g. user:alice@example.com) (required)")
	c.Flags().StringVar(role, "role", "", "IAM role to bind (required)")
	c.Flags().StringVar(condExpr, "condition-expression", "", "CEL expression for a conditional binding")
	c.Flags().StringVar(condTitle, "condition-title", "", "Title for a conditional binding")
	c.Flags().StringVar(condDesc, "condition-description", "", "Description for a conditional binding")
	_ = c.MarkFlagRequired("member")
	_ = c.MarkFlagRequired("role")
}

func dpBuildCondition(expr, title, desc string) *dataproc.Expr {
	if expr == "" && title == "" && desc == "" {
		return nil
	}
	return &dataproc.Expr{Expression: expr, Title: title, Description: desc}
}

func dpCondsEqual(a, b *dataproc.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

func dpAddBinding(policy *dataproc.Policy, role, member string, cond *dataproc.Expr) {
	for _, b := range policy.Bindings {
		if b.Role != role || !dpCondsEqual(b.Condition, cond) {
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
	policy.Bindings = append(policy.Bindings, &dataproc.Binding{
		Role:      role,
		Members:   []string{member},
		Condition: cond,
	})
}

func dpRemoveBinding(policy *dataproc.Policy, role, member string, cond *dataproc.Expr, allConds bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConds || dpCondsEqual(b.Condition, cond))
		if !match {
			kept = append(kept, b)
			continue
		}
		newMembers := b.Members[:0]
		for _, m := range b.Members {
			if m == member {
				changed = true
				continue
			}
			newMembers = append(newMembers, m)
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

func dpUpdatedIam(who string) {
	fmt.Fprintf(os.Stderr, "Updated IAM policy for %s.\n", who)
}
