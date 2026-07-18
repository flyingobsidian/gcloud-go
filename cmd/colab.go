package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud colab (#316) ---
//
// Colab Enterprise is served by the aiplatform API surface. Every subgroup
// accepts a --region flag; the aiplatform client is regional (endpoint of
// the form https://<region>-aiplatform.googleapis.com/).

var colabCmd = &cobra.Command{Use: "colab", Short: "Manage Colab Enterprise"}

func init() {
	rootCmd.AddCommand(colabCmd)
}

func colabParent(region string) (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if region == "" {
		return "", fmt.Errorf("--region is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, region), nil
}

func colabChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func colabIamMemberFlags(c *cobra.Command, member, role, condExpr, condTitle, condDesc *string) {
	c.Flags().StringVar(member, "member", "", "IAM member (e.g. user:alice@example.com) (required)")
	c.Flags().StringVar(role, "role", "", "IAM role to bind (required)")
	c.Flags().StringVar(condExpr, "condition-expression", "", "CEL expression for a conditional binding")
	c.Flags().StringVar(condTitle, "condition-title", "", "Title for a conditional binding")
	c.Flags().StringVar(condDesc, "condition-description", "", "Description for a conditional binding")
	_ = c.MarkFlagRequired("member")
	_ = c.MarkFlagRequired("role")
}

func colabBuildCondition(expr, title, desc string) *aiplatform.GoogleTypeExpr {
	if expr == "" && title == "" && desc == "" {
		return nil
	}
	return &aiplatform.GoogleTypeExpr{Expression: expr, Title: title, Description: desc}
}

func colabIamCondsEqual(a, b *aiplatform.GoogleTypeExpr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

func colabAddBinding(policy *aiplatform.GoogleIamV1Policy, role, member string, cond *aiplatform.GoogleTypeExpr) {
	for _, b := range policy.Bindings {
		if b.Role != role || !colabIamCondsEqual(b.Condition, cond) {
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
	policy.Bindings = append(policy.Bindings, &aiplatform.GoogleIamV1Binding{
		Role:      role,
		Members:   []string{member},
		Condition: cond,
	})
}

func colabRemoveBinding(policy *aiplatform.GoogleIamV1Policy, role, member string, cond *aiplatform.GoogleTypeExpr, allConds bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConds || colabIamCondsEqual(b.Condition, cond))
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

func writeUpdatedIam(who string) {
	fmt.Fprintf(os.Stderr, "Updated IAM policy for %s.\n", who)
}
