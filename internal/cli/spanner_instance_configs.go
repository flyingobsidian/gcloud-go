package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	spanner "google.golang.org/api/spanner/v1"
)

// --- gcloud spanner instance-configs (#1208) ---

func (s *spannerCmds) initSpannerInstanceConfigs() {
	all := []*cobra.Command{s.spannerICCreateCmd, s.spannerICDeleteCmd, s.spannerICDescribeCmd, s.spannerICListCmd, s.spannerICUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagSpICFormat, "format", "", "Output format")
	}
	s.spannerICCreateCmd.Flags().StringVar(&s.flagSpICConfigFile, "config-file", "", "YAML/JSON file for the InstanceConfig body (required)")
	_ = s.spannerICCreateCmd.MarkFlagRequired("config-file")
	s.spannerICUpdateCmd.Flags().StringVar(&s.flagSpICConfigFile, "config-file", "", "YAML/JSON file for the InstanceConfig body (required)")
	_ = s.spannerICUpdateCmd.MarkFlagRequired("config-file")
	s.spannerICUpdateCmd.Flags().StringVar(&s.flagSpICUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	s.spannerICListCmd.Flags().Int64Var(&s.flagSpICPageSize, "page-size", 0, "Maximum results per page")

	s.spannerInstanceConfigsCmd.AddCommand(all...)
	s.spannerCmd.AddCommand(s.spannerInstanceConfigsCmd)
}

func (s *spannerCmds) runSpICCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.spannerProject()
	if err != nil {
		return err
	}
	cfg := &spanner.InstanceConfig{}
	if err := loadYAMLOrJSONInto(s.flagSpICConfigFile, cfg); err != nil {
		return err
	}
	body := &spanner.CreateInstanceConfigRequest{
		InstanceConfig:   cfg,
		InstanceConfigId: args[0],
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.InstanceConfigs.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance config: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create instance config [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagSpICFormat)
}

func (s *spannerCmds) runSpICDelete(cmd *cobra.Command, args []string) error {
	name, err := s.spannerInstanceConfig(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.InstanceConfigs.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting instance config: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted instance config [%s].\n", args[0])
	return nil
}

func (s *spannerCmds) runSpICDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.spannerInstanceConfig(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.InstanceConfigs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSpICFormat)
}

func (s *spannerCmds) runSpICList(cmd *cobra.Command, args []string) error {
	parent, err := s.spannerProject()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	var all []*spanner.InstanceConfig
	pageToken := ""
	for {
		call := svc.Projects.InstanceConfigs.List(parent).Context(ctx)
		if s.flagSpICPageSize > 0 {
			call = call.PageSize(s.flagSpICPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing instance configs: %w", err)
		}
		all = append(all, resp.InstanceConfigs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagSpICFormat)
}

func (s *spannerCmds) runSpICUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.spannerInstanceConfig(args[0])
	if err != nil {
		return err
	}
	cfg := &spanner.InstanceConfig{}
	if err := loadYAMLOrJSONInto(s.flagSpICConfigFile, cfg); err != nil {
		return err
	}
	cfg.Name = name
	mask := s.flagSpICUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(cfg))
	}
	body := &spanner.UpdateInstanceConfigRequest{
		InstanceConfig: cfg,
		UpdateMask:     mask,
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.InstanceConfigs.Patch(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating instance config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagSpICFormat)
}
