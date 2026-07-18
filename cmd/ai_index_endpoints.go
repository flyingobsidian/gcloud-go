package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai index-endpoints (#1454) ---

var aiIECmd = &cobra.Command{Use: "index-endpoints", Short: "Manage Vertex AI index endpoints"}

var (
	flagAIIERegion     string
	flagAIIEFormat     string
	flagAIIEConfigFile string
	flagAIIEUpdateMask string
	flagAIIEFilter     string
	flagAIIEPageSize   int64
	flagAIIEReadMask   string
)

var (
	aiIECreateCmd = &cobra.Command{
		Use: "create", Short: "Create an index endpoint",
		Args: cobra.NoArgs, RunE: runAIIECreate,
	}
	aiIEDeleteCmd = &cobra.Command{
		Use: "delete INDEX_ENDPOINT", Short: "Delete an index endpoint",
		Args: cobra.ExactArgs(1), RunE: runAIIEDelete,
	}
	aiIEDeployIndexCmd = &cobra.Command{
		Use: "deploy-index INDEX_ENDPOINT", Short: "Deploy an index to an index endpoint",
		Args: cobra.ExactArgs(1), RunE: runAIIEDeployIndex,
	}
	aiIEDescribeCmd = &cobra.Command{
		Use: "describe INDEX_ENDPOINT", Short: "Describe an index endpoint",
		Args: cobra.ExactArgs(1), RunE: runAIIEDescribe,
	}
	aiIEListCmd = &cobra.Command{
		Use: "list", Short: "List index endpoints",
		Args: cobra.NoArgs, RunE: runAIIEList,
	}
	aiIEMutateDeployedIndexCmd = &cobra.Command{
		Use: "mutate-deployed-index INDEX_ENDPOINT",
		Short: "Update a deployed index on an index endpoint",
		Args: cobra.ExactArgs(1), RunE: runAIIEMutateDeployedIndex,
	}
	aiIEUndeployIndexCmd = &cobra.Command{
		Use: "undeploy-index INDEX_ENDPOINT", Short: "Undeploy an index from an index endpoint",
		Args: cobra.ExactArgs(1), RunE: runAIIEUndeployIndex,
	}
	aiIEUpdateCmd = &cobra.Command{
		Use: "update INDEX_ENDPOINT", Short: "Update an index endpoint",
		Args: cobra.ExactArgs(1), RunE: runAIIEUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		aiIECreateCmd, aiIEDeleteCmd, aiIEDeployIndexCmd, aiIEDescribeCmd, aiIEListCmd,
		aiIEMutateDeployedIndexCmd, aiIEUndeployIndexCmd, aiIEUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagAIIERegion, "region", "", "Region where the index endpoint lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagAIIEFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{
		aiIECreateCmd, aiIEDeployIndexCmd, aiIEMutateDeployedIndexCmd,
		aiIEUndeployIndexCmd, aiIEUpdateCmd,
	} {
		c.Flags().StringVar(&flagAIIEConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	aiIEUpdateCmd.Flags().StringVar(&flagAIIEUpdateMask, "update-mask", "",
		"Comma-separated field mask; defaults to top-level fields in --config-file")
	aiIEListCmd.Flags().StringVar(&flagAIIEFilter, "filter", "", "Server-side filter expression")
	aiIEListCmd.Flags().Int64Var(&flagAIIEPageSize, "page-size", 0, "Maximum results per page")
	aiIEListCmd.Flags().StringVar(&flagAIIEReadMask, "read-mask", "", "Field mask for reads")

	aiIECmd.AddCommand(all...)
	aiCmd.AddCommand(aiIECmd)
}

func aiIEParent() (string, error) { return aiParent(flagAIIERegion) }

func aiIEName(id string) (string, error) {
	parent, err := aiIEParent()
	if err != nil {
		return "", err
	}
	return aiChild("indexEndpoints", id, parent), nil
}

func runAIIECreate(cmd *cobra.Command, args []string) error {
	parent, err := aiIEParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1IndexEndpoint{}
	if err := loadYAMLOrJSONInto(flagAIIEConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIIERegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.IndexEndpoints.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating index endpoint: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Create request issued (operation: %s).\n", op.Name)
	return emitFormatted(op, flagAIIEFormat)
}

func runAIIEDelete(cmd *cobra.Command, args []string) error {
	name, err := aiIEName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIIERegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.IndexEndpoints.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting index endpoint: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Delete request issued for index endpoint [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagAIIEFormat)
}

func runAIIEDeployIndex(cmd *cobra.Command, args []string) error {
	name, err := aiIEName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1DeployIndexRequest{}
	if err := loadYAMLOrJSONInto(flagAIIEConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIIERegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.IndexEndpoints.DeployIndex(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deploying index: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Deploy request issued for index endpoint [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagAIIEFormat)
}

func runAIIEDescribe(cmd *cobra.Command, args []string) error {
	name, err := aiIEName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIIERegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.IndexEndpoints.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing index endpoint: %w", err)
	}
	return emitFormatted(got, flagAIIEFormat)
}

func runAIIEList(cmd *cobra.Command, args []string) error {
	parent, err := aiIEParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIIERegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1IndexEndpoint
	pageToken := ""
	for {
		call := svc.Projects.Locations.IndexEndpoints.List(parent).Context(ctx)
		if flagAIIEFilter != "" {
			call = call.Filter(flagAIIEFilter)
		}
		if flagAIIEPageSize > 0 {
			call = call.PageSize(flagAIIEPageSize)
		}
		if flagAIIEReadMask != "" {
			call = call.ReadMask(flagAIIEReadMask)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing index endpoints: %w", err)
		}
		all = append(all, resp.IndexEndpoints...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagAIIEFormat)
}

func runAIIEMutateDeployedIndex(cmd *cobra.Command, args []string) error {
	name, err := aiIEName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1DeployedIndex{}
	if err := loadYAMLOrJSONInto(flagAIIEConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIIERegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.IndexEndpoints.MutateDeployedIndex(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("mutating deployed index: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Mutate request issued for index endpoint [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagAIIEFormat)
}

func runAIIEUndeployIndex(cmd *cobra.Command, args []string) error {
	name, err := aiIEName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1UndeployIndexRequest{}
	if err := loadYAMLOrJSONInto(flagAIIEConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIIERegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.IndexEndpoints.UndeployIndex(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeploying index: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Undeploy request issued for index endpoint [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagAIIEFormat)
}

func runAIIEUpdate(cmd *cobra.Command, args []string) error {
	name, err := aiIEName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1IndexEndpoint{}
	if err := loadYAMLOrJSONInto(flagAIIEConfigFile, body); err != nil {
		return err
	}
	mask := flagAIIEUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIIERegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.IndexEndpoints.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating index endpoint: %w", err)
	}
	return emitFormatted(got, flagAIIEFormat)
}
