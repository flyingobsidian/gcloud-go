package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware private-clouds (#1123) ---

func (s *vmwareCmds) initVmwarePrivateClouds() {
	all := []*cobra.Command{
		s.vmwarePcCreateCmd, s.vmwarePcDeleteCmd, s.vmwarePcDescribeCmd,
		s.vmwarePcListCmd, s.vmwarePcUpdateCmd, s.vmwarePcMigrateMgmtVmsCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagVmwarePcLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagVmwarePcFormat, "format", "", "Output format")
	}
	s.vmwarePcCreateCmd.Flags().StringVar(&s.flagVmwarePcConfigFile, "config-file", "", "YAML/JSON file with private cloud body (required)")
	_ = s.vmwarePcCreateCmd.MarkFlagRequired("config-file")
	s.vmwarePcUpdateCmd.Flags().StringVar(&s.flagVmwarePcConfigFile, "config-file", "", "YAML/JSON file with private cloud body (required)")
	_ = s.vmwarePcUpdateCmd.MarkFlagRequired("config-file")
	s.vmwarePcUpdateCmd.Flags().StringVar(&s.flagVmwarePcUpdateMask, "update-mask", "", "Update mask (comma-separated field paths)")
	s.vmwarePcListCmd.Flags().Int64Var(&s.flagVmwarePcPageSize, "page-size", 0, "Maximum results per page")
	s.vmwarePcMigrateMgmtVmsCmd.Flags().StringVar(&s.flagVmwarePcClusterID, "cluster-id", "",
		"ID of the destination workload cluster (must already exist in the private cloud) (required)")
	_ = s.vmwarePcMigrateMgmtVmsCmd.MarkFlagRequired("cluster-id")
	s.vmwarePcMigrateMgmtVmsCmd.Flags().StringVar(&s.flagVmwarePcRequestID, "request-id", "",
		"Optional idempotency UUID")
	s.vmwarePcMigrateMgmtVmsCmd.Flags().StringVar(&s.flagVmwarePcEtag, "etag", "",
		"Optional resource etag for optimistic concurrency")

	s.vmwarePrivateCloudsCmd.AddCommand(all...)
	s.vmwareCmd.AddCommand(s.vmwarePrivateCloudsCmd)
}

func (s *vmwareCmds) vmwarePcName(id string) (string, error) {
	return s.vmwareResource(s.flagVmwarePcLocation, "privateClouds", id)
}

func (s *vmwareCmds) runVmwarePcCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.vmwareLocationParent(s.flagVmwarePcLocation)
	if err != nil {
		return err
	}
	body := &vmwareengine.PrivateCloud{}
	if err := loadYAMLOrJSONInto(s.flagVmwarePcConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateClouds.Create(parent, body).PrivateCloudId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating private cloud: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create private cloud [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwarePcFormat)
}

func (s *vmwareCmds) runVmwarePcDelete(cmd *cobra.Command, args []string) error {
	name, err := s.vmwarePcName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateClouds.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting private cloud: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete private cloud [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwarePcFormat)
}

func (s *vmwareCmds) runVmwarePcDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.vmwarePcName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.PrivateClouds.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing private cloud: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagVmwarePcFormat)
}

func (s *vmwareCmds) runVmwarePcList(cmd *cobra.Command, args []string) error {
	parent, err := s.vmwareLocationParent(s.flagVmwarePcLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	var all []*vmwareengine.PrivateCloud
	pageToken := ""
	for {
		call := svc.Projects.Locations.PrivateClouds.List(parent).Context(ctx)
		if s.flagVmwarePcPageSize > 0 {
			call = call.PageSize(s.flagVmwarePcPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing private clouds: %w", err)
		}
		all = append(all, resp.PrivateClouds...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagVmwarePcFormat)
}

func (s *vmwareCmds) runVmwarePcMigrateMgmtVms(cmd *cobra.Command, args []string) error {
	name, err := s.vmwarePcName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	req := &vmwareengine.MigrateManagementVmsRequest{
		ClusterId: s.flagVmwarePcClusterID,
		RequestId: s.flagVmwarePcRequestID,
		Etag:      s.flagVmwarePcEtag,
	}
	op, err := svc.Projects.Locations.PrivateClouds.MigrateManagementVms(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("migrating management VMs: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Migrate management VMs on private cloud [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwarePcFormat)
}

func (s *vmwareCmds) runVmwarePcUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.vmwarePcName(args[0])
	if err != nil {
		return err
	}
	body := &vmwareengine.PrivateCloud{}
	if err := loadYAMLOrJSONInto(s.flagVmwarePcConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagVmwarePcUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateClouds.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating private cloud: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update private cloud [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwarePcFormat)
}
