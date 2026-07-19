package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
)

// --- gcloud apigee applications (#1376) ---

var apigeeApplicationsCmd = &cobra.Command{Use: "applications", Short: "Manage Apigee applications"}

var (
	flagApigeeAppOrganization string
	flagApigeeAppFormat       string
	flagApigeeAppPageSize     int64
)

var (
	apigeeAppDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an Apigee application",
		Args: cobra.ExactArgs(1), RunE: runApigeeAppDescribe,
	}
	apigeeAppListCmd = &cobra.Command{
		Use: "list", Short: "List Apigee applications in an organization",
		Args: cobra.NoArgs, RunE: runApigeeAppList,
	}
)

func init() {
	all := []*cobra.Command{apigeeAppDescribeCmd, apigeeAppListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagApigeeAppOrganization, "organization", "", "Apigee organization (required)")
		_ = c.MarkFlagRequired("organization")
		c.Flags().StringVar(&flagApigeeAppFormat, "format", "", "Output format")
	}
	apigeeAppListCmd.Flags().Int64Var(&flagApigeeAppPageSize, "page-size", 0, "Maximum results per page")

	apigeeApplicationsCmd.AddCommand(all...)
	apigeeCmd.AddCommand(apigeeApplicationsCmd)
}

func apigeeAppName(id string) (string, error) {
	return apigeeResource(flagApigeeAppOrganization, "apps", id)
}

func runApigeeAppDescribe(cmd *cobra.Command, args []string) error {
	name, err := apigeeAppName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Apps.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing application: %w", err)
	}
	return emitFormatted(got, flagApigeeAppFormat)
}

func runApigeeAppList(cmd *cobra.Command, args []string) error {
	parent, err := apigeeOrgName(flagApigeeAppOrganization)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*apigee.GoogleCloudApigeeV1App
	pageToken := ""
	for {
		call := svc.Organizations.Apps.List(parent).Context(ctx)
		if flagApigeeAppPageSize > 0 {
			call = call.PageSize(flagApigeeAppPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing applications: %w", err)
		}
		all = append(all, resp.App...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagApigeeAppFormat)
}
