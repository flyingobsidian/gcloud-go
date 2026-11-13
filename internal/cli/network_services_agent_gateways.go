package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	networkservicesbeta "google.golang.org/api/networkservices/v1beta1"
)

// --- gcloud network-services agent-gateways (#1006) ---
//
// agent-gateways is only exposed on the v1beta1 surface as of this build,
// so this file uses gcp.NetworkServicesBetaService rather than the v1
// client used by its siblings.

func (s *networkServicesCmds) initNetworkServicesAgentGateways() {
	all := []*cobra.Command{
		s.networkServicesAgDeleteCmd, s.networkServicesAgDescribeCmd,
		s.networkServicesAgExportCmd, s.networkServicesAgImportCmd,
		s.networkServicesAgListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNsAgLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNsAgFormat, "format", "", "Output format")
	}
	nsBindExportFlags(s.networkServicesAgExportCmd, &s.flagNsAgDestination)
	nsBindImportFlags(s.networkServicesAgImportCmd, &s.flagNsAgSource)
	s.networkServicesAgListCmd.Flags().Int64Var(&s.flagNsAgPageSize, "page-size", 0, "Maximum results per page")

	s.networkServicesAgentGatewaysCmd.AddCommand(all...)
	s.networkServicesCmd.AddCommand(s.networkServicesAgentGatewaysCmd)
}

func (s *networkServicesCmds) nsAgName(id string) (string, error) {
	return s.nsResourceName(s.flagNsAgLocation, "agentGateways", id)
}

func (s *networkServicesCmds) runNsAgDelete(cmd *cobra.Command, args []string) error {
	name, err := s.nsAgName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServicesBeta(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AgentGateways.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting agent gateway: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete agent gateway [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsAgFormat)
}

func (s *networkServicesCmds) runNsAgDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.nsAgName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServicesBeta(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AgentGateways.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing agent gateway: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNsAgFormat)
}

func (s *networkServicesCmds) runNsAgExport(cmd *cobra.Command, args []string) error {
	name, err := s.nsAgName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServicesBeta(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AgentGateways.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting agent gateway: %w", err)
	}
	return saveAsYAML(s.flagNsAgDestination, got)
}

func (s *networkServicesCmds) runNsAgImport(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsAgLocation)
	if err != nil {
		return err
	}
	name, err := s.nsAgName(args[0])
	if err != nil {
		return err
	}
	body := &networkservicesbeta.AgentGateway{}
	if err := loadYAMLOrJSONInto(s.flagNsAgSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServicesBeta(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.AgentGateways.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.AgentGateways.Create(parent, body).AgentGatewayId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating agent gateway: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Create agent gateway [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagNsAgFormat)
	}
	op, err := svc.Projects.Locations.AgentGateways.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating agent gateway: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update agent gateway [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsAgFormat)
}

func (s *networkServicesCmds) runNsAgList(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsAgLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServicesBeta(ctx)
	if err != nil {
		return err
	}
	var all []*networkservicesbeta.AgentGateway
	pageToken := ""
	for {
		call := svc.Projects.Locations.AgentGateways.List(parent).Context(ctx)
		if s.flagNsAgPageSize > 0 {
			call = call.PageSize(s.flagNsAgPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing agent gateways: %w", err)
		}
		all = append(all, resp.AgentGateways...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNsAgFormat)
}
