package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai tensorboards (#1461) ---

func (s *aiCmds) initAiTensorboards() {
	all := []*cobra.Command{
		s.aiTBCreateCmd, s.aiTBDeleteCmd, s.aiTBDescribeCmd, s.aiTBListCmd, s.aiTBUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagAITBRegion, "region", "", "Region where the Tensorboard lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagAITBFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.aiTBCreateCmd, s.aiTBUpdateCmd} {
		c.Flags().StringVar(&s.flagAITBConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Tensorboard body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.aiTBUpdateCmd.Flags().StringVar(&s.flagAITBUpdateMask, "update-mask", "",
		"Comma-separated field mask; defaults to top-level fields in --config-file")
	s.aiTBListCmd.Flags().StringVar(&s.flagAITBFilter, "filter", "", "Server-side filter expression")
	s.aiTBListCmd.Flags().StringVar(&s.flagAITBOrderBy, "order-by", "", "Order-by expression")
	s.aiTBListCmd.Flags().Int64Var(&s.flagAITBPageSize, "page-size", 0, "Maximum results per page")
	s.aiTBListCmd.Flags().StringVar(&s.flagAITBReadMask, "read-mask", "", "Field mask for reads")

	s.aiTBCmd.AddCommand(all...)
	s.aiCmd.AddCommand(s.aiTBCmd)
}

func (s *aiCmds) aiTBParent() (string, error) { return s.aiParent(s.flagAITBRegion) }

func (s *aiCmds) aiTBName(id string) (string, error) {
	parent, err := s.aiTBParent()
	if err != nil {
		return "", err
	}
	return aiChild("tensorboards", id, parent), nil
}

func (s *aiCmds) runAITBCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.aiTBParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1Tensorboard{}
	if err := loadYAMLOrJSONInto(s.flagAITBConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAITBRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Tensorboards.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating Tensorboard: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Create request issued (operation: %s).\n", op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAITBFormat)
}

func (s *aiCmds) runAITBDelete(cmd *cobra.Command, args []string) error {
	name, err := s.aiTBName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAITBRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Tensorboards.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting Tensorboard: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Delete request issued for Tensorboard [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAITBFormat)
}

func (s *aiCmds) runAITBDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.aiTBName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAITBRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Tensorboards.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing Tensorboard: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAITBFormat)
}

func (s *aiCmds) runAITBList(cmd *cobra.Command, args []string) error {
	parent, err := s.aiTBParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAITBRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1Tensorboard
	pageToken := ""
	for {
		call := svc.Projects.Locations.Tensorboards.List(parent).Context(ctx)
		if s.flagAITBFilter != "" {
			call = call.Filter(s.flagAITBFilter)
		}
		if s.flagAITBOrderBy != "" {
			call = call.OrderBy(s.flagAITBOrderBy)
		}
		if s.flagAITBPageSize > 0 {
			call = call.PageSize(s.flagAITBPageSize)
		}
		if s.flagAITBReadMask != "" {
			call = call.ReadMask(s.flagAITBReadMask)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagAITBFormat)
}

func (s *aiCmds) runAITBUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.aiTBName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1Tensorboard{}
	if err := loadYAMLOrJSONInto(s.flagAITBConfigFile, body); err != nil {
		return err
	}
	mask := s.flagAITBUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAITBRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Tensorboards.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating Tensorboard: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagAITBFormat)
}
