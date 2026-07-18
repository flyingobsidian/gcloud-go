package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud memorystore operations (#980) ---

var memstoreOpCmd = &cobra.Command{Use: "operations", Short: "Manage Memorystore operations"}

var (
	flagMemstoreOpLocation string
	flagMemstoreOpFormat   string
	flagMemstoreOpPageSize int64
)

var (
	memstoreOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Memorystore operation",
		Args: cobra.ExactArgs(1), RunE: runMemstoreOpCancel,
	}
	memstoreOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a Memorystore operation record",
		Args: cobra.ExactArgs(1), RunE: runMemstoreOpDelete,
	}
	memstoreOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Memorystore operation",
		Args: cobra.ExactArgs(1), RunE: runMemstoreOpDescribe,
	}
	memstoreOpListCmd = &cobra.Command{
		Use: "list", Short: "List Memorystore operations",
		Args: cobra.NoArgs, RunE: runMemstoreOpList,
	}
)

func init() {
	all := []*cobra.Command{memstoreOpCancelCmd, memstoreOpDeleteCmd, memstoreOpDescribeCmd, memstoreOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagMemstoreOpLocation, "location", "", "Memorystore location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagMemstoreOpFormat, "format", "", "Output format")
	}
	memstoreOpListCmd.Flags().Int64Var(&flagMemstoreOpPageSize, "page-size", 0, "Maximum results per page")

	memstoreOpCmd.AddCommand(all...)
	memorystoreCmd.AddCommand(memstoreOpCmd)
}

func memstoreOpParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagMemstoreOpLocation), nil
}

func memstoreOpName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := memstoreOpParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/operations/%s", parent, id), nil
}

func runMemstoreOpCancel(cmd *cobra.Command, args []string) error {
	name, err := memstoreOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := memorystoreRest.do(ctx, http.MethodPost, "/"+name+":cancel", nil, map[string]any{}, nil); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancel request issued for operation [%s].\n", args[0])
	return nil
}

func runMemstoreOpDelete(cmd *cobra.Command, args []string) error {
	name, err := memstoreOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := memorystoreRest.do(ctx, http.MethodDelete, "/"+name, nil, nil, nil); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation [%s].\n", args[0])
	return nil
}

func runMemstoreOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := memstoreOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := memorystoreRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagMemstoreOpFormat)
}

func runMemstoreOpList(cmd *cobra.Command, args []string) error {
	parent, err := memstoreOpParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	items, err := memorystoreRest.paginate(ctx, "/"+parent+"/operations", nil, "operations", flagMemstoreOpPageSize)
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	return emitFormatted(items, flagMemstoreOpFormat)
}
