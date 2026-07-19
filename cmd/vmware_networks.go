package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware networks (#1120) ---

var vmwareNetworksCmd = &cobra.Command{Use: "networks", Short: "Manage VMware Engine networks"}

var (
	flagVmwareNetLocation   string
	flagVmwareNetFormat     string
	flagVmwareNetConfigFile string
	flagVmwareNetUpdateMask string
	flagVmwareNetPageSize   int64
)

var (
	vmwareNetCreateCmd = &cobra.Command{
		Use: "create NETWORK", Short: "Create a VMware Engine network",
		Args: cobra.ExactArgs(1), RunE: runVmwareNetCreate,
	}
	vmwareNetDeleteCmd = &cobra.Command{
		Use: "delete NETWORK", Short: "Delete a VMware Engine network",
		Args: cobra.ExactArgs(1), RunE: runVmwareNetDelete,
	}
	vmwareNetDescribeCmd = &cobra.Command{
		Use: "describe NETWORK", Short: "Describe a VMware Engine network",
		Args: cobra.ExactArgs(1), RunE: runVmwareNetDescribe,
	}
	vmwareNetListCmd = &cobra.Command{
		Use: "list", Short: "List VMware Engine networks in a location",
		Args: cobra.NoArgs, RunE: runVmwareNetList,
	}
	vmwareNetUpdateCmd = &cobra.Command{
		Use: "update NETWORK", Short: "Update a VMware Engine network",
		Args: cobra.ExactArgs(1), RunE: runVmwareNetUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		vmwareNetCreateCmd, vmwareNetDeleteCmd, vmwareNetDescribeCmd,
		vmwareNetListCmd, vmwareNetUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagVmwareNetLocation, "location", "", "Location (required, typically \"global\")")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagVmwareNetFormat, "format", "", "Output format")
	}
	vmwareNetCreateCmd.Flags().StringVar(&flagVmwareNetConfigFile, "config-file", "", "YAML/JSON file with VMware Engine network body (required)")
	_ = vmwareNetCreateCmd.MarkFlagRequired("config-file")
	vmwareNetUpdateCmd.Flags().StringVar(&flagVmwareNetConfigFile, "config-file", "", "YAML/JSON file with VMware Engine network body (required)")
	_ = vmwareNetUpdateCmd.MarkFlagRequired("config-file")
	vmwareNetUpdateCmd.Flags().StringVar(&flagVmwareNetUpdateMask, "update-mask", "", "Update mask (comma-separated field paths)")
	vmwareNetListCmd.Flags().Int64Var(&flagVmwareNetPageSize, "page-size", 0, "Maximum results per page")

	vmwareNetworksCmd.AddCommand(all...)
	vmwareCmd.AddCommand(vmwareNetworksCmd)
}

func vmwareNetName(id string) (string, error) {
	return vmwareResource(flagVmwareNetLocation, "vmwareEngineNetworks", id)
}

func runVmwareNetCreate(cmd *cobra.Command, args []string) error {
	parent, err := vmwareLocationParent(flagVmwareNetLocation)
	if err != nil {
		return err
	}
	body := &vmwareengine.VmwareEngineNetwork{}
	if err := loadYAMLOrJSONInto(flagVmwareNetConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.VmwareEngineNetworks.Create(parent, body).VmwareEngineNetworkId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating vmware engine network: %w", err)
	}
	fmt.Printf("Create vmware engine network [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagVmwareNetFormat)
}

func runVmwareNetDelete(cmd *cobra.Command, args []string) error {
	name, err := vmwareNetName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.VmwareEngineNetworks.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting vmware engine network: %w", err)
	}
	fmt.Printf("Delete vmware engine network [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagVmwareNetFormat)
}

func runVmwareNetDescribe(cmd *cobra.Command, args []string) error {
	name, err := vmwareNetName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.VmwareEngineNetworks.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing vmware engine network: %w", err)
	}
	return emitFormatted(got, flagVmwareNetFormat)
}

func runVmwareNetList(cmd *cobra.Command, args []string) error {
	parent, err := vmwareLocationParent(flagVmwareNetLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*vmwareengine.VmwareEngineNetwork
	pageToken := ""
	for {
		call := svc.Projects.Locations.VmwareEngineNetworks.List(parent).Context(ctx)
		if flagVmwareNetPageSize > 0 {
			call = call.PageSize(flagVmwareNetPageSize)
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
	return emitFormatted(all, flagVmwareNetFormat)
}

func runVmwareNetUpdate(cmd *cobra.Command, args []string) error {
	name, err := vmwareNetName(args[0])
	if err != nil {
		return err
	}
	body := &vmwareengine.VmwareEngineNetwork{}
	if err := loadYAMLOrJSONInto(flagVmwareNetConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagVmwareNetUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.VmwareEngineNetworks.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating vmware engine network: %w", err)
	}
	fmt.Printf("Update vmware engine network [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagVmwareNetFormat)
}
