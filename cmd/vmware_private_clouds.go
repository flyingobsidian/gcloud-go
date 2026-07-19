package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware private-clouds (#1123) ---

var vmwarePrivateCloudsCmd = &cobra.Command{Use: "private-clouds", Short: "Manage VMware Engine private clouds"}

var (
	flagVmwarePcLocation   string
	flagVmwarePcFormat     string
	flagVmwarePcConfigFile string
	flagVmwarePcUpdateMask string
	flagVmwarePcPageSize   int64
)

var (
	vmwarePcCreateCmd = &cobra.Command{
		Use: "create PRIVATE_CLOUD", Short: "Create a VMware Engine private cloud",
		Args: cobra.ExactArgs(1), RunE: runVmwarePcCreate,
	}
	vmwarePcDeleteCmd = &cobra.Command{
		Use: "delete PRIVATE_CLOUD", Short: "Delete a VMware Engine private cloud",
		Args: cobra.ExactArgs(1), RunE: runVmwarePcDelete,
	}
	vmwarePcDescribeCmd = &cobra.Command{
		Use: "describe PRIVATE_CLOUD", Short: "Describe a VMware Engine private cloud",
		Args: cobra.ExactArgs(1), RunE: runVmwarePcDescribe,
	}
	vmwarePcListCmd = &cobra.Command{
		Use: "list", Short: "List VMware Engine private clouds in a location",
		Args: cobra.NoArgs, RunE: runVmwarePcList,
	}
	vmwarePcUpdateCmd = &cobra.Command{
		Use: "update PRIVATE_CLOUD", Short: "Update a VMware Engine private cloud",
		Args: cobra.ExactArgs(1), RunE: runVmwarePcUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		vmwarePcCreateCmd, vmwarePcDeleteCmd, vmwarePcDescribeCmd,
		vmwarePcListCmd, vmwarePcUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagVmwarePcLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagVmwarePcFormat, "format", "", "Output format")
	}
	vmwarePcCreateCmd.Flags().StringVar(&flagVmwarePcConfigFile, "config-file", "", "YAML/JSON file with private cloud body (required)")
	_ = vmwarePcCreateCmd.MarkFlagRequired("config-file")
	vmwarePcUpdateCmd.Flags().StringVar(&flagVmwarePcConfigFile, "config-file", "", "YAML/JSON file with private cloud body (required)")
	_ = vmwarePcUpdateCmd.MarkFlagRequired("config-file")
	vmwarePcUpdateCmd.Flags().StringVar(&flagVmwarePcUpdateMask, "update-mask", "", "Update mask (comma-separated field paths)")
	vmwarePcListCmd.Flags().Int64Var(&flagVmwarePcPageSize, "page-size", 0, "Maximum results per page")

	vmwarePrivateCloudsCmd.AddCommand(all...)
	vmwareCmd.AddCommand(vmwarePrivateCloudsCmd)
}

func vmwarePcName(id string) (string, error) {
	return vmwareResource(flagVmwarePcLocation, "privateClouds", id)
}

func runVmwarePcCreate(cmd *cobra.Command, args []string) error {
	parent, err := vmwareLocationParent(flagVmwarePcLocation)
	if err != nil {
		return err
	}
	body := &vmwareengine.PrivateCloud{}
	if err := loadYAMLOrJSONInto(flagVmwarePcConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateClouds.Create(parent, body).PrivateCloudId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating private cloud: %w", err)
	}
	fmt.Printf("Create private cloud [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagVmwarePcFormat)
}

func runVmwarePcDelete(cmd *cobra.Command, args []string) error {
	name, err := vmwarePcName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateClouds.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting private cloud: %w", err)
	}
	fmt.Printf("Delete private cloud [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagVmwarePcFormat)
}

func runVmwarePcDescribe(cmd *cobra.Command, args []string) error {
	name, err := vmwarePcName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.PrivateClouds.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing private cloud: %w", err)
	}
	return emitFormatted(got, flagVmwarePcFormat)
}

func runVmwarePcList(cmd *cobra.Command, args []string) error {
	parent, err := vmwareLocationParent(flagVmwarePcLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*vmwareengine.PrivateCloud
	pageToken := ""
	for {
		call := svc.Projects.Locations.PrivateClouds.List(parent).Context(ctx)
		if flagVmwarePcPageSize > 0 {
			call = call.PageSize(flagVmwarePcPageSize)
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
	return emitFormatted(all, flagVmwarePcFormat)
}

func runVmwarePcUpdate(cmd *cobra.Command, args []string) error {
	name, err := vmwarePcName(args[0])
	if err != nil {
		return err
	}
	body := &vmwareengine.PrivateCloud{}
	if err := loadYAMLOrJSONInto(flagVmwarePcConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagVmwarePcUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateClouds.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating private cloud: %w", err)
	}
	fmt.Printf("Update private cloud [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagVmwarePcFormat)
}
