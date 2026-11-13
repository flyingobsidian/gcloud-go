package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware datastores (#1115) ---

func (s *vmwareCmds) initVmwareDatastores() {
	all := []*cobra.Command{
		s.vmwareDsCreateCmd, s.vmwareDsDeleteCmd, s.vmwareDsDescribeCmd,
		s.vmwareDsListCmd, s.vmwareDsUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagVmwareDsLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagVmwareDsFormat, "format", "", "Output format")
	}
	s.vmwareDsCreateCmd.Flags().StringVar(&s.flagVmwareDsConfigFile, "config-file", "", "YAML/JSON file with datastore body (required)")
	_ = s.vmwareDsCreateCmd.MarkFlagRequired("config-file")
	s.vmwareDsUpdateCmd.Flags().StringVar(&s.flagVmwareDsConfigFile, "config-file", "", "YAML/JSON file with datastore body (required)")
	_ = s.vmwareDsUpdateCmd.MarkFlagRequired("config-file")
	s.vmwareDsUpdateCmd.Flags().StringVar(&s.flagVmwareDsUpdateMask, "update-mask", "", "Update mask (comma-separated field paths)")
	s.vmwareDsListCmd.Flags().Int64Var(&s.flagVmwareDsPageSize, "page-size", 0, "Maximum results per page")

	s.vmwareDatastoresCmd.AddCommand(all...)
	s.vmwareCmd.AddCommand(s.vmwareDatastoresCmd)
}

func (s *vmwareCmds) vmwareDsName(id string) (string, error) {
	return s.vmwareResource(s.flagVmwareDsLocation, "datastores", id)
}

func (s *vmwareCmds) runVmwareDsCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.vmwareLocationParent(s.flagVmwareDsLocation)
	if err != nil {
		return err
	}
	body := &vmwareengine.Datastore{}
	if err := loadYAMLOrJSONInto(s.flagVmwareDsConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Datastores.Create(parent, body).DatastoreId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating datastore: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create datastore [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwareDsFormat)
}

func (s *vmwareCmds) runVmwareDsDelete(cmd *cobra.Command, args []string) error {
	name, err := s.vmwareDsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Datastores.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting datastore: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete datastore [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwareDsFormat)
}

func (s *vmwareCmds) runVmwareDsDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.vmwareDsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datastores.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing datastore: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagVmwareDsFormat)
}

func (s *vmwareCmds) runVmwareDsList(cmd *cobra.Command, args []string) error {
	parent, err := s.vmwareLocationParent(s.flagVmwareDsLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	var all []*vmwareengine.Datastore
	pageToken := ""
	for {
		call := svc.Projects.Locations.Datastores.List(parent).Context(ctx)
		if s.flagVmwareDsPageSize > 0 {
			call = call.PageSize(s.flagVmwareDsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing datastores: %w", err)
		}
		all = append(all, resp.Datastores...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagVmwareDsFormat)
}

func (s *vmwareCmds) runVmwareDsUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.vmwareDsName(args[0])
	if err != nil {
		return err
	}
	body := &vmwareengine.Datastore{}
	if err := loadYAMLOrJSONInto(s.flagVmwareDsConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagVmwareDsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Datastores.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating datastore: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update datastore [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwareDsFormat)
}
