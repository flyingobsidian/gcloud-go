package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	ml "google.golang.org/api/ml/v1"
)

// --- gcloud ai-platform (#982-#987) ---
//
// Legacy Cloud ML Engine surface (predecessor to Vertex AI). Backed by
// the ml.googleapis.com API (google.golang.org/api/ml/v1). Distinct from
// `gcloud ai ...`, which uses the newer regional aiplatform.googleapis.com
// endpoint.

func (s *aiPlatformCmds) initAiPlatform() {
	s.aiPlatformPredictCmd.Flags().StringVar(&s.flagAIPlatformPredictModel, "model", "", "Model name (required)")
	s.aiPlatformPredictCmd.Flags().StringVar(&s.flagAIPlatformPredictVersion, "version", "", "Version name; defaults to the model's default version")
	s.aiPlatformPredictCmd.Flags().StringVar(&s.flagAIPlatformPredictReqFile, "json-request", "",
		"Path to a JSON file with the prediction request body (required)")
	s.aiPlatformPredictCmd.Flags().StringVar(&s.flagAIPlatformPredictFormat, "format", "", "Output format")
	_ = s.aiPlatformPredictCmd.MarkFlagRequired("model")
	_ = s.aiPlatformPredictCmd.MarkFlagRequired("json-request")

	s.aiPlatformCmd.AddCommand(s.aiPlatformPredictCmd)

}

// --- Resource name helpers ---

// mlProjectPath returns "projects/PROJ".
func mlProjectPath(project string) string {
	return fmt.Sprintf("projects/%s", project)
}

// mlModelName returns "projects/PROJ/models/MODEL", passing through a
// caller-supplied full resource name unchanged.
func mlModelName(project, model string) string {
	if len(model) >= len("projects/") && model[:len("projects/")] == "projects/" {
		return model
	}
	return fmt.Sprintf("projects/%s/models/%s", project, model)
}

// mlVersionName returns "projects/PROJ/models/MODEL/versions/VERSION",
// passing through a caller-supplied full resource name unchanged.
func mlVersionName(project, model, version string) string {
	if len(version) >= len("projects/") && version[:len("projects/")] == "projects/" {
		return version
	}
	return fmt.Sprintf("projects/%s/models/%s/versions/%s", project, model, version)
}

// mlJobName returns "projects/PROJ/jobs/JOB", passing through a
// caller-supplied full resource name unchanged.
func mlJobName(project, job string) string {
	if len(job) >= len("projects/") && job[:len("projects/")] == "projects/" {
		return job
	}
	return fmt.Sprintf("projects/%s/jobs/%s", project, job)
}

// mlOperationName returns "projects/PROJ/operations/OP", passing through
// a caller-supplied full resource name unchanged.
func mlOperationName(project, op string) string {
	if len(op) >= len("projects/") && op[:len("projects/")] == "projects/" {
		return op
	}
	return fmt.Sprintf("projects/%s/operations/%s", project, op)
}

// --- IAM helpers (mirror the pattern used in cmd/dns.go, but bound to the
// ml v1 IAM type names — GoogleIamV1__Policy / GoogleIamV1__Binding /
// GoogleType__Expr — which use double-underscore prefixes). ---

func mlIamFlags(c *cobra.Command, member, role, condExpr, condTitle, condDesc *string) {
	c.Flags().StringVar(member, "member", "", "IAM member (required)")
	c.Flags().StringVar(role, "role", "", "IAM role to bind (required)")
	c.Flags().StringVar(condExpr, "condition-expression", "", "CEL expression for a conditional binding")
	c.Flags().StringVar(condTitle, "condition-title", "", "Title for a conditional binding")
	c.Flags().StringVar(condDesc, "condition-description", "", "Description for a conditional binding")
	_ = c.MarkFlagRequired("member")
	_ = c.MarkFlagRequired("role")
}

