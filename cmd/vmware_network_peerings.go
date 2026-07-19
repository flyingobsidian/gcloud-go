package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware network-peerings (#1118) ---

var vmwareNetworkPeeringsCmd = &cobra.Command{Use: "network-peerings", Short: "Manage VMware Engine network peerings"}

var (
	flagVmwareNpLocation   string
	flagVmwareNpFormat     string
	flagVmwareNpConfigFile string
	flagVmwareNpUpdateMask string
	flagVmwareNpPageSize   int64
)

var (
	vmwareNpCreateCmd = &cobra.Command{
		Use: "create PEERING", Short: "Create a VMware Engine network peering",
		Args: cobra.ExactArgs(1), RunE: runVmwareNpCreate,
	}
	vmwareNpDeleteCmd = &cobra.Command{
		Use: "delete PEERING", Short: "Delete a VMware Engine network peering",
		Args: cobra.ExactArgs(1), RunE: runVmwareNpDelete,
	}
	vmwareNpDescribeCmd = &cobra.Command{
		Use: "describe PEERING", Short: "Describe a VMware Engine network peering",
		Args: cobra.ExactArgs(1), RunE: runVmwareNpDescribe,
	}
	vmwareNpListCmd = &cobra.Command{
		Use: "list", Short: "List VMware Engine network peerings in a location",
		Args: cobra.NoArgs, RunE: runVmwareNpList,
	}
	vmwareNpUpdateCmd = &cobra.Command{
		Use: "update PEERING", Short: "Update a VMware Engine network peering",
		Args: cobra.ExactArgs(1), RunE: runVmwareNpUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		vmwareNpCreateCmd, vmwareNpDeleteCmd, vmwareNpDescribeCmd,
		vmwareNpListCmd, vmwareNpUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagVmwareNpLocation, "location", "", "Location (required, typically \"global\")")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagVmwareNpFormat, "format", "", "Output format")
	}
	vmwareNpCreateCmd.Flags().StringVar(&flagVmwareNpConfigFile, "config-file", "", "YAML/JSON file with network peering body (required)")
	_ = vmwareNpCreateCmd.MarkFlagRequired("config-file")
	vmwareNpUpdateCmd.Flags().StringVar(&flagVmwareNpConfigFile, "config-file", "", "YAML/JSON file with network peering body (required)")
	_ = vmwareNpUpdateCmd.MarkFlagRequired("config-file")
	vmwareNpUpdateCmd.Flags().StringVar(&flagVmwareNpUpdateMask, "update-mask", "", "Update mask (comma-separated field paths)")
	vmwareNpListCmd.Flags().Int64Var(&flagVmwareNpPageSize, "page-size", 0, "Maximum results per page")

	vmwareNetworkPeeringsCmd.AddCommand(all...)
	vmwareCmd.AddCommand(vmwareNetworkPeeringsCmd)
}

func vmwareNpName(id string) (string, error) {
	return vmwareResource(flagVmwareNpLocation, "networkPeerings", id)
}

func runVmwareNpCreate(cmd *cobra.Command, args []string) error {
	parent, err := vmwareLocationParent(flagVmwareNpLocation)
	if err != nil {
		return err
	}
	body := &vmwareengine.NetworkPeering{}
	if err := loadYAMLOrJSONInto(flagVmwareNpConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NetworkPeerings.Create(parent, body).NetworkPeeringId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating network peering: %w", err)
	}
	fmt.Printf("Create network peering [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagVmwareNpFormat)
}

func runVmwareNpDelete(cmd *cobra.Command, args []string) error {
	name, err := vmwareNpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NetworkPeerings.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting network peering: %w", err)
	}
	fmt.Printf("Delete network peering [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagVmwareNpFormat)
}

func runVmwareNpDescribe(cmd *cobra.Command, args []string) error {
	name, err := vmwareNpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.NetworkPeerings.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing network peering: %w", err)
	}
	return emitFormatted(got, flagVmwareNpFormat)
}

func runVmwareNpList(cmd *cobra.Command, args []string) error {
	parent, err := vmwareLocationParent(flagVmwareNpLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*vmwareengine.NetworkPeering
	pageToken := ""
	for {
		call := svc.Projects.Locations.NetworkPeerings.List(parent).Context(ctx)
		if flagVmwareNpPageSize > 0 {
			call = call.PageSize(flagVmwareNpPageSize)
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
	return emitFormatted(all, flagVmwareNpFormat)
}

func runVmwareNpUpdate(cmd *cobra.Command, args []string) error {
	name, err := vmwareNpName(args[0])
	if err != nil {
		return err
	}
	body := &vmwareengine.NetworkPeering{}
	if err := loadYAMLOrJSONInto(flagVmwareNpConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagVmwareNpUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NetworkPeerings.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating network peering: %w", err)
	}
	fmt.Printf("Update network peering [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagVmwareNpFormat)
}
