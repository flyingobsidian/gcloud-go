package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	dns "google.golang.org/api/dns/v1"
)

// --- gcloud dns (#331) ---

var dnsCmd = &cobra.Command{Use: "dns", Short: "Manage Cloud DNS"}

func init() {
	rootCmd.AddCommand(dnsCmd)
}

func dnsIamMemberFlags(c *cobra.Command, member, role, condExpr, condTitle, condDesc *string) {
	c.Flags().StringVar(member, "member", "", "IAM member (required)")
	c.Flags().StringVar(role, "role", "", "IAM role to bind (required)")
	c.Flags().StringVar(condExpr, "condition-expression", "", "CEL expression for a conditional binding")
	c.Flags().StringVar(condTitle, "condition-title", "", "Title for a conditional binding")
	c.Flags().StringVar(condDesc, "condition-description", "", "Description for a conditional binding")
	_ = c.MarkFlagRequired("member")
	_ = c.MarkFlagRequired("role")
}

func dnsBuildCondition(expr, title, desc string) *dns.Expr {
	if expr == "" && title == "" && desc == "" {
		return nil
	}
	return &dns.Expr{Expression: expr, Title: title, Description: desc}
}

func dnsCondsEqual(a, b *dns.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

func dnsAddBinding(policy *dns.GoogleIamV1Policy, role, member string, cond *dns.Expr) {
	for _, b := range policy.Bindings {
		if b.Role != role || !dnsCondsEqual(b.Condition, cond) {
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
	policy.Bindings = append(policy.Bindings, &dns.GoogleIamV1Binding{
		Role: role, Members: []string{member}, Condition: cond,
	})
}

func dnsRemoveBinding(policy *dns.GoogleIamV1Policy, role, member string, cond *dns.Expr, allConds bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConds || dnsCondsEqual(b.Condition, cond))
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

func dnsUpdatedIam(who string) {
	fmt.Fprintf(os.Stderr, "Updated IAM policy for %s.\n", who)
}
