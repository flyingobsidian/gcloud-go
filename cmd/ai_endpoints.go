package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai endpoints (#1452) ---

var aiEndpointsCmd = &cobra.Command{Use: "endpoints", Short: "Manage Vertex AI endpoints"}

var (
	flagAIEPRegion     string
	flagAIEPFormat     string
	flagAIEPConfigFile string
	flagAIEPUpdateMask string
	flagAIEPEndpointID string
	flagAIEPFilter     string
	flagAIEPOrderBy    string
	flagAIEPPageSize   int64
	flagAIEPReadMask   string
)

var (
	aiEPCreateCmd = &cobra.Command{
		Use: "create", Short: "Create an endpoint",
		Args: cobra.NoArgs, RunE: runAIEPCreate,
	}
	aiEPDeleteCmd = &cobra.Command{
		Use: "delete ENDPOINT", Short: "Delete an endpoint",
		Args: cobra.ExactArgs(1), RunE: runAIEPDelete,
	}
	aiEPDeployModelCmd = &cobra.Command{
		Use: "deploy-model ENDPOINT", Short: "Deploy a model to an endpoint",
		Args: cobra.ExactArgs(1), RunE: runAIEPDeployModel,
	}
	aiEPUndeployModelCmd = &cobra.Command{
		Use: "undeploy-model ENDPOINT", Short: "Undeploy a model from an endpoint",
		Args: cobra.ExactArgs(1), RunE: runAIEPUndeployModel,
	}
	aiEPDescribeCmd = &cobra.Command{
		Use: "describe ENDPOINT", Short: "Describe an endpoint",
		Args: cobra.ExactArgs(1), RunE: runAIEPDescribe,
	}
	aiEPListCmd = &cobra.Command{
		Use: "list", Short: "List endpoints",
		Args: cobra.NoArgs, RunE: runAIEPList,
	}
	aiEPUpdateCmd = &cobra.Command{
		Use: "update ENDPOINT", Short: "Update an endpoint",
		Args: cobra.ExactArgs(1), RunE: runAIEPUpdate,
	}
	aiEPPredictCmd = &cobra.Command{
		Use: "predict ENDPOINT", Short: "Send a prediction request to an endpoint",
		Args: cobra.ExactArgs(1), RunE: runAIEPPredict,
	}
	aiEPRawPredictCmd = &cobra.Command{
		Use: "raw-predict ENDPOINT", Short: "Send a raw prediction request to an endpoint",
		Args: cobra.ExactArgs(1), RunE: runAIEPRawPredict,
	}
	aiEPExplainCmd = &cobra.Command{
		Use: "explain ENDPOINT", Short: "Send an explanation request to an endpoint",
		Args: cobra.ExactArgs(1), RunE: runAIEPExplain,
	}
	aiEPDirectPredictCmd = &cobra.Command{
		Use: "direct-predict ENDPOINT", Short: "Send a direct prediction request to a dedicated endpoint",
		Args: cobra.ExactArgs(1), RunE: runAIEPDirectPredict,
	}
	aiEPDirectRawPredictCmd = &cobra.Command{
		Use: "direct-raw-predict ENDPOINT", Short: "Send a direct raw prediction request to a dedicated endpoint",
		Args: cobra.ExactArgs(1), RunE: runAIEPDirectRawPredict,
	}
	aiEPStreamDirectPredictCmd = &cobra.Command{
		Use: "stream-direct-predict ENDPOINT",
		Short: "Send a streaming direct prediction request " +
			"(REST fallback: returns the single response emitted by DirectPredict)",
		Long: "Vertex AI's stream-direct-predict is gRPC-only. Over the REST " +
			"transport used by gcloud-go this issues a single DirectPredict " +
			"call and prints the one response returned by the server.",
		Args: cobra.ExactArgs(1), RunE: runAIEPDirectPredict,
	}
	aiEPStreamDirectRawPredictCmd = &cobra.Command{
		Use: "stream-direct-raw-predict ENDPOINT",
		Short: "Send a streaming direct raw prediction request " +
			"(REST fallback: returns the single response emitted by DirectRawPredict)",
		Long: "Vertex AI's stream-direct-raw-predict is gRPC-only. Over the REST " +
			"transport used by gcloud-go this issues a single DirectRawPredict " +
			"call and prints the one response returned by the server.",
		Args: cobra.ExactArgs(1), RunE: runAIEPDirectRawPredict,
	}
	aiEPStreamRawPredictCmd = &cobra.Command{
		Use: "stream-raw-predict ENDPOINT",
		Short: "Send a streaming raw prediction request " +
			"(returns the full aggregated response body over REST)",
		Args: cobra.ExactArgs(1), RunE: runAIEPStreamRawPredict,
	}
)

func init() {
	all := []*cobra.Command{
		aiEPCreateCmd, aiEPDeleteCmd, aiEPDeployModelCmd, aiEPUndeployModelCmd,
		aiEPDescribeCmd, aiEPListCmd, aiEPUpdateCmd,
		aiEPPredictCmd, aiEPRawPredictCmd, aiEPExplainCmd,
		aiEPDirectPredictCmd, aiEPDirectRawPredictCmd,
		aiEPStreamDirectPredictCmd, aiEPStreamDirectRawPredictCmd, aiEPStreamRawPredictCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagAIEPRegion, "region", "", "Region where the endpoint lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagAIEPFormat, "format", "", "Output format")
	}

	for _, c := range []*cobra.Command{
		aiEPCreateCmd, aiEPDeployModelCmd, aiEPUndeployModelCmd, aiEPUpdateCmd,
		aiEPPredictCmd, aiEPRawPredictCmd, aiEPExplainCmd,
		aiEPDirectPredictCmd, aiEPDirectRawPredictCmd,
		aiEPStreamDirectPredictCmd, aiEPStreamDirectRawPredictCmd, aiEPStreamRawPredictCmd,
	} {
		c.Flags().StringVar(&flagAIEPConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}

	aiEPCreateCmd.Flags().StringVar(&flagAIEPEndpointID, "endpoint-id", "",
		"Optional caller-supplied endpoint ID (immutable)")
	aiEPUpdateCmd.Flags().StringVar(&flagAIEPUpdateMask, "update-mask", "",
		"Comma-separated field mask; defaults to top-level fields in --config-file")

	aiEPListCmd.Flags().StringVar(&flagAIEPFilter, "filter", "", "Server-side filter expression")
	aiEPListCmd.Flags().StringVar(&flagAIEPOrderBy, "order-by", "", "Order-by expression")
	aiEPListCmd.Flags().Int64Var(&flagAIEPPageSize, "page-size", 0, "Maximum results per page")
	aiEPListCmd.Flags().StringVar(&flagAIEPReadMask, "read-mask", "", "Field mask for reads")

	aiEndpointsCmd.AddCommand(all...)
	aiCmd.AddCommand(aiEndpointsCmd)
}

func aiEPParent() (string, error) { return aiParent(flagAIEPRegion) }

func aiEPName(id string) (string, error) {
	parent, err := aiEPParent()
	if err != nil {
		return "", err
	}
	return aiChild("endpoints", id, parent), nil
}

func runAIEPCreate(cmd *cobra.Command, args []string) error {
	parent, err := aiEPParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1Endpoint{}
	if err := loadYAMLOrJSONInto(flagAIEPConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIEPRegion)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Endpoints.Create(parent, body).Context(ctx)
	if flagAIEPEndpointID != "" {
		call = call.EndpointId(flagAIEPEndpointID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating endpoint: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Create request issued (operation: %s).\n", op.Name)
	return emitFormatted(op, flagAIEPFormat)
}

func runAIEPDelete(cmd *cobra.Command, args []string) error {
	name, err := aiEPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIEPRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Endpoints.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting endpoint: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Delete request issued for endpoint [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagAIEPFormat)
}

func runAIEPDeployModel(cmd *cobra.Command, args []string) error {
	name, err := aiEPName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1DeployModelRequest{}
	if err := loadYAMLOrJSONInto(flagAIEPConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIEPRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Endpoints.DeployModel(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deploying model: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Deploy request issued for endpoint [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagAIEPFormat)
}

func runAIEPUndeployModel(cmd *cobra.Command, args []string) error {
	name, err := aiEPName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1UndeployModelRequest{}
	if err := loadYAMLOrJSONInto(flagAIEPConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIEPRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Endpoints.UndeployModel(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeploying model: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Undeploy request issued for endpoint [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagAIEPFormat)
}

func runAIEPDescribe(cmd *cobra.Command, args []string) error {
	name, err := aiEPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIEPRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Endpoints.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing endpoint: %w", err)
	}
	return emitFormatted(got, flagAIEPFormat)
}

func runAIEPList(cmd *cobra.Command, args []string) error {
	parent, err := aiEPParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIEPRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1Endpoint
	pageToken := ""
	for {
		call := svc.Projects.Locations.Endpoints.List(parent).Context(ctx)
		if flagAIEPFilter != "" {
			call = call.Filter(flagAIEPFilter)
		}
		if flagAIEPOrderBy != "" {
			call = call.OrderBy(flagAIEPOrderBy)
		}
		if flagAIEPPageSize > 0 {
			call = call.PageSize(flagAIEPPageSize)
		}
		if flagAIEPReadMask != "" {
			call = call.ReadMask(flagAIEPReadMask)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing endpoints: %w", err)
		}
		all = append(all, resp.Endpoints...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagAIEPFormat)
}

func runAIEPUpdate(cmd *cobra.Command, args []string) error {
	name, err := aiEPName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1Endpoint{}
	if err := loadYAMLOrJSONInto(flagAIEPConfigFile, body); err != nil {
		return err
	}
	mask := flagAIEPUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIEPRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Endpoints.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating endpoint: %w", err)
	}
	return emitFormatted(got, flagAIEPFormat)
}

func runAIEPPredict(cmd *cobra.Command, args []string) error {
	name, err := aiEPName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1PredictRequest{}
	if err := loadYAMLOrJSONInto(flagAIEPConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIEPRegion)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Endpoints.Predict(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running predict: %w", err)
	}
	return emitFormatted(resp, flagAIEPFormat)
}

func runAIEPRawPredict(cmd *cobra.Command, args []string) error {
	name, err := aiEPName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1RawPredictRequest{}
	if err := loadYAMLOrJSONInto(flagAIEPConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIEPRegion)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Endpoints.RawPredict(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running raw-predict: %w", err)
	}
	return emitFormatted(resp, flagAIEPFormat)
}

func runAIEPExplain(cmd *cobra.Command, args []string) error {
	name, err := aiEPName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1ExplainRequest{}
	if err := loadYAMLOrJSONInto(flagAIEPConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIEPRegion)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Endpoints.Explain(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running explain: %w", err)
	}
	return emitFormatted(resp, flagAIEPFormat)
}

func runAIEPDirectPredict(cmd *cobra.Command, args []string) error {
	name, err := aiEPName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1DirectPredictRequest{}
	if err := loadYAMLOrJSONInto(flagAIEPConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIEPRegion)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Endpoints.DirectPredict(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running direct-predict: %w", err)
	}
	return emitFormatted(resp, flagAIEPFormat)
}

func runAIEPDirectRawPredict(cmd *cobra.Command, args []string) error {
	name, err := aiEPName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1DirectRawPredictRequest{}
	if err := loadYAMLOrJSONInto(flagAIEPConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIEPRegion)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Endpoints.DirectRawPredict(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running direct-raw-predict: %w", err)
	}
	return emitFormatted(resp, flagAIEPFormat)
}

func runAIEPStreamRawPredict(cmd *cobra.Command, args []string) error {
	name, err := aiEPName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1StreamRawPredictRequest{}
	if err := loadYAMLOrJSONInto(flagAIEPConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIEPRegion)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Endpoints.StreamRawPredict(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("running stream-raw-predict: %w", err)
	}
	return emitFormatted(resp, flagAIEPFormat)
}