func mlBuildCondition(expr, title, desc string) *ml.GoogleType__Expr {
	if expr == "" && title == "" && desc == "" {
		return nil
	}
	return &ml.GoogleType__Expr{Expression: expr, Title: title, Description: desc}
}

func mlCondsEqual(a, b *ml.GoogleType__Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

func mlAddBinding(policy *ml.GoogleIamV1__Policy, role, member string, cond *ml.GoogleType__Expr) {
	for _, b := range policy.Bindings {
		if b.Role != role || !mlCondsEqual(b.Condition, cond) {
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
	policy.Bindings = append(policy.Bindings, &ml.GoogleIamV1__Binding{
		Role: role, Members: []string{member}, Condition: cond,
	})
}

func mlRemoveBinding(policy *ml.GoogleIamV1__Policy, role, member string, cond *ml.GoogleType__Expr, allConds bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConds || mlCondsEqual(b.Condition, cond))
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

func (s *aiPlatformCmds) mlUpdatedIam(who string) {
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for %s.\n", who)
}

// --- ai-platform predict ---

func (s *aiPlatformCmds) runAIPlatformPredict(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(s.flagAIPlatformPredictReqFile)
	if err != nil {
		return fmt.Errorf("reading request body: %w", err)
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	var name string
	if s.flagAIPlatformPredictVersion != "" {
		name = mlVersionName(project, s.flagAIPlatformPredictModel, s.flagAIPlatformPredictVersion)
	} else {
		name = mlModelName(project, s.flagAIPlatformPredictModel)
	}
	// The ml v1 client sends HttpBody.Data verbatim as the POST body
	// (see (*ProjectsPredictCall).doRequest), so Data must be the raw
	// JSON string — not a base64-encoded copy of it. The generated
	// PredictCall bypasses the standard HttpBody JSON marshaller that
	// would otherwise base64 the bytes.
	req := &ml.GoogleCloudMlV1__PredictRequest{
		HttpBody: &ml.GoogleApi__HttpBody{
			ContentType: "application/json",
			Data:        string(data),
		},
	}
	resp, err := svc.Projects.Predict(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running prediction: %w", err)
	}
	// The prediction response body is returned in Data as a raw JSON string.
	// Print it directly for the default (empty) format so callers see the
	// server's JSON response; honour --format for structured output.
	if s.flagAIPlatformPredictFormat == "" {
		fmt.Fprintln(s.inv.stdout, resp.Data)
		return nil
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagAIPlatformPredictFormat)
}

// aiPlatformCmds holds the commands and flag values of one invocation's "aiPlatform"
// command group, so invocations don't share state (#2070).
type aiPlatformCmds struct {
	inv *invocation

	aiPlatformCmd                    *cobra.Command
	flagAIPlatformPredictModel       string
	flagAIPlatformPredictVersion     string
	flagAIPlatformPredictReqFile     string
	flagAIPlatformPredictFormat      string
	aiPlatformPredictCmd             *cobra.Command
	aiPlatformJobsCmd                *cobra.Command
	flagAIPlatformJobsFormat         string
	flagAIPlatformJobsConfigFile     string
	flagAIPlatformJobsPageSize       int64
	flagAIPlatformJobsFilter         string
	flagAIPlatformJobsUpdateMask     string
	flagAIPlatformJobsJobID          string
	flagAIPlatformJobsPollEvery      time.Duration
	flagAIPlatformJobsTimeout        time.Duration
	aiPlatformJobsCancelCmd          *cobra.Command
	aiPlatformJobsDescribeCmd        *cobra.Command
	aiPlatformJobsListCmd            *cobra.Command
	aiPlatformJobsStreamCmd          *cobra.Command
	aiPlatformJobsSubmitCmd          *cobra.Command
	aiPlatformJobsUpdateCmd          *cobra.Command
	aiPlatformLocalCmd               *cobra.Command
	flagAIPlatformLocalFormat        string
	flagAIPlatformLocalModelDir      string
	flagAIPlatformLocalJSONInstances string
	flagAIPlatformLocalTextInstances string
	flagAIPlatformLocalFramework     string
	flagAIPlatformLocalSignatureName string
	flagAIPlatformLocalPackagePath   string
	flagAIPlatformLocalModuleName    string
	flagAIPlatformLocalJobDir        string
	flagAIPlatformLocalDistributed   bool
	aiPlatformLocalPredictCmd        *cobra.Command
	aiPlatformLocalTrainCmd          *cobra.Command
	aiPlatformModelsCmd              *cobra.Command
	flagAIPlatformModelsFormat       string
	flagAIPlatformModelsConfigFile   string
	flagAIPlatformModelsPageSize     int64
	flagAIPlatformModelsFilter       string
	flagAIPlatformModelsUpdateMask   string
	flagAIPlatformModelsIamMember    string
	flagAIPlatformModelsIamRole      string
	flagAIPlatformModelsIamCondExpr  string
	flagAIPlatformModelsIamCondT     string
	flagAIPlatformModelsIamCondD     string
	flagAIPlatformModelsIamAllCond   bool
	aiPlatformModelsCreateCmd        *cobra.Command
	aiPlatformModelsDeleteCmd        *cobra.Command
	aiPlatformModelsDescribeCmd      *cobra.Command
	aiPlatformModelsListCmd          *cobra.Command
	aiPlatformModelsUpdateCmd        *cobra.Command
	aiPlatformModelsGetIamCmd        *cobra.Command
	aiPlatformModelsSetIamCmd        *cobra.Command
	aiPlatformModelsAddIamCmd        *cobra.Command
	aiPlatformModelsRemoveIamCmd     *cobra.Command
	aiPlatformOpsCmd                 *cobra.Command
	flagAIPlatformOpsFormat          string
	flagAIPlatformOpsPageSize        int64
	flagAIPlatformOpsFilter          string
	flagAIPlatformOpsPollEvery       time.Duration
	flagAIPlatformOpsTimeout         time.Duration
	aiPlatformOpsCancelCmd           *cobra.Command
	aiPlatformOpsDescribeCmd         *cobra.Command
	aiPlatformOpsListCmd             *cobra.Command
	aiPlatformOpsWaitCmd             *cobra.Command
	aiPlatformVersionsCmd            *cobra.Command
	flagAIPlatformVersionsFormat     string
	flagAIPlatformVersionsModel      string
	flagAIPlatformVersionsConfigFile string
	flagAIPlatformVersionsPageSize   int64
	flagAIPlatformVersionsFilter     string
	flagAIPlatformVersionsUpdateMask string
	aiPlatformVersionsCreateCmd      *cobra.Command
	aiPlatformVersionsDeleteCmd      *cobra.Command
	aiPlatformVersionsDescribeCmd    *cobra.Command
	aiPlatformVersionsListCmd        *cobra.Command
	aiPlatformVersionsSetDefaultCmd  *cobra.Command
	aiPlatformVersionsUpdateCmd      *cobra.Command
}

// newAiPlatformCmd returns the "aiPlatform" command group for inv.
func newAiPlatformCmd(inv *invocation) *cobra.Command {
	return newAiPlatformCmds(inv).aiPlatformCmd
}

// newAiPlatformCmds builds the "aiPlatform" command group for inv and returns its state.
func newAiPlatformCmds(inv *invocation) *aiPlatformCmds {
	s := &aiPlatformCmds{inv: inv}
	s.aiPlatformCmd = &cobra.Command{Use: "ai-platform", Short: "Manage AI Platform (Cloud ML Engine legacy)"}
	s.aiPlatformPredictCmd = &cobra.Command{
		Use:   "predict",
		Short: "Run online prediction against an AI Platform deployed model",
		Args:  cobra.NoArgs,
		RunE:  s.runAIPlatformPredict,
	}
	s.aiPlatformJobsCmd = &cobra.Command{Use: "jobs", Short: "Manage AI Platform jobs"}
	s.aiPlatformJobsCancelCmd = &cobra.Command{
		Use: "cancel JOB", Short: "Cancel a running AI Platform job",
		Args: cobra.ExactArgs(1), RunE: s.runAIPlatformJobsCancel,
	}
	s.aiPlatformJobsDescribeCmd = &cobra.Command{
		Use: "describe JOB", Short: "Describe an AI Platform job",
		Args: cobra.ExactArgs(1), RunE: s.runAIPlatformJobsDescribe,
	}
	s.aiPlatformJobsListCmd = &cobra.Command{
		Use: "list", Short: "List AI Platform jobs",
		Args: cobra.NoArgs, RunE: s.runAIPlatformJobsList,
	}
	s.aiPlatformJobsStreamCmd = &cobra.Command{
		Use: "stream-logs JOB", Short: "Poll an AI Platform job's state until it reaches a terminal state",
		Args: cobra.ExactArgs(1), RunE: s.runAIPlatformJobsStream,
	}
	s.aiPlatformJobsSubmitCmd = &cobra.Command{
		Use: "submit", Short: "Submit an AI Platform training or prediction job",
		Args: cobra.NoArgs, RunE: s.runAIPlatformJobsSubmit,
	}
	s.aiPlatformJobsUpdateCmd = &cobra.Command{
		Use: "update JOB", Short: "Update an AI Platform job",
		Args: cobra.ExactArgs(1), RunE: s.runAIPlatformJobsUpdate,
	}
	s.aiPlatformLocalCmd = &cobra.Command{
		Use:   "local",
		Short: "AI Platform local (client-side) commands",
		Long: "Client-side prediction and training commands from upstream gcloud. " +
			"These require a local Python runtime with TensorFlow/scikit-learn (predict) or " +
			"a user-supplied training package (train) and are not supported by gcloud-go. " +
			"Use `ai-platform predict` or `ai-platform jobs submit` for the server-side equivalents.",
	}
	s.aiPlatformLocalPredictCmd = &cobra.Command{
		Use:   "predict",
		Short: "(Unsupported) Run client-side prediction against a locally exported model",
		Long: "Runs prediction against a locally exported TensorFlow/scikit-learn model. " +
			"This is a client-side flow that requires a Python runtime and the model framework " +
			"and is not supported by gcloud-go. Use `ai-platform predict` to run against a deployed model.",
		Args: cobra.NoArgs,
		RunE: runAIPlatformLocalPredict,
	}
	s.aiPlatformLocalTrainCmd = &cobra.Command{
		Use:   "train",
		Short: "(Unsupported) Run a training job locally, using the same code the AI Platform job runners would use",
		Long: "Runs a training package locally so users can iterate before submitting to AI Platform. " +
			"This is a client-side flow that requires a Python runtime and the user's training package " +
			"and is not supported by gcloud-go. Use `ai-platform jobs submit` to run training on the service.",
		Args: cobra.NoArgs,
		RunE: runAIPlatformLocalTrain,
	}
	s.aiPlatformModelsCmd = &cobra.Command{Use: "models", Short: "Manage AI Platform models"}
	s.aiPlatformModelsCreateCmd = &cobra.Command{
		Use: "create MODEL", Short: "Create an AI Platform model",
		Args: cobra.ExactArgs(1), RunE: s.runAIPlatformModelsCreate,
	}
	s.aiPlatformModelsDeleteCmd = &cobra.Command{
		Use: "delete MODEL", Short: "Delete an AI Platform model",
		Args: cobra.ExactArgs(1), RunE: s.runAIPlatformModelsDelete,
	}
	s.aiPlatformModelsDescribeCmd = &cobra.Command{
		Use: "describe MODEL", Short: "Describe an AI Platform model",
		Args: cobra.ExactArgs(1), RunE: s.runAIPlatformModelsDescribe,
	}
	s.aiPlatformModelsListCmd = &cobra.Command{
		Use: "list", Short: "List AI Platform models",
		Args: cobra.NoArgs, RunE: s.runAIPlatformModelsList,
	}
	s.aiPlatformModelsUpdateCmd = &cobra.Command{
		Use: "update MODEL", Short: "Update an AI Platform model",
		Args: cobra.ExactArgs(1), RunE: s.runAIPlatformModelsUpdate,
	}
	s.aiPlatformModelsGetIamCmd = &cobra.Command{
		Use: "get-iam-policy MODEL", Short: "Get the IAM policy for an AI Platform model",
		Args: cobra.ExactArgs(1), RunE: s.runAIPlatformModelsGetIam,
	}
	s.aiPlatformModelsSetIamCmd = &cobra.Command{
		Use: "set-iam-policy MODEL POLICY_FILE", Short: "Set the IAM policy for an AI Platform model",
		Args: cobra.ExactArgs(2), RunE: s.runAIPlatformModelsSetIam,
	}
	s.aiPlatformModelsAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding MODEL", Short: "Add an IAM policy binding to an AI Platform model",
		Args: cobra.ExactArgs(1), RunE: s.runAIPlatformModelsAddIam,
	}
	s.aiPlatformModelsRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding MODEL", Short: "Remove an IAM policy binding from an AI Platform model",
		Args: cobra.ExactArgs(1), RunE: s.runAIPlatformModelsRemoveIam,
	}
	s.aiPlatformOpsCmd = &cobra.Command{Use: "operations", Short: "Manage AI Platform long-running operations"}
	s.aiPlatformOpsCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel an AI Platform operation",
		Args: cobra.ExactArgs(1), RunE: s.runAIPlatformOpsCancel,
	}
	s.aiPlatformOpsDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an AI Platform operation",
		Args: cobra.ExactArgs(1), RunE: s.runAIPlatformOpsDescribe,
	}
	s.aiPlatformOpsListCmd = &cobra.Command{
		Use: "list", Short: "List AI Platform operations",
		Args: cobra.NoArgs, RunE: s.runAIPlatformOpsList,
	}
	s.aiPlatformOpsWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Poll an AI Platform operation until it completes",
		Args: cobra.ExactArgs(1), RunE: s.runAIPlatformOpsWait,
	}
	s.aiPlatformVersionsCmd = &cobra.Command{Use: "versions", Short: "Manage AI Platform model versions"}
	s.aiPlatformVersionsCreateCmd = &cobra.Command{
		Use: "create VERSION", Short: "Create an AI Platform model version",
		Args: cobra.ExactArgs(1), RunE: s.runAIPlatformVersionsCreate,
	}
	s.aiPlatformVersionsDeleteCmd = &cobra.Command{
		Use: "delete VERSION", Short: "Delete an AI Platform model version",
		Args: cobra.ExactArgs(1), RunE: s.runAIPlatformVersionsDelete,
	}
	s.aiPlatformVersionsDescribeCmd = &cobra.Command{
		Use: "describe VERSION", Short: "Describe an AI Platform model version",
		Args: cobra.ExactArgs(1), RunE: s.runAIPlatformVersionsDescribe,
	}
	s.aiPlatformVersionsListCmd = &cobra.Command{
		Use: "list", Short: "List AI Platform model versions",
		Args: cobra.NoArgs, RunE: s.runAIPlatformVersionsList,
	}
	s.aiPlatformVersionsSetDefaultCmd = &cobra.Command{
		Use: "set-default VERSION", Short: "Set an AI Platform model version as the default",
		Args: cobra.ExactArgs(1), RunE: s.runAIPlatformVersionsSetDefault,
	}
	s.aiPlatformVersionsUpdateCmd = &cobra.Command{
		Use: "update VERSION", Short: "Update an AI Platform model version",
		Args: cobra.ExactArgs(1), RunE: s.runAIPlatformVersionsUpdate,
	}
	s.initAiPlatform()
	s.initAiPlatformJobs()
	s.initAiPlatformLocal()
	s.initAiPlatformModels()
	s.initAiPlatformOperations()
	s.initAiPlatformVersions()
	return s
}
