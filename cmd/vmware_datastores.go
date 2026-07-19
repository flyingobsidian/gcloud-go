package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware datastores (#1115) ---

var vmwareDatastoresCmd = &cobra.Command{Use: "datastores", Short: "Manage VMware Engine datastores"}

var (
	flagVmwareDsLocation   string
	flagVmwareDsFormat     string
	flagVmwareDsConfigFile string
	flagVmwareDsUpdateMask string
	flagVmwareDsPageSize   int64
)

var (
	vmwareDsCreateCmd = &cobra.Command{
		Use: "create DATASTORE", Short: "Create a VMware Engine datastore",
		Args: cobra.ExactArgs(1), RunE: runVmwareDsCreate,
	}
	vmwareDsDeleteCmd = &cobra.Command{
		Use: "delete DATASTORE", Short: "Delete a VMware Engine datastore",
		Args: cobra.ExactArgs(1), RunE: runVmwareDsDelete,
	}
	vmwareDsDescribeCmd = &cobra.Command{
		Use: "describe DATASTORE", Short: "Describe a VMware Engine datastore",
		Args: cobra.ExactArgs(1), RunE: runVmwareDsDescribe,
	}
	vmwareDsListCmd = &cobra.Command{
		Use: "list", Short: "List VMware Engine datastores in a location",
		Args: cobra.NoArgs, RunE: runVmwareDsList,
	}
	vmwareDsUpdateCmd = &cobra.Command{
		Use: "update DATASTORE", Short: "Update a VMware Engine datastore",
		Args: cobra.ExactArgs(1), RunE: runVmwareDsUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		vmwareDsCreateCmd, vmwareDsDeleteCmd, vmwareDsDescribeCmd,
		vmwareDsListCmd, vmwareDsUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagVmwareDsLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagVmwareDsFormat, "format", "", "Output format")
	}
	vmwareDsCreateCmd.Flags().StringVar(&flagVmwareDsConfigFile, "config-file", "", "YAML/JSON file with datastore body (required)")
	_ = vmwareDsCreateCmd.MarkFlagRequired("config-file")
	vmwareDsUpdateCmd.Flags().StringVar(&flagVmwareDsConfigFile, "config-file", "", "YAML/JSON file with datastore body (required)")
	_ = vmwareDsUpdateCmd.MarkFlagRequired("config-file")
	vmwareDsUpdateCmd.Flags().StringVar(&flagVmwareDsUpdateMask, "update-mask", "", "Update mask (comma-separated field paths)")
	vmwareDsListCmd.Flags().Int64Var(&flagVmwareDsPageSize, "page-size", 0, "Maximum results per page")

	vmwareDatastoresCmd.AddCommand(all...)
	vmwareCmd.AddCommand(vmwareDatastoresCmd)
}

func vmwareDsName(id string) (string, error) {
	return vmwareResource(flagVmwareDsLocation, "datastores", id)
}

func runVmwareDsCreate(cmd *cobra.Command, args []string) error {
	parent, err := vmwareLocationParent(flagVmwareDsLocation)
	if err != nil {
		return err
	}
	body := &vmwareengine.Datastore{}
	if err := loadYAMLOrJSONInto(flagVmwareDsConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Datastores.Create(parent, body).DatastoreId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating datastore: %w", err)
	}
	fmt.Printf("Create datastore [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagVmwareDsFormat)
}

func runVmwareDsDelete(cmd *cobra.Command, args []string) error {
	name, err := vmwareDsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Datastores.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting datastore: %w", err)
	}
	fmt.Printf("Delete datastore [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagVmwareDsFormat)
}

func runVmwareDsDescribe(cmd *cobra.Command, args []string) error {
	name, err := vmwareDsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datastores.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing datastore: %w", err)
	}
	return emitFormatted(got, flagVmwareDsFormat)
}

func runVmwareDsList(cmd *cobra.Command, args []string) error {
	parent, err := vmwareLocationParent(flagVmwareDsLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*vmwareengine.Datastore
	pageToken := ""
	for {
		call := svc.Projects.Locations.Datastores.List(parent).Context(ctx)
		if flagVmwareDsPageSize > 0 {
			call = call.PageSize(flagVmwareDsPageSize)
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
	return emitFormatted(all, flagVmwareDsFormat)
}

func runVmwareDsUpdate(cmd *cobra.Command, args []string) error {
	name, err := vmwareDsName(args[0])
	if err != nil {
		return err
	}
	body := &vmwareengine.Datastore{}
	if err := loadYAMLOrJSONInto(flagVmwareDsConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagVmwareDsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Datastores.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating datastore: %w", err)
	}
	fmt.Printf("Update datastore [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagVmwareDsFormat)
}
