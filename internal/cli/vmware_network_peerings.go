package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware network-peerings (#1118) ---

func (s *vmwareCmds) initVmwareNetworkPeerings() {
	all := []*cobra.Command{
		s.vmwareNpCreateCmd, s.vmwareNpDeleteCmd, s.vmwareNpDescribeCmd,
		s.vmwareNpListCmd, s.vmwareNpUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagVmwareNpLocation, "location", "", "Location (required, typically \"global\")")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagVmwareNpFormat, "format", "", "Output format")
	}
	s.vmwareNpCreateCmd.Flags().StringVar(&s.flagVmwareNpConfigFile, "config-file", "", "YAML/JSON file with network peering body (required)")
	_ = s.vmwareNpCreateCmd.MarkFlagRequired("config-file")
	s.vmwareNpUpdateCmd.Flags().StringVar(&s.flagVmwareNpConfigFile, "config-file", "", "YAML/JSON file with network peering body (required)")
	_ = s.vmwareNpUpdateCmd.MarkFlagRequired("config-file")
	s.vmwareNpUpdateCmd.Flags().StringVar(&s.flagVmwareNpUpdateMask, "update-mask", "", "Update mask (comma-separated field paths)")
	s.vmwareNpListCmd.Flags().Int64Var(&s.flagVmwareNpPageSize, "page-size", 0, "Maximum results per page")

	s.vmwareNetworkPeeringsCmd.AddCommand(all...)
	s.vmwareCmd.AddCommand(s.vmwareNetworkPeeringsCmd)
}

func (s *vmwareCmds) vmwareNpName(id string) (string, error) {
	return s.vmwareResource(s.flagVmwareNpLocation, "networkPeerings", id)
}

func (s *vmwareCmds) runVmwareNpCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.vmwareLocationParent(s.flagVmwareNpLocation)
	if err != nil {
		return err
	}
	body := &vmwareengine.NetworkPeering{}
	if err := loadYAMLOrJSONInto(s.flagVmwareNpConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NetworkPeerings.Create(parent, body).NetworkPeeringId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating network peering: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create network peering [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwareNpFormat)
}

func (s *vmwareCmds) runVmwareNpDelete(cmd *cobra.Command, args []string) error {
	name, err := s.vmwareNpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NetworkPeerings.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting network peering: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete network peering [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwareNpFormat)
}

func (s *vmwareCmds) runVmwareNpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.vmwareNpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.NetworkPeerings.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing network peering: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagVmwareNpFormat)
}

func (s *vmwareCmds) runVmwareNpList(cmd *cobra.Command, args []string) error {
	parent, err := s.vmwareLocationParent(s.flagVmwareNpLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	var all []*vmwareengine.NetworkPeering
	pageToken := ""
	for {
		call := svc.Projects.Locations.NetworkPeerings.List(parent).Context(ctx)
		if s.flagVmwareNpPageSize > 0 {
			call = call.PageSize(s.flagVmwareNpPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing network peerings: %w", err)
		}
		all = append(all, resp.NetworkPeerings...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagVmwareNpFormat)
}

func (s *vmwareCmds) runVmwareNpUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.vmwareNpName(args[0])
	if err != nil {
		return err
	}
	body := &vmwareengine.NetworkPeering{}
	if err := loadYAMLOrJSONInto(s.flagVmwareNpConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagVmwareNpUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NetworkPeerings.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating network peering: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update network peering [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwareNpFormat)
}
