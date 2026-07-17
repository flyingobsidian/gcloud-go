package cmd

import (
	"context"
	"fmt"
	"path"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	netapp "google.golang.org/api/netapp/v1"
)

// --- gcloud netapp operations (#1203) ---

var netappOpCmd = &cobra.Command{Use: "operations", Short: "Manage NetApp Files operations"}

var (
	flagNetAppOpLocation string
	flagNetAppOpFormat   string
	flagNetAppOpFilter   string
	flagNetAppOpPageSize int64
)

var (
	netappOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a NetApp operation",
		Args: cobra.ExactArgs(1), RunE: runNetAppOpCancel,
	}
	netappOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a NetApp operation record",
		Args: cobra.ExactArgs(1), RunE: runNetAppOpDelete,
	}
	netappOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a NetApp operation",
		Args: cobra.ExactArgs(1), RunE: runNetAppOpDescribe,
	}
	netappOpListCmd = &cobra.Command{
		Use: "list", Short: "List NetApp operations",
		Args: cobra.NoArgs, RunE: runNetAppOpList,
	}
)

func init() {
	all := []*cobra.Command{netappOpCancelCmd, netappOpDeleteCmd, netappOpDescribeCmd, netappOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagNetAppOpLocation, "location", "", "Location for the operation (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNetAppOpFormat, "format", "", "Output format")
	}
	netappOpListCmd.Flags().StringVar(&flagNetAppOpFilter, "filter", "", "Server-side filter expression")
	netappOpListCmd.Flags().Int64Var(&flagNetAppOpPageSize, "page-size", 0, "Maximum number of results per page")

	netappOpCmd.AddCommand(all...)
	netappCmd.AddCommand(netappOpCmd)
}

func netappOpParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return netappLocationParent(project, flagNetAppOpLocation), nil
}

func netappOpName(id string) (string, error) {
	parent, err := netappOpParent()
	if err != nil {
		return "", err
	}
	return netappChild("operations", id, parent), nil
}

func runNetAppOpCancel(cmd *cobra.Command, args []string) error {
	name, err := netappOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(name, &netapp.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancel request issued for operation %s.\n", args[0])
	return nil
}

func runNetAppOpDelete(cmd *cobra.Command, args []string) error {
	name, err := netappOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation %s.\n", args[0])
	return nil
}

func runNetAppOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := netappOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagNetAppOpFormat)
}

func runNetAppOpList(cmd *cobra.Command, args []string) error {
	parent, err := netappOpParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*netapp.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
		if flagNetAppOpFilter != "" {
			call = call.Filter(flagNetAppOpFilter)
		}
		if flagNetAppOpPageSize > 0 {
			call = call.PageSize(flagNetAppOpPageSize)
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
	if flagNetAppOpFormat != "" {
		return emitFormatted(all, flagNetAppOpFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range all {
		fmt.Printf("%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}
