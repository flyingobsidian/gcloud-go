package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	iap "google.golang.org/api/iap/v1"
)

// --- gcloud iap (#345) ---
//
// gcloud-go already ships an internal iap package under internal/iap for SSH
// tunneling. This command surface backs the gcloud-python subcommand set
// against the Cloud Identity-Aware Proxy v1 API.

var iapCmd = &cobra.Command{Use: "iap", Short: "Manage Identity-Aware Proxy resources"}

// iapTcpCmd is the parent for `gcloud iap tcp` and hosts dest-groups.
var iapTcpCmd = &cobra.Command{Use: "tcp", Short: "IAP TCP forwarding"}

// IapIamResourceTypes lists the IAM resource types accepted by
// `gcloud iap web` IAM commands. `agent-registry` was promoted to GA in
// gcloud-python 576.0.0 (#1781), so it is included here regardless of track.
var IapIamResourceTypes = []string{
	"app-engine",
	"backend-services",
	"forwarding-rule",
	"cloud-run",
	"agent-registry",
}

func init() {
	iapCmd.AddCommand(iapTcpCmd)
	// `iap web` remains a stub surface pending full IAM integration for the
	// web application-level policy management commands. Each stub still
	// accepts the same flag set as gcloud-python so scripts written against
	// the GA `--resource-type=agent-registry` option parse cleanly.
	web := &cobra.Command{Use: "web", Short: "Manage IAP web policies"}
	iamStubNames := []string{"get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding"}
	for _, n := range iamStubNames {
		c := registerStubCommand(web, n, "Not yet implemented")
		addIapWebIamStubFlags(c)
	}
	for _, n := range []string{"enable", "disable"} {
		c := registerStubCommand(web, n, "Not yet implemented")
		addIapWebEnableStubFlags(c)
	}
	iapCmd.AddCommand(web)
	rootCmd.AddCommand(iapCmd)
}

func addIapWebIamStubFlags(c *cobra.Command) {
	c.Flags().String("resource-type", "", "Resource type of the IAP resource. One of: "+strings.Join(IapIamResourceTypes, ", "))
	c.Flags().String("service", "", "Service name")
	c.Flags().String("region", "", "Region name")
	c.Flags().String("version", "", "Service version (app-engine only)")
	// Agent-registry-specific selectors (GA since 576.0.0).
	c.Flags().String("agent", "", "Agent ID for the agent-registry resource type")
	c.Flags().String("mcp-server", "", "MCP server ID for the agent-registry resource type")
	c.Flags().String("endpoint", "", "Endpoint ID for the agent-registry resource type")
}

func addIapWebEnableStubFlags(c *cobra.Command) {
	c.Flags().String("resource-type", "", "Resource type of the IAP resource")
	c.Flags().String("service", "", "Service name")
	c.Flags().String("region", "", "Region name")
}

func iapIamMemberFlags(c *cobra.Command, member, role, condExpr, condTitle, condDesc *string) {
	c.Flags().StringVar(member, "member", "", "IAM member (required)")
	c.Flags().StringVar(role, "role", "", "IAM role to bind (required)")
	c.Flags().StringVar(condExpr, "condition-expression", "", "CEL expression for a conditional binding")
	c.Flags().StringVar(condTitle, "condition-title", "", "Title for a conditional binding")
	c.Flags().StringVar(condDesc, "condition-description", "", "Description for a conditional binding")
	_ = c.MarkFlagRequired("member")
	_ = c.MarkFlagRequired("role")
}

func iapIamBuildCondition(expr, title, desc string) *iap.Expr {
	if expr == "" && title == "" && desc == "" {
		return nil
	}
	return &iap.Expr{Expression: expr, Title: title, Description: desc}
}

func iapIamCondsEqual(a, b *iap.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

func iapIamAddBinding(policy *iap.Policy, role, member string, cond *iap.Expr) {
	for _, b := range policy.Bindings {
		if b.Role != role || !iapIamCondsEqual(b.Condition, cond) {
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
	policy.Bindings = append(policy.Bindings, &iap.Binding{
		Role: role, Members: []string{member}, Condition: cond,
	})
}

func iapIamRemoveBinding(policy *iap.Policy, role, member string, cond *iap.Expr, allConds bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConds || iapIamCondsEqual(b.Condition, cond))
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

func iapUpdatedIam(who string) {
	fmt.Fprintf(os.Stderr, "Updated IAM policy for %s.\n", who)
}
