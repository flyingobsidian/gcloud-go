package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware network-policies (#1119) ---

func (s *vmwareCmds) initVmwareNetworkPolicies() {
	all := []*cobra.Command{
		s.vmwareNpolCreateCmd, s.vmwareNpolDeleteCmd, s.vmwareNpolDescribeCmd,
		s.vmwareNpolListCmd, s.vmwareNpolUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagVmwareNpolLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagVmwareNpolFormat, "format", "", "Output format")
	}
	s.vmwareNpolCreateCmd.Flags().StringVar(&s.flagVmwareNpolConfigFile, "config-file", "", "YAML/JSON file with network policy body (required)")
	_ = s.vmwareNpolCreateCmd.MarkFlagRequired("config-file")
	s.vmwareNpolUpdateCmd.Flags().StringVar(&s.flagVmwareNpolConfigFile, "config-file", "", "YAML/JSON file with network policy body (required)")
	_ = s.vmwareNpolUpdateCmd.MarkFlagRequired("config-file")
	s.vmwareNpolUpdateCmd.Flags().StringVar(&s.flagVmwareNpolUpdateMask, "update-mask", "", "Update mask (comma-separated field paths)")
	s.vmwareNpolListCmd.Flags().Int64Var(&s.flagVmwareNpolPageSize, "page-size", 0, "Maximum results per page")

	s.vmwareNetworkPoliciesCmd.AddCommand(all...)
	s.vmwareCmd.AddCommand(s.vmwareNetworkPoliciesCmd)
}

func (s *vmwareCmds) vmwareNpolName(id string) (string, error) {
	return s.vmwareResource(s.flagVmwareNpolLocation, "networkPolicies", id)
}

func (s *vmwareCmds) runVmwareNpolCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.vmwareLocationParent(s.flagVmwareNpolLocation)
	if err != nil {
		return err
	}
	body := &vmwareengine.NetworkPolicy{}
	if err := loadYAMLOrJSONInto(s.flagVmwareNpolConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NetworkPolicies.Create(parent, body).NetworkPolicyId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating network policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create network policy [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwareNpolFormat)
}

func (s *vmwareCmds) runVmwareNpolDelete(cmd *cobra.Command, args []string) error {
	name, err := s.vmwareNpolName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NetworkPolicies.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting network policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete network policy [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwareNpolFormat)
}

func (s *vmwareCmds) runVmwareNpolDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.vmwareNpolName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.NetworkPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing network policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagVmwareNpolFormat)
}

func (s *vmwareCmds) runVmwareNpolList(cmd *cobra.Command, args []string) error {
	parent, err := s.vmwareLocationParent(s.flagVmwareNpolLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	var all []*vmwareengine.NetworkPolicy
	pageToken := ""
	for {
		call := svc.Projects.Locations.NetworkPolicies.List(parent).Context(ctx)
		if s.flagVmwareNpolPageSize > 0 {
			call = call.PageSize(s.flagVmwareNpolPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing network policies: %w", err)
		}
		all = append(all, resp.NetworkPolicies...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagVmwareNpolFormat)
}

func (s *vmwareCmds) runVmwareNpolUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.vmwareNpolName(args[0])
	if err != nil {
		return err
	}
	body := &vmwareengine.NetworkPolicy{}
	if err := loadYAMLOrJSONInto(s.flagVmwareNpolConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagVmwareNpolUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NetworkPolicies.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating network policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update network policy [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwareNpolFormat)
}
