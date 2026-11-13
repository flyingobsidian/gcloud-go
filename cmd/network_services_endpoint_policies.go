package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services endpoint-policies (#988) ---

func (s *networkServicesCmds) initNetworkServicesEndpointPolicies() {
	all := []*cobra.Command{
		s.networkServicesEpDeleteCmd, s.networkServicesEpDescribeCmd,
		s.networkServicesEpExportCmd, s.networkServicesEpImportCmd,
		s.networkServicesEpListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNsEpLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNsEpFormat, "format", "", "Output format")
	}
	nsBindExportFlags(s.networkServicesEpExportCmd, &s.flagNsEpDestination)
	nsBindImportFlags(s.networkServicesEpImportCmd, &s.flagNsEpSource)
	s.networkServicesEpListCmd.Flags().Int64Var(&s.flagNsEpPageSize, "page-size", 0, "Maximum results per page")

	s.networkServicesEndpointPoliciesCmd.AddCommand(all...)
	s.networkServicesCmd.AddCommand(s.networkServicesEndpointPoliciesCmd)
}

func (s *networkServicesCmds) nsEpName(id string) (string, error) {
	return s.nsResourceName(s.flagNsEpLocation, "endpointPolicies", id)
}

func (s *networkServicesCmds) runNsEpDelete(cmd *cobra.Command, args []string) error {
	name, err := s.nsEpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.EndpointPolicies.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting endpoint policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete endpoint policy [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsEpFormat)
}

func (s *networkServicesCmds) runNsEpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.nsEpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.EndpointPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing endpoint policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNsEpFormat)
}

func (s *networkServicesCmds) runNsEpExport(cmd *cobra.Command, args []string) error {
	name, err := s.nsEpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.EndpointPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting endpoint policy: %w", err)
	}
	return saveAsYAML(s.flagNsEpDestination, got)
}

func (s *networkServicesCmds) runNsEpImport(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsEpLocation)
	if err != nil {
		return err
	}
	name, err := s.nsEpName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.EndpointPolicy{}
	if err := loadYAMLOrJSONInto(s.flagNsEpSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.EndpointPolicies.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.EndpointPolicies.Create(parent, body).EndpointPolicyId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating endpoint policy: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Create endpoint policy [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagNsEpFormat)
	}
	op, err := svc.Projects.Locations.EndpointPolicies.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating endpoint policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update endpoint policy [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsEpFormat)
}

func (s *networkServicesCmds) runNsEpList(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsEpLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	var all []*networkservices.EndpointPolicy
	pageToken := ""
	for {
		call := svc.Projects.Locations.EndpointPolicies.List(parent).Context(ctx)
		if s.flagNsEpPageSize > 0 {
			call = call.PageSize(s.flagNsEpPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing endpoint policies: %w", err)
		}
		all = append(all, resp.EndpointPolicies...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNsEpFormat)
}
