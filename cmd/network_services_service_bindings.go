package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services service-bindings (#1002) ---

func (s *networkServicesCmds) initNetworkServicesServiceBindings() {
	all := []*cobra.Command{
		s.networkServicesSbCreateCmd, s.networkServicesSbDeleteCmd,
		s.networkServicesSbDescribeCmd, s.networkServicesSbExportCmd,
		s.networkServicesSbImportCmd, s.networkServicesSbListCmd,
		s.networkServicesSbUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNsSbLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNsSbFormat, "format", "", "Output format")
	}
	s.networkServicesSbCreateCmd.Flags().StringVar(&s.flagNsSbConfigFile, "config-file", "", "YAML/JSON file with the ServiceBinding body (required)")
	_ = s.networkServicesSbCreateCmd.MarkFlagRequired("config-file")
	nsBindExportFlags(s.networkServicesSbExportCmd, &s.flagNsSbDestination)
	nsBindImportFlags(s.networkServicesSbImportCmd, &s.flagNsSbSource)
	s.networkServicesSbListCmd.Flags().Int64Var(&s.flagNsSbPageSize, "page-size", 0, "Maximum results per page")
	s.networkServicesSbUpdateCmd.Flags().StringVar(&s.flagNsSbConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = s.networkServicesSbUpdateCmd.MarkFlagRequired("config-file")
	s.networkServicesSbUpdateCmd.Flags().StringVar(&s.flagNsSbUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	s.networkServicesServiceBindingsCmd.AddCommand(all...)
	s.networkServicesCmd.AddCommand(s.networkServicesServiceBindingsCmd)
}

func (s *networkServicesCmds) nsSbName(id string) (string, error) {
	return s.nsResourceName(s.flagNsSbLocation, "serviceBindings", id)
}

func (s *networkServicesCmds) runNsSbCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsSbLocation)
	if err != nil {
		return err
	}
	body := &networkservices.ServiceBinding{}
	if err := loadYAMLOrJSONInto(s.flagNsSbConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ServiceBindings.Create(parent, body).ServiceBindingId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating service binding: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create service binding [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsSbFormat)
}

func (s *networkServicesCmds) runNsSbDelete(cmd *cobra.Command, args []string) error {
	name, err := s.nsSbName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ServiceBindings.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting service binding: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete service binding [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsSbFormat)
}

func (s *networkServicesCmds) runNsSbDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.nsSbName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ServiceBindings.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service binding: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNsSbFormat)
}

func (s *networkServicesCmds) runNsSbExport(cmd *cobra.Command, args []string) error {
	name, err := s.nsSbName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ServiceBindings.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting service binding: %w", err)
	}
	return saveAsYAML(s.flagNsSbDestination, got)
}

func (s *networkServicesCmds) runNsSbImport(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsSbLocation)
	if err != nil {
		return err
	}
	name, err := s.nsSbName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.ServiceBinding{}
	if err := loadYAMLOrJSONInto(s.flagNsSbSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.ServiceBindings.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.ServiceBindings.Create(parent, body).ServiceBindingId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating service binding: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Create service binding [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagNsSbFormat)
	}
	op, err := svc.Projects.Locations.ServiceBindings.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating service binding: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update service binding [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsSbFormat)
}

func (s *networkServicesCmds) runNsSbList(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsSbLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	var all []*networkservices.ServiceBinding
	pageToken := ""
	for {
		call := svc.Projects.Locations.ServiceBindings.List(parent).Context(ctx)
		if s.flagNsSbPageSize > 0 {
			call = call.PageSize(s.flagNsSbPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing service bindings: %w", err)
		}
		all = append(all, resp.ServiceBindings...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNsSbFormat)
}

func (s *networkServicesCmds) runNsSbUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.nsSbName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.ServiceBinding{}
	if err := loadYAMLOrJSONInto(s.flagNsSbConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagNsSbUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.ServiceBindings.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating service binding: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update service binding [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsSbFormat)
}
