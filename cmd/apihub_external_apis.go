package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub external-apis (#1161) ---

var apihubExtCmd = &cobra.Command{Use: "external-apis", Short: "Manage API Hub external APIs"}

var (
	flagApihubExtLocation    string
	flagApihubExtFormat      string
	flagApihubExtDestination string
	flagApihubExtSource      string
	flagApihubExtPageSize    int64
)

var (
	apihubExtDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an external API",
		Args: cobra.ExactArgs(1), RunE: runApihubExtDelete,
	}
	apihubExtDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an external API",
		Args: cobra.ExactArgs(1), RunE: runApihubExtDescribe,
	}
	apihubExtExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export an external API to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runApihubExtExport,
	}
	apihubExtImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import an external API from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runApihubExtImport,
	}
	apihubExtListCmd = &cobra.Command{
		Use: "list", Short: "List external APIs in a location",
		Args: cobra.NoArgs, RunE: runApihubExtList,
	}
)

func init() {
	all := []*cobra.Command{
		apihubExtDeleteCmd, apihubExtDescribeCmd,
		apihubExtExportCmd, apihubExtImportCmd,
		apihubExtListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagApihubExtLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagApihubExtFormat, "format", "", "Output format")
	}
	nsBindExportFlags(apihubExtExportCmd, &flagApihubExtDestination)
	nsBindImportFlags(apihubExtImportCmd, &flagApihubExtSource)
	apihubExtListCmd.Flags().Int64Var(&flagApihubExtPageSize, "page-size", 0, "Maximum results per page")

	apihubExtCmd.AddCommand(all...)
	apihubCmd.AddCommand(apihubExtCmd)
}

func apihubExtName(id string) (string, error) {
	return apihubResource(flagApihubExtLocation, "externalApis", id)
}

func runApihubExtDelete(cmd *cobra.Command, args []string) error {
	name, err := apihubExtName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.ExternalApis.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting external api: %w", err)
	}
	fmt.Printf("Deleted external api [%s].\n", args[0])
	return nil
}

func runApihubExtDescribe(cmd *cobra.Command, args []string) error {
	name, err := apihubExtName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ExternalApis.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing external api: %w", err)
	}
	return emitFormatted(got, flagApihubExtFormat)
}

func runApihubExtExport(cmd *cobra.Command, args []string) error {
	name, err := apihubExtName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ExternalApis.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting external api: %w", err)
	}
	return saveAsYAML(flagApihubExtDestination, got)
}

func runApihubExtImport(cmd *cobra.Command, args []string) error {
	parent, err := apihubLocationParent(flagApihubExtLocation)
	if err != nil {
		return err
	}
	name, err := apihubExtName(args[0])
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1ExternalApi{}
	if err := loadYAMLOrJSONInto(flagApihubExtSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.ExternalApis.Get(name).Context(ctx).Do(); err != nil {
		got, err := svc.Projects.Locations.ExternalApis.Create(parent, body).ExternalApiId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating external api: %w", err)
		}
		return emitFormatted(got, flagApihubExtFormat)
	}
	got, err := svc.Projects.Locations.ExternalApis.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating external api: %w", err)
	}
	return emitFormatted(got, flagApihubExtFormat)
}

func runApihubExtList(cmd *cobra.Command, args []string) error {
	parent, err := apihubLocationParent(flagApihubExtLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1ExternalApi
	pageToken := ""
	for {
		call := svc.Projects.Locations.ExternalApis.List(parent).Context(ctx)
		if flagApihubExtPageSize > 0 {
			call = call.PageSize(flagApihubExtPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing external apis: %w", err)
		}
		all = append(all, resp.ExternalApis...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagApihubExtFormat)
}
