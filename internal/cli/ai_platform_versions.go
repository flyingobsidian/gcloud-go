package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	ml "google.golang.org/api/ml/v1"
)

// --- gcloud ai-platform versions (#986) ---

func (s *aiPlatformCmds) initAiPlatformVersions() {
	all := []*cobra.Command{
		s.aiPlatformVersionsCreateCmd, s.aiPlatformVersionsDeleteCmd, s.aiPlatformVersionsDescribeCmd,
		s.aiPlatformVersionsListCmd, s.aiPlatformVersionsSetDefaultCmd, s.aiPlatformVersionsUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagAIPlatformVersionsFormat, "format", "", "Output format")
		c.Flags().StringVar(&s.flagAIPlatformVersionsModel, "model", "", "Parent model name (required)")
		_ = c.MarkFlagRequired("model")
	}
	for _, c := range []*cobra.Command{s.aiPlatformVersionsCreateCmd, s.aiPlatformVersionsUpdateCmd} {
		c.Flags().StringVar(&s.flagAIPlatformVersionsConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Version body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.aiPlatformVersionsUpdateCmd.Flags().StringVar(&s.flagAIPlatformVersionsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update; defaults to the populated top-level fields in --config-file")

	s.aiPlatformVersionsListCmd.Flags().Int64Var(&s.flagAIPlatformVersionsPageSize, "page-size", 0, "Maximum results per page")
	s.aiPlatformVersionsListCmd.Flags().StringVar(&s.flagAIPlatformVersionsFilter, "filter", "", "List filter expression")

	s.aiPlatformVersionsCmd.AddCommand(all...)
	s.aiPlatformCmd.AddCommand(s.aiPlatformVersionsCmd)
}

func (s *aiPlatformCmds) runAIPlatformVersionsCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &ml.GoogleCloudMlV1__Version{}
	if err := loadYAMLOrJSONInto(s.flagAIPlatformVersionsConfigFile, body); err != nil {
		return err
	}
	body.Name = args[0]
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Models.Versions.Create(mlModelName(project, s.flagAIPlatformVersionsModel), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating version: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for version [%s] under model [%s].\n", args[0], s.flagAIPlatformVersionsModel)
	return emitFormattedTo(s.inv.stdout, op, s.flagAIPlatformVersionsFormat)
}

func (s *aiPlatformCmds) runAIPlatformVersionsDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Models.Versions.Delete(mlVersionName(project, s.flagAIPlatformVersionsModel, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting version: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for version [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagAIPlatformVersionsFormat)
}

func (s *aiPlatformCmds) runAIPlatformVersionsDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Models.Versions.Get(mlVersionName(project, s.flagAIPlatformVersionsModel, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing version: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAIPlatformVersionsFormat)
}

func (s *aiPlatformCmds) runAIPlatformVersionsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	var all []*ml.GoogleCloudMlV1__Version
	pageToken := ""
	for {
		call := svc.Projects.Models.Versions.List(mlModelName(project, s.flagAIPlatformVersionsModel)).Context(ctx)
		if s.flagAIPlatformVersionsPageSize > 0 {
			call = call.PageSize(s.flagAIPlatformVersionsPageSize)
		}
		if s.flagAIPlatformVersionsFilter != "" {
			call = call.Filter(s.flagAIPlatformVersionsFilter)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagAIPlatformVersionsFormat)
}

func (s *aiPlatformCmds) runAIPlatformVersionsSetDefault(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Models.Versions.SetDefault(
		mlVersionName(project, s.flagAIPlatformVersionsModel, args[0]),
		&ml.GoogleCloudMlV1__SetDefaultVersionRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting default version: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Set version [%s] as the default for model [%s].\n", args[0], s.flagAIPlatformVersionsModel)
	return emitFormattedTo(s.inv.stdout, got, s.flagAIPlatformVersionsFormat)
}

func (s *aiPlatformCmds) runAIPlatformVersionsUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &ml.GoogleCloudMlV1__Version{}
	if err := loadYAMLOrJSONInto(s.flagAIPlatformVersionsConfigFile, body); err != nil {
		return err
	}
	mask := s.flagAIPlatformVersionsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Models.Versions.Patch(mlVersionName(project, s.flagAIPlatformVersionsModel, args[0]), body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating version: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for version [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagAIPlatformVersionsFormat)
}
