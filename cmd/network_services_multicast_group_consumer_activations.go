package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services multicast-group-consumer-activations (#995) ---

func (s *networkServicesCmds) initNetworkServicesMulticastGroupConsumerActivations() {
	all := []*cobra.Command{
		s.networkServicesMgcaDeleteCmd, s.networkServicesMgcaDescribeCmd,
		s.networkServicesMgcaExportCmd, s.networkServicesMgcaImportCmd,
		s.networkServicesMgcaListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNsMgcaLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNsMgcaFormat, "format", "", "Output format")
	}
	nsBindExportFlags(s.networkServicesMgcaExportCmd, &s.flagNsMgcaDestination)
	nsBindImportFlags(s.networkServicesMgcaImportCmd, &s.flagNsMgcaSource)
	s.networkServicesMgcaListCmd.Flags().Int64Var(&s.flagNsMgcaPageSize, "page-size", 0, "Maximum results per page")

	s.networkServicesMulticastGroupConsumerActivationsCmd.AddCommand(all...)
	s.networkServicesCmd.AddCommand(s.networkServicesMulticastGroupConsumerActivationsCmd)
}

func (s *networkServicesCmds) nsMgcaName(id string) (string, error) {
	return s.nsResourceName(s.flagNsMgcaLocation, "multicastGroupConsumerActivations", id)
}

func (s *networkServicesCmds) runNsMgcaDelete(cmd *cobra.Command, args []string) error {
	name, err := s.nsMgcaName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.MulticastGroupConsumerActivations.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting multicast group consumer activation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete multicast group consumer activation [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsMgcaFormat)
}

func (s *networkServicesCmds) runNsMgcaDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.nsMgcaName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.MulticastGroupConsumerActivations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing multicast group consumer activation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNsMgcaFormat)
}

func (s *networkServicesCmds) runNsMgcaExport(cmd *cobra.Command, args []string) error {
	name, err := s.nsMgcaName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.MulticastGroupConsumerActivations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting multicast group consumer activation: %w", err)
	}
	return saveAsYAML(s.flagNsMgcaDestination, got)
}

func (s *networkServicesCmds) runNsMgcaImport(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsMgcaLocation)
	if err != nil {
		return err
	}
	name, err := s.nsMgcaName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.MulticastGroupConsumerActivation{}
	if err := loadYAMLOrJSONInto(s.flagNsMgcaSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.MulticastGroupConsumerActivations.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.MulticastGroupConsumerActivations.Create(parent, body).MulticastGroupConsumerActivationId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating multicast group consumer activation: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Create multicast group consumer activation [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagNsMgcaFormat)
	}
	op, err := svc.Projects.Locations.MulticastGroupConsumerActivations.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating multicast group consumer activation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update multicast group consumer activation [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsMgcaFormat)
}

func (s *networkServicesCmds) runNsMgcaList(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsMgcaLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	var all []*networkservices.MulticastGroupConsumerActivation
	pageToken := ""
	for {
		call := svc.Projects.Locations.MulticastGroupConsumerActivations.List(parent).Context(ctx)
		if s.flagNsMgcaPageSize > 0 {
			call = call.PageSize(s.flagNsMgcaPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing multicast group consumer activations: %w", err)
		}
		all = append(all, resp.MulticastGroupConsumerActivations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNsMgcaFormat)
}
