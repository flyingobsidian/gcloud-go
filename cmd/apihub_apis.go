package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub apis (#1155) ---

var apihubApisCmd = &cobra.Command{Use: "apis", Short: "Manage API Hub APIs"}

var (
	flagApihubApisLocation    string
	flagApihubApisFormat      string
	flagApihubApisDestination string
	flagApihubApisSource      string
	flagApihubApisPageSize    int64
)

var (
	apihubApisDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an API",
		Args: cobra.ExactArgs(1), RunE: runApihubApisDelete,
	}
	apihubApisDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an API",
		Args: cobra.ExactArgs(1), RunE: runApihubApisDescribe,
	}
	apihubApisExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export an API to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runApihubApisExport,
	}
	apihubApisImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import an API from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runApihubApisImport,
	}
	apihubApisListCmd = &cobra.Command{
		Use: "list", Short: "List APIs in a location",
		Args: cobra.NoArgs, RunE: runApihubApisList,
	}
)

func init() {
	all := []*cobra.Command{
		apihubApisDeleteCmd, apihubApisDescribeCmd,
		apihubApisExportCmd, apihubApisImportCmd,
		apihubApisListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagApihubApisLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagApihubApisFormat, "format", "", "Output format")
	}
	nsBindExportFlags(apihubApisExportCmd, &flagApihubApisDestination)
	nsBindImportFlags(apihubApisImportCmd, &flagApihubApisSource)
	apihubApisListCmd.Flags().Int64Var(&flagApihubApisPageSize, "page-size", 0, "Maximum results per page")

	apihubApisCmd.AddCommand(all...)
	apihubCmd.AddCommand(apihubApisCmd)
}

func apihubApisName(id string) (string, error) {
	return apihubResource(flagApihubApisLocation, "apis", id)
}

func runApihubApisDelete(cmd *cobra.Command, args []string) error {
	name, err := apihubApisName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Apis.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting api: %w", err)
	}
	fmt.Printf("Deleted api [%s].\n", args[0])
	return nil
}

func runApihubApisDescribe(cmd *cobra.Command, args []string) error {
	name, err := apihubApisName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Apis.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing api: %w", err)
	}
	return emitFormatted(got, flagApihubApisFormat)
}

func runApihubApisExport(cmd *cobra.Command, args []string) error {
	name, err := apihubApisName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Apis.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting api: %w", err)
	}
	return saveAsYAML(flagApihubApisDestination, got)
}

func runApihubApisImport(cmd *cobra.Command, args []string) error {
	parent, err := apihubLocationParent(flagApihubApisLocation)
	if err != nil {
		return err
	}
	name, err := apihubApisName(args[0])
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1Api{}
	if err := loadYAMLOrJSONInto(flagApihubApisSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Apis.Get(name).Context(ctx).Do(); err != nil {
		got, err := svc.Projects.Locations.Apis.Create(parent, body).ApiId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating api: %w", err)
		}
		return emitFormatted(got, flagApihubApisFormat)
	}
	got, err := svc.Projects.Locations.Apis.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating api: %w", err)
	}
	return emitFormatted(got, flagApihubApisFormat)
}

func runApihubApisList(cmd *cobra.Command, args []string) error {
	parent, err := apihubLocationParent(flagApihubApisLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1Api
	pageToken := ""
	for {
		call := svc.Projects.Locations.Apis.List(parent).Context(ctx)
		if flagApihubApisPageSize > 0 {
			call = call.PageSize(flagApihubApisPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing apis: %w", err)
		}
		all = append(all, resp.Apis...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagApihubApisFormat)
}
