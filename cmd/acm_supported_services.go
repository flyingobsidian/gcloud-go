package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	accesscontextmanager "google.golang.org/api/accesscontextmanager/v1"
)

// --- gcloud access-context-manager supported-services (#1447) ---

var acmSSCmd = &cobra.Command{Use: "supported-services", Short: "VPC-SC supported services"}

var (
	flagACMSSFormat   string
	flagACMSSPageSize int64
)

var (
	acmSSListCmd = &cobra.Command{
		Use: "list", Short: "List VPC-SC supported services",
		Args: cobra.NoArgs, RunE: runACMSSList,
	}
	acmSSDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe a VPC-SC supported service",
		Args: cobra.ExactArgs(1), RunE: runACMSSDescribe,
	}
)

func init() {
	for _, c := range []*cobra.Command{acmSSListCmd, acmSSDescribeCmd} {
		c.Flags().StringVar(&flagACMSSFormat, "format", "", "Output format")
	}
	acmSSListCmd.Flags().Int64Var(&flagACMSSPageSize, "page-size", 0, "Maximum results per page")

	acmSSCmd.AddCommand(acmSSListCmd, acmSSDescribeCmd)
	accessContextManagerCmd.AddCommand(acmSSCmd)
}

func runACMSSList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*accesscontextmanager.SupportedService
	pageToken := ""
	for {
		call := svc.Services.List().Context(ctx)
		if flagACMSSPageSize > 0 {
			call = call.PageSize(flagACMSSPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing supported services: %w", err)
		}
		all = append(all, resp.SupportedServices...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagACMSSFormat)
}

func runACMSSDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Services.Get(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing supported service: %w", err)
	}
	return emitFormatted(got, flagACMSSFormat)
}
