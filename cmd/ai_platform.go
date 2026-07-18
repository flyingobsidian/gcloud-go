package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	ml "google.golang.org/api/ml/v1"
)

// --- gcloud ai-platform (#982-#987) ---
//
// Legacy Cloud ML Engine surface (predecessor to Vertex AI). Backed by
// the ml.googleapis.com API (google.golang.org/api/ml/v1). Distinct from
// `gcloud ai ...`, which uses the newer regional aiplatform.googleapis.com
// endpoint.

var aiPlatformCmd = &cobra.Command{Use: "ai-platform", Short: "Manage AI Platform (Cloud ML Engine legacy)"}

var (
	flagAIPlatformPredictModel   string
	flagAIPlatformPredictVersion string
	flagAIPlatformPredictReqFile string
	flagAIPlatformPredictFormat  string
)

var aiPlatformPredictCmd = &cobra.Command{
	Use:   "predict",
	Short: "Run online prediction against an AI Platform deployed model",
	Args:  cobra.NoArgs,
	RunE:  runAIPlatformPredict,
}

func init() {
	aiPlatformPredictCmd.Flags().StringVar(&flagAIPlatformPredictModel, "model", "", "Model name (required)")
	aiPlatformPredictCmd.Flags().StringVar(&flagAIPlatformPredictVersion, "version", "", "Version name; defaults to the model's default version")
	aiPlatformPredictCmd.Flags().StringVar(&flagAIPlatformPredictReqFile, "json-request", "",
		"Path to a JSON file with the prediction request body (required)")
	aiPlatformPredictCmd.Flags().StringVar(&flagAIPlatformPredictFormat, "format", "", "Output format")
	_ = aiPlatformPredictCmd.MarkFlagRequired("model")
	_ = aiPlatformPredictCmd.MarkFlagRequired("json-request")

	aiPlatformCmd.AddCommand(aiPlatformPredictCmd)
	rootCmd.AddCommand(aiPlatformCmd)
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

func mlUpdatedIam(who string) {
	fmt.Fprintf(os.Stderr, "Updated IAM policy for %s.\n", who)
}

// --- ai-platform predict ---

func runAIPlatformPredict(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(flagAIPlatformPredictReqFile)
	if err != nil {
		return fmt.Errorf("reading request body: %w", err)
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var name string
	if flagAIPlatformPredictVersion != "" {
		name = mlVersionName(project, flagAIPlatformPredictModel, flagAIPlatformPredictVersion)
	} else {
		name = mlModelName(project, flagAIPlatformPredictModel)
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
	if flagAIPlatformPredictFormat == "" {
		fmt.Println(resp.Data)
		return nil
	}
	return emitFormatted(resp, flagAIPlatformPredictFormat)
}
