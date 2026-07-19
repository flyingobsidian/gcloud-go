package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms (#349) ---

var kmsCmd = &cobra.Command{Use: "kms", Short: "Manage Cloud KMS"}

func init() {
	rootCmd.AddCommand(kmsCmd)
}

// kmsLocationParent returns "projects/PROJECT/locations/LOCATION".
func kmsLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

// kmsKeyringParent returns
// "projects/PROJECT/locations/LOCATION/keyRings/KEYRING".
func kmsKeyringParent(project, location, keyring string) string {
	return fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", project, location, keyring)
}

// kmsKeyName returns
// "projects/PROJECT/locations/LOCATION/keyRings/KEYRING/cryptoKeys/KEY".
func kmsKeyName(project, location, keyring, key string) string {
	return fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
		project, location, keyring, key)
}

// kmsFullName returns raw if it already looks like a fully-qualified resource
// name; otherwise it joins parent and raw.
func kmsFullName(parent, raw string) string {
	if strings.HasPrefix(raw, "projects/") || strings.HasPrefix(raw, "folders/") ||
		strings.HasPrefix(raw, "organizations/") {
		return raw
	}
	return parent + "/" + raw
}

// --- shared IAM helpers ---

// kmsIamBuildCondition returns a *cloudkms.Expr or nil.
func kmsIamBuildCondition(expr, title, desc string) *cloudkms.Expr {
	if expr == "" && title == "" && desc == "" {
		return nil
	}
	return &cloudkms.Expr{Expression: expr, Title: title, Description: desc}
}

// kmsIamCondsEqual reports whether two IAM condition exprs are equivalent.
func kmsIamCondsEqual(a, b *cloudkms.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

// kmsIamAddBinding adds (role, member[, cond]) to policy in-place.
func kmsIamAddBinding(policy *cloudkms.Policy, role, member string, cond *cloudkms.Expr) {
	for _, b := range policy.Bindings {
		if b.Role != role || !kmsIamCondsEqual(b.Condition, cond) {
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
	policy.Bindings = append(policy.Bindings, &cloudkms.Binding{
		Role: role, Members: []string{member}, Condition: cond,
	})
}

// kmsIamRemoveBinding removes (role, member[, cond]) from policy in-place. If
// allConds is true, matching bindings across all conditions are cleared.
func kmsIamRemoveBinding(policy *cloudkms.Policy, role, member string, cond *cloudkms.Expr, allConds bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConds || kmsIamCondsEqual(b.Condition, cond))
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

// kmsUpdatedIam prints the standard "Updated IAM policy for ..." message.
func kmsUpdatedIam(who string) {
	fmt.Fprintf(os.Stderr, "Updated IAM policy for %s.\n", who)
}

// kmsIamMemberFlags binds the common --member/--role/--condition-* flags for a
// (add|remove)-iam-policy-binding subcommand.
func kmsIamMemberFlags(c *cobra.Command, member, role, condExpr, condTitle, condDesc *string) {
	c.Flags().StringVar(member, "member", "", "IAM member (required)")
	c.Flags().StringVar(role, "role", "", "IAM role to bind (required)")
	c.Flags().StringVar(condExpr, "condition-expression", "", "CEL expression for a conditional binding")
	c.Flags().StringVar(condTitle, "condition-title", "", "Title for a conditional binding")
	c.Flags().StringVar(condDesc, "condition-description", "", "Description for a conditional binding")
	_ = c.MarkFlagRequired("member")
	_ = c.MarkFlagRequired("role")
}
