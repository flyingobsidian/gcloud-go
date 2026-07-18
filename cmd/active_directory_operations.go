package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	managedidentities "google.golang.org/api/managedidentities/v1"
)

// --- gcloud active-directory operations (#1449) ---

var adOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Managed AD long-running operations"}

var (
	flagADOpFormat   string
	flagADOpPageSize int64
	flagADOpFilter   string
)

var (
	adOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Managed AD operation",
		Args: cobra.ExactArgs(1), RunE: runADOpCancel,
	}
	adOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Managed AD operation",
		Args: cobra.ExactArgs(1), RunE: runADOpDescribe,
	}
	adOpListCmd = &cobra.Command{
		Use: "list", Short: "List Managed AD operations",
		Args: cobra.NoArgs, RunE: runADOpList,
	}
)

func init() {
	all := []*cobra.Command{adOpCancelCmd, adOpDescribeCmd, adOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagADOpFormat, "format", "", "Output format")
	}
	adOpListCmd.Flags().Int64Var(&flagADOpPageSize, "page-size", 0, "Maximum results per page")
	adOpListCmd.Flags().StringVar(&flagADOpFilter, "filter", "", "Server-side list filter")

	adOperationsCmd.AddCommand(all...)
	activeDirectoryCmd.AddCommand(adOperationsCmd)
}

func adOpResource(project, op string) string {
	return fmt.Sprintf("projects/%s/locations/global/operations/%s", project, op)
}

func adOpParent(project string) string {
	return fmt.Sprintf("projects/%s/locations/global/operations", project)
}

func runADOpCancel(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Global.Operations.
		Cancel(adOpResource(project, args[0]),
			&managedidentities.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancel request issued for operation [%s].\n", args[0])
	return nil
}

func runADOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Global.Operations.
		Get(adOpResource(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagADOpFormat)
}

func runADOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*managedidentities.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Global.Operations.List(adOpParent(project)).Context(ctx)
		if flagADOpPageSize > 0 {
			call = call.PageSize(flagADOpPageSize)
		}
		if flagADOpFilter != "" {
			call = call.Filter(flagADOpFilter)
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
	return emitFormatted(all, flagADOpFormat)
}
