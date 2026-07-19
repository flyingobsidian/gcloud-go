package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	sourcerepo "google.golang.org/api/sourcerepo/v1"
)

// --- gcloud source project-configs (#1153) ---

var sourceProjectConfigsCmd = &cobra.Command{Use: "project-configs", Short: "Manage per-project source repository configuration"}

var (
	flagSourcePCFormat     string
	flagSourcePCConfigFile string
	flagSourcePCUpdateMask string
)

var (
	sourcePCDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe the source repositories project config",
		Args: cobra.NoArgs, RunE: runSourcePCDescribe,
	}
	sourcePCUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update the source repositories project config",
		Args: cobra.NoArgs, RunE: runSourcePCUpdate,
	}
)

func init() {
	for _, c := range []*cobra.Command{sourcePCDescribeCmd, sourcePCUpdateCmd} {
		c.Flags().StringVar(&flagSourcePCFormat, "format", "", "Output format")
	}
	sourcePCUpdateCmd.Flags().StringVar(&flagSourcePCConfigFile, "config-file", "", "YAML/JSON file with the ProjectConfig body (required)")
	_ = sourcePCUpdateCmd.MarkFlagRequired("config-file")
	sourcePCUpdateCmd.Flags().StringVar(&flagSourcePCUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	sourceProjectConfigsCmd.AddCommand(sourcePCDescribeCmd, sourcePCUpdateCmd)
	sourceCmd.AddCommand(sourceProjectConfigsCmd)
}

func runSourcePCDescribe(cmd *cobra.Command, args []string) error {
	name, err := sourceProjectName()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SourceRepoService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.GetConfig(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing project config: %w", err)
	}
	return emitFormatted(got, flagSourcePCFormat)
}

func runSourcePCUpdate(cmd *cobra.Command, args []string) error {
	name, err := sourceProjectName()
	if err != nil {
		return err
	}
	body := &sourcerepo.ProjectConfig{}
	if err := loadYAMLOrJSONInto(flagSourcePCConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagSourcePCUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.SourceRepoService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.UpdateConfig(name, &sourcerepo.UpdateProjectConfigRequest{
		ProjectConfig: body,
		UpdateMask:    mask,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating project config: %w", err)
	}
	return emitFormatted(got, flagSourcePCFormat)
}
