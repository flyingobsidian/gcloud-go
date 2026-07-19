package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware operations (#1122) ---

var vmwareOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage VMware Engine long-running operations"}

var (
	flagVmwareOpLocation string
	flagVmwareOpFormat   string
	flagVmwareOpFilter   string
	flagVmwareOpPageSize int64
)

var (
	vmwareOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a VMware Engine operation",
		Args: cobra.ExactArgs(1), RunE: runVmwareOpDelete,
	}
	vmwareOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a VMware Engine operation",
		Args: cobra.ExactArgs(1), RunE: runVmwareOpDescribe,
	}
	vmwareOpListCmd = &cobra.Command{
		Use: "list", Short: "List VMware Engine operations in a location",
		Args: cobra.NoArgs, RunE: runVmwareOpList,
	}
)

func init() {
	all := []*cobra.Command{vmwareOpDeleteCmd, vmwareOpDescribeCmd, vmwareOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagVmwareOpLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagVmwareOpFormat, "format", "", "Output format")
	}
	vmwareOpListCmd.Flags().StringVar(&flagVmwareOpFilter, "filter", "", "Server-side filter expression")
	vmwareOpListCmd.Flags().Int64Var(&flagVmwareOpPageSize, "page-size", 0, "Maximum results per page")

	vmwareOperationsCmd.AddCommand(all...)
	vmwareCmd.AddCommand(vmwareOperationsCmd)
}

func vmwareOpName(id string) (string, error) {
	return vmwareResource(flagVmwareOpLocation, "operations", id)
}

func runVmwareOpDelete(cmd *cobra.Command, args []string) error {
	name, err := vmwareOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation [%s].\n", args[0])
	return nil
}

func runVmwareOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := vmwareOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagVmwareOpFormat)
}

func runVmwareOpList(cmd *cobra.Command, args []string) error {
	parent, err := vmwareLocationParent(flagVmwareOpLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*vmwareengine.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
		if flagVmwareOpFilter != "" {
			call = call.Filter(flagVmwareOpFilter)
		}
		if flagVmwareOpPageSize > 0 {
			call = call.PageSize(flagVmwareOpPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagVmwareOpFormat)
}
