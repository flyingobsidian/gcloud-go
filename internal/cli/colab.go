package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud colab (#316) ---
//
// Colab Enterprise is served by the aiplatform API surface. Every subgroup
// accepts a --region flag; the aiplatform client is regional (endpoint of
// the form https://<region>-aiplatform.googleapis.com/).

func (s *colabCmds) initColab() {

}

func (s *colabCmds) colabParent(region string) (string, error) {
	project, err := s.inv.project()
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

func (s *colabCmds) writeUpdatedIam(who string) {
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for %s.\n", who)
}

// colabCmds holds the commands and flag values of one invocation's "colab"
// command group, so invocations don't share state (#2070).
type colabCmds struct {
	inv *invocation

	colabCmd                 *cobra.Command
	colabExecCmd             *cobra.Command
	flagColabExecRegion      string
	flagColabExecFormat      string
	flagColabExecConfigFile  string
	flagColabExecFilter      string
	flagColabExecOrderBy     string
	flagColabExecPageSize    int64
	flagColabExecView        string
	colabExecCreateCmd       *cobra.Command
	colabExecDeleteCmd       *cobra.Command
	colabExecDescribeCmd     *cobra.Command
	colabExecListCmd         *cobra.Command
	colabRTTCmd              *cobra.Command
	flagColabRTTRegion       string
	flagColabRTTFormat       string
	flagColabRTTConfigFile   string
	flagColabRTTFilter       string
	flagColabRTTOrderBy      string
	flagColabRTTPageSize     int64
	flagColabRTTReadMask     string
	flagColabRTTIamMember    string
	flagColabRTTIamRole      string
	flagColabRTTIamCondExpr  string
	flagColabRTTIamCondT     string
	flagColabRTTIamCondD     string
	flagColabRTTIamAllCond   bool
	colabRTTCreateCmd        *cobra.Command
	colabRTTDeleteCmd        *cobra.Command
	colabRTTDescribeCmd      *cobra.Command
	colabRTTListCmd          *cobra.Command
	colabRTTGetIamCmd        *cobra.Command
	colabRTTSetIamCmd        *cobra.Command
	colabRTTAddIamCmd        *cobra.Command
	colabRTTRemoveIamCmd     *cobra.Command
	colabRTCmd               *cobra.Command
	flagColabRTRegion        string
	flagColabRTFormat        string
	flagColabRTConfigFile    string
	flagColabRTFilter        string
	flagColabRTOrderBy       string
	flagColabRTPageSize      int64
	flagColabRTReadMask      string
	flagColabRTVersionID     string
	colabRTCreateCmd         *cobra.Command
	colabRTDeleteCmd         *cobra.Command
	colabRTDescribeCmd       *cobra.Command
	colabRTListCmd           *cobra.Command
	colabRTStartCmd          *cobra.Command
	colabRTStopCmd           *cobra.Command
	colabRTUpgradeCmd        *cobra.Command
	colabSchedCmd            *cobra.Command
	flagColabSchedRegion     string
	flagColabSchedFormat     string
	flagColabSchedConfigFile string
	flagColabSchedUpdateMask string
	flagColabSchedFilter     string
	flagColabSchedOrderBy    string
	flagColabSchedPageSize   int64
	colabSchedCreateCmd      *cobra.Command
	colabSchedDeleteCmd      *cobra.Command
	colabSchedDescribeCmd    *cobra.Command
	colabSchedListCmd        *cobra.Command
	colabSchedUpdateCmd      *cobra.Command
	colabSchedPauseCmd       *cobra.Command
	colabSchedResumeCmd      *cobra.Command
}

// newColabCmd returns the "colab" command group for inv.
func newColabCmd(inv *invocation) *cobra.Command {
	return newColabCmds(inv).colabCmd
}

// newColabCmds builds the "colab" command group for inv and returns its state.
func newColabCmds(inv *invocation) *colabCmds {
	s := &colabCmds{inv: inv}
	s.colabCmd = &cobra.Command{Use: "colab", Short: "Manage Colab Enterprise"}
	s.colabExecCmd = &cobra.Command{Use: "executions", Short: "Manage Colab Enterprise notebook executions"}
	s.colabExecCreateCmd = &cobra.Command{
		Use: "create EXECUTION", Short: "Create a notebook execution job",
		Args: cobra.ExactArgs(1), RunE: s.runColabExecCreate,
	}
	s.colabExecDeleteCmd = &cobra.Command{
		Use: "delete EXECUTION", Short: "Delete a notebook execution job",
		Args: cobra.ExactArgs(1), RunE: s.runColabExecDelete,
	}
	s.colabExecDescribeCmd = &cobra.Command{
		Use: "describe EXECUTION", Short: "Describe a notebook execution job",
		Args: cobra.ExactArgs(1), RunE: s.runColabExecDescribe,
	}
	s.colabExecListCmd = &cobra.Command{
		Use: "list", Short: "List notebook execution jobs",
		Args: cobra.NoArgs, RunE: s.runColabExecList,
	}
	s.colabRTTCmd = &cobra.Command{Use: "runtime-templates", Short: "Manage Colab Enterprise runtime templates"}
	s.colabRTTCreateCmd = &cobra.Command{
		Use: "create RUNTIME_TEMPLATE", Short: "Create a runtime template",
		Args: cobra.ExactArgs(1), RunE: s.runColabRTTCreate,
	}
	s.colabRTTDeleteCmd = &cobra.Command{
		Use: "delete RUNTIME_TEMPLATE", Short: "Delete a runtime template",
		Args: cobra.ExactArgs(1), RunE: s.runColabRTTDelete,
	}
	s.colabRTTDescribeCmd = &cobra.Command{
		Use: "describe RUNTIME_TEMPLATE", Short: "Describe a runtime template",
		Args: cobra.ExactArgs(1), RunE: s.runColabRTTDescribe,
	}
	s.colabRTTListCmd = &cobra.Command{
		Use: "list", Short: "List runtime templates",
		Args: cobra.NoArgs, RunE: s.runColabRTTList,
	}
	s.colabRTTGetIamCmd = &cobra.Command{
		Use: "get-iam-policy RUNTIME_TEMPLATE", Short: "Get the IAM policy for a runtime template",
		Args: cobra.ExactArgs(1), RunE: s.runColabRTTGetIam,
	}
	s.colabRTTSetIamCmd = &cobra.Command{
		Use:   "set-iam-policy RUNTIME_TEMPLATE POLICY_FILE",
		Short: "Set the IAM policy for a runtime template",
		Args:  cobra.ExactArgs(2), RunE: s.runColabRTTSetIam,
	}
	s.colabRTTAddIamCmd = &cobra.Command{
		Use:   "add-iam-policy-binding RUNTIME_TEMPLATE",
		Short: "Add an IAM policy binding to a runtime template",
		Args:  cobra.ExactArgs(1), RunE: s.runColabRTTAddIam,
	}
	s.colabRTTRemoveIamCmd = &cobra.Command{
		Use:   "remove-iam-policy-binding RUNTIME_TEMPLATE",
		Short: "Remove an IAM policy binding from a runtime template",
		Args:  cobra.ExactArgs(1), RunE: s.runColabRTTRemoveIam,
	}
	s.colabRTCmd = &cobra.Command{Use: "runtimes", Short: "Manage Colab Enterprise runtimes"}
	s.colabRTCreateCmd = &cobra.Command{
		Use: "create", Short: "Create (assign) a Colab Enterprise notebook runtime",
		Args: cobra.NoArgs, RunE: s.runColabRTCreate,
	}
	s.colabRTDeleteCmd = &cobra.Command{
		Use: "delete RUNTIME", Short: "Delete a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: s.runColabRTDelete,
	}
	s.colabRTDescribeCmd = &cobra.Command{
		Use: "describe RUNTIME", Short: "Describe a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: s.runColabRTDescribe,
	}
	s.colabRTListCmd = &cobra.Command{
		Use: "list", Short: "List notebook runtimes",
		Args: cobra.NoArgs, RunE: s.runColabRTList,
	}
	s.colabRTStartCmd = &cobra.Command{
		Use: "start RUNTIME", Short: "Start a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: s.runColabRTStart,
	}
	s.colabRTStopCmd = &cobra.Command{
		Use: "stop RUNTIME", Short: "Stop a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: s.runColabRTStop,
	}
	s.colabRTUpgradeCmd = &cobra.Command{
		Use: "upgrade RUNTIME", Short: "Upgrade a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: s.runColabRTUpgrade,
	}
	s.colabSchedCmd = &cobra.Command{Use: "schedules", Short: "Manage Colab Enterprise execution schedules"}
	s.colabSchedCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a schedule",
		Args: cobra.NoArgs, RunE: s.runColabSchedCreate,
	}
	s.colabSchedDeleteCmd = &cobra.Command{
		Use: "delete SCHEDULE", Short: "Delete a schedule",
		Args: cobra.ExactArgs(1), RunE: s.runColabSchedDelete,
	}
	s.colabSchedDescribeCmd = &cobra.Command{
		Use: "describe SCHEDULE", Short: "Describe a schedule",
		Args: cobra.ExactArgs(1), RunE: s.runColabSchedDescribe,
	}
	s.colabSchedListCmd = &cobra.Command{
		Use: "list", Short: "List schedules",
		Args: cobra.NoArgs, RunE: s.runColabSchedList,
	}
	s.colabSchedUpdateCmd = &cobra.Command{
		Use: "update SCHEDULE", Short: "Update a schedule",
		Args: cobra.ExactArgs(1), RunE: s.runColabSchedUpdate,
	}
	s.colabSchedPauseCmd = &cobra.Command{
		Use: "pause SCHEDULE", Short: "Pause a schedule",
		Args: cobra.ExactArgs(1), RunE: s.runColabSchedPause,
	}
	s.colabSchedResumeCmd = &cobra.Command{
		Use: "resume SCHEDULE", Short: "Resume a schedule",
		Args: cobra.ExactArgs(1), RunE: s.runColabSchedResume,
	}
	s.initColab()
	s.initColabExecutions()
	s.initColabRuntimeTemplates()
	s.initColabRuntimes()
	s.initColabSchedules()
	return s
}
