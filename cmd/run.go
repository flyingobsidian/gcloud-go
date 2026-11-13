package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	runv2 "google.golang.org/api/run/v2"
)

// --- gcloud run (#380, #1049-#1056) ---

func (s *runCmds) initRun() {
	// `run compose` is not in this batch's scope; keep it as a stub so the
	// help surface still lists it.
	registerStubGroup(s.runCmd, "compose", "Docker Compose workflows", "run", "list")

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

func (s *runCmds) runIamUpdatedIam(who string) {
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for %s.\n", who)
}

// runCmds holds the commands and flag values of one invocation's "run"
// command group, so invocations don't share state (#2070).
type runCmds struct {
	inv *invocation

	runCmd                          *cobra.Command
	flagRunDeployRegion             string
	flagRunDeployImage              string
	flagRunDeployFormat             string
	flagRunDeployAllowUnauth        bool
	flagRunDeployNoAllowUnauth      bool
	flagRunDeployServiceAcct        string
	flagRunDeployMemory             string
	flagRunDeployCPU                string
	flagRunDeployMaxInstances       int64
	flagRunDeployMinInstances       int64
	flagRunDeployConcurrency        int64
	flagRunDeployTimeout            string
	flagRunDeployPort               int64
	flagRunDeployEnvVars            map[string]string
	flagRunDeploySetEnvVars         map[string]string
	flagRunDeployCommand            []string
	flagRunDeployArgs               []string
	flagRunDeployVpcConnector       string
	flagRunDeployIngress            string
	flagRunDeployLabels             map[string]string
	flagRunDeployPlatform           string
	runDeployCmd                    *cobra.Command
	runDomainMappingsCmd            *cobra.Command
	flagRunDomainMappingsRegion     string
	flagRunDomainMappingsFormat     string
	flagRunDomainMappingsConfigFile string
	flagRunDomainMappingsLimit      int64
	runDomainMappingsCreateCmd      *cobra.Command
	runDomainMappingsDeleteCmd      *cobra.Command
	runDomainMappingsDescribeCmd    *cobra.Command
	runDomainMappingsListCmd        *cobra.Command
	runJobsCmd                      *cobra.Command
	runJobsExecutionsCmd            *cobra.Command
	flagRunJobsRegion               string
	flagRunJobsFormat               string
	flagRunJobsConfigFile           string
	flagRunJobsUpdateMask           string
	flagRunJobsPageSize             int64
	flagRunJobsShowDel              bool
	flagRunJobsLimit                int64
	flagRunJobsWait                 bool
	flagRunJobsImage                string
	flagRunJobsCommand              []string
	flagRunJobsArgs                 []string
	flagRunJobsEnvVars              map[string]string
	flagRunJobsServiceAcct          string
	flagRunJobsMemory               string
	flagRunJobsCPU                  string
	flagRunJobsTaskCount            int64
	flagRunJobsMaxRetries           int64
	flagRunJobsParallelism          int64
	flagRunJobsTaskTimeout          string
	flagRunJobsVpcConnector         string
	flagRunJobsIamMember            string
	flagRunJobsIamRole              string
	flagRunJobsIamCondExpr          string
	flagRunJobsIamCondT             string
	flagRunJobsIamCondD             string
	flagRunJobsIamAllCond           bool
	// executions subgroup
	flagRunJobsExecJob                    string
	flagRunJobsExecPageSize               int64
	flagRunJobsExecLimit                  int64
	runJobsCreateCmd                      *cobra.Command
	runJobsDeleteCmd                      *cobra.Command
	runJobsDeployCmd                      *cobra.Command
	runJobsDescribeCmd                    *cobra.Command
	runJobsExecuteCmd                     *cobra.Command
	runJobsListCmd                        *cobra.Command
	runJobsLogsCmd                        *cobra.Command
	runJobsReplaceCmd                     *cobra.Command
	runJobsUpdateCmd                      *cobra.Command
	runJobsGetIamCmd                      *cobra.Command
	runJobsSetIamCmd                      *cobra.Command
	runJobsAddIamCmd                      *cobra.Command
	runJobsRemoveIamCmd                   *cobra.Command
	runJobsExecDescribeCmd                *cobra.Command
	runJobsExecListCmd                    *cobra.Command
	runJobsExecDeleteCmd                  *cobra.Command
	runJobsExecCancelCmd                  *cobra.Command
	runMultiRegionServicesCmd             *cobra.Command
	flagRunMRSRegion                      string
	flagRunMRSFormat                      string
	flagRunMRSConfigFile                  string
	flagRunMRSUpdateMask                  string
	flagRunMRSPageSize                    int64
	runMRSDeleteCmd                       *cobra.Command
	runMRSDescribeCmd                     *cobra.Command
	runMRSListCmd                         *cobra.Command
	runMRSReplaceCmd                      *cobra.Command
	runMRSUpdateCmd                       *cobra.Command
	runRegionsCmd                         *cobra.Command
	flagRunRegionsFormat                  string
	flagRunRegionsPageSize                int64
	runRegionsListCmd                     *cobra.Command
	runRevisionsCmd                       *cobra.Command
	flagRunRevisionsRegion                string
	flagRunRevisionsService               string
	flagRunRevisionsFormat                string
	flagRunRevisionsPageSize              int64
	flagRunRevisionsShowDel               bool
	runRevisionsDeleteCmd                 *cobra.Command
	runRevisionsDescribeCmd               *cobra.Command
	runRevisionsListCmd                   *cobra.Command
	flagRunDeployScalingCPUTarget         string
	flagRunDeployScalingConcurrencyTarget string
	runServicesCmd                        *cobra.Command
	flagRunServicesRegion                 string
	flagRunServicesFormat                 string
	flagRunServicesConfigFile             string
	flagRunServicesUpdateMask             string
	flagRunServicesPageSize               int64
	flagRunServicesShowDel                bool
	flagRunServicesLimit                  int64
	flagRunServicesIamMember              string
	flagRunServicesIamRole                string
	flagRunServicesIamCondExpr            string
	flagRunServicesIamCondT               string
	flagRunServicesIamCondD               string
	flagRunServicesIamAllCond             bool
	flagRunServicesTrafficRevs            map[string]string
	flagRunServicesTrafficLatest          bool
	flagRunServicesProxyPort              int
	flagRunServicesProxyBind              string
	runServicesDeleteCmd                  *cobra.Command
	runServicesDescribeCmd                *cobra.Command
	runServicesListCmd                    *cobra.Command
	runServicesReplaceCmd                 *cobra.Command
	runServicesUpdateCmd                  *cobra.Command
	runServicesUpdateTrafficCmd           *cobra.Command
	runServicesLogsCmd                    *cobra.Command
	runServicesProxyCmd                   *cobra.Command
	runServicesGetIamCmd                  *cobra.Command
	runServicesSetIamCmd                  *cobra.Command
	runServicesAddIamCmd                  *cobra.Command
	runServicesRemoveIamCmd               *cobra.Command
	runWorkerPoolsCmd                     *cobra.Command
	runWorkerPoolsRevisionsCmd            *cobra.Command
	flagRunWPRegion                       string
	flagRunWPFormat                       string
	flagRunWPConfigFile                   string
	flagRunWPUpdateMask                   string
	flagRunWPPageSize                     int64
	flagRunWPShowDel                      bool
	flagRunWPLimit                        int64
	flagRunWPImage                        string
	flagRunWPCommand                      []string
	flagRunWPArgs                         []string
	flagRunWPEnvVars                      map[string]string
	flagRunWPServiceAcct                  string
	flagRunWPMemory                       string
	flagRunWPCPU                          string
	flagRunWPVpcConnector                 string
	flagRunWPIamMember                    string
	flagRunWPIamRole                      string
	flagRunWPIamCondExpr                  string
	flagRunWPIamCondT                     string
	flagRunWPIamCondD                     string
	flagRunWPIamAllCond                   bool
	flagRunWPSplitRevs                    map[string]string
	flagRunWPSplitLatest                  bool
	// revisions subgroup
	flagRunWPRevPool     string
	flagRunWPRevPageSize int64
	flagRunWPRevShowDel  bool
	runWPDeleteCmd       *cobra.Command
	runWPDeployCmd       *cobra.Command
	runWPDescribeCmd     *cobra.Command
	runWPListCmd         *cobra.Command
	runWPLogsCmd         *cobra.Command
	runWPReplaceCmd      *cobra.Command
	runWPUpdateCmd       *cobra.Command
	runWPUpdateSplitCmd  *cobra.Command
	runWPGetIamCmd       *cobra.Command
	runWPSetIamCmd       *cobra.Command
	runWPAddIamCmd       *cobra.Command
	runWPRemoveIamCmd    *cobra.Command
	runWPRevDescribeCmd  *cobra.Command
	runWPRevListCmd      *cobra.Command
	runWPRevDeleteCmd    *cobra.Command
}

// newRunCmd returns the "run" command group for inv.
func newRunCmd(inv *invocation) *cobra.Command {
	return newRunCmds(inv).runCmd
}

// newRunCmds builds the "run" command group for inv and returns its state.
func newRunCmds(inv *invocation) *runCmds {
	s := &runCmds{inv: inv}
	s.runCmd = &cobra.Command{Use: "run", Short: "Manage Cloud Run"}
	s.runDeployCmd = &cobra.Command{
		Use:   "deploy SERVICE",
		Short: "Deploy a Cloud Run service",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runDeployRun,
	}
	s.runDomainMappingsCmd = &cobra.Command{Use: "domain-mappings", Short: "Manage Cloud Run domain mappings"}
	s.runDomainMappingsCreateCmd = &cobra.Command{
		Use: "create DOMAIN", Short: "Create a Cloud Run domain mapping",
		Args: cobra.ExactArgs(1), RunE: s.runDomainMappingsCreate,
	}
	s.runDomainMappingsDeleteCmd = &cobra.Command{
		Use: "delete DOMAIN", Short: "Delete a Cloud Run domain mapping",
		Args: cobra.ExactArgs(1), RunE: s.runDomainMappingsDelete,
	}
	s.runDomainMappingsDescribeCmd = &cobra.Command{
		Use: "describe DOMAIN", Short: "Describe a Cloud Run domain mapping",
		Args: cobra.ExactArgs(1), RunE: s.runDomainMappingsDescribe,
	}
	s.runDomainMappingsListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Run domain mappings",
		Args: cobra.NoArgs, RunE: s.runDomainMappingsList,
	}
	s.runJobsCmd = &cobra.Command{Use: "jobs", Short: "Manage Cloud Run jobs"}
	s.runJobsExecutionsCmd = &cobra.Command{Use: "executions", Short: "Manage Cloud Run job executions"}
	s.runJobsCreateCmd = &cobra.Command{
		Use: "create JOB", Short: "Create a Cloud Run job",
		Args: cobra.ExactArgs(1), RunE: s.runJobsCreate,
	}
	s.runJobsDeleteCmd = &cobra.Command{
		Use: "delete JOB", Short: "Delete a Cloud Run job",
		Args: cobra.ExactArgs(1), RunE: s.runJobsDelete,
	}
	s.runJobsDeployCmd = &cobra.Command{
		Use: "deploy JOB", Short: "Deploy a Cloud Run job (create-or-update)",
		Args: cobra.ExactArgs(1), RunE: s.runJobsDeploy,
	}
	s.runJobsDescribeCmd = &cobra.Command{
		Use: "describe JOB", Short: "Describe a Cloud Run job",
		Args: cobra.ExactArgs(1), RunE: s.runJobsDescribe,
	}
	s.runJobsExecuteCmd = &cobra.Command{
		Use: "execute JOB", Short: "Execute a Cloud Run job",
		Args: cobra.ExactArgs(1), RunE: s.runJobsExecute,
	}
	s.runJobsListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Run jobs",
		Args: cobra.NoArgs, RunE: s.runJobsList,
	}
	s.runJobsLogsCmd = &cobra.Command{
		Use: "logs JOB", Short: "Read Cloud Logging entries for a Cloud Run job",
		Args: cobra.ExactArgs(1), RunE: s.runJobsLogs,
	}
	s.runJobsReplaceCmd = &cobra.Command{
		Use: "replace JOB", Short: "Replace a Cloud Run job from a config file",
		Args: cobra.ExactArgs(1), RunE: s.runJobsReplace,
	}
	s.runJobsUpdateCmd = &cobra.Command{
		Use: "update JOB", Short: "Update a Cloud Run job from a config file",
		Args: cobra.ExactArgs(1), RunE: s.runJobsUpdate,
	}
	s.runJobsGetIamCmd = &cobra.Command{
		Use: "get-iam-policy JOB", Short: "Get the IAM policy for a Cloud Run job",
		Args: cobra.ExactArgs(1), RunE: s.runJobsGetIam,
	}
	s.runJobsSetIamCmd = &cobra.Command{
		Use: "set-iam-policy JOB POLICY_FILE", Short: "Set the IAM policy for a Cloud Run job",
		Args: cobra.ExactArgs(2), RunE: s.runJobsSetIam,
	}
	s.runJobsAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding JOB", Short: "Add an IAM policy binding to a Cloud Run job",
		Args: cobra.ExactArgs(1), RunE: s.runJobsAddIam,
	}
	s.runJobsRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding JOB", Short: "Remove an IAM policy binding from a Cloud Run job",
		Args: cobra.ExactArgs(1), RunE: s.runJobsRemoveIam,
	}
	s.runJobsExecDescribeCmd = &cobra.Command{
		Use: "describe EXECUTION", Short: "Describe a Cloud Run job execution",
		Args: cobra.ExactArgs(1), RunE: s.runJobsExecDescribe,
	}
	s.runJobsExecListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Run job executions",
		Args: cobra.NoArgs, RunE: s.runJobsExecList,
	}
	s.runJobsExecDeleteCmd = &cobra.Command{
		Use: "delete EXECUTION", Short: "Delete a Cloud Run job execution",
		Args: cobra.ExactArgs(1), RunE: s.runJobsExecDelete,
	}
	s.runJobsExecCancelCmd = &cobra.Command{
		Use: "cancel EXECUTION", Short: "Cancel a running Cloud Run job execution",
		Args: cobra.ExactArgs(1), RunE: s.runJobsExecCancel,
	}
	s.runMultiRegionServicesCmd = &cobra.Command{Use: "multi-region-services", Short: "Manage Cloud Run multi-region services"}
	s.runMRSDeleteCmd = &cobra.Command{
		Use: "delete SERVICE", Short: "Delete a Cloud Run multi-region service",
		Args: cobra.ExactArgs(1), RunE: s.runMRSDelete,
	}
	s.runMRSDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe a Cloud Run multi-region service",
		Args: cobra.ExactArgs(1), RunE: s.runMRSDescribe,
	}
	s.runMRSListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Run multi-region services",
		Args: cobra.NoArgs, RunE: s.runMRSList,
	}
	s.runMRSReplaceCmd = &cobra.Command{
		Use: "replace SERVICE", Short: "Replace a Cloud Run multi-region service (PUT)",
		Args: cobra.ExactArgs(1), RunE: s.runMRSReplace,
	}
	s.runMRSUpdateCmd = &cobra.Command{
		Use: "update SERVICE", Short: "Update a Cloud Run multi-region service (PATCH)",
		Args: cobra.ExactArgs(1), RunE: s.runMRSUpdate,
	}
	s.runRegionsCmd = &cobra.Command{Use: "regions", Short: "View Cloud Run regions"}
	s.runRegionsListCmd = &cobra.Command{
		Use: "list", Short: "List the regions where Cloud Run is available",
		Args: cobra.NoArgs, RunE: s.runRegionsList,
	}
	s.runRevisionsCmd = &cobra.Command{Use: "revisions", Short: "Manage Cloud Run revisions"}
	s.runRevisionsDeleteCmd = &cobra.Command{
		Use: "delete REVISION", Short: "Delete a Cloud Run revision",
		Args: cobra.ExactArgs(1), RunE: s.runRevisionsDelete,
	}
	s.runRevisionsDescribeCmd = &cobra.Command{
		Use: "describe REVISION", Short: "Describe a Cloud Run revision",
		Args: cobra.ExactArgs(1), RunE: s.runRevisionsDescribe,
	}
	s.runRevisionsListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Run revisions for a service",
		Args: cobra.NoArgs, RunE: s.runRevisionsList,
	}
	s.runServicesCmd = &cobra.Command{Use: "services", Short: "Manage Cloud Run services"}
	s.runServicesDeleteCmd = &cobra.Command{
		Use: "delete SERVICE", Short: "Delete a Cloud Run service",
		Args: cobra.ExactArgs(1), RunE: s.runSvcDelete,
	}
	s.runServicesDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe a Cloud Run service",
		Args: cobra.ExactArgs(1), RunE: s.runSvcDescribe,
	}
	s.runServicesListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Run services",
		Args: cobra.NoArgs, RunE: s.runSvcList,
	}
	s.runServicesReplaceCmd = &cobra.Command{
		Use: "replace SERVICE", Short: "Replace a Cloud Run service from a config file",
		Args: cobra.ExactArgs(1), RunE: s.runSvcReplace,
	}
	s.runServicesUpdateCmd = &cobra.Command{
		Use: "update SERVICE", Short: "Update a Cloud Run service from a config file",
		Args: cobra.ExactArgs(1), RunE: s.runSvcUpdate,
	}
	s.runServicesUpdateTrafficCmd = &cobra.Command{
		Use: "update-traffic SERVICE", Short: "Update Cloud Run service traffic split",
		Args: cobra.ExactArgs(1), RunE: s.runSvcUpdateTraffic,
	}
	s.runServicesLogsCmd = &cobra.Command{
		Use: "logs SERVICE", Short: "Read Cloud Logging entries for a Cloud Run service",
		Args: cobra.ExactArgs(1), RunE: s.runSvcLogs,
	}
	s.runServicesProxyCmd = &cobra.Command{
		Use: "proxy SERVICE", Short: "Proxy authenticated local traffic to a Cloud Run service",
		Args: cobra.ExactArgs(1), RunE: s.runSvcProxy,
	}
	s.runServicesGetIamCmd = &cobra.Command{
		Use: "get-iam-policy SERVICE", Short: "Get the IAM policy for a Cloud Run service",
		Args: cobra.ExactArgs(1), RunE: s.runSvcGetIam,
	}
	s.runServicesSetIamCmd = &cobra.Command{
		Use: "set-iam-policy SERVICE POLICY_FILE", Short: "Set the IAM policy for a Cloud Run service",
		Args: cobra.ExactArgs(2), RunE: s.runSvcSetIam,
	}
	s.runServicesAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding SERVICE", Short: "Add an IAM policy binding to a Cloud Run service",
		Args: cobra.ExactArgs(1), RunE: s.runSvcAddIam,
	}
	s.runServicesRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding SERVICE", Short: "Remove an IAM policy binding from a Cloud Run service",
		Args: cobra.ExactArgs(1), RunE: s.runSvcRemoveIam,
	}
	s.runWorkerPoolsCmd = &cobra.Command{Use: "worker-pools", Short: "Manage Cloud Run worker pools"}
	s.runWorkerPoolsRevisionsCmd = &cobra.Command{Use: "revisions", Short: "Manage Cloud Run worker pool revisions"}
	s.runWPDeleteCmd = &cobra.Command{
		Use: "delete POOL", Short: "Delete a Cloud Run worker pool",
		Args: cobra.ExactArgs(1), RunE: s.runWPDelete,
	}
	s.runWPDeployCmd = &cobra.Command{
		Use: "deploy POOL", Short: "Deploy a Cloud Run worker pool (create-or-update)",
		Args: cobra.ExactArgs(1), RunE: s.runWPDeploy,
	}
	s.runWPDescribeCmd = &cobra.Command{
		Use: "describe POOL", Short: "Describe a Cloud Run worker pool",
		Args: cobra.ExactArgs(1), RunE: s.runWPDescribe,
	}
	s.runWPListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Run worker pools",
		Args: cobra.NoArgs, RunE: s.runWPList,
	}
	s.runWPLogsCmd = &cobra.Command{
		Use: "logs POOL", Short: "Read Cloud Logging entries for a Cloud Run worker pool",
		Args: cobra.ExactArgs(1), RunE: s.runWPLogs,
	}
	s.runWPReplaceCmd = &cobra.Command{
		Use: "replace POOL", Short: "Replace a Cloud Run worker pool from a config file",
		Args: cobra.ExactArgs(1), RunE: s.runWPReplace,
	}
	s.runWPUpdateCmd = &cobra.Command{
		Use: "update POOL", Short: "Update a Cloud Run worker pool from a config file",
		Args: cobra.ExactArgs(1), RunE: s.runWPUpdate,
	}
	s.runWPUpdateSplitCmd = &cobra.Command{
		Use: "update-instance-split POOL", Short: "Update the instance split for a Cloud Run worker pool",
		Args: cobra.ExactArgs(1), RunE: s.runWPUpdateSplit,
	}
	s.runWPGetIamCmd = &cobra.Command{
		Use: "get-iam-policy POOL", Short: "Get the IAM policy for a Cloud Run worker pool",
		Args: cobra.ExactArgs(1), RunE: s.runWPGetIam,
	}
	s.runWPSetIamCmd = &cobra.Command{
		Use: "set-iam-policy POOL POLICY_FILE", Short: "Set the IAM policy for a Cloud Run worker pool",
		Args: cobra.ExactArgs(2), RunE: s.runWPSetIam,
	}
	s.runWPAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding POOL", Short: "Add an IAM policy binding to a Cloud Run worker pool",
		Args: cobra.ExactArgs(1), RunE: s.runWPAddIam,
	}
	s.runWPRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding POOL", Short: "Remove an IAM policy binding from a Cloud Run worker pool",
		Args: cobra.ExactArgs(1), RunE: s.runWPRemoveIam,
	}
	s.runWPRevDescribeCmd = &cobra.Command{
		Use: "describe REVISION", Short: "Describe a Cloud Run worker pool revision",
		Args: cobra.ExactArgs(1), RunE: s.runWPRevDescribe,
	}
	s.runWPRevListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Run worker pool revisions",
		Args: cobra.NoArgs, RunE: s.runWPRevList,
	}
	s.runWPRevDeleteCmd = &cobra.Command{
		Use: "delete REVISION", Short: "Delete a Cloud Run worker pool revision",
		Args: cobra.ExactArgs(1), RunE: s.runWPRevDelete,
	}
	s.initRun()
	s.initRunDeploy()
	s.initRunDomainMappings()
	s.initRunJobs()
	s.initRunMultiRegionServices()
	s.initRunRegions()
	s.initRunRevisions()
	s.initRunScalingTargets()
	s.initRunServices()
	s.initRunWorkerPools()
	return s
}
