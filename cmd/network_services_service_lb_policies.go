package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services service-lb-policies (#1003) ---

func (s *networkServicesCmds) initNetworkServicesServiceLbPolicies() {
	all := []*cobra.Command{
		s.networkServicesSlCreateCmd, s.networkServicesSlDeleteCmd,
		s.networkServicesSlDescribeCmd, s.networkServicesSlExportCmd,
		s.networkServicesSlImportCmd, s.networkServicesSlListCmd,
		s.networkServicesSlUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNsSlLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNsSlFormat, "format", "", "Output format")
	}
	s.networkServicesSlCreateCmd.Flags().StringVar(&s.flagNsSlConfigFile, "config-file", "", "YAML/JSON file with the ServiceLbPolicy body (required)")
	_ = s.networkServicesSlCreateCmd.MarkFlagRequired("config-file")
	nsBindExportFlags(s.networkServicesSlExportCmd, &s.flagNsSlDestination)
	nsBindImportFlags(s.networkServicesSlImportCmd, &s.flagNsSlSource)
	s.networkServicesSlListCmd.Flags().Int64Var(&s.flagNsSlPageSize, "page-size", 0, "Maximum results per page")
	s.networkServicesSlUpdateCmd.Flags().StringVar(&s.flagNsSlConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = s.networkServicesSlUpdateCmd.MarkFlagRequired("config-file")
	s.networkServicesSlUpdateCmd.Flags().StringVar(&s.flagNsSlUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	s.networkServicesServiceLbPoliciesCmd.AddCommand(all...)
	s.networkServicesCmd.AddCommand(s.networkServicesServiceLbPoliciesCmd)
}

func (s *networkServicesCmds) nsSlName(id string) (string, error) {
	return s.nsResourceName(s.flagNsSlLocation, "serviceLbPolicies", id)
}

func (s *networkServicesCmds) runNsSlCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsSlLocation)
	if err != nil {
		return err
	}
	body := &networkservices.ServiceLbPolicy{}
	if err := loadYAMLOrJSONInto(s.flagNsSlConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ServiceLbPolicies.Create(parent, body).ServiceLbPolicyId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating service LB policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create service LB policy [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsSlFormat)
}

func (s *networkServicesCmds) runNsSlDelete(cmd *cobra.Command, args []string) error {
	name, err := s.nsSlName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ServiceLbPolicies.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting service LB policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete service LB policy [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsSlFormat)
}

func (s *networkServicesCmds) runNsSlDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.nsSlName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ServiceLbPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service LB policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNsSlFormat)
}

func (s *networkServicesCmds) runNsSlExport(cmd *cobra.Command, args []string) error {
	name, err := s.nsSlName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ServiceLbPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting service LB policy: %w", err)
	}
	return saveAsYAML(s.flagNsSlDestination, got)
}

func (s *networkServicesCmds) runNsSlImport(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsSlLocation)
	if err != nil {
		return err
	}
	name, err := s.nsSlName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.ServiceLbPolicy{}
	if err := loadYAMLOrJSONInto(s.flagNsSlSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.ServiceLbPolicies.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.ServiceLbPolicies.Create(parent, body).ServiceLbPolicyId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating service LB policy: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Create service LB policy [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagNsSlFormat)
	}
	op, err := svc.Projects.Locations.ServiceLbPolicies.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating service LB policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update service LB policy [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsSlFormat)
}

func (s *networkServicesCmds) runNsSlList(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsSlLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	var all []*networkservices.ServiceLbPolicy
	pageToken := ""
	for {
		call := svc.Projects.Locations.ServiceLbPolicies.List(parent).Context(ctx)
		if s.flagNsSlPageSize > 0 {
			call = call.PageSize(s.flagNsSlPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing service LB policies: %w", err)
		}
		all = append(all, resp.ServiceLbPolicies...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNsSlFormat)
}

func (s *networkServicesCmds) runNsSlUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.nsSlName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.ServiceLbPolicy{}
	if err := loadYAMLOrJSONInto(s.flagNsSlConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagNsSlUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.ServiceLbPolicies.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating service LB policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update service LB policy [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsSlFormat)
}
