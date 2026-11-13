package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware private-connections (#1124) ---

func (s *vmwareCmds) initVmwarePrivateConnections() {
	all := []*cobra.Command{
		s.vmwarePconCreateCmd, s.vmwarePconDeleteCmd, s.vmwarePconDescribeCmd,
		s.vmwarePconListCmd, s.vmwarePconUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagVmwarePconLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagVmwarePconFormat, "format", "", "Output format")
	}
	s.vmwarePconCreateCmd.Flags().StringVar(&s.flagVmwarePconConfigFile, "config-file", "", "YAML/JSON file with private connection body (required)")
	_ = s.vmwarePconCreateCmd.MarkFlagRequired("config-file")
	s.vmwarePconUpdateCmd.Flags().StringVar(&s.flagVmwarePconConfigFile, "config-file", "", "YAML/JSON file with private connection body (required)")
	_ = s.vmwarePconUpdateCmd.MarkFlagRequired("config-file")
	s.vmwarePconUpdateCmd.Flags().StringVar(&s.flagVmwarePconUpdateMask, "update-mask", "", "Update mask (comma-separated field paths)")
	s.vmwarePconListCmd.Flags().Int64Var(&s.flagVmwarePconPageSize, "page-size", 0, "Maximum results per page")

	s.vmwarePrivateConnectionsCmd.AddCommand(all...)
	s.vmwareCmd.AddCommand(s.vmwarePrivateConnectionsCmd)
}

func (s *vmwareCmds) vmwarePconName(id string) (string, error) {
	return s.vmwareResource(s.flagVmwarePconLocation, "privateConnections", id)
}

func (s *vmwareCmds) runVmwarePconCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.vmwareLocationParent(s.flagVmwarePconLocation)
	if err != nil {
		return err
	}
	body := &vmwareengine.PrivateConnection{}
	if err := loadYAMLOrJSONInto(s.flagVmwarePconConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateConnections.Create(parent, body).PrivateConnectionId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating private connection: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create private connection [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwarePconFormat)
}

func (s *vmwareCmds) runVmwarePconDelete(cmd *cobra.Command, args []string) error {
	name, err := s.vmwarePconName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateConnections.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting private connection: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete private connection [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwarePconFormat)
}

func (s *vmwareCmds) runVmwarePconDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.vmwarePconName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.PrivateConnections.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing private connection: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagVmwarePconFormat)
}

func (s *vmwareCmds) runVmwarePconList(cmd *cobra.Command, args []string) error {
	parent, err := s.vmwareLocationParent(s.flagVmwarePconLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	var all []*vmwareengine.PrivateConnection
	pageToken := ""
	for {
		call := svc.Projects.Locations.PrivateConnections.List(parent).Context(ctx)
		if s.flagVmwarePconPageSize > 0 {
			call = call.PageSize(s.flagVmwarePconPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing private connections: %w", err)
		}
		all = append(all, resp.PrivateConnections...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagVmwarePconFormat)
}

func (s *vmwareCmds) runVmwarePconUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.vmwarePconName(args[0])
	if err != nil {
		return err
	}
	body := &vmwareengine.PrivateConnection{}
	if err := loadYAMLOrJSONInto(s.flagVmwarePconConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagVmwarePconUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateConnections.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating private connection: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update private connection [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwarePconFormat)
}
