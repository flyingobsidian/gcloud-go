package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai tensorboards (#1461) ---

var aiTBCmd = &cobra.Command{Use: "tensorboards", Short: "Manage Vertex AI Tensorboards"}

var (
	flagAITBRegion     string
	flagAITBFormat     string
	flagAITBConfigFile string
	flagAITBUpdateMask string
	flagAITBFilter     string
	flagAITBOrderBy    string
	flagAITBPageSize   int64
	flagAITBReadMask   string
)

var (
	aiTBCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a Tensorboard",
		Args: cobra.NoArgs, RunE: runAITBCreate,
	}
	aiTBDeleteCmd = &cobra.Command{
		Use: "delete TENSORBOARD", Short: "Delete a Tensorboard",
		Args: cobra.ExactArgs(1), RunE: runAITBDelete,
	}
	aiTBDescribeCmd = &cobra.Command{
		Use: "describe TENSORBOARD", Short: "Describe a Tensorboard",
		Args: cobra.ExactArgs(1), RunE: runAITBDescribe,
	}
	aiTBListCmd = &cobra.Command{
		Use: "list", Short: "List Tensorboards",
		Args: cobra.NoArgs, RunE: runAITBList,
	}
	aiTBUpdateCmd = &cobra.Command{
		Use: "update TENSORBOARD", Short: "Update a Tensorboard",
		Args: cobra.ExactArgs(1), RunE: runAITBUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		aiTBCreateCmd, aiTBDeleteCmd, aiTBDescribeCmd, aiTBListCmd, aiTBUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagAITBRegion, "region", "", "Region where the Tensorboard lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagAITBFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{aiTBCreateCmd, aiTBUpdateCmd} {
		c.Flags().StringVar(&flagAITBConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Tensorboard body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	aiTBUpdateCmd.Flags().StringVar(&flagAITBUpdateMask, "update-mask", "",
		"Comma-separated field mask; defaults to top-level fields in --config-file")
	aiTBListCmd.Flags().StringVar(&flagAITBFilter, "filter", "", "Server-side filter expression")
	aiTBListCmd.Flags().StringVar(&flagAITBOrderBy, "order-by", "", "Order-by expression")
	aiTBListCmd.Flags().Int64Var(&flagAITBPageSize, "page-size", 0, "Maximum results per page")
	aiTBListCmd.Flags().StringVar(&flagAITBReadMask, "read-mask", "", "Field mask for reads")

	aiTBCmd.AddCommand(all...)
	aiCmd.AddCommand(aiTBCmd)
}

func aiTBParent() (string, error) { return aiParent(flagAITBRegion) }

func aiTBName(id string) (string, error) {
	parent, err := aiTBParent()
	if err != nil {
		return "", err
	}
	return aiChild("tensorboards", id, parent), nil
}

func runAITBCreate(cmd *cobra.Command, args []string) error {
	parent, err := aiTBParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1Tensorboard{}
	if err := loadYAMLOrJSONInto(flagAITBConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAITBRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Tensorboards.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating Tensorboard: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Create request issued (operation: %s).\n", op.Name)
	return emitFormatted(op, flagAITBFormat)
}

func runAITBDelete(cmd *cobra.Command, args []string) error {
	name, err := aiTBName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAITBRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Tensorboards.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting Tensorboard: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Delete request issued for Tensorboard [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagAITBFormat)
}

func runAITBDescribe(cmd *cobra.Command, args []string) error {
	name, err := aiTBName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAITBRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Tensorboards.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing Tensorboard: %w", err)
	}
	return emitFormatted(got, flagAITBFormat)
}

func runAITBList(cmd *cobra.Command, args []string) error {
	parent, err := aiTBParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAITBRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1Tensorboard
	pageToken := ""
	for {
		call := svc.Projects.Locations.Tensorboards.List(parent).Context(ctx)
		if flagAITBFilter != "" {
			call = call.Filter(flagAITBFilter)
		}
		if flagAITBOrderBy != "" {
			call = call.OrderBy(flagAITBOrderBy)
		}
		if flagAITBPageSize > 0 {
			call = call.PageSize(flagAITBPageSize)
		}
		if flagAITBReadMask != "" {
			call = call.ReadMask(flagAITBReadMask)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing Tensorboards: %w", err)
		}
		all = append(all, resp.Tensorboards...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagAITBFormat)
}

func runAITBUpdate(cmd *cobra.Command, args []string) error {
	name, err := aiTBName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1Tensorboard{}
	if err := loadYAMLOrJSONInto(flagAITBConfigFile, body); err != nil {
		return err
	}
	mask := flagAITBUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagAITBRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Tensorboards.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating Tensorboard: %w", err)
	}
	return emitFormatted(op, flagAITBFormat)
}
