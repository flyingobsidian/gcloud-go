package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai semantic-governance-policies /
//     gcloud ai semantic-governance-policy-engine (#1748) ---
//
// Vertex AI Semantic Governance surface. Ports the GA promotion shipped in
// gcloud-python 581.0.0 (both command groups moved from beta to GA), plus the
// `semantic-governance-policy-engine deprovision` command added in 574.0.0
// and re-affirmed at 581.0.0. Backed by the v1 aiplatform service; the
// v1beta1 discovery document does not yet expose SemanticGovernance* in the
// Go SDK, so gcloud-go does not currently ship a beta alias.

var aiSGPCmd = &cobra.Command{
	Use:   "semantic-governance-policies",
	Short: "Manage Vertex AI Semantic Governance policies",
}

var aiSGPECmd = &cobra.Command{
	Use:   "semantic-governance-policy-engine",
	Short: "Manage the Vertex AI Semantic Governance policy engine",
}

var (
	flagAISGPRegion     string
	flagAISGPFormat     string
	flagAISGPConfigFile string
	flagAISGPUpdateMask string
	flagAISGPPolicyID   string
	flagAISGPPageSize   int64
)

var (
	aiSGPCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a semantic governance policy",
		Args: cobra.NoArgs, RunE: runAISGPCreate,
	}
	aiSGPDeleteCmd = &cobra.Command{
		Use: "delete POLICY", Short: "Delete a semantic governance policy",
		Args: cobra.ExactArgs(1), RunE: runAISGPDelete,
	}
	aiSGPDescribeCmd = &cobra.Command{
		Use: "describe POLICY", Short: "Describe a semantic governance policy",
		Args: cobra.ExactArgs(1), RunE: runAISGPDescribe,
	}
	aiSGPListCmd = &cobra.Command{
		Use: "list", Short: "List semantic governance policies",
		Args: cobra.NoArgs, RunE: runAISGPList,
	}
	aiSGPUpdateCmd = &cobra.Command{
		Use: "update POLICY", Short: "Update a semantic governance policy",
		Args: cobra.ExactArgs(1), RunE: runAISGPUpdate,
	}

	aiSGPEDeprovisionCmd = &cobra.Command{
		Use: "deprovision", Short: "Tear down the semantic governance policy engine (tenant project, GKE cluster, PSC service attachments)",
		Args: cobra.NoArgs, RunE: runAISGPEDeprovision,
	}
)

func init() {
	policyCmds := []*cobra.Command{
		aiSGPCreateCmd, aiSGPDeleteCmd, aiSGPDescribeCmd, aiSGPListCmd, aiSGPUpdateCmd,
	}
	engineCmds := []*cobra.Command{aiSGPEDeprovisionCmd}

	for _, c := range append(policyCmds, engineCmds...) {
		c.Flags().StringVar(&flagAISGPRegion, "region", "", "Region for the semantic governance resource (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagAISGPFormat, "format", "", "Output format")
	}

	aiSGPCreateCmd.Flags().StringVar(&flagAISGPConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the SemanticGovernancePolicy body (required)")
	_ = aiSGPCreateCmd.MarkFlagRequired("config-file")
	aiSGPCreateCmd.Flags().StringVar(&flagAISGPPolicyID, "policy-id", "",
		"Caller-supplied policy ID (required)")
	_ = aiSGPCreateCmd.MarkFlagRequired("policy-id")

	aiSGPUpdateCmd.Flags().StringVar(&flagAISGPConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the SemanticGovernancePolicy patch body (required)")
	_ = aiSGPUpdateCmd.MarkFlagRequired("config-file")
	aiSGPUpdateCmd.Flags().StringVar(&flagAISGPUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")

	aiSGPListCmd.Flags().Int64Var(&flagAISGPPageSize, "page-size", 0, "Maximum results per page")

	aiSGPCmd.AddCommand(policyCmds...)
	aiSGPECmd.AddCommand(engineCmds...)
	aiCmd.AddCommand(aiSGPCmd, aiSGPECmd)
}

func aiSGPParent() (string, error) { return aiParent(flagAISGPRegion) }

func aiSGPName(id string) (string, error) {
	parent, err := aiSGPParent()
	if err != nil {
		return "", err
	}
	return aiChild("semanticGovernancePolicies", id, parent), nil
}

func runAISGPCreate(cmd *cobra.Command, args []string) error {
	parent, err := aiSGPParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1SemanticGovernancePolicy{}
	if err := loadYAMLOrJSONInto(flagAISGPConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAISGPRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.SemanticGovernancePolicies.Create(parent, body).
		SemanticGovernancePolicyId(flagAISGPPolicyID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating semantic governance policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Create request issued for semantic governance policy [%s] (operation: %s).\n",
		flagAISGPPolicyID, op.Name)
	return emitFormatted(op, flagAISGPFormat)
}

func runAISGPDelete(cmd *cobra.Command, args []string) error {
	name, err := aiSGPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAISGPRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.SemanticGovernancePolicies.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting semantic governance policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Delete request issued for semantic governance policy [%s] (operation: %s).\n",
		args[0], op.Name)
	return emitFormatted(op, flagAISGPFormat)
}

func runAISGPDescribe(cmd *cobra.Command, args []string) error {
	name, err := aiSGPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAISGPRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.SemanticGovernancePolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing semantic governance policy: %w", err)
	}
	return emitFormatted(got, flagAISGPFormat)
}

func runAISGPList(cmd *cobra.Command, args []string) error {
	parent, err := aiSGPParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAISGPRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1SemanticGovernancePolicy
	pageToken := ""
	for {
		call := svc.Projects.Locations.SemanticGovernancePolicies.List(parent).Context(ctx)
		if flagAISGPPageSize > 0 {
			call = call.PageSize(flagAISGPPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing semantic governance policies: %w", err)
		}
		all = append(all, resp.SemanticGovernancePolicies...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagAISGPFormat)
}

func runAISGPUpdate(cmd *cobra.Command, args []string) error {
	name, err := aiSGPName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1SemanticGovernancePolicy{}
	if err := loadYAMLOrJSONInto(flagAISGPConfigFile, body); err != nil {
		return err
	}
	mask := flagAISGPUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAISGPRegion)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.SemanticGovernancePolicies.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating semantic governance policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Update request issued for semantic governance policy [%s] (operation: %s).\n",
		args[0], op.Name)
	return emitFormatted(op, flagAISGPFormat)
}

// aiSGPEName returns the fixed engine resource name for a region. The engine
// is a singleton per (project, location) in the aiplatform v1 discovery
// document.
func aiSGPEName() (string, error) {
	parent, err := aiSGPParent()
	if err != nil {
		return "", err
	}
	return parent + "/semanticGovernancePolicyEngine", nil
}

func runAISGPEDeprovision(cmd *cobra.Command, args []string) error {
	name, err := aiSGPEName()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAISGPRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.SemanticGovernancePolicyEngine.Deprovision(name,
		&aiplatform.GoogleCloudAiplatformV1DeprovisionSemanticGovernancePolicyEngineRequest{}).
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deprovisioning semantic governance policy engine: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Deprovision request issued for semantic governance policy engine (operation: %s).\n",
		op.Name)
	return emitFormatted(op, flagAISGPFormat)
}
