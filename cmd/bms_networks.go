package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	baremetalsolution "google.golang.org/api/baremetalsolution/v2"
)

// --- gcloud bms networks (#1227) ---

func (s *bmsCmds) initBmsNetworks() {
	all := []*cobra.Command{s.bmsNetDescribeCmd, s.bmsNetListCmd, s.bmsNetUpdateCmd, s.bmsNetRenameCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagBmsNetLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagBmsNetFormat, "format", "", "Output format")
	}
	s.bmsNetListCmd.Flags().Int64Var(&s.flagBmsNetPageSize, "page-size", 0, "Maximum results per page")
	s.bmsNetUpdateCmd.Flags().StringVar(&s.flagBmsNetConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = s.bmsNetUpdateCmd.MarkFlagRequired("config-file")
	s.bmsNetUpdateCmd.Flags().StringVar(&s.flagBmsNetUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	s.bmsNetRenameCmd.Flags().StringVar(&s.flagBmsNetNewName, "new-name", "", "New network id (required)")
	_ = s.bmsNetRenameCmd.MarkFlagRequired("new-name")

	s.bmsNetworksCmd.AddCommand(all...)
	s.bmsCmd.AddCommand(s.bmsNetworksCmd)
}

func (s *bmsCmds) bmsNetName(id string) (string, error) {
	return s.bmsResource(s.flagBmsNetLocation, "networks", id)
}

func (s *bmsCmds) runBmsNetDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.bmsNetName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Networks.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing network: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBmsNetFormat)
}

func (s *bmsCmds) runBmsNetList(cmd *cobra.Command, args []string) error {
	parent, err := s.bmsLocationParent(s.flagBmsNetLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	var all []*baremetalsolution.Network
	pageToken := ""
	for {
		call := svc.Projects.Locations.Networks.List(parent).Context(ctx)
		if s.flagBmsNetPageSize > 0 {
			call = call.PageSize(s.flagBmsNetPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing networks: %w", err)
		}
		all = append(all, resp.Networks...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagBmsNetFormat)
}

func (s *bmsCmds) runBmsNetUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.bmsNetName(args[0])
	if err != nil {
		return err
	}
	body := &baremetalsolution.Network{}
	if err := loadYAMLOrJSONInto(s.flagBmsNetConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagBmsNetUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Networks.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating network: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update network [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBmsNetFormat)
}

func (s *bmsCmds) runBmsNetRename(cmd *cobra.Command, args []string) error {
	name, err := s.bmsNetName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Networks.Rename(name, &baremetalsolution.RenameNetworkRequest{NewNetworkId: s.flagBmsNetNewName}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("renaming network: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Renamed network [%s] to [%s].\n", args[0], s.flagBmsNetNewName)
	return emitFormattedTo(s.inv.stdout, got, s.flagBmsNetFormat)
}
