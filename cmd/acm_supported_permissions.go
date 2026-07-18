package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
)

// --- gcloud access-context-manager supported-permissions (#1446) ---

var acmSPCmd = &cobra.Command{Use: "supported-permissions", Short: "VPC-SC supported permissions"}

var (
	flagACMSPFormat   string
	flagACMSPPageSize int64
)

var (
	acmSPListCmd = &cobra.Command{
		Use: "list", Short: "List VPC-SC supported permissions",
		Args: cobra.NoArgs, RunE: runACMSPList,
	}
	acmSPDescribeCmd = &cobra.Command{
		Use: "describe PERMISSION", Short: "Describe a VPC-SC supported permission",
		Args: cobra.ExactArgs(1), RunE: runACMSPDescribe,
	}
)

func init() {
	for _, c := range []*cobra.Command{acmSPListCmd, acmSPDescribeCmd} {
		c.Flags().StringVar(&flagACMSPFormat, "format", "", "Output format")
	}
	acmSPListCmd.Flags().Int64Var(&flagACMSPPageSize, "page-size", 0, "Maximum results per page")

	acmSPCmd.AddCommand(acmSPListCmd, acmSPDescribeCmd)
	accessContextManagerCmd.AddCommand(acmSPCmd)
}

func acmSPListAll(ctx context.Context) ([]string, error) {
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return nil, err
	}
	var all []string
	pageToken := ""
	for {
		call := svc.Permissions.List().Context(ctx)
		if flagACMSPPageSize > 0 {
			call = call.PageSize(flagACMSPPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("listing supported permissions: %w", err)
		}
		all = append(all, resp.SupportedPermissions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return all, nil
}

func runACMSPList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	all, err := acmSPListAll(ctx)
	if err != nil {
		return err
	}
	return emitFormatted(all, flagACMSPFormat)
}

func runACMSPDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	all, err := acmSPListAll(ctx)
	if err != nil {
		return err
	}
	for _, p := range all {
		if p == args[0] {
			return emitFormatted(p, flagACMSPFormat)
		}
	}
	return fmt.Errorf("supported permission %q not found", args[0])
}
