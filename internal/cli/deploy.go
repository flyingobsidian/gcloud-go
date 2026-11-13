package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
)

// --- gcloud deploy (#327) ---

func (s *deployCmds) initDeploy() {
	for _, name := range []string{"apply", "delete", "get-config"} {
		registerStubCommand(s.deployCmd, name, "Not yet implemented")
	}

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

func (s *deployCmds) deployUpdatedIam(who string) {
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for %s.\n", who)
}

// deployCmds holds the commands and flag values of one invocation's "deploy"
// command group, so invocations don't share state (#2070).
type deployCmds struct {
	inv *invocation

	deployCmd                 *cobra.Command
	deployARCmd               *cobra.Command
	flagDeployARRegion        string
	flagDeployARPipeline      string
	flagDeployARFormat        string
	flagDeployARPageSize      int64
	deployARCancelCmd         *cobra.Command
	deployARDescribeCmd       *cobra.Command
	deployARListCmd           *cobra.Command
	deployAutoCmd             *cobra.Command
	flagDeployAutoRegion      string
	flagDeployAutoPipeline    string
	flagDeployAutoFormat      string
	flagDeployAutoPageSize    int64
	deployAutoDeleteCmd       *cobra.Command
	deployAutoDescribeCmd     *cobra.Command
	deployAutoExportCmd       *cobra.Command
	deployAutoListCmd         *cobra.Command
	deployCTTCmd              *cobra.Command
	flagDeployCTTRegion       string
	flagDeployCTTFormat       string
	flagDeployCTTPageSize     int64
	flagDeployCTTIamMember    string
	flagDeployCTTIamRole      string
	flagDeployCTTIamCondExpr  string
	flagDeployCTTIamCondT     string
	flagDeployCTTIamCondD     string
	flagDeployCTTIamAllCond   bool
	deployCTTDeleteCmd        *cobra.Command
	deployCTTDescribeCmd      *cobra.Command
	deployCTTExportCmd        *cobra.Command
	deployCTTListCmd          *cobra.Command
	deployCTTGetIamCmd        *cobra.Command
	deployCTTSetIamCmd        *cobra.Command
	deployCTTAddIamCmd        *cobra.Command
	deployCTTRemoveIamCmd     *cobra.Command
	deployDPCmd               *cobra.Command
	flagDeployDPRegion        string
	flagDeployDPFormat        string
	flagDeployDPForce         bool
	flagDeployDPPageSize      int64
	flagDeployDPIamMember     string
	flagDeployDPIamRole       string
	flagDeployDPIamCondExpr   string
	flagDeployDPIamCondT      string
	flagDeployDPIamCondD      string
	flagDeployDPIamAllCond    bool
	deployDPDeleteCmd         *cobra.Command
	deployDPDescribeCmd       *cobra.Command
	deployDPExportCmd         *cobra.Command
	deployDPListCmd           *cobra.Command
	deployDPGetIamCmd         *cobra.Command
	deployDPSetIamCmd         *cobra.Command
	deployDPAddIamCmd         *cobra.Command
	deployDPRemoveIamCmd      *cobra.Command
	deployDPolCmd             *cobra.Command
	flagDeployDPolRegion      string
	flagDeployDPolFormat      string
	flagDeployDPolIamMember   string
	flagDeployDPolIamRole     string
	flagDeployDPolIamCondExpr string
	flagDeployDPolIamCondT    string
	flagDeployDPolIamCondD    string
	flagDeployDPolIamAllCond  bool
	deployDPolDeleteCmd       *cobra.Command
	deployDPolDescribeCmd     *cobra.Command
	deployDPolExportCmd       *cobra.Command
	deployDPolGetIamCmd       *cobra.Command
	deployDPolSetIamCmd       *cobra.Command
	deployDPolAddIamCmd       *cobra.Command
	deployDPolRemoveIamCmd    *cobra.Command
	deployJRCmd               *cobra.Command
	flagDeployJRRegion        string
	flagDeployJRPipeline      string
	flagDeployJRRelease       string
	flagDeployJRRollout       string
	flagDeployJRFormat        string
	flagDeployJRPageSize      int64
	deployJRDescribeCmd       *cobra.Command
	deployJRListCmd           *cobra.Command
	deployJRTerminateCmd      *cobra.Command
	deployRelCmd              *cobra.Command
	flagDeployRelRegion       string
	flagDeployRelPipeline     string
	flagDeployRelFormat       string
	flagDeployRelConfigFile   string
	flagDeployRelPageSize     int64
	flagDeployRelOverride     []string
	flagDeployRelValidateOnly bool
	flagDeployRelPromoteTgt   string
	flagDeployRelPromoteID    string
	deployRelAbandonCmd       *cobra.Command
	deployRelCreateCmd        *cobra.Command
	deployRelDescribeCmd      *cobra.Command
	deployRelListCmd          *cobra.Command
	deployRelPromoteCmd       *cobra.Command
	deployRollCmd             *cobra.Command
	flagDeployRollRegion      string
	flagDeployRollPipeline    string
	flagDeployRollRelease     string
	flagDeployRollFormat      string
	flagDeployRollPageSize    int64
	flagDeployRollPhaseID     string
	flagDeployRollJobID       string
	flagDeployRollOverride    []string
	deployRollAdvanceCmd      *cobra.Command
	deployRollApproveCmd      *cobra.Command
	deployRollCancelCmd       *cobra.Command
	deployRollDescribeCmd     *cobra.Command
	deployRollListCmd         *cobra.Command
	deployRollRejectCmd       *cobra.Command
	deployRollIgnoreJobCmd    *cobra.Command
	deployRollRetryJobCmd     *cobra.Command
	deployTargetsCmd          *cobra.Command
	flagDeployTgtRegion       string
	flagDeployTgtFormat       string
	flagDeployTgtPageSize     int64
	// rollback/redeploy shared flags
	flagDeployTgtPipeline    string
	flagDeployTgtRelease     string
	flagDeployTgtRolloutID   string
	flagDeployTgtDescription string
	flagDeployTgtStartPhase  string
	flagDeployTgtOverridePol []string
	flagDeployTgtIamMember   string
	flagDeployTgtIamRole     string
	flagDeployTgtIamCondExpr string
	flagDeployTgtIamCondT    string
	flagDeployTgtIamCondD    string
	flagDeployTgtIamAllCond  bool
	deployTgtDeleteCmd       *cobra.Command
	deployTgtDescribeCmd     *cobra.Command
	deployTgtExportCmd       *cobra.Command
	deployTgtListCmd         *cobra.Command
	deployTgtRedeployCmd     *cobra.Command
	deployTgtRollbackCmd     *cobra.Command
	deployTgtGetIamCmd       *cobra.Command
	deployTgtSetIamCmd       *cobra.Command
	deployTgtAddIamCmd       *cobra.Command
	deployTgtRemoveIamCmd    *cobra.Command
}

// newDeployCmd returns the "deploy" command group for inv.
func newDeployCmd(inv *invocation) *cobra.Command {
	return newDeployCmds(inv).deployCmd
}

// newDeployCmds builds the "deploy" command group for inv and returns its state.
func newDeployCmds(inv *invocation) *deployCmds {
	s := &deployCmds{inv: inv}
	s.deployCmd = &cobra.Command{Use: "deploy", Short: "Manage Cloud Deploy"}
	s.deployARCmd = &cobra.Command{Use: "automation-runs", Short: "Manage Cloud Deploy automation runs"}
	s.deployARCancelCmd = &cobra.Command{
		Use: "cancel AUTOMATION_RUN", Short: "Cancel an automation run",
		Args: cobra.ExactArgs(1), RunE: s.runDeployARCancel,
	}
	s.deployARDescribeCmd = &cobra.Command{
		Use: "describe AUTOMATION_RUN", Short: "Describe an automation run",
		Args: cobra.ExactArgs(1), RunE: s.runDeployARDescribe,
	}
	s.deployARListCmd = &cobra.Command{
		Use: "list", Short: "List automation runs for a delivery pipeline",
		Args: cobra.NoArgs, RunE: s.runDeployARList,
	}
	s.deployAutoCmd = &cobra.Command{Use: "automations", Short: "Manage Cloud Deploy automations"}
	s.deployAutoDeleteCmd = &cobra.Command{
		Use: "delete AUTOMATION", Short: "Delete an automation",
		Args: cobra.ExactArgs(1), RunE: s.runDeployAutoDelete,
	}
	s.deployAutoDescribeCmd = &cobra.Command{
		Use: "describe AUTOMATION", Short: "Describe an automation",
		Args: cobra.ExactArgs(1), RunE: s.runDeployAutoDescribe,
	}
	s.deployAutoExportCmd = &cobra.Command{
		Use: "export AUTOMATION", Short: "Export an automation",
		Args: cobra.ExactArgs(1), RunE: s.runDeployAutoExport,
	}
	s.deployAutoListCmd = &cobra.Command{
		Use: "list", Short: "List automations for a delivery pipeline",
		Args: cobra.NoArgs, RunE: s.runDeployAutoList,
	}
	s.deployCTTCmd = &cobra.Command{Use: "custom-target-types", Short: "Manage Cloud Deploy custom target types"}
	s.deployCTTDeleteCmd = &cobra.Command{
		Use: "delete CUSTOM_TARGET_TYPE", Short: "Delete a custom target type",
		Args: cobra.ExactArgs(1), RunE: s.runDeployCTTDelete,
	}
	s.deployCTTDescribeCmd = &cobra.Command{
		Use: "describe CUSTOM_TARGET_TYPE", Short: "Describe a custom target type",
		Args: cobra.ExactArgs(1), RunE: s.runDeployCTTDescribe,
	}
	s.deployCTTExportCmd = &cobra.Command{
		Use: "export CUSTOM_TARGET_TYPE", Short: "Export a custom target type",
		Args: cobra.ExactArgs(1), RunE: s.runDeployCTTExport,
	}
	s.deployCTTListCmd = &cobra.Command{
		Use: "list", Short: "List custom target types",
		Args: cobra.NoArgs, RunE: s.runDeployCTTList,
	}
	s.deployCTTGetIamCmd = &cobra.Command{
		Use: "get-iam-policy CUSTOM_TARGET_TYPE", Short: "Get the IAM policy",
		Args: cobra.ExactArgs(1), RunE: s.runDeployCTTGetIam,
	}
	s.deployCTTSetIamCmd = &cobra.Command{
		Use: "set-iam-policy CUSTOM_TARGET_TYPE POLICY_FILE", Short: "Set the IAM policy",
		Args: cobra.ExactArgs(2), RunE: s.runDeployCTTSetIam,
	}
	s.deployCTTAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding CUSTOM_TARGET_TYPE", Short: "Add an IAM policy binding",
		Args: cobra.ExactArgs(1), RunE: s.runDeployCTTAddIam,
	}
	s.deployCTTRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding CUSTOM_TARGET_TYPE", Short: "Remove an IAM policy binding",
		Args: cobra.ExactArgs(1), RunE: s.runDeployCTTRemoveIam,
	}
	s.deployDPCmd = &cobra.Command{Use: "delivery-pipelines", Short: "Manage Cloud Deploy delivery pipelines"}
	s.deployDPDeleteCmd = &cobra.Command{
		Use: "delete PIPELINE", Short: "Delete a delivery pipeline",
		Args: cobra.ExactArgs(1), RunE: s.runDeployDPDelete,
	}
	s.deployDPDescribeCmd = &cobra.Command{
		Use: "describe PIPELINE", Short: "Describe a delivery pipeline",
		Args: cobra.ExactArgs(1), RunE: s.runDeployDPDescribe,
	}
	s.deployDPExportCmd = &cobra.Command{
		Use: "export PIPELINE", Short: "Export a delivery pipeline (YAML/JSON)",
		Args: cobra.ExactArgs(1), RunE: s.runDeployDPExport,
	}
	s.deployDPListCmd = &cobra.Command{
		Use: "list", Short: "List delivery pipelines",
		Args: cobra.NoArgs, RunE: s.runDeployDPList,
	}
	s.deployDPGetIamCmd = &cobra.Command{
		Use: "get-iam-policy PIPELINE", Short: "Get the IAM policy for a delivery pipeline",
		Args: cobra.ExactArgs(1), RunE: s.runDeployDPGetIam,
	}
	s.deployDPSetIamCmd = &cobra.Command{
		Use: "set-iam-policy PIPELINE POLICY_FILE", Short: "Set the IAM policy for a delivery pipeline",
		Args: cobra.ExactArgs(2), RunE: s.runDeployDPSetIam,
	}
	s.deployDPAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding PIPELINE", Short: "Add an IAM policy binding to a delivery pipeline",
		Args: cobra.ExactArgs(1), RunE: s.runDeployDPAddIam,
	}
	s.deployDPRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding PIPELINE", Short: "Remove an IAM policy binding from a delivery pipeline",
		Args: cobra.ExactArgs(1), RunE: s.runDeployDPRemoveIam,
	}
	s.deployDPolCmd = &cobra.Command{Use: "deploy-policies", Short: "Manage Cloud Deploy deploy policies"}
	s.deployDPolDeleteCmd = &cobra.Command{
		Use: "delete DEPLOY_POLICY", Short: "Delete a deploy policy",
		Args: cobra.ExactArgs(1), RunE: s.runDeployDPolDelete,
	}
	s.deployDPolDescribeCmd = &cobra.Command{
		Use: "describe DEPLOY_POLICY", Short: "Describe a deploy policy",
		Args: cobra.ExactArgs(1), RunE: s.runDeployDPolDescribe,
	}
	s.deployDPolExportCmd = &cobra.Command{
		Use: "export DEPLOY_POLICY", Short: "Export a deploy policy",
		Args: cobra.ExactArgs(1), RunE: s.runDeployDPolExport,
	}
	s.deployDPolGetIamCmd = &cobra.Command{
		Use: "get-iam-policy DEPLOY_POLICY", Short: "Get the IAM policy",
		Args: cobra.ExactArgs(1), RunE: s.runDeployDPolGetIam,
	}
	s.deployDPolSetIamCmd = &cobra.Command{
		Use: "set-iam-policy DEPLOY_POLICY POLICY_FILE", Short: "Set the IAM policy",
		Args: cobra.ExactArgs(2), RunE: s.runDeployDPolSetIam,
	}
	s.deployDPolAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding DEPLOY_POLICY", Short: "Add an IAM policy binding",
		Args: cobra.ExactArgs(1), RunE: s.runDeployDPolAddIam,
	}
	s.deployDPolRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding DEPLOY_POLICY", Short: "Remove an IAM policy binding",
		Args: cobra.ExactArgs(1), RunE: s.runDeployDPolRemoveIam,
	}
	s.deployJRCmd = &cobra.Command{Use: "job-runs", Short: "Manage Cloud Deploy job runs"}
	s.deployJRDescribeCmd = &cobra.Command{
		Use: "describe JOB_RUN", Short: "Describe a job run",
		Args: cobra.ExactArgs(1), RunE: s.runDeployJRDescribe,
	}
	s.deployJRListCmd = &cobra.Command{
		Use: "list", Short: "List job runs for a rollout",
		Args: cobra.NoArgs, RunE: s.runDeployJRList,
	}
	s.deployJRTerminateCmd = &cobra.Command{
		Use: "terminate JOB_RUN", Short: "Terminate a running job run",
		Args: cobra.ExactArgs(1), RunE: s.runDeployJRTerminate,
	}
	s.deployRelCmd = &cobra.Command{Use: "releases", Short: "Manage Cloud Deploy releases"}
	s.deployRelAbandonCmd = &cobra.Command{
		Use: "abandon RELEASE", Short: "Abandon a release",
		Args: cobra.ExactArgs(1), RunE: s.runDeployRelAbandon,
	}
	s.deployRelCreateCmd = &cobra.Command{
		Use: "create RELEASE", Short: "Create a release (body from --config-file)",
		Args: cobra.ExactArgs(1), RunE: s.runDeployRelCreate,
	}
	s.deployRelDescribeCmd = &cobra.Command{
		Use: "describe RELEASE", Short: "Describe a release",
		Args: cobra.ExactArgs(1), RunE: s.runDeployRelDescribe,
	}
	s.deployRelListCmd = &cobra.Command{
		Use: "list", Short: "List releases",
		Args: cobra.NoArgs, RunE: s.runDeployRelList,
	}
	s.deployRelPromoteCmd = &cobra.Command{
		Use: "promote RELEASE", Short: "Promote a release (create a rollout targeting a new target)",
		Args: cobra.ExactArgs(1), RunE: s.runDeployRelPromote,
	}
	s.deployRollCmd = &cobra.Command{Use: "rollouts", Short: "Manage Cloud Deploy rollouts"}
	s.deployRollAdvanceCmd = &cobra.Command{
		Use: "advance ROLLOUT", Short: "Advance a rollout to the specified phase",
		Args: cobra.ExactArgs(1), RunE: s.runDeployRollAdvance,
	}
	s.deployRollApproveCmd = &cobra.Command{
		Use: "approve ROLLOUT", Short: "Approve a rollout",
		Args: cobra.ExactArgs(1), RunE: s.runDeployRollApprove,
	}
	s.deployRollCancelCmd = &cobra.Command{
		Use: "cancel ROLLOUT", Short: "Cancel a rollout",
		Args: cobra.ExactArgs(1), RunE: s.runDeployRollCancel,
	}
	s.deployRollDescribeCmd = &cobra.Command{
		Use: "describe ROLLOUT", Short: "Describe a rollout",
		Args: cobra.ExactArgs(1), RunE: s.runDeployRollDescribe,
	}
	s.deployRollListCmd = &cobra.Command{
		Use: "list", Short: "List rollouts for a release",
		Args: cobra.NoArgs, RunE: s.runDeployRollList,
	}
	s.deployRollRejectCmd = &cobra.Command{
		Use: "reject ROLLOUT", Short: "Reject a pending rollout",
		Args: cobra.ExactArgs(1), RunE: s.runDeployRollReject,
	}
	s.deployRollIgnoreJobCmd = &cobra.Command{
		Use: "ignore-job ROLLOUT", Short: "Ignore a job in a rollout phase",
		Args: cobra.ExactArgs(1), RunE: s.runDeployRollIgnoreJob,
	}
	s.deployRollRetryJobCmd = &cobra.Command{
		Use: "retry-job ROLLOUT", Short: "Retry a failed job in a rollout phase",
		Args: cobra.ExactArgs(1), RunE: s.runDeployRollRetryJob,
	}
	s.deployTargetsCmd = &cobra.Command{Use: "targets", Short: "Manage Cloud Deploy targets"}
	s.deployTgtDeleteCmd = &cobra.Command{
		Use: "delete TARGET", Short: "Delete a target",
		Args: cobra.ExactArgs(1), RunE: s.runDeployTgtDelete,
	}
	s.deployTgtDescribeCmd = &cobra.Command{
		Use: "describe TARGET", Short: "Describe a target",
		Args: cobra.ExactArgs(1), RunE: s.runDeployTgtDescribe,
	}
	s.deployTgtExportCmd = &cobra.Command{
		Use: "export TARGET", Short: "Export a target",
		Args: cobra.ExactArgs(1), RunE: s.runDeployTgtExport,
	}
	s.deployTgtListCmd = &cobra.Command{
		Use: "list", Short: "List targets",
		Args: cobra.NoArgs, RunE: s.runDeployTgtList,
	}
	s.deployTgtRedeployCmd = &cobra.Command{
		Use:   "redeploy TARGET",
		Short: "Redeploy the most recent release for a target by creating a new rollout",
		Args:  cobra.ExactArgs(1), RunE: s.runDeployTgtRedeploy,
	}
	s.deployTgtRollbackCmd = &cobra.Command{
		Use:   "rollback TARGET",
		Short: "Roll a target back to a previous release",
		Args:  cobra.ExactArgs(1), RunE: s.runDeployTgtRollback,
	}
	s.deployTgtGetIamCmd = &cobra.Command{
		Use: "get-iam-policy TARGET", Short: "Get the IAM policy",
		Args: cobra.ExactArgs(1), RunE: s.runDeployTgtGetIam,
	}
	s.deployTgtSetIamCmd = &cobra.Command{
		Use: "set-iam-policy TARGET POLICY_FILE", Short: "Set the IAM policy",
		Args: cobra.ExactArgs(2), RunE: s.runDeployTgtSetIam,
	}
	s.deployTgtAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding TARGET", Short: "Add an IAM policy binding",
		Args: cobra.ExactArgs(1), RunE: s.runDeployTgtAddIam,
	}
	s.deployTgtRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding TARGET", Short: "Remove an IAM policy binding",
		Args: cobra.ExactArgs(1), RunE: s.runDeployTgtRemoveIam,
	}
	s.initDeploy()
	s.initDeployAutomationRuns()
	s.initDeployAutomations()
	s.initDeployCustomTargetTypes()
	s.initDeployDeliveryPipelines()
	s.initDeployDeployPolicies()
	s.initDeployJobRuns()
	s.initDeployReleases()
	s.initDeployRollouts()
	s.initDeployTargets()
	return s
}
