package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	netapp "google.golang.org/api/netapp/v1"
)

// --- gcloud netapp active-directories (#1197) ---

func (s *netappCmds) initNetappActiveDirectories() {
	all := []*cobra.Command{s.netappADCreateCmd, s.netappADDeleteCmd, s.netappADDescribeCmd, s.netappADListCmd, s.netappADUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNetAppADLocation, "location", "", "Location for the Active Directory (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNetAppADFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.netappADCreateCmd, s.netappADUpdateCmd} {
		c.Flags().StringVar(&s.flagNetAppADConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the ActiveDirectory body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.netappADUpdateCmd.Flags().StringVar(&s.flagNetAppADUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.netappADListCmd.Flags().StringVar(&s.flagNetAppADFilter, "filter", "", "Server-side filter expression")
	s.netappADListCmd.Flags().Int64Var(&s.flagNetAppADPageSize, "page-size", 0, "Maximum number of results per page")

	s.netappADCmd.AddCommand(all...)
	s.netappCmd.AddCommand(s.netappADCmd)
}

func (s *netappCmds) netappADParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return netappLocationParent(project, s.flagNetAppADLocation), nil
}

func (s *netappCmds) netappADName(id string) (string, error) {
	parent, err := s.netappADParent()
	if err != nil {
		return "", err
	}
	return netappChild("activeDirectories", id, parent), nil
}

func (s *netappCmds) runNetAppADCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.netappADParent()
	if err != nil {
		return err
	}
	body := &netapp.ActiveDirectory{}
	if err := loadYAMLOrJSONInto(s.flagNetAppADConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ActiveDirectories.Create(parent, body).ActiveDirectoryId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating active directory: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for active directory [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppADFormat)
}

func (s *netappCmds) runNetAppADDelete(cmd *cobra.Command, args []string) error {
	name, err := s.netappADName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ActiveDirectories.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting active directory: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for active directory [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppADFormat)
}

func (s *netappCmds) runNetAppADDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.netappADName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ActiveDirectories.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing active directory: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNetAppADFormat)
}

func (s *netappCmds) runNetAppADList(cmd *cobra.Command, args []string) error {
	parent, err := s.netappADParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	var all []*netapp.ActiveDirectory
	pageToken := ""
	for {
		call := svc.Projects.Locations.ActiveDirectories.List(parent).Context(ctx)
		if s.flagNetAppADFilter != "" {
			call = call.Filter(s.flagNetAppADFilter)
		}
		if s.flagNetAppADPageSize > 0 {
			call = call.PageSize(s.flagNetAppADPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing active directories: %w", err)
		}
		all = append(all, resp.ActiveDirectories...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNetAppADFormat)
}

func (s *netappCmds) runNetAppADUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.netappADName(args[0])
	if err != nil {
		return err
	}
	body := &netapp.ActiveDirectory{}
	if err := loadYAMLOrJSONInto(s.flagNetAppADConfigFile, body); err != nil {
		return err
	}
	mask := s.flagNetAppADUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.ActiveDirectories.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating active directory: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for active directory [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppADFormat)
}
