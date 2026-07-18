package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai indexes (#1455) ---

var aiIdxCmd = &cobra.Command{Use: "indexes", Short: "Manage Vertex AI indexes"}

var (
	flagAIIdxRegion     string
	flagAIIdxFormat     string
	flagAIIdxConfigFile string
	flagAIIdxUpdateMask string
	flagAIIdxFilter     string
	flagAIIdxPageSize   int64
	flagAIIdxReadMask   string
)

var (
	aiIdxCreateCmd = &cobra.Command{
		Use: "create", Short: "Create an index",
		Args: cobra.NoArgs, RunE: runAIIdxCreate,
	}
	aiIdxDeleteCmd = &cobra.Command{
		Use: "delete INDEX", Short: "Delete an index",
		Args: cobra.ExactArgs(1), RunE: runAIIdxDelete,
	}
	aiIdxDescribeCmd = &cobra.Command{
		Use: "describe INDEX", Short: "Describe an index",
		Args: cobra.ExactArgs(1), RunE: runAIIdxDescribe,
	}
	aiIdxListCmd = &cobra.Command{
		Use: "list", Short: "List indexes",
		Args: cobra.NoArgs, RunE: runAIIdxList,
	}
	aiIdxUpdateCmd = &cobra.Command{
		Use: "update INDEX", Short: "Update an index",
		Args: cobra.ExactArgs(1), RunE: runAIIdxUpdate,
	}
	aiIdxUpsertCmd = &cobra.Command{
		Use: "upsert-datapoints INDEX", Short: "Upsert datapoints in an index",
		Args: cobra.ExactArgs(1), RunE: runAIIdxUpsert,
	}
	aiIdxRemoveCmd = &cobra.Command{
		Use: "remove-datapoints INDEX", Short: "Remove datapoints from an index",
		Args: cobra.ExactArgs(1), RunE: runAIIdxRemove,
	}
)

func init() {
	all := []*cobra.Command{
		aiIdxCreateCmd, aiIdxDeleteCmd, aiIdxDescribeCmd, aiIdxListCmd,
		aiIdxUpdateCmd, aiIdxUpsertCmd, aiIdxRemoveCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagAIIdxRegion, "region", "", "Region where the index lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagAIIdxFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{
		aiIdxCreateCmd, aiIdxUpdateCmd, aiIdxUpsertCmd, aiIdxRemoveCmd,
	} {
		c.Flags().StringVar(&flagAIIdxConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	aiIdxUpdateCmd.Flags().StringVar(&flagAIIdxUpdateMask, "update-mask", "",
		"Comma-separated field mask; defaults to top-level fields in --config-file")
	aiIdxListCmd.Flags().StringVar(&flagAIIdxFilter, "filter", "", "Server-side filter expression")
	aiIdxListCmd.Flags().Int64Var(&flagAIIdxPageSize, "page-size", 0, "Maximum results per page")
	aiIdxListCmd.Flags().StringVar(&flagAIIdxReadMask, "read-mask", "", "Field mask for reads")

	aiIdxCmd.AddCommand(all...)
	aiCmd.AddCommand(aiIdxCmd)
}

func aiIdxParent() (string, error) { return aiParent(flagAIIdxRegion) }

func aiIdxName(id string) (string, error) {
	parent, err := aiIdxParent()
	if err != nil {
		return "", err
	}
	return aiChild("indexes", id, parent), nil
}

func runAIIdxCreate(cmd *cobra.Command, args []string) error {
	parent, err := aiIdxParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1Index{}
	if err := loadYAMLOrJSONInto(flagAIIdxConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIIdxRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Indexes.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating index: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Create request issued (operation: %s).\n", op.Name)
	return emitFormatted(op, flagAIIdxFormat)
}

func runAIIdxDelete(cmd *cobra.Command, args []string) error {
	name, err := aiIdxName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIIdxRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Indexes.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting index: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Delete request issued for index [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagAIIdxFormat)
}

func runAIIdxDescribe(cmd *cobra.Command, args []string) error {
	name, err := aiIdxName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIIdxRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Indexes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing index: %w", err)
	}
	return emitFormatted(got, flagAIIdxFormat)
}

func runAIIdxList(cmd *cobra.Command, args []string) error {
	parent, err := aiIdxParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIIdxRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1Index
	pageToken := ""
	for {
		call := svc.Projects.Locations.Indexes.List(parent).Context(ctx)
		if flagAIIdxFilter != "" {
			call = call.Filter(flagAIIdxFilter)
		}
		if flagAIIdxPageSize > 0 {
			call = call.PageSize(flagAIIdxPageSize)
		}
		if flagAIIdxReadMask != "" {
			call = call.ReadMask(flagAIIdxReadMask)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing indexes: %w", err)
		}
		all = append(all, resp.Indexes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagAIIdxFormat)
}

func runAIIdxUpdate(cmd *cobra.Command, args []string) error {
	name, err := aiIdxName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1Index{}
	if err := loadYAMLOrJSONInto(flagAIIdxConfigFile, body); err != nil {
		return err
	}
	mask := flagAIIdxUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIIdxRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Indexes.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating index: %w", err)
	}
	return emitFormatted(op, flagAIIdxFormat)
}

func runAIIdxUpsert(cmd *cobra.Command, args []string) error {
	name, err := aiIdxName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1UpsertDatapointsRequest{}
	if err := loadYAMLOrJSONInto(flagAIIdxConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIIdxRegion)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Indexes.UpsertDatapoints(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("upserting datapoints: %w", err)
	}
	return emitFormatted(resp, flagAIIdxFormat)
}

func runAIIdxRemove(cmd *cobra.Command, args []string) error {
	name, err := aiIdxName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1RemoveDatapointsRequest{}
	if err := loadYAMLOrJSONInto(flagAIIdxConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAIIdxRegion)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Indexes.RemoveDatapoints(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing datapoints: %w", err)
	}
	return emitFormatted(resp, flagAIIdxFormat)
}
