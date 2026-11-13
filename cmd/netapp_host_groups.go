package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	netapp "google.golang.org/api/netapp/v1"
)

// --- gcloud netapp host-groups (#1200) ---

func (s *netappCmds) initNetappHostGroups() {
	all := []*cobra.Command{s.netappHGCreateCmd, s.netappHGDeleteCmd, s.netappHGDescribeCmd, s.netappHGListCmd, s.netappHGUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNetAppHGLocation, "location", "", "Location for the host group (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNetAppHGFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.netappHGCreateCmd, s.netappHGUpdateCmd} {
		c.Flags().StringVar(&s.flagNetAppHGConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the HostGroup body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.netappHGUpdateCmd.Flags().StringVar(&s.flagNetAppHGUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.netappHGListCmd.Flags().StringVar(&s.flagNetAppHGFilter, "filter", "", "Server-side filter expression")
	s.netappHGListCmd.Flags().Int64Var(&s.flagNetAppHGPageSize, "page-size", 0, "Maximum number of results per page")

	s.netappHGCmd.AddCommand(all...)
	s.netappCmd.AddCommand(s.netappHGCmd)
}

func (s *netappCmds) netappHGParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return netappLocationParent(project, s.flagNetAppHGLocation), nil
}

func (s *netappCmds) netappHGName(id string) (string, error) {
	parent, err := s.netappHGParent()
	if err != nil {
		return "", err
	}
	return netappChild("hostGroups", id, parent), nil
}

func (s *netappCmds) runNetAppHGCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.netappHGParent()
	if err != nil {
		return err
	}
	body := &netapp.HostGroup{}
	if err := loadYAMLOrJSONInto(s.flagNetAppHGConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.HostGroups.Create(parent, body).HostGroupId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating host group: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for host group [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppHGFormat)
}

func (s *netappCmds) runNetAppHGDelete(cmd *cobra.Command, args []string) error {
	name, err := s.netappHGName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.HostGroups.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting host group: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for host group [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppHGFormat)
}

func (s *netappCmds) runNetAppHGDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.netappHGName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.HostGroups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing host group: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNetAppHGFormat)
}

func (s *netappCmds) runNetAppHGList(cmd *cobra.Command, args []string) error {
	parent, err := s.netappHGParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	var all []*netapp.HostGroup
	pageToken := ""
	for {
		call := svc.Projects.Locations.HostGroups.List(parent).Context(ctx)
		if s.flagNetAppHGFilter != "" {
			call = call.Filter(s.flagNetAppHGFilter)
		}
		if s.flagNetAppHGPageSize > 0 {
			call = call.PageSize(s.flagNetAppHGPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing host groups: %w", err)
		}
		all = append(all, resp.HostGroups...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNetAppHGFormat)
}

func (s *netappCmds) runNetAppHGUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.netappHGName(args[0])
	if err != nil {
		return err
	}
	body := &netapp.HostGroup{}
	if err := loadYAMLOrJSONInto(s.flagNetAppHGConfigFile, body); err != nil {
		return err
	}
	mask := s.flagNetAppHGUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.HostGroups.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating host group: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for host group [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppHGFormat)
}
