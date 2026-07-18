package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// --- gcloud pam operations (#964) ---

var pamOpCmd = &cobra.Command{Use: "operations", Short: "Manage Privileged Access Manager operations"}

var (
	flagPAMOpLocation string
	flagPAMOpFormat   string
	flagPAMOpPageSize int64
	flagPAMOpTimeout  time.Duration
)

var (
	pamOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a Privileged Access Manager operation record",
		Args: cobra.ExactArgs(1), RunE: runPAMOpDelete,
	}
	pamOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Privileged Access Manager operation",
		Args: cobra.ExactArgs(1), RunE: runPAMOpDescribe,
	}
	pamOpListCmd = &cobra.Command{
		Use: "list", Short: "List Privileged Access Manager operations",
		Args: cobra.NoArgs, RunE: runPAMOpList,
	}
	pamOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for a Privileged Access Manager operation to finish",
		Args: cobra.ExactArgs(1), RunE: runPAMOpWait,
	}
)

func init() {
	all := []*cobra.Command{pamOpDeleteCmd, pamOpDescribeCmd, pamOpListCmd, pamOpWaitCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagPAMOpLocation, "location", "",
			"Location containing the operation (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagPAMOpFormat, "format", "", "Output format")
	}
	pamOpListCmd.Flags().Int64Var(&flagPAMOpPageSize, "page-size", 0, "Maximum results per page")
	pamOpWaitCmd.Flags().DurationVar(&flagPAMOpTimeout, "timeout", 30*time.Minute,
		"Maximum time to wait for the operation to finish")

	pamOpCmd.AddCommand(all...)
	pamCmd.AddCommand(pamOpCmd)
}

func pamOpParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagPAMOpLocation), nil
}

func pamOpName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") ||
		strings.HasPrefix(id, "folders/") ||
		strings.HasPrefix(id, "organizations/") {
		return id, nil
	}
	parent, err := pamOpParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/operations/%s", parent, id), nil
}

func runPAMOpDelete(cmd *cobra.Command, args []string) error {
	name, err := pamOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := pamRest.do(ctx, http.MethodDelete, "/"+name, nil, nil, nil); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation [%s].\n", args[0])
	return nil
}

func runPAMOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := pamOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := pamRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagPAMOpFormat)
}

func runPAMOpList(cmd *cobra.Command, args []string) error {
	parent, err := pamOpParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	items, err := pamRest.paginate(ctx, "/"+parent+"/operations", nil, "operations", flagPAMOpPageSize)
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	return emitFormatted(items, flagPAMOpFormat)
}

func runPAMOpWait(cmd *cobra.Command, args []string) error {
	name, err := pamOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	op, err := pamRest.waitOperation(ctx, name, flagPAMOpTimeout)
	if err != nil {
		return err
	}
	fmt.Printf("Operation %s completed.\n", args[0])
	return emitFormatted(op, flagPAMOpFormat)
}
