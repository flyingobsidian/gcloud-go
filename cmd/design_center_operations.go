package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// --- gcloud design-center operations (#1534) ---

var dcOpCmd = &cobra.Command{Use: "operations", Short: "Manage Design Center operations"}

var (
	flagDCOpLocation string
	flagDCOpFormat   string
	flagDCOpFilter   string
	flagDCOpPageSize int64
	flagDCOpTimeout  time.Duration
)

var (
	dcOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Design Center operation",
		Args: cobra.ExactArgs(1), RunE: runDCOpCancel,
	}
	dcOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a Design Center operation record",
		Args: cobra.ExactArgs(1), RunE: runDCOpDelete,
	}
	dcOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Design Center operation",
		Args: cobra.ExactArgs(1), RunE: runDCOpDescribe,
	}
	dcOpListCmd = &cobra.Command{
		Use: "list", Short: "List Design Center operations",
		Args: cobra.NoArgs, RunE: runDCOpList,
	}
	dcOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for a Design Center operation to finish",
		Args: cobra.ExactArgs(1), RunE: runDCOpWait,
	}
)

func init() {
	all := []*cobra.Command{dcOpCancelCmd, dcOpDeleteCmd, dcOpDescribeCmd, dcOpListCmd, dcOpWaitCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagDCOpLocation, "location", "", "Design Center location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagDCOpFormat, "format", "", "Output format")
	}
	dcOpListCmd.Flags().StringVar(&flagDCOpFilter, "filter", "", "Server-side filter expression")
	dcOpListCmd.Flags().Int64Var(&flagDCOpPageSize, "page-size", 0, "Maximum results per page")
	dcOpWaitCmd.Flags().DurationVar(&flagDCOpTimeout, "timeout", 30*time.Minute,
		"Maximum time to wait for the operation to finish")

	dcOpCmd.AddCommand(all...)
	designCenterCmd.AddCommand(dcOpCmd)
}

func dcOpQualifiedName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/operations/%s", dcLocationName(project, flagDCOpLocation), id), nil
}

func runDCOpCancel(cmd *cobra.Command, args []string) error {
	name, err := dcOpQualifiedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := designCenterRest.do(ctx, http.MethodPost, "/"+name+":cancel", nil, map[string]any{}, nil); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancel request issued for operation %s.\n", args[0])
	return nil
}

func runDCOpDelete(cmd *cobra.Command, args []string) error {
	name, err := dcOpQualifiedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := designCenterRest.do(ctx, http.MethodDelete, "/"+name, nil, nil, nil); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation %s.\n", args[0])
	return nil
}

func runDCOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := dcOpQualifiedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := designCenterRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagDCOpFormat)
}

func runDCOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	base := url.Values{}
	if flagDCOpFilter != "" {
		base.Set("filter", flagDCOpFilter)
	}
	ctx := context.Background()
	parent := fmt.Sprintf("%s/operations", dcLocationName(project, flagDCOpLocation))
	items, err := designCenterRest.paginate(ctx, "/"+parent, base, "operations", flagDCOpPageSize)
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if flagDCOpFormat != "" {
		return emitFormatted(items, flagDCOpFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range items {
		name, _ := o["name"].(string)
		done, _ := o["done"].(bool)
		fmt.Printf("%-40s %v\n", path.Base(name), done)
	}
	return nil
}

func runDCOpWait(cmd *cobra.Command, args []string) error {
	name, err := dcOpQualifiedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	op, err := designCenterRest.waitOperation(ctx, name, flagDCOpTimeout)
	if err != nil {
		return err
	}
	fmt.Printf("Operation %s completed.\n", args[0])
	return emitFormatted(op, flagDCOpFormat)
}
