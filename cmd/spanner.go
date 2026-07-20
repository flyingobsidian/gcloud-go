package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	spanner "google.golang.org/api/spanner/v1"
)

// --- gcloud spanner (#387) ---

var spannerCmd = &cobra.Command{Use: "spanner", Short: "Manage Cloud Spanner"}

func init() {
	// backups (#1206), databases (#1207), instance-configs (#1208),
	// instance-partitions (#1209), instances (#1210), operations (#1211),
	// and rows (#1212) live in their own files. samples, cli, and
	// backup-schedules remain stubs pending their own issues.
	registerStubGroup(spannerCmd, "samples", "Sample apps", "list", "run")
	registerStubCommand(spannerCmd, "cli", "Interactive Spanner shell")
	rootCmd.AddCommand(spannerCmd)
}

// spannerProject returns the `projects/PROJECT` prefix, resolving the project
// from the standard --project/env/config chain.
func spannerProject() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return "projects/" + project, nil
}

// spannerInstance returns a fully qualified instance name. If the input is
// already a full URI (starts with "projects/"), it is returned as-is.
func spannerInstance(instance string) (string, error) {
	if strings.HasPrefix(instance, "projects/") {
		return instance, nil
	}
	project, err := spannerProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/instances/%s", project, instance), nil
}

// spannerDatabase returns a fully qualified database name.
func spannerDatabase(instance, database string) (string, error) {
	if strings.HasPrefix(database, "projects/") {
		return database, nil
	}
	inst, err := spannerInstance(instance)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/databases/%s", inst, database), nil
}

// spannerBackup returns a fully qualified backup name.
func spannerBackup(instance, backup string) (string, error) {
	if strings.HasPrefix(backup, "projects/") {
		return backup, nil
	}
	inst, err := spannerInstance(instance)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/backups/%s", inst, backup), nil
}

// spannerInstanceConfig returns a fully qualified instance config name.
func spannerInstanceConfig(cfg string) (string, error) {
	if strings.HasPrefix(cfg, "projects/") {
		return cfg, nil
	}
	project, err := spannerProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/instanceConfigs/%s", project, cfg), nil
}

// spannerInstancePartition returns a fully qualified instance partition name.
func spannerInstancePartition(instance, partition string) (string, error) {
	if strings.HasPrefix(partition, "projects/") {
		return partition, nil
	}
	inst, err := spannerInstance(instance)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/instancePartitions/%s", inst, partition), nil
}

// spannerBackupSchedule returns a fully qualified backup schedule name.
func spannerBackupSchedule(instance, database, schedule string) (string, error) {
	if strings.HasPrefix(schedule, "projects/") {
		return schedule, nil
	}
	db, err := spannerDatabase(instance, database)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/backupSchedules/%s", db, schedule), nil
}

// spIamMemberFlags binds the standard IAM member/role/condition flags shared by
// all Cloud Spanner add/remove-iam-policy-binding commands.
func spIamMemberFlags(c *cobra.Command, member, role, condExpr, condTitle, condDesc *string) {
	c.Flags().StringVar(member, "member", "", "IAM member (required)")
	c.Flags().StringVar(role, "role", "", "IAM role to bind (required)")
	c.Flags().StringVar(condExpr, "condition-expression", "", "CEL expression for a conditional binding")
	c.Flags().StringVar(condTitle, "condition-title", "", "Title for a conditional binding")
	c.Flags().StringVar(condDesc, "condition-description", "", "Description for a conditional binding")
	_ = c.MarkFlagRequired("member")
	_ = c.MarkFlagRequired("role")
}

func spIamBuildCondition(expr, title, desc string) *spanner.Expr {
	if expr == "" && title == "" && desc == "" {
		return nil
	}
	return &spanner.Expr{Expression: expr, Title: title, Description: desc}
}

func spIamCondsEqual(a, b *spanner.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

func spIamAddBinding(policy *spanner.Policy, role, member string, cond *spanner.Expr) {
	for _, b := range policy.Bindings {
		if b.Role != role || !spIamCondsEqual(b.Condition, cond) {
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
	policy.Bindings = append(policy.Bindings, &spanner.Binding{
		Role: role, Members: []string{member}, Condition: cond,
	})
}

func spIamRemoveBinding(policy *spanner.Policy, role, member string, cond *spanner.Expr, allConds bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConds || spIamCondsEqual(b.Condition, cond))
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

func spUpdatedIam(who string) {
	fmt.Fprintf(os.Stderr, "Updated IAM policy for %s.\n", who)
}
