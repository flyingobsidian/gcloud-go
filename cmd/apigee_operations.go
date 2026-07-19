package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
)

// --- gcloud apigee operations (#1381) ---

var apigeeOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Apigee long-running operations"}

var (
	flagApigeeOpOrganization string
	flagApigeeOpFormat       string
	flagApigeeOpPageSize     int64
)

var (
	apigeeOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an Apigee operation",
		Args: cobra.ExactArgs(1), RunE: runApigeeOpDescribe,
	}
	apigeeOpListCmd = &cobra.Command{
		Use: "list", Short: "List Apigee operations in an organization",
		Args: cobra.NoArgs, RunE: runApigeeOpList,
	}
)

func init() {
	all := []*cobra.Command{apigeeOpDescribeCmd, apigeeOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagApigeeOpOrganization, "organization", "", "Apigee organization (required)")
		_ = c.MarkFlagRequired("organization")
		c.Flags().StringVar(&flagApigeeOpFormat, "format", "", "Output format")
	}
	apigeeOpListCmd.Flags().Int64Var(&flagApigeeOpPageSize, "page-size", 0, "Maximum results per page")

	apigeeOperationsCmd.AddCommand(all...)
	apigeeCmd.AddCommand(apigeeOperationsCmd)
}

func apigeeOpName(id string) (string, error) {
	return apigeeResource(flagApigeeOpOrganization, "operations", id)
}

func runApigeeOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := apigeeOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagApigeeOpFormat)
}

func runApigeeOpList(cmd *cobra.Command, args []string) error {
	parent, err := apigeeOrgName(flagApigeeOpOrganization)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*apigee.GoogleLongrunningOperation
	pageToken := ""
	for {
		call := svc.Organizations.Operations.List(parent).Context(ctx)
		if flagApigeeOpPageSize > 0 {
			call = call.PageSize(flagApigeeOpPageSize)
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
	return emitFormatted(all, flagApigeeOpFormat)
}
