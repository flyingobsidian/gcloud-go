package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai indexes (#1455) ---

func (s *aiCmds) initAiIndexes() {
	all := []*cobra.Command{
		s.aiIdxCreateCmd, s.aiIdxDeleteCmd, s.aiIdxDescribeCmd, s.aiIdxListCmd,
		s.aiIdxUpdateCmd, s.aiIdxUpsertCmd, s.aiIdxRemoveCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagAIIdxRegion, "region", "", "Region where the index lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagAIIdxFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{
		s.aiIdxCreateCmd, s.aiIdxUpdateCmd, s.aiIdxUpsertCmd, s.aiIdxRemoveCmd,
	} {
		c.Flags().StringVar(&s.flagAIIdxConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.aiIdxUpdateCmd.Flags().StringVar(&s.flagAIIdxUpdateMask, "update-mask", "",
		"Comma-separated field mask; defaults to top-level fields in --config-file")
	s.aiIdxListCmd.Flags().StringVar(&s.flagAIIdxFilter, "filter", "", "Server-side filter expression")
	s.aiIdxListCmd.Flags().Int64Var(&s.flagAIIdxPageSize, "page-size", 0, "Maximum results per page")
	s.aiIdxListCmd.Flags().StringVar(&s.flagAIIdxReadMask, "read-mask", "", "Field mask for reads")

	s.aiIdxCmd.AddCommand(all...)
	s.aiCmd.AddCommand(s.aiIdxCmd)
}

func (s *aiCmds) aiIdxParent() (string, error) { return s.aiParent(s.flagAIIdxRegion) }

func (s *aiCmds) aiIdxName(id string) (string, error) {
	parent, err := s.aiIdxParent()
	if err != nil {
		return "", err
	}
	return aiChild("indexes", id, parent), nil
}

func (s *aiCmds) runAIIdxCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.aiIdxParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1Index{}
	if err := loadYAMLOrJSONInto(s.flagAIIdxConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIIdxRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Indexes.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating index: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Create request issued (operation: %s).\n", op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAIIdxFormat)
}

func (s *aiCmds) runAIIdxDelete(cmd *cobra.Command, args []string) error {
	name, err := s.aiIdxName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIIdxRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Indexes.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting index: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Delete request issued for index [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAIIdxFormat)
}

func (s *aiCmds) runAIIdxDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.aiIdxName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIIdxRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Indexes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing index: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAIIdxFormat)
}

func (s *aiCmds) runAIIdxList(cmd *cobra.Command, args []string) error {
	parent, err := s.aiIdxParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIIdxRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1Index
	pageToken := ""
	for {
		call := svc.Projects.Locations.Indexes.List(parent).Context(ctx)
		if s.flagAIIdxFilter != "" {
			call = call.Filter(s.flagAIIdxFilter)
		}
		if s.flagAIIdxPageSize > 0 {
			call = call.PageSize(s.flagAIIdxPageSize)
		}
		if s.flagAIIdxReadMask != "" {
			call = call.ReadMask(s.flagAIIdxReadMask)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagAIIdxFormat)
}

func (s *aiCmds) runAIIdxUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.aiIdxName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1Index{}
	if err := loadYAMLOrJSONInto(s.flagAIIdxConfigFile, body); err != nil {
		return err
	}
	mask := s.flagAIIdxUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIIdxRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Indexes.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating index: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagAIIdxFormat)
}

func (s *aiCmds) runAIIdxUpsert(cmd *cobra.Command, args []string) error {
	name, err := s.aiIdxName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1UpsertDatapointsRequest{}
	if err := loadYAMLOrJSONInto(s.flagAIIdxConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIIdxRegion)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Indexes.UpsertDatapoints(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("upserting datapoints: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagAIIdxFormat)
}

func (s *aiCmds) runAIIdxRemove(cmd *cobra.Command, args []string) error {
	name, err := s.aiIdxName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1RemoveDatapointsRequest{}
	if err := loadYAMLOrJSONInto(s.flagAIIdxConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIIdxRegion)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Indexes.RemoveDatapoints(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing datapoints: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagAIIdxFormat)
}
