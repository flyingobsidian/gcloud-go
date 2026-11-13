package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	baremetalsolution "google.golang.org/api/baremetalsolution/v2"
)

// --- gcloud bms nfs-shares (#1228) ---

func (s *bmsCmds) initBmsNfsShares() {
	all := []*cobra.Command{s.bmsNfsDescribeCmd, s.bmsNfsListCmd, s.bmsNfsUpdateCmd, s.bmsNfsRenameCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagBmsNfsLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagBmsNfsFormat, "format", "", "Output format")
	}
	s.bmsNfsListCmd.Flags().Int64Var(&s.flagBmsNfsPageSize, "page-size", 0, "Maximum results per page")
	s.bmsNfsUpdateCmd.Flags().StringVar(&s.flagBmsNfsConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = s.bmsNfsUpdateCmd.MarkFlagRequired("config-file")
	s.bmsNfsUpdateCmd.Flags().StringVar(&s.flagBmsNfsUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	s.bmsNfsRenameCmd.Flags().StringVar(&s.flagBmsNfsNewName, "new-name", "", "New NFS share id (required)")
	_ = s.bmsNfsRenameCmd.MarkFlagRequired("new-name")

	s.bmsNfsSharesCmd.AddCommand(all...)
	s.bmsCmd.AddCommand(s.bmsNfsSharesCmd)
}

func (s *bmsCmds) bmsNfsName(id string) (string, error) {
	return s.bmsResource(s.flagBmsNfsLocation, "nfsShares", id)
}

func (s *bmsCmds) runBmsNfsDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.bmsNfsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.NfsShares.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing nfs share: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBmsNfsFormat)
}

func (s *bmsCmds) runBmsNfsList(cmd *cobra.Command, args []string) error {
	parent, err := s.bmsLocationParent(s.flagBmsNfsLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	var all []*baremetalsolution.NfsShare
	pageToken := ""
	for {
		call := svc.Projects.Locations.NfsShares.List(parent).Context(ctx)
		if s.flagBmsNfsPageSize > 0 {
			call = call.PageSize(s.flagBmsNfsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing nfs shares: %w", err)
		}
		all = append(all, resp.NfsShares...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagBmsNfsFormat)
}

func (s *bmsCmds) runBmsNfsUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.bmsNfsName(args[0])
	if err != nil {
		return err
	}
	body := &baremetalsolution.NfsShare{}
	if err := loadYAMLOrJSONInto(s.flagBmsNfsConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagBmsNfsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.NfsShares.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating nfs share: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update nfs share [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBmsNfsFormat)
}

func (s *bmsCmds) runBmsNfsRename(cmd *cobra.Command, args []string) error {
	name, err := s.bmsNfsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.NfsShares.Rename(name, &baremetalsolution.RenameNfsShareRequest{NewNfsshareId: s.flagBmsNfsNewName}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("renaming nfs share: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Renamed nfs share [%s] to [%s].\n", args[0], s.flagBmsNfsNewName)
	return emitFormattedTo(s.inv.stdout, got, s.flagBmsNfsFormat)
}
