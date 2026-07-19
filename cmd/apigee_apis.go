package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
)

// --- gcloud apigee apis (#1375) ---

var apigeeApisCmd = &cobra.Command{Use: "apis", Short: "Manage Apigee API proxies"}

var (
	flagApigeeApiOrganization string
	flagApigeeApiFormat       string
)

var (
	apigeeApisDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an API proxy",
		Args: cobra.ExactArgs(1), RunE: runApigeeApisDelete,
	}
	apigeeApisDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an API proxy",
		Args: cobra.ExactArgs(1), RunE: runApigeeApisDescribe,
	}
	apigeeApisListCmd = &cobra.Command{
		Use: "list", Short: "List API proxies in an organization",
		Args: cobra.NoArgs, RunE: runApigeeApisList,
	}
)

func init() {
	all := []*cobra.Command{apigeeApisDeleteCmd, apigeeApisDescribeCmd, apigeeApisListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagApigeeApiOrganization, "organization", "", "Apigee organization (required)")
		_ = c.MarkFlagRequired("organization")
		c.Flags().StringVar(&flagApigeeApiFormat, "format", "", "Output format")
	}

	apigeeApisCmd.AddCommand(all...)
	apigeeCmd.AddCommand(apigeeApisCmd)
}

func apigeeApiName(id string) (string, error) {
	return apigeeResource(flagApigeeApiOrganization, "apis", id)
}

func runApigeeApisDelete(cmd *cobra.Command, args []string) error {
	name, err := apigeeApiName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Organizations.Apis.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting api: %w", err)
	}
	fmt.Printf("Deleted api [%s].\n", args[0])
	return nil
}

func runApigeeApisDescribe(cmd *cobra.Command, args []string) error {
	name, err := apigeeApiName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Apis.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing api: %w", err)
	}
	return emitFormatted(got, flagApigeeApiFormat)
}

func runApigeeApisList(cmd *cobra.Command, args []string) error {
	parent, err := apigeeOrgName(flagApigeeApiOrganization)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Organizations.Apis.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing apis: %w", err)
	}
	var all []*apigee.GoogleCloudApigeeV1ApiProxy
	all = append(all, resp.Proxies...)
	return emitFormatted(all, flagApigeeApiFormat)
}
