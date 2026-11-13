package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	developerconnect "google.golang.org/api/developerconnect/v1"
)

// --- gcloud developer-connect insights-configs (#1024) ---

func (s *developerConnectCmds) initDeveloperConnectInsightsConfigs() {
	all := []*cobra.Command{
		s.developerConnectIcCreateCmd, s.developerConnectIcDeleteCmd,
		s.developerConnectIcDescribeCmd, s.developerConnectIcListCmd,
		s.developerConnectIcUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDcIcLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagDcIcFormat, "format", "", "Output format")
	}
	s.developerConnectIcCreateCmd.Flags().StringVar(&s.flagDcIcConfigFile, "config-file", "", "YAML/JSON file with the InsightsConfig body (required)")
	_ = s.developerConnectIcCreateCmd.MarkFlagRequired("config-file")
	s.developerConnectIcListCmd.Flags().Int64Var(&s.flagDcIcPageSize, "page-size", 0, "Maximum results per page")
	s.developerConnectIcUpdateCmd.Flags().StringVar(&s.flagDcIcConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = s.developerConnectIcUpdateCmd.MarkFlagRequired("config-file")

	s.developerConnectInsightsConfigsCmd.AddCommand(all...)
	s.developerConnectCmd.AddCommand(s.developerConnectInsightsConfigsCmd)
}

func (s *developerConnectCmds) dcIcName(id string) (string, error) {
	return s.devConnResourceName(s.flagDcIcLocation, "insightsConfigs", id)
}

func (s *developerConnectCmds) runDcIcCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.devConnLocationParent(s.flagDcIcLocation)
	if err != nil {
		return err
	}
	body := &developerconnect.InsightsConfig{}
	if err := loadYAMLOrJSONInto(s.flagDcIcConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeveloperConnect(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.InsightsConfigs.Create(parent, body).InsightsConfigId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating insights config: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create insights config [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDcIcFormat)
}

func (s *developerConnectCmds) runDcIcDelete(cmd *cobra.Command, args []string) error {
	name, err := s.dcIcName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeveloperConnect(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.InsightsConfigs.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting insights config: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete insights config [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDcIcFormat)
}

func (s *developerConnectCmds) runDcIcDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.dcIcName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeveloperConnect(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.InsightsConfigs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing insights config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDcIcFormat)
}

func (s *developerConnectCmds) runDcIcList(cmd *cobra.Command, args []string) error {
	parent, err := s.devConnLocationParent(s.flagDcIcLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeveloperConnect(ctx)
	if err != nil {
		return err
	}
	var all []*developerconnect.InsightsConfig
	pageToken := ""
	for {
		call := svc.Projects.Locations.InsightsConfigs.List(parent).Context(ctx)
		if s.flagDcIcPageSize > 0 {
			call = call.PageSize(s.flagDcIcPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing insights configs: %w", err)
		}
		all = append(all, resp.InsightsConfigs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagDcIcFormat)
}

func (s *developerConnectCmds) runDcIcUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.dcIcName(args[0])
	if err != nil {
		return err
	}
	body := &developerconnect.InsightsConfig{}
	if err := loadYAMLOrJSONInto(s.flagDcIcConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().DeveloperConnect(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.InsightsConfigs.Patch(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating insights config: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update insights config [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDcIcFormat)
}
