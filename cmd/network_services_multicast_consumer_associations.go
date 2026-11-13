package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services multicast-consumer-associations (#993) ---

func (s *networkServicesCmds) initNetworkServicesMulticastConsumerAssociations() {
	all := []*cobra.Command{
		s.networkServicesMcaDeleteCmd, s.networkServicesMcaDescribeCmd,
		s.networkServicesMcaExportCmd, s.networkServicesMcaImportCmd,
		s.networkServicesMcaListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNsMcaLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNsMcaFormat, "format", "", "Output format")
	}
	nsBindExportFlags(s.networkServicesMcaExportCmd, &s.flagNsMcaDestination)
	nsBindImportFlags(s.networkServicesMcaImportCmd, &s.flagNsMcaSource)
	s.networkServicesMcaListCmd.Flags().Int64Var(&s.flagNsMcaPageSize, "page-size", 0, "Maximum results per page")

	s.networkServicesMulticastConsumerAssociationsCmd.AddCommand(all...)
	s.networkServicesCmd.AddCommand(s.networkServicesMulticastConsumerAssociationsCmd)
}

func (s *networkServicesCmds) nsMcaName(id string) (string, error) {
	return s.nsResourceName(s.flagNsMcaLocation, "multicastConsumerAssociations", id)
}

func (s *networkServicesCmds) runNsMcaDelete(cmd *cobra.Command, args []string) error {
	name, err := s.nsMcaName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.MulticastConsumerAssociations.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting multicast consumer association: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete multicast consumer association [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsMcaFormat)
}

func (s *networkServicesCmds) runNsMcaDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.nsMcaName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.MulticastConsumerAssociations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing multicast consumer association: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNsMcaFormat)
}

func (s *networkServicesCmds) runNsMcaExport(cmd *cobra.Command, args []string) error {
	name, err := s.nsMcaName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.MulticastConsumerAssociations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting multicast consumer association: %w", err)
	}
	return saveAsYAML(s.flagNsMcaDestination, got)
}

func (s *networkServicesCmds) runNsMcaImport(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsMcaLocation)
	if err != nil {
		return err
	}
	name, err := s.nsMcaName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.MulticastConsumerAssociation{}
	if err := loadYAMLOrJSONInto(s.flagNsMcaSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.MulticastConsumerAssociations.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.MulticastConsumerAssociations.Create(parent, body).MulticastConsumerAssociationId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating multicast consumer association: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Create multicast consumer association [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagNsMcaFormat)
	}
	op, err := svc.Projects.Locations.MulticastConsumerAssociations.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating multicast consumer association: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update multicast consumer association [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNsMcaFormat)
}

func (s *networkServicesCmds) runNsMcaList(cmd *cobra.Command, args []string) error {
	parent, err := s.nsLocationParent(s.flagNsMcaLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkServices(ctx)
	if err != nil {
		return err
	}
	var all []*networkservices.MulticastConsumerAssociation
	pageToken := ""
	for {
		call := svc.Projects.Locations.MulticastConsumerAssociations.List(parent).Context(ctx)
		if s.flagNsMcaPageSize > 0 {
			call = call.PageSize(s.flagNsMcaPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing multicast consumer associations: %w", err)
		}
		all = append(all, resp.MulticastConsumerAssociations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNsMcaFormat)
}
