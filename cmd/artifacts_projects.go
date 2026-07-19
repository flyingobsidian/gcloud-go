package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts projects (#1078) ---

var artifactsProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage per-project Artifact Registry settings",
}

var (
	flagArtProjectsFormat     string
	flagArtProjectsConfigFile string
	flagArtProjectsMask       string
)

var artifactsProjectsDescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe the Artifact Registry project settings",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsProjectsDescribe,
}

var artifactsProjectsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the Artifact Registry project settings",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsProjectsUpdate,
}

func init() {
	artifactsProjectsDescribeCmd.Flags().StringVar(&flagArtProjectsFormat, "format", "", "Output format")

	artifactsProjectsUpdateCmd.Flags().StringVar(&flagArtProjectsFormat, "format", "", "Output format")
	artifactsProjectsUpdateCmd.Flags().StringVar(&flagArtProjectsConfigFile, "config-file", "", "YAML/JSON file with ProjectSettings body")
	artifactsProjectsUpdateCmd.Flags().StringVar(&flagArtProjectsMask, "update-mask", "", "Fields to update (comma-separated)")

	artifactsProjectsCmd.AddCommand(artifactsProjectsDescribeCmd)
	artifactsProjectsCmd.AddCommand(artifactsProjectsUpdateCmd)
	artifactsCmd.AddCommand(artifactsProjectsCmd)
}

func runArtifactsProjectsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	settings, err := svc.Projects.GetProjectSettings(fmt.Sprintf("projects/%s/projectSettings", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing project settings: %w", err)
	}
	return emitFormatted(settings, flagArtProjectsFormat)
}

func runArtifactsProjectsUpdate(cmd *cobra.Command, args []string) error {
	if flagArtProjectsConfigFile == "" {
		return fmt.Errorf("--config-file is required")
	}
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &artifactregistry.ProjectSettings{}
	if err := loadYAMLOrJSONInto(flagArtProjectsConfigFile, body); err != nil {
		return err
	}
	mask := flagArtProjectsMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("projects/%s/projectSettings", project)
	call := svc.Projects.UpdateProjectSettings(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating project settings: %w", err)
	}
	return emitFormatted(out, flagArtProjectsFormat)
}
