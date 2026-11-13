package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	notebooksv1 "google.golang.org/api/notebooks/v1"
)

// --- gcloud notebooks (#365) ---
//
// Cloud AI Platform Notebooks (`gcloud notebooks`) is served by the
// notebooks.googleapis.com API. The v1 surface covers user-managed Environments,
// Instances and managed Runtimes; the subcommand files below hang off this
// parent command.

func (s *notebooksCmds) initNotebooks() {

}

func notebooksIamMemberFlags(c *cobra.Command, member, role, condExpr, condTitle, condDesc *string) {
	c.Flags().StringVar(member, "member", "", "IAM member (required)")
	c.Flags().StringVar(role, "role", "", "IAM role to bind (required)")
	c.Flags().StringVar(condExpr, "condition-expression", "", "CEL expression for a conditional binding")
	c.Flags().StringVar(condTitle, "condition-title", "", "Title for a conditional binding")
	c.Flags().StringVar(condDesc, "condition-description", "", "Description for a conditional binding")
	_ = c.MarkFlagRequired("member")
	_ = c.MarkFlagRequired("role")
}

func notebooksIamBuildCondition(expr, title, desc string) *notebooksv1.Expr {
	if expr == "" && title == "" && desc == "" {
		return nil
	}
	return &notebooksv1.Expr{Expression: expr, Title: title, Description: desc}
}

