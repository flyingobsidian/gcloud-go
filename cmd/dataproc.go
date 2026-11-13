package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	dataproc "google.golang.org/api/dataproc/v1"
)

// --- gcloud dataproc (#324) ---

func (s *dataprocCmds) initDataproc() {

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

func (s *dataprocCmds) dpUpdatedIam(who string) {
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for %s.\n", who)
}

// dataprocCmds holds the commands and flag values of one invocation's "dataproc"
// command group, so invocations don't share state (#2070).
type dataprocCmds struct {
	inv *invocation

	dataprocCmd                *cobra.Command
	dpASPCmd                   *cobra.Command
	flagDPASPRegion            string
	flagDPASPFormat            string
	flagDPASPConfigFile        string
	flagDPASPPageSize          int64
	dpASPDeleteCmd             *cobra.Command
	dpASPDescribeCmd           *cobra.Command
	dpASPExportCmd             *cobra.Command
	dpASPImportCmd             *cobra.Command
	dpASPListCmd               *cobra.Command
	dpASPGetIamCmd             *cobra.Command
	dpASPSetIamCmd             *cobra.Command
	dpBatchesCmd               *cobra.Command
	flagDPBatchRegion          string
	flagDPBatchFormat          string
	flagDPBatchConfigFile      string
	flagDPBatchRequestID       string
	flagDPBatchFilter          string
	flagDPBatchOrderBy         string
	flagDPBatchPageSize        int64
	flagDPBatchWaitTO          time.Duration
	dpBatchCancelCmd           *cobra.Command
	dpBatchDeleteCmd           *cobra.Command
	dpBatchDescribeCmd         *cobra.Command
	dpBatchListCmd             *cobra.Command
	dpBatchSubmitCmd           *cobra.Command
	dpBatchWaitCmd             *cobra.Command
	dpClustersCmd              *cobra.Command
	dpClustersGKECmd           *cobra.Command
	flagDPClusterRegion        string
	flagDPClusterFormat        string
	flagDPClusterConfigFile    string
	flagDPClusterUpdateMask    string
	flagDPClusterFilter        string
	flagDPClusterPageSize      int64
	flagDPClusterRequestID     string
	flagDPClusterGracefulTO    string
	flagDPClusterDiagnoseJobs  string
	flagDPClusterDiagnoseYARN  string
	flagDPClusterDiagnoseTS    string
	dpClusterCreateCmd         *cobra.Command
	dpClusterDeleteCmd         *cobra.Command
	dpClusterDescribeCmd       *cobra.Command
	dpClusterDiagnoseCmd       *cobra.Command
	dpClusterExportCmd         *cobra.Command
	dpClusterImportCmd         *cobra.Command
	dpClusterListCmd           *cobra.Command
	dpClusterStartCmd          *cobra.Command
	dpClusterStopCmd           *cobra.Command
	dpClusterUpdateCmd         *cobra.Command
	dpClusterGetIamCmd         *cobra.Command
	dpClusterSetIamCmd         *cobra.Command
	dpClusterGKECreateCmd      *cobra.Command
	dpJobsCmd                  *cobra.Command
	flagDPJobRegion            string
	flagDPJobFormat            string
	flagDPJobConfigFile        string
	flagDPJobUpdateMask        string
	flagDPJobRequestID         string
	flagDPJobCluster           string
	flagDPJobFilter            string
	flagDPJobStateMatch        string
	flagDPJobPageSize          int64
	flagDPJobWaitTimeout       time.Duration
	dpJobDeleteCmd             *cobra.Command
	dpJobDescribeCmd           *cobra.Command
	dpJobKillCmd               *cobra.Command
	dpJobListCmd               *cobra.Command
	dpJobSubmitCmd             *cobra.Command
	dpJobUpdateCmd             *cobra.Command
	dpJobWaitCmd               *cobra.Command
	dpJobGetIamCmd             *cobra.Command
	dpJobSetIamCmd             *cobra.Command
	dpNGCmd                    *cobra.Command
	flagDPNGRegion             string
	flagDPNGCluster            string
	flagDPNGFormat             string
	flagDPNGSize               int64
	flagDPNGGraceful           string
	flagDPNGRequestID          string
	dpNGDescribeCmd            *cobra.Command
	dpNGResizeCmd              *cobra.Command
	dpOpCmd                    *cobra.Command
	flagDPOpRegion             string
	flagDPOpFormat             string
	flagDPOpFilter             string
	flagDPOpPageSize           int64
	dpOpCancelCmd              *cobra.Command
	dpOpDeleteCmd              *cobra.Command
	dpOpDescribeCmd            *cobra.Command
	dpOpListCmd                *cobra.Command
	dpOpGetIamCmd              *cobra.Command
	dpOpSetIamCmd              *cobra.Command
	dpWFCmd                    *cobra.Command
	flagDPWFRegion             string
	flagDPWFFormat             string
	flagDPWFConfigFile         string
	flagDPWFPageSize           int64
	flagDPWFParams             map[string]string
	flagDPWFRequestID          string
	flagDPWFAddJobFile         string
	flagDPWFSelectorFile       string
	flagDPWFManagedFile        string
	flagDPWFDagTimeout         string
	flagDPWFStepID             string
	flagDPWFVersion            int64
	dpWFCreateCmd              *cobra.Command
	dpWFDeleteCmd              *cobra.Command
	dpWFDescribeCmd            *cobra.Command
	dpWFExportCmd              *cobra.Command
	dpWFImportCmd              *cobra.Command
	dpWFListCmd                *cobra.Command
	dpWFGetIamCmd              *cobra.Command
	dpWFSetIamCmd              *cobra.Command
	dpWFInstantiateCmd         *cobra.Command
	dpWFInstantiateFromFileCmd *cobra.Command
	dpWFAddJobCmd              *cobra.Command
	dpWFRemoveJobCmd           *cobra.Command
	dpWFSetClusterSelectorCmd  *cobra.Command
	dpWFSetManagedClusterCmd   *cobra.Command
	dpWFSetDagTimeoutCmd       *cobra.Command
	dpWFRemoveDagTimeoutCmd    *cobra.Command
}

// newDataprocCmd returns the "dataproc" command group for inv.
func newDataprocCmd(inv *invocation) *cobra.Command {
	return newDataprocCmds(inv).dataprocCmd
}

// newDataprocCmds builds the "dataproc" command group for inv and returns its state.
func newDataprocCmds(inv *invocation) *dataprocCmds {
	s := &dataprocCmds{inv: inv}
	s.dataprocCmd = &cobra.Command{Use: "dataproc", Short: "Manage Dataproc"}
	s.dpASPCmd = &cobra.Command{Use: "autoscaling-policies", Short: "Manage Dataproc autoscaling policies"}
	s.dpASPDeleteCmd = &cobra.Command{
		Use: "delete POLICY", Short: "Delete an autoscaling policy",
		Args: cobra.ExactArgs(1), RunE: s.runDPASPDelete,
	}
	s.dpASPDescribeCmd = &cobra.Command{
		Use: "describe POLICY", Short: "Describe an autoscaling policy",
		Args: cobra.ExactArgs(1), RunE: s.runDPASPDescribe,
	}
	s.dpASPExportCmd = &cobra.Command{
		Use: "export POLICY", Short: "Export an autoscaling policy (returns JSON)",
		Args: cobra.ExactArgs(1), RunE: s.runDPASPExport,
	}
	s.dpASPImportCmd = &cobra.Command{
		Use: "import POLICY", Short: "Import (create or update) an autoscaling policy from a YAML/JSON file",
		Args: cobra.ExactArgs(1), RunE: s.runDPASPImport,
	}
	s.dpASPListCmd = &cobra.Command{
		Use: "list", Short: "List autoscaling policies",
		Args: cobra.NoArgs, RunE: s.runDPASPList,
	}
	s.dpASPGetIamCmd = &cobra.Command{
		Use: "get-iam-policy POLICY", Short: "Get the IAM policy for an autoscaling policy",
		Args: cobra.ExactArgs(1), RunE: s.runDPASPGetIam,
	}
	s.dpASPSetIamCmd = &cobra.Command{
		Use: "set-iam-policy POLICY POLICY_FILE", Short: "Set the IAM policy for an autoscaling policy",
		Args: cobra.ExactArgs(2), RunE: s.runDPASPSetIam,
	}
	s.dpBatchesCmd = &cobra.Command{Use: "batches", Short: "Manage Dataproc serverless batches"}
	s.dpBatchCancelCmd = &cobra.Command{
		Use: "cancel BATCH", Short: "Cancel a batch (via its underlying operation)",
		Args: cobra.ExactArgs(1), RunE: s.runDPBatchCancel,
	}
	s.dpBatchDeleteCmd = &cobra.Command{
		Use: "delete BATCH", Short: "Delete a batch",
		Args: cobra.ExactArgs(1), RunE: s.runDPBatchDelete,
	}
	s.dpBatchDescribeCmd = &cobra.Command{
		Use: "describe BATCH", Short: "Describe a batch",
		Args: cobra.ExactArgs(1), RunE: s.runDPBatchDescribe,
	}
	s.dpBatchListCmd = &cobra.Command{
		Use: "list", Short: "List batches",
		Args: cobra.NoArgs, RunE: s.runDPBatchList,
	}
	s.dpBatchSubmitCmd = &cobra.Command{
		Use: "submit BATCH", Short: "Submit a batch (body loaded from --config-file)",
		Args: cobra.ExactArgs(1), RunE: s.runDPBatchSubmit,
	}
	s.dpBatchWaitCmd = &cobra.Command{
		Use: "wait BATCH", Short: "Wait for a batch to reach a terminal state",
		Args: cobra.ExactArgs(1), RunE: s.runDPBatchWait,
	}
	s.dpClustersCmd = &cobra.Command{Use: "clusters", Short: "Manage Dataproc clusters"}
	s.dpClustersGKECmd = &cobra.Command{Use: "gke", Short: "Manage GKE-based Dataproc clusters"}
	s.dpClusterCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a Dataproc cluster",
		Args: cobra.ExactArgs(1), RunE: s.runDPClusterCreate,
	}
	s.dpClusterDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a Dataproc cluster",
		Args: cobra.ExactArgs(1), RunE: s.runDPClusterDelete,
	}
	s.dpClusterDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a Dataproc cluster",
		Args: cobra.ExactArgs(1), RunE: s.runDPClusterDescribe,
	}
	s.dpClusterDiagnoseCmd = &cobra.Command{
		Use: "diagnose CLUSTER", Short: "Diagnose a Dataproc cluster",
		Args: cobra.ExactArgs(1), RunE: s.runDPClusterDiagnose,
	}
	s.dpClusterExportCmd = &cobra.Command{
		Use: "export CLUSTER", Short: "Export a Dataproc cluster (returns YAML/JSON)",
		Args: cobra.ExactArgs(1), RunE: s.runDPClusterExport,
	}
	s.dpClusterImportCmd = &cobra.Command{
		Use: "import CLUSTER", Short: "Import (create or update) a cluster from a YAML/JSON file",
		Args: cobra.ExactArgs(1), RunE: s.runDPClusterImport,
	}
	s.dpClusterListCmd = &cobra.Command{
		Use: "list", Short: "List Dataproc clusters",
		Args: cobra.NoArgs, RunE: s.runDPClusterList,
	}
	s.dpClusterStartCmd = &cobra.Command{
		Use: "start CLUSTER", Short: "Start a stopped Dataproc cluster",
		Args: cobra.ExactArgs(1), RunE: s.runDPClusterStart,
	}
	s.dpClusterStopCmd = &cobra.Command{
		Use: "stop CLUSTER", Short: "Stop a Dataproc cluster",
		Args: cobra.ExactArgs(1), RunE: s.runDPClusterStop,
	}
	s.dpClusterUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update a Dataproc cluster",
		Args: cobra.ExactArgs(1), RunE: s.runDPClusterUpdate,
	}
	s.dpClusterGetIamCmd = &cobra.Command{
		Use: "get-iam-policy CLUSTER", Short: "Get the IAM policy for a Dataproc cluster",
		Args: cobra.ExactArgs(1), RunE: s.runDPClusterGetIam,
	}
	s.dpClusterSetIamCmd = &cobra.Command{
		Use: "set-iam-policy CLUSTER POLICY_FILE", Short: "Set the IAM policy for a Dataproc cluster",
		Args: cobra.ExactArgs(2), RunE: s.runDPClusterSetIam,
	}
	s.dpClusterGKECreateCmd = &cobra.Command{
		Use:   "create CLUSTER",
		Short: "Create a GKE-based Dataproc cluster (--config-file must set virtualClusterConfig)",
		Args:  cobra.ExactArgs(1), RunE: s.runDPClusterCreate,
	}
	s.dpJobsCmd = &cobra.Command{Use: "jobs", Short: "Manage Dataproc jobs"}
	s.dpJobDeleteCmd = &cobra.Command{
		Use: "delete JOB", Short: "Delete a Dataproc job",
		Args: cobra.ExactArgs(1), RunE: s.runDPJobDelete,
	}
	s.dpJobDescribeCmd = &cobra.Command{
		Use: "describe JOB", Short: "Describe a Dataproc job",
		Args: cobra.ExactArgs(1), RunE: s.runDPJobDescribe,
	}
	s.dpJobKillCmd = &cobra.Command{
		Use: "kill JOB", Short: "Kill (cancel) a Dataproc job",
		Args: cobra.ExactArgs(1), RunE: s.runDPJobKill,
	}
	s.dpJobListCmd = &cobra.Command{
		Use: "list", Short: "List Dataproc jobs",
		Args: cobra.NoArgs, RunE: s.runDPJobList,
	}
	s.dpJobSubmitCmd = &cobra.Command{
		Use: "submit", Short: "Submit a Dataproc job (body loaded from --config-file)",
		Args: cobra.NoArgs, RunE: s.runDPJobSubmit,
	}
	s.dpJobUpdateCmd = &cobra.Command{
		Use: "update JOB", Short: "Update a Dataproc job",
		Args: cobra.ExactArgs(1), RunE: s.runDPJobUpdate,
	}
	s.dpJobWaitCmd = &cobra.Command{
		Use: "wait JOB", Short: "Wait for a job to reach a terminal state",
		Args: cobra.ExactArgs(1), RunE: s.runDPJobWait,
	}
	s.dpJobGetIamCmd = &cobra.Command{
		Use: "get-iam-policy JOB", Short: "Get the IAM policy for a Dataproc job",
		Args: cobra.ExactArgs(1), RunE: s.runDPJobGetIam,
	}
	s.dpJobSetIamCmd = &cobra.Command{
		Use: "set-iam-policy JOB POLICY_FILE", Short: "Set the IAM policy for a Dataproc job",
		Args: cobra.ExactArgs(2), RunE: s.runDPJobSetIam,
	}
	s.dpNGCmd = &cobra.Command{Use: "node-groups", Short: "Manage Dataproc cluster node groups"}
	s.dpNGDescribeCmd = &cobra.Command{
		Use: "describe NODE_GROUP", Short: "Describe a Dataproc node group",
		Args: cobra.ExactArgs(1), RunE: s.runDPNGDescribe,
	}
	s.dpNGResizeCmd = &cobra.Command{
		Use: "resize NODE_GROUP", Short: "Resize a Dataproc node group",
		Args: cobra.ExactArgs(1), RunE: s.runDPNGResize,
	}
	s.dpOpCmd = &cobra.Command{Use: "operations", Short: "Manage Dataproc operations"}
	s.dpOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Dataproc operation",
		Args: cobra.ExactArgs(1), RunE: s.runDPOpCancel,
	}
	s.dpOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a Dataproc operation record",
		Args: cobra.ExactArgs(1), RunE: s.runDPOpDelete,
	}
	s.dpOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Dataproc operation",
		Args: cobra.ExactArgs(1), RunE: s.runDPOpDescribe,
	}
	s.dpOpListCmd = &cobra.Command{
		Use: "list", Short: "List Dataproc operations",
		Args: cobra.NoArgs, RunE: s.runDPOpList,
	}
	s.dpOpGetIamCmd = &cobra.Command{
		Use: "get-iam-policy OPERATION", Short: "Get the IAM policy for a Dataproc operation",
		Args: cobra.ExactArgs(1), RunE: s.runDPOpGetIam,
	}
	s.dpOpSetIamCmd = &cobra.Command{
		Use: "set-iam-policy OPERATION POLICY_FILE", Short: "Set the IAM policy for a Dataproc operation",
		Args: cobra.ExactArgs(2), RunE: s.runDPOpSetIam,
	}
	s.dpWFCmd = &cobra.Command{Use: "workflow-templates", Short: "Manage Dataproc workflow templates"}
	s.dpWFCreateCmd = &cobra.Command{
		Use: "create TEMPLATE", Short: "Create a workflow template",
		Args: cobra.ExactArgs(1), RunE: s.runDPWFCreate,
	}
	s.dpWFDeleteCmd = &cobra.Command{
		Use: "delete TEMPLATE", Short: "Delete a workflow template",
		Args: cobra.ExactArgs(1), RunE: s.runDPWFDelete,
	}
	s.dpWFDescribeCmd = &cobra.Command{
		Use: "describe TEMPLATE", Short: "Describe a workflow template",
		Args: cobra.ExactArgs(1), RunE: s.runDPWFDescribe,
	}
	s.dpWFExportCmd = &cobra.Command{
		Use: "export TEMPLATE", Short: "Export a workflow template (YAML/JSON)",
		Args: cobra.ExactArgs(1), RunE: s.runDPWFExport,
	}
	s.dpWFImportCmd = &cobra.Command{
		Use: "import TEMPLATE", Short: "Import (create or update) a workflow template from a file",
		Args: cobra.ExactArgs(1), RunE: s.runDPWFImport,
	}
	s.dpWFListCmd = &cobra.Command{
		Use: "list", Short: "List workflow templates",
		Args: cobra.NoArgs, RunE: s.runDPWFList,
	}
	s.dpWFGetIamCmd = &cobra.Command{
		Use: "get-iam-policy TEMPLATE", Short: "Get the IAM policy for a workflow template",
		Args: cobra.ExactArgs(1), RunE: s.runDPWFGetIam,
	}
	s.dpWFSetIamCmd = &cobra.Command{
		Use: "set-iam-policy TEMPLATE POLICY_FILE", Short: "Set the IAM policy for a workflow template",
		Args: cobra.ExactArgs(2), RunE: s.runDPWFSetIam,
	}
	s.dpWFInstantiateCmd = &cobra.Command{
		Use: "instantiate TEMPLATE", Short: "Instantiate a workflow template",
		Args: cobra.ExactArgs(1), RunE: s.runDPWFInstantiate,
	}
	s.dpWFInstantiateFromFileCmd = &cobra.Command{
		Use: "instantiate-from-file", Short: "Instantiate an inline workflow template loaded from a file",
		Args: cobra.NoArgs, RunE: s.runDPWFInstantiateFromFile,
	}
	s.dpWFAddJobCmd = &cobra.Command{
		Use: "add-job TEMPLATE", Short: "Add a job to a workflow template",
		Args: cobra.ExactArgs(1), RunE: s.runDPWFAddJob,
	}
	s.dpWFRemoveJobCmd = &cobra.Command{
		Use: "remove-job TEMPLATE", Short: "Remove a job from a workflow template",
		Args: cobra.ExactArgs(1), RunE: s.runDPWFRemoveJob,
	}
	s.dpWFSetClusterSelectorCmd = &cobra.Command{
		Use: "set-cluster-selector TEMPLATE", Short: "Set the ClusterSelector for a workflow template",
		Args: cobra.ExactArgs(1), RunE: s.runDPWFSetClusterSelector,
	}
	s.dpWFSetManagedClusterCmd = &cobra.Command{
		Use: "set-managed-cluster TEMPLATE", Short: "Set the ManagedCluster for a workflow template",
		Args: cobra.ExactArgs(1), RunE: s.runDPWFSetManagedCluster,
	}
	s.dpWFSetDagTimeoutCmd = &cobra.Command{
		Use: "set-dag-timeout TEMPLATE", Short: "Set the DAG timeout for a workflow template",
		Args: cobra.ExactArgs(1), RunE: s.runDPWFSetDagTimeout,
	}
	s.dpWFRemoveDagTimeoutCmd = &cobra.Command{
		Use: "remove-dag-timeout TEMPLATE", Short: "Remove the DAG timeout from a workflow template",
		Args: cobra.ExactArgs(1), RunE: s.runDPWFRemoveDagTimeout,
	}
	s.initDataproc()
	s.initDataprocAutoscalingPolicies()
	s.initDataprocBatches()
	s.initDataprocClusters()
	s.initDataprocJobs()
	s.initDataprocNodeGroups()
	s.initDataprocOperations()
	s.initDataprocWorkflowTemplates()
	return s
}
