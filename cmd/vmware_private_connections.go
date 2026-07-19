package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware private-connections (#1124) ---

var vmwarePrivateConnectionsCmd = &cobra.Command{Use: "private-connections", Short: "Manage VMware Engine private connections"}

var (
	flagVmwarePconLocation   string
	flagVmwarePconFormat     string
	flagVmwarePconConfigFile string
	flagVmwarePconUpdateMask string
	flagVmwarePconPageSize   int64
)

var (
	vmwarePconCreateCmd = &cobra.Command{
		Use: "create PRIVATE_CONNECTION", Short: "Create a VMware Engine private connection",
		Args: cobra.ExactArgs(1), RunE: runVmwarePconCreate,
	}
	vmwarePconDeleteCmd = &cobra.Command{
		Use: "delete PRIVATE_CONNECTION", Short: "Delete a VMware Engine private connection",
		Args: cobra.ExactArgs(1), RunE: runVmwarePconDelete,
	}
	vmwarePconDescribeCmd = &cobra.Command{
		Use: "describe PRIVATE_CONNECTION", Short: "Describe a VMware Engine private connection",
		Args: cobra.ExactArgs(1), RunE: runVmwarePconDescribe,
	}
	vmwarePconListCmd = &cobra.Command{
		Use: "list", Short: "List VMware Engine private connections in a location",
		Args: cobra.NoArgs, RunE: runVmwarePconList,
	}
	vmwarePconUpdateCmd = &cobra.Command{
		Use: "update PRIVATE_CONNECTION", Short: "Update a VMware Engine private connection",
		Args: cobra.ExactArgs(1), RunE: runVmwarePconUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		vmwarePconCreateCmd, vmwarePconDeleteCmd, vmwarePconDescribeCmd,
		vmwarePconListCmd, vmwarePconUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagVmwarePconLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagVmwarePconFormat, "format", "", "Output format")
	}
	vmwarePconCreateCmd.Flags().StringVar(&flagVmwarePconConfigFile, "config-file", "", "YAML/JSON file with private connection body (required)")
	_ = vmwarePconCreateCmd.MarkFlagRequired("config-file")
	vmwarePconUpdateCmd.Flags().StringVar(&flagVmwarePconConfigFile, "config-file", "", "YAML/JSON file with private connection body (required)")
	_ = vmwarePconUpdateCmd.MarkFlagRequired("config-file")
	vmwarePconUpdateCmd.Flags().StringVar(&flagVmwarePconUpdateMask, "update-mask", "", "Update mask (comma-separated field paths)")
	vmwarePconListCmd.Flags().Int64Var(&flagVmwarePconPageSize, "page-size", 0, "Maximum results per page")

	vmwarePrivateConnectionsCmd.AddCommand(all...)
	vmwareCmd.AddCommand(vmwarePrivateConnectionsCmd)
}

func vmwarePconName(id string) (string, error) {
	return vmwareResource(flagVmwarePconLocation, "privateConnections", id)
}

func runVmwarePconCreate(cmd *cobra.Command, args []string) error {
	parent, err := vmwareLocationParent(flagVmwarePconLocation)
	if err != nil {
		return err
	}
	body := &vmwareengine.PrivateConnection{}
	if err := loadYAMLOrJSONInto(flagVmwarePconConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateConnections.Create(parent, body).PrivateConnectionId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating private connection: %w", err)
	}
	fmt.Printf("Create private connection [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagVmwarePconFormat)
}

func runVmwarePconDelete(cmd *cobra.Command, args []string) error {
	name, err := vmwarePconName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateConnections.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting private connection: %w", err)
	}
	fmt.Printf("Delete private connection [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagVmwarePconFormat)
}

func runVmwarePconDescribe(cmd *cobra.Command, args []string) error {
	name, err := vmwarePconName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.PrivateConnections.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing private connection: %w", err)
	}
	return emitFormatted(got, flagVmwarePconFormat)
}

func runVmwarePconList(cmd *cobra.Command, args []string) error {
	parent, err := vmwareLocationParent(flagVmwarePconLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*vmwareengine.PrivateConnection
	pageToken := ""
	for {
		call := svc.Projects.Locations.PrivateConnections.List(parent).Context(ctx)
		if flagVmwarePconPageSize > 0 {
			call = call.PageSize(flagVmwarePconPageSize)
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
	return emitFormatted(all, flagVmwarePconFormat)
}

func runVmwarePconUpdate(cmd *cobra.Command, args []string) error {
	name, err := vmwarePconName(args[0])
	if err != nil {
		return err
	}
	body := &vmwareengine.PrivateConnection{}
	if err := loadYAMLOrJSONInto(flagVmwarePconConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagVmwarePconUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateConnections.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating private connection: %w", err)
	}
	fmt.Printf("Update private connection [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagVmwarePconFormat)
}
