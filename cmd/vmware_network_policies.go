package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware network-policies (#1119) ---

var vmwareNetworkPoliciesCmd = &cobra.Command{Use: "network-policies", Short: "Manage VMware Engine network policies"}

var (
	flagVmwareNpolLocation   string
	flagVmwareNpolFormat     string
	flagVmwareNpolConfigFile string
	flagVmwareNpolUpdateMask string
	flagVmwareNpolPageSize   int64
)

var (
	vmwareNpolCreateCmd = &cobra.Command{
		Use: "create POLICY", Short: "Create a VMware Engine network policy",
		Args: cobra.ExactArgs(1), RunE: runVmwareNpolCreate,
	}
	vmwareNpolDeleteCmd = &cobra.Command{
		Use: "delete POLICY", Short: "Delete a VMware Engine network policy",
		Args: cobra.ExactArgs(1), RunE: runVmwareNpolDelete,
	}
	vmwareNpolDescribeCmd = &cobra.Command{
		Use: "describe POLICY", Short: "Describe a VMware Engine network policy",
		Args: cobra.ExactArgs(1), RunE: runVmwareNpolDescribe,
	}
	vmwareNpolListCmd = &cobra.Command{
		Use: "list", Short: "List VMware Engine network policies in a location",
		Args: cobra.NoArgs, RunE: runVmwareNpolList,
	}
	vmwareNpolUpdateCmd = &cobra.Command{
		Use: "update POLICY", Short: "Update a VMware Engine network policy",
		Args: cobra.ExactArgs(1), RunE: runVmwareNpolUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		vmwareNpolCreateCmd, vmwareNpolDeleteCmd, vmwareNpolDescribeCmd,
		vmwareNpolListCmd, vmwareNpolUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagVmwareNpolLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagVmwareNpolFormat, "format", "", "Output format")
	}
	vmwareNpolCreateCmd.Flags().StringVar(&flagVmwareNpolConfigFile, "config-file", "", "YAML/JSON file with network policy body (required)")
	_ = vmwareNpolCreateCmd.MarkFlagRequired("config-file")
	vmwareNpolUpdateCmd.Flags().StringVar(&flagVmwareNpolConfigFile, "config-file", "", "YAML/JSON file with network policy body (required)")
	_ = vmwareNpolUpdateCmd.MarkFlagRequired("config-file")
	vmwareNpolUpdateCmd.Flags().StringVar(&flagVmwareNpolUpdateMask, "update-mask", "", "Update mask (comma-separated field paths)")
	vmwareNpolListCmd.Flags().Int64Var(&flagVmwareNpolPageSize, "page-size", 0, "Maximum results per page")

	vmwareNetworkPoliciesCmd.AddCommand(all...)
	vmwareCmd.AddCommand(vmwareNetworkPoliciesCmd)
}

func vmwareNpolName(id string) (string, error) {
	return vmwareResource(flagVmwareNpolLocation, "networkPolicies", id)
}

func runVmwareNpolCreate(cmd *cobra.Command, args []string) error {
	parent, err := vmwareLocationParent(flagVmwareNpolLocation)
	if err != nil {
		return err
	}
	body := &vmwareengine.NetworkPolicy{}
	if err := loadYAMLOrJSONInto(flagVmwareNpolConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NetworkPolicies.Create(parent, body).NetworkPolicyId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating network policy: %w", err)
	}
	fmt.Printf("Create network policy [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagVmwareNpolFormat)
}

func runVmwareNpolDelete(cmd *cobra.Command, args []string) error {
	name, err := vmwareNpolName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NetworkPolicies.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting network policy: %w", err)
	}
	fmt.Printf("Delete network policy [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagVmwareNpolFormat)
}

func runVmwareNpolDescribe(cmd *cobra.Command, args []string) error {
	name, err := vmwareNpolName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.NetworkPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing network policy: %w", err)
	}
	return emitFormatted(got, flagVmwareNpolFormat)
}

func runVmwareNpolList(cmd *cobra.Command, args []string) error {
	parent, err := vmwareLocationParent(flagVmwareNpolLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*vmwareengine.NetworkPolicy
	pageToken := ""
	for {
		call := svc.Projects.Locations.NetworkPolicies.List(parent).Context(ctx)
		if flagVmwareNpolPageSize > 0 {
			call = call.PageSize(flagVmwareNpolPageSize)
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
	return emitFormatted(all, flagVmwareNpolFormat)
}

func runVmwareNpolUpdate(cmd *cobra.Command, args []string) error {
	name, err := vmwareNpolName(args[0])
	if err != nil {
		return err
	}
	body := &vmwareengine.NetworkPolicy{}
	if err := loadYAMLOrJSONInto(flagVmwareNpolConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagVmwareNpolUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NetworkPolicies.Patch(name, body).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating network policy: %w", err)
	}
	fmt.Printf("Update network policy [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagVmwareNpolFormat)
}
