package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	runv2 "google.golang.org/api/run/v2"
)

// --- gcloud run (#380, #1049-#1056) ---

var runCmd = &cobra.Command{Use: "run", Short: "Manage Cloud Run"}

func init() {
	// `run compose` is not in this batch's scope; keep it as a stub so the
	// help surface still lists it.
	registerStubGroup(runCmd, "compose", "Docker Compose workflows", "run", "list")
	rootCmd.AddCommand(runCmd)
}

// runParent returns "projects/PROJ/locations/REGION" used as the parent for
// most v2 collection RPCs. Both segments are required.
func runParent(project, region string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, region)
}

// runResourceName joins a parent, collection segment ("services", "jobs", ...)
// and a resource id. If the caller passes an already-qualified name
// ("projects/.../foo/bar"), it is returned unchanged.
func runResourceName(project, region, collection, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", runParent(project, region), collection, id)
}

// runNamespaceName joins a namespace ("namespaces/PROJ") and a domain mapping
// or v1 resource id, passing through fully-qualified names unchanged.
func runNamespaceName(project, collection, id string) string {
	if strings.HasPrefix(id, "namespaces/") {
		return id
	}
	return fmt.Sprintf("namespaces/%s/%s/%s", project, collection, id)
}

// --- Shared IAM helpers (mirror cmd/dns.go, but wired to run/v2's
// GoogleIamV1Policy/GoogleIamV1Binding types with GoogleTypeExpr conditions). ---

func runIamFlags(c *cobra.Command, member, role, condExpr, condTitle, condDesc *string) {
	c.Flags().StringVar(member, "member", "", "IAM member (required)")
	c.Flags().StringVar(role, "role", "", "IAM role to bind (required)")
	c.Flags().StringVar(condExpr, "condition-expression", "", "CEL expression for a conditional binding")
	c.Flags().StringVar(condTitle, "condition-title", "", "Title for a conditional binding")
	c.Flags().StringVar(condDesc, "condition-description", "", "Description for a conditional binding")
	_ = c.MarkFlagRequired("member")
	_ = c.MarkFlagRequired("role")
}

func runIamBuildCondition(expr, title, desc string) *runv2.GoogleTypeExpr {
	if expr == "" && title == "" && desc == "" {
		return nil
	}
	return &runv2.GoogleTypeExpr{Expression: expr, Title: title, Description: desc}
}

func runIamCondsEqual(a, b *runv2.GoogleTypeExpr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

func runIamAddBinding(policy *runv2.GoogleIamV1Policy, role, member string, cond *runv2.GoogleTypeExpr) {
	for _, b := range policy.Bindings {
		if b.Role != role || !runIamCondsEqual(b.Condition, cond) {
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
	policy.Bindings = append(policy.Bindings, &runv2.GoogleIamV1Binding{
		Role: role, Members: []string{member}, Condition: cond,
	})
}

func runIamRemoveBinding(policy *runv2.GoogleIamV1Policy, role, member string, cond *runv2.GoogleTypeExpr, allConds bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConds || runIamCondsEqual(b.Condition, cond))
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

func runIamUpdatedIam(who string) {
	fmt.Fprintf(os.Stderr, "Updated IAM policy for %s.\n", who)
}
