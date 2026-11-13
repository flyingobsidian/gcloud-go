package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware networks (#1120) ---

func (s *vmwareCmds) initVmwareNetworks() {
	all := []*cobra.Command{
		s.vmwareNetCreateCmd, s.vmwareNetDeleteCmd, s.vmwareNetDescribeCmd,
		s.vmwareNetListCmd, s.vmwareNetUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagVmwareNetLocation, "location", "", "Location (required, typically \"global\")")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagVmwareNetFormat, "format", "", "Output format")
	}
	s.vmwareNetCreateCmd.Flags().StringVar(&s.flagVmwareNetConfigFile, "config-file", "", "YAML/JSON file with VMware Engine network body (required)")
	_ = s.vmwareNetCreateCmd.MarkFlagRequired("config-file")
	s.vmwareNetUpdateCmd.Flags().StringVar(&s.flagVmwareNetConfigFile, "config-file", "", "YAML/JSON file with VMware Engine network body (required)")
	_ = s.vmwareNetUpdateCmd.MarkFlagRequired("config-file")
	s.vmwareNetUpdateCmd.Flags().StringVar(&s.flagVmwareNetUpdateMask, "update-mask", "", "Update mask (comma-separated field paths)")
	s.vmwareNetListCmd.Flags().Int64Var(&s.flagVmwareNetPageSize, "page-size", 0, "Maximum results per page")

	s.vmwareNetworksCmd.AddCommand(all...)
	s.vmwareCmd.AddCommand(s.vmwareNetworksCmd)
}

func (s *vmwareCmds) vmwareNetName(id string) (string, error) {
	return s.vmwareResource(s.flagVmwareNetLocation, "vmwareEngineNetworks", id)
}

func (s *vmwareCmds) runVmwareNetCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.vmwareLocationParent(s.flagVmwareNetLocation)
	if err != nil {
		return err
	}
	body := &vmwareengine.VmwareEngineNetwork{}
	if err := loadYAMLOrJSONInto(s.flagVmwareNetConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.VmwareEngineNetworks.Create(parent, body).VmwareEngineNetworkId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating vmware engine network: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create vmware engine network [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwareNetFormat)
}

func (s *vmwareCmds) runVmwareNetDelete(cmd *cobra.Command, args []string) error {
	name, err := s.vmwareNetName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.VmwareEngineNetworks.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting vmware engine network: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete vmware engine network [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwareNetFormat)
}

func (s *vmwareCmds) runVmwareNetDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.vmwareNetName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.VmwareEngineNetworks.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing vmware engine network: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagVmwareNetFormat)
}

func (s *vmwareCmds) runVmwareNetList(cmd *cobra.Command, args []string) error {
	parent, err := s.vmwareLocationParent(s.flagVmwareNetLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	var all []*vmwareengine.VmwareEngineNetwork
	pageToken := ""
	for {
		call := svc.Projects.Locations.VmwareEngineNetworks.List(parent).Context(ctx)
		if s.flagVmwareNetPageSize > 0 {
			call = call.PageSize(s.flagVmwareNetPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing vmware engine networks: %w", err)
		}
		all = append(all, resp.VmwareEngineNetworks...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagVmwareNetFormat)
}

func (s *vmwareCmds) runVmwareNetUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.vmwareNetName(args[0])
	if err != nil {
		return err
	}
	body := &vmwareengine.VmwareEngineNetwork{}
	if err := loadYAMLOrJSONInto(s.flagVmwareNetConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagVmwareNetUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.VmwareEngineNetworks.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating vmware engine network: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update vmware engine network [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwareNetFormat)
}
