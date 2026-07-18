package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	ml "google.golang.org/api/ml/v1"
)

// --- gcloud ai-platform versions (#986) ---

var aiPlatformVersionsCmd = &cobra.Command{Use: "versions", Short: "Manage AI Platform model versions"}

var (
	flagAIPlatformVersionsFormat     string
	flagAIPlatformVersionsModel      string
	flagAIPlatformVersionsConfigFile string
	flagAIPlatformVersionsPageSize   int64
	flagAIPlatformVersionsFilter     string
	flagAIPlatformVersionsUpdateMask string
)

var (
	aiPlatformVersionsCreateCmd = &cobra.Command{
		Use: "create VERSION", Short: "Create an AI Platform model version",
		Args: cobra.ExactArgs(1), RunE: runAIPlatformVersionsCreate,
	}
	aiPlatformVersionsDeleteCmd = &cobra.Command{
		Use: "delete VERSION", Short: "Delete an AI Platform model version",
		Args: cobra.ExactArgs(1), RunE: runAIPlatformVersionsDelete,
	}
	aiPlatformVersionsDescribeCmd = &cobra.Command{
		Use: "describe VERSION", Short: "Describe an AI Platform model version",
		Args: cobra.ExactArgs(1), RunE: runAIPlatformVersionsDescribe,
	}
	aiPlatformVersionsListCmd = &cobra.Command{
		Use: "list", Short: "List AI Platform model versions",
		Args: cobra.NoArgs, RunE: runAIPlatformVersionsList,
	}
	aiPlatformVersionsSetDefaultCmd = &cobra.Command{
		Use: "set-default VERSION", Short: "Set an AI Platform model version as the default",
		Args: cobra.ExactArgs(1), RunE: runAIPlatformVersionsSetDefault,
	}
	aiPlatformVersionsUpdateCmd = &cobra.Command{
		Use: "update VERSION", Short: "Update an AI Platform model version",
		Args: cobra.ExactArgs(1), RunE: runAIPlatformVersionsUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		aiPlatformVersionsCreateCmd, aiPlatformVersionsDeleteCmd, aiPlatformVersionsDescribeCmd,
		aiPlatformVersionsListCmd, aiPlatformVersionsSetDefaultCmd, aiPlatformVersionsUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagAIPlatformVersionsFormat, "format", "", "Output format")
		c.Flags().StringVar(&flagAIPlatformVersionsModel, "model", "", "Parent model name (required)")
		_ = c.MarkFlagRequired("model")
	}
	for _, c := range []*cobra.Command{aiPlatformVersionsCreateCmd, aiPlatformVersionsUpdateCmd} {
		c.Flags().StringVar(&flagAIPlatformVersionsConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Version body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	aiPlatformVersionsUpdateCmd.Flags().StringVar(&flagAIPlatformVersionsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update; defaults to the populated top-level fields in --config-file")

	aiPlatformVersionsListCmd.Flags().Int64Var(&flagAIPlatformVersionsPageSize, "page-size", 0, "Maximum results per page")
	aiPlatformVersionsListCmd.Flags().StringVar(&flagAIPlatformVersionsFilter, "filter", "", "List filter expression")

	aiPlatformVersionsCmd.AddCommand(all...)
	aiPlatformCmd.AddCommand(aiPlatformVersionsCmd)
}

func runAIPlatformVersionsCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &ml.GoogleCloudMlV1__Version{}
	if err := loadYAMLOrJSONInto(flagAIPlatformVersionsConfigFile, body); err != nil {
		return err
	}
	body.Name = args[0]
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Models.Versions.Create(mlModelName(project, flagAIPlatformVersionsModel), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating version: %w", err)
	}
	fmt.Printf("Create request issued for version [%s] under model [%s].\n", args[0], flagAIPlatformVersionsModel)
	return emitFormatted(op, flagAIPlatformVersionsFormat)
}

func runAIPlatformVersionsDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Models.Versions.Delete(mlVersionName(project, flagAIPlatformVersionsModel, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting version: %w", err)
	}
	fmt.Printf("Delete request issued for version [%s].\n", args[0])
	return emitFormatted(op, flagAIPlatformVersionsFormat)
}

func runAIPlatformVersionsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Models.Versions.Get(mlVersionName(project, flagAIPlatformVersionsModel, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing version: %w", err)
	}
	return emitFormatted(got, flagAIPlatformVersionsFormat)
}

func runAIPlatformVersionsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*ml.GoogleCloudMlV1__Version
	pageToken := ""
	for {
		call := svc.Projects.Models.Versions.List(mlModelName(project, flagAIPlatformVersionsModel)).Context(ctx)
		if flagAIPlatformVersionsPageSize > 0 {
			call = call.PageSize(flagAIPlatformVersionsPageSize)
		}
		if flagAIPlatformVersionsFilter != "" {
			call = call.Filter(flagAIPlatformVersionsFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing versions: %w", err)
		}
		all = append(all, resp.Versions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagAIPlatformVersionsFormat)
}

func runAIPlatformVersionsSetDefault(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Models.Versions.SetDefault(
		mlVersionName(project, flagAIPlatformVersionsModel, args[0]),
		&ml.GoogleCloudMlV1__SetDefaultVersionRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting default version: %w", err)
	}
	fmt.Printf("Set version [%s] as the default for model [%s].\n", args[0], flagAIPlatformVersionsModel)
	return emitFormatted(got, flagAIPlatformVersionsFormat)
}

func runAIPlatformVersionsUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &ml.GoogleCloudMlV1__Version{}
	if err := loadYAMLOrJSONInto(flagAIPlatformVersionsConfigFile, body); err != nil {
		return err
	}
	mask := flagAIPlatformVersionsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.MLService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Models.Versions.Patch(mlVersionName(project, flagAIPlatformVersionsModel, args[0]), body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating version: %w", err)
	}
	fmt.Printf("Update request issued for version [%s].\n", args[0])
	return emitFormatted(op, flagAIPlatformVersionsFormat)
}
