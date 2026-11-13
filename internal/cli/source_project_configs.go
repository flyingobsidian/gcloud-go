package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	sourcerepo "google.golang.org/api/sourcerepo/v1"
)

// --- gcloud source project-configs (#1153) ---

func (s *sourceCmds) initSourceProjectConfigs() {
	for _, c := range []*cobra.Command{s.sourcePCDescribeCmd, s.sourcePCUpdateCmd} {
		c.Flags().StringVar(&s.flagSourcePCFormat, "format", "", "Output format")
	}
	s.sourcePCUpdateCmd.Flags().StringVar(&s.flagSourcePCConfigFile, "config-file", "", "YAML/JSON file with the ProjectConfig body (required)")
	_ = s.sourcePCUpdateCmd.MarkFlagRequired("config-file")
	s.sourcePCUpdateCmd.Flags().StringVar(&s.flagSourcePCUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	s.sourceProjectConfigsCmd.AddCommand(s.sourcePCDescribeCmd, s.sourcePCUpdateCmd)
	s.sourceCmd.AddCommand(s.sourceProjectConfigsCmd)
}

func (s *sourceCmds) runSourcePCDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.sourceProjectName()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SourceRepo(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.GetConfig(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing project config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSourcePCFormat)
}

func (s *sourceCmds) runSourcePCUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.sourceProjectName()
	if err != nil {
		return err
	}
	body := &sourcerepo.ProjectConfig{}
	if err := loadYAMLOrJSONInto(s.flagSourcePCConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagSourcePCUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SourceRepo(ctx)
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
	return emitFormattedTo(s.inv.stdout, got, s.flagSourcePCFormat)
}
