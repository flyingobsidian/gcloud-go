package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts projects (#1078) ---

func (s *artifactsCmds) initArtifactsProjects() {
	s.artifactsProjectsDescribeCmd.Flags().StringVar(&s.flagArtProjectsFormat, "format", "", "Output format")

	s.artifactsProjectsUpdateCmd.Flags().StringVar(&s.flagArtProjectsFormat, "format", "", "Output format")
	s.artifactsProjectsUpdateCmd.Flags().StringVar(&s.flagArtProjectsConfigFile, "config-file", "", "YAML/JSON file with ProjectSettings body")
	s.artifactsProjectsUpdateCmd.Flags().StringVar(&s.flagArtProjectsMask, "update-mask", "", "Fields to update (comma-separated)")

	s.artifactsProjectsCmd.AddCommand(s.artifactsProjectsDescribeCmd)
	s.artifactsProjectsCmd.AddCommand(s.artifactsProjectsUpdateCmd)
	s.artifactsCmd.AddCommand(s.artifactsProjectsCmd)
}

func (s *artifactsCmds) runArtifactsProjectsDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	settings, err := svc.Projects.GetProjectSettings(fmt.Sprintf("projects/%s/projectSettings", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing project settings: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, settings, s.flagArtProjectsFormat)
}

func (s *artifactsCmds) runArtifactsProjectsUpdate(cmd *cobra.Command, args []string) error {
	if s.flagArtProjectsConfigFile == "" {
		return fmt.Errorf("--config-file is required")
	}
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &artifactregistry.ProjectSettings{}
	if err := loadYAMLOrJSONInto(s.flagArtProjectsConfigFile, body); err != nil {
		return err
	}
	mask := s.flagArtProjectsMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
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
	return emitFormattedTo(s.inv.stdout, out, s.flagArtProjectsFormat)
}
