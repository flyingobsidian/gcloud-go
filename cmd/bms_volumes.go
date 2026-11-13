package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	baremetalsolution "google.golang.org/api/baremetalsolution/v2"
)

// --- gcloud bms volumes (#1232) ---

func (s *bmsCmds) initBmsVolumes() {
	all := []*cobra.Command{s.bmsVolDescribeCmd, s.bmsVolListCmd, s.bmsVolUpdateCmd, s.bmsVolRenameCmd, s.bmsVolResizeCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagBmsVolLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagBmsVolFormat, "format", "", "Output format")
	}
	s.bmsVolListCmd.Flags().Int64Var(&s.flagBmsVolPageSize, "page-size", 0, "Maximum results per page")
	s.bmsVolUpdateCmd.Flags().StringVar(&s.flagBmsVolConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = s.bmsVolUpdateCmd.MarkFlagRequired("config-file")
	s.bmsVolUpdateCmd.Flags().StringVar(&s.flagBmsVolUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	s.bmsVolRenameCmd.Flags().StringVar(&s.flagBmsVolNewName, "new-name", "", "New volume id (required)")
	_ = s.bmsVolRenameCmd.MarkFlagRequired("new-name")
	s.bmsVolResizeCmd.Flags().Int64Var(&s.flagBmsVolSizeGib, "size-gib", 0, "New size in GiB (required)")
	_ = s.bmsVolResizeCmd.MarkFlagRequired("size-gib")

	s.bmsVolumesCmd.AddCommand(all...)
	s.bmsCmd.AddCommand(s.bmsVolumesCmd)
}

func (s *bmsCmds) bmsVolName(id string) (string, error) {
	return s.bmsResource(s.flagBmsVolLocation, "volumes", id)
}

func (s *bmsCmds) runBmsVolDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.bmsVolName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Volumes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing volume: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBmsVolFormat)
}

func (s *bmsCmds) runBmsVolList(cmd *cobra.Command, args []string) error {
	parent, err := s.bmsLocationParent(s.flagBmsVolLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	var all []*baremetalsolution.Volume
	pageToken := ""
	for {
		call := svc.Projects.Locations.Volumes.List(parent).Context(ctx)
		if s.flagBmsVolPageSize > 0 {
			call = call.PageSize(s.flagBmsVolPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing volumes: %w", err)
		}
		all = append(all, resp.Volumes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagBmsVolFormat)
}

func (s *bmsCmds) runBmsVolUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.bmsVolName(args[0])
	if err != nil {
		return err
	}
	body := &baremetalsolution.Volume{}
	if err := loadYAMLOrJSONInto(s.flagBmsVolConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagBmsVolUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Volumes.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating volume: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update volume [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBmsVolFormat)
}

func (s *bmsCmds) runBmsVolRename(cmd *cobra.Command, args []string) error {
	name, err := s.bmsVolName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Volumes.Rename(name, &baremetalsolution.RenameVolumeRequest{NewVolumeId: s.flagBmsVolNewName}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("renaming volume: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Renamed volume [%s] to [%s].\n", args[0], s.flagBmsVolNewName)
	return emitFormattedTo(s.inv.stdout, got, s.flagBmsVolFormat)
}

func (s *bmsCmds) runBmsVolResize(cmd *cobra.Command, args []string) error {
	name, err := s.bmsVolName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Resize(name, &baremetalsolution.ResizeVolumeRequest{SizeGib: s.flagBmsVolSizeGib}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resizing volume: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Resize volume [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBmsVolFormat)
}
