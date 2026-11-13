package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	managedidentities "google.golang.org/api/managedidentities/v1"
)

// --- gcloud active-directory peerings (#1450) ---

func (s *activeDirectoryCmds) initActiveDirectoryPeerings() {
	all := []*cobra.Command{
		s.adPeeringCreateCmd, s.adPeeringDeleteCmd, s.adPeeringDescribeCmd,
		s.adPeeringListCmd, s.adPeeringUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagADPeeringFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.adPeeringCreateCmd, s.adPeeringUpdateCmd} {
		c.Flags().StringVar(&s.flagADPeeringConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Peering body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.adPeeringUpdateCmd.Flags().StringVar(&s.flagADPeeringUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.adPeeringListCmd.Flags().Int64Var(&s.flagADPeeringPageSize, "page-size", 0, "Maximum results per page")
	s.adPeeringListCmd.Flags().StringVar(&s.flagADPeeringFilter, "filter", "", "Server-side list filter")
	s.adPeeringListCmd.Flags().StringVar(&s.flagADPeeringOrderBy, "order-by", "", "Server-side ordering")

	s.adPeeringsCmd.AddCommand(all...)
	s.activeDirectoryCmd.AddCommand(s.adPeeringsCmd)
}

func adPeeringParent(project string) string {
	return fmt.Sprintf("projects/%s/locations/global", project)
}

func adPeeringResource(project, peering string) string {
	return fmt.Sprintf("projects/%s/locations/global/peerings/%s", project, peering)
}

func (s *activeDirectoryCmds) runADPeeringCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &managedidentities.Peering{}
	if err := loadYAMLOrJSONInto(s.flagADPeeringConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Peerings.
		Create(adPeeringParent(project), body).
		PeeringId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating peering: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for peering [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagADPeeringFormat)
}

func (s *activeDirectoryCmds) runADPeeringDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Peerings.
		Delete(adPeeringResource(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting peering: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for peering [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagADPeeringFormat)
}

func (s *activeDirectoryCmds) runADPeeringDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Global.Peerings.
		Get(adPeeringResource(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing peering: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagADPeeringFormat)
}

func (s *activeDirectoryCmds) runADPeeringList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	var all []*managedidentities.Peering
	pageToken := ""
	for {
		call := svc.Projects.Locations.Global.Peerings.List(adPeeringParent(project)).Context(ctx)
		if s.flagADPeeringPageSize > 0 {
			call = call.PageSize(s.flagADPeeringPageSize)
		}
		if s.flagADPeeringFilter != "" {
			call = call.Filter(s.flagADPeeringFilter)
		}
		if s.flagADPeeringOrderBy != "" {
			call = call.OrderBy(s.flagADPeeringOrderBy)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing peerings: %w", err)
		}
		all = append(all, resp.Peerings...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagADPeeringFormat)
}

func (s *activeDirectoryCmds) runADPeeringUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &managedidentities.Peering{}
	if err := loadYAMLOrJSONInto(s.flagADPeeringConfigFile, body); err != nil {
		return err
	}
	mask := s.flagADPeeringUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Global.Peerings.
		Patch(adPeeringResource(project, args[0]), body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating peering: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for peering [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagADPeeringFormat)
}
