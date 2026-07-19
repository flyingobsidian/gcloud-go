package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware node-types (#1121) ---

var vmwareNodeTypesCmd = &cobra.Command{Use: "node-types", Short: "Show VMware Engine node types"}

var (
	flagVmwareNtLocation string
	flagVmwareNtFormat   string
	flagVmwareNtPageSize int64
)

var (
	vmwareNtDescribeCmd = &cobra.Command{
		Use: "describe NODE_TYPE", Short: "Describe a VMware Engine node type",
		Args: cobra.ExactArgs(1), RunE: runVmwareNtDescribe,
	}
	vmwareNtListCmd = &cobra.Command{
		Use: "list", Short: "List VMware Engine node types in a location",
		Args: cobra.NoArgs, RunE: runVmwareNtList,
	}
)

func init() {
	all := []*cobra.Command{vmwareNtDescribeCmd, vmwareNtListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagVmwareNtLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagVmwareNtFormat, "format", "", "Output format")
	}
	vmwareNtListCmd.Flags().Int64Var(&flagVmwareNtPageSize, "page-size", 0, "Maximum results per page")

	vmwareNodeTypesCmd.AddCommand(all...)
	vmwareCmd.AddCommand(vmwareNodeTypesCmd)
}

func vmwareNtName(id string) (string, error) {
	return vmwareResource(flagVmwareNtLocation, "nodeTypes", id)
}

func runVmwareNtDescribe(cmd *cobra.Command, args []string) error {
	name, err := vmwareNtName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.NodeTypes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing node type: %w", err)
	}
	return emitFormatted(got, flagVmwareNtFormat)
}

func runVmwareNtList(cmd *cobra.Command, args []string) error {
	parent, err := vmwareLocationParent(flagVmwareNtLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*vmwareengine.NodeType
	pageToken := ""
	for {
		call := svc.Projects.Locations.NodeTypes.List(parent).Context(ctx)
		if flagVmwareNtPageSize > 0 {
			call = call.PageSize(flagVmwareNtPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing node types: %w", err)
		}
		all = append(all, resp.NodeTypes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagVmwareNtFormat)
}