func notebooksIamCondsEqual(a, b *notebooksv1.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

func notebooksIamAddBinding(policy *notebooksv1.Policy, role, member string, cond *notebooksv1.Expr) {
	for _, b := range policy.Bindings {
		if b.Role != role || !notebooksIamCondsEqual(b.Condition, cond) {
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
	policy.Bindings = append(policy.Bindings, &notebooksv1.Binding{
		Role: role, Members: []string{member}, Condition: cond,
	})
}

func notebooksIamRemoveBinding(policy *notebooksv1.Policy, role, member string, cond *notebooksv1.Expr, allConds bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConds || notebooksIamCondsEqual(b.Condition, cond))
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

func (s *notebooksCmds) notebooksUpdatedIam(who string) {
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for %s.\n", who)
}

// notebooksCmds holds the commands and flag values of one invocation's "notebooks"
// command group, so invocations don't share state (#2070).
type notebooksCmds struct {
	inv *invocation

	notebooksCmd                   *cobra.Command
	notebooksEnvCmd                *cobra.Command
	flagNotebooksEnvLocation       string
	flagNotebooksEnvFormat         string
	flagNotebooksEnvConfigFile     string
	flagNotebooksEnvPageSize       int64
	notebooksEnvCreateCmd          *cobra.Command
	notebooksEnvDeleteCmd          *cobra.Command
	notebooksEnvDescribeCmd        *cobra.Command
	notebooksEnvListCmd            *cobra.Command
	notebooksInstCmd               *cobra.Command
	flagNotebooksInstLocation      string
	flagNotebooksInstFormat        string
	flagNotebooksInstConfigFile    string
	flagNotebooksInstUpdateMask    string
	flagNotebooksInstPageSize      int64
	flagNotebooksInstIamMember     string
	flagNotebooksInstIamRole       string
	flagNotebooksInstIamCondExpr   string
	flagNotebooksInstIamCondT      string
	flagNotebooksInstIamCondD      string
	flagNotebooksInstIamAllCond    bool
	notebooksInstCreateCmd         *cobra.Command
	notebooksInstDeleteCmd         *cobra.Command
	notebooksInstDescribeCmd       *cobra.Command
	notebooksInstListCmd           *cobra.Command
	notebooksInstUpdateCmd         *cobra.Command
	notebooksInstAddIamCmd         *cobra.Command
	notebooksInstRemoveIamCmd      *cobra.Command
	notebooksInstGetIamCmd         *cobra.Command
	notebooksInstSetIamCmd         *cobra.Command
	notebooksInstDiagnoseCmd       *cobra.Command
	notebooksInstGetHealthCmd      *cobra.Command
	notebooksInstIsUpgradeableCmd  *cobra.Command
	notebooksInstMigrateCmd        *cobra.Command
	notebooksInstRegisterCmd       *cobra.Command
	notebooksInstResetCmd          *cobra.Command
	notebooksInstRollbackCmd       *cobra.Command
	notebooksInstStartCmd          *cobra.Command
	notebooksInstStopCmd           *cobra.Command
	notebooksInstUpgradeCmd        *cobra.Command
	notebooksLocCmd                *cobra.Command
	flagNotebooksLocFormat         string
	flagNotebooksLocPageSize       int64
	notebooksLocListCmd            *cobra.Command
	notebooksRuntimeCmd            *cobra.Command
	flagNotebooksRuntimeLocation   string
	flagNotebooksRuntimeFormat     string
	flagNotebooksRuntimeConfigFile string
	flagNotebooksRuntimePageSize   int64
	notebooksRuntimeCreateCmd      *cobra.Command
	notebooksRuntimeDeleteCmd      *cobra.Command
	notebooksRuntimeDescribeCmd    *cobra.Command
	notebooksRuntimeListCmd        *cobra.Command
	notebooksRuntimeDiagnoseCmd    *cobra.Command
	notebooksRuntimeMigrateCmd     *cobra.Command
	notebooksRuntimeResetCmd       *cobra.Command
	notebooksRuntimeStartCmd       *cobra.Command
	notebooksRuntimeStopCmd        *cobra.Command
	notebooksRuntimeSwitchCmd      *cobra.Command
}

// newNotebooksCmd returns the "notebooks" command group for inv.
func newNotebooksCmd(inv *invocation) *cobra.Command {
	return newNotebooksCmds(inv).notebooksCmd
}

// newNotebooksCmds builds the "notebooks" command group for inv and returns its state.
func newNotebooksCmds(inv *invocation) *notebooksCmds {
	s := &notebooksCmds{inv: inv}
	s.notebooksCmd = &cobra.Command{Use: "notebooks", Short: "Manage Cloud AI Platform Notebooks"}
	s.notebooksEnvCmd = &cobra.Command{Use: "environments", Short: "Manage notebook environments"}
	s.notebooksEnvCreateCmd = &cobra.Command{
		Use: "create ENVIRONMENT", Short: "Create a notebook environment",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksEnvCreate,
	}
	s.notebooksEnvDeleteCmd = &cobra.Command{
		Use: "delete ENVIRONMENT", Short: "Delete a notebook environment",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksEnvDelete,
	}
	s.notebooksEnvDescribeCmd = &cobra.Command{
		Use: "describe ENVIRONMENT", Short: "Describe a notebook environment",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksEnvDescribe,
	}
	s.notebooksEnvListCmd = &cobra.Command{
		Use: "list", Short: "List notebook environments",
		Args: cobra.NoArgs, RunE: s.runNotebooksEnvList,
	}
	s.notebooksInstCmd = &cobra.Command{Use: "instances", Short: "Manage notebook instances"}
	s.notebooksInstCreateCmd = &cobra.Command{
		Use: "create INSTANCE", Short: "Create a notebook instance",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksInstCreate,
	}
	s.notebooksInstDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete a notebook instance",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksInstDelete,
	}
	s.notebooksInstDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe a notebook instance",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksInstDescribe,
	}
	s.notebooksInstListCmd = &cobra.Command{
		Use: "list", Short: "List notebook instances",
		Args: cobra.NoArgs, RunE: s.runNotebooksInstList,
	}
	s.notebooksInstUpdateCmd = &cobra.Command{
		Use: "update INSTANCE", Short: "Update a notebook instance configuration",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksInstUpdate,
	}
	s.notebooksInstAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding INSTANCE", Short: "Add an IAM policy binding to a notebook instance",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksInstAddIam,
	}
	s.notebooksInstRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding INSTANCE", Short: "Remove an IAM policy binding from a notebook instance",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksInstRemoveIam,
	}
	s.notebooksInstGetIamCmd = &cobra.Command{
		Use: "get-iam-policy INSTANCE", Short: "Get the IAM policy for a notebook instance",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksInstGetIam,
	}
	s.notebooksInstSetIamCmd = &cobra.Command{
		Use: "set-iam-policy INSTANCE POLICY_FILE", Short: "Set the IAM policy for a notebook instance",
		Args: cobra.ExactArgs(2), RunE: s.runNotebooksInstSetIam,
	}
	s.notebooksInstDiagnoseCmd = &cobra.Command{
		Use: "diagnose INSTANCE", Short: "Run diagnostics for a notebook instance",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksInstDiagnose,
	}
	s.notebooksInstGetHealthCmd = &cobra.Command{
		Use: "get-health INSTANCE", Short: "Get the health state of a notebook instance",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksInstGetHealth,
	}
	s.notebooksInstIsUpgradeableCmd = &cobra.Command{
		Use: "is-upgradeable INSTANCE", Short: "Check whether a notebook instance can be upgraded",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksInstIsUpgradeable,
	}
	s.notebooksInstMigrateCmd = &cobra.Command{
		Use: "migrate INSTANCE", Short: "Migrate a notebook instance to the Workbench Instances API",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksInstMigrate,
	}
	s.notebooksInstRegisterCmd = &cobra.Command{
		Use: "register INSTANCE", Short: "Register a legacy notebook instance under Managed Notebooks",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksInstRegister,
	}
	s.notebooksInstResetCmd = &cobra.Command{
		Use: "reset INSTANCE", Short: "Reset a notebook instance",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksInstReset,
	}
	s.notebooksInstRollbackCmd = &cobra.Command{
		Use: "rollback INSTANCE", Short: "Rollback a notebook instance to a prior VM image",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksInstRollback,
	}
	s.notebooksInstStartCmd = &cobra.Command{
		Use: "start INSTANCE", Short: "Start a notebook instance",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksInstStart,
	}
	s.notebooksInstStopCmd = &cobra.Command{
		Use: "stop INSTANCE", Short: "Stop a notebook instance",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksInstStop,
	}
	s.notebooksInstUpgradeCmd = &cobra.Command{
		Use: "upgrade INSTANCE", Short: "Upgrade a notebook instance",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksInstUpgrade,
	}
	s.notebooksLocCmd = &cobra.Command{Use: "locations", Short: "View notebook locations"}
	s.notebooksLocListCmd = &cobra.Command{
		Use: "list", Short: "List notebook locations for the current project",
		Args: cobra.NoArgs, RunE: s.runNotebooksLocList,
	}
	s.notebooksRuntimeCmd = &cobra.Command{Use: "runtimes", Short: "Manage notebook runtimes"}
	s.notebooksRuntimeCreateCmd = &cobra.Command{
		Use: "create RUNTIME", Short: "Create a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksRuntimeCreate,
	}
	s.notebooksRuntimeDeleteCmd = &cobra.Command{
		Use: "delete RUNTIME", Short: "Delete a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksRuntimeDelete,
	}
	s.notebooksRuntimeDescribeCmd = &cobra.Command{
		Use: "describe RUNTIME", Short: "Describe a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksRuntimeDescribe,
	}
	s.notebooksRuntimeListCmd = &cobra.Command{
		Use: "list", Short: "List notebook runtimes",
		Args: cobra.NoArgs, RunE: s.runNotebooksRuntimeList,
	}
	s.notebooksRuntimeDiagnoseCmd = &cobra.Command{
		Use: "diagnose RUNTIME", Short: "Run diagnostics for a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksRuntimeDiagnose,
	}
	s.notebooksRuntimeMigrateCmd = &cobra.Command{
		Use: "migrate RUNTIME", Short: "Migrate a notebook runtime to the Workbench Instances API",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksRuntimeMigrate,
	}
	s.notebooksRuntimeResetCmd = &cobra.Command{
		Use: "reset RUNTIME", Short: "Reset a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksRuntimeReset,
	}
	s.notebooksRuntimeStartCmd = &cobra.Command{
		Use: "start RUNTIME", Short: "Start a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksRuntimeStart,
	}
	s.notebooksRuntimeStopCmd = &cobra.Command{
		Use: "stop RUNTIME", Short: "Stop a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksRuntimeStop,
	}
	s.notebooksRuntimeSwitchCmd = &cobra.Command{
		Use: "switch RUNTIME", Short: "Switch a notebook runtime to a different machine type or accelerator",
		Args: cobra.ExactArgs(1), RunE: s.runNotebooksRuntimeSwitch,
	}
	s.initNotebooks()
	s.initNotebooksEnvironments()
	s.initNotebooksInstances()
	s.initNotebooksLocations()
	s.initNotebooksRuntimes()
	return s
}
