package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub dependencies (#1158) ---

var apihubDepCmd = &cobra.Command{Use: "dependencies", Short: "Manage API Hub dependencies"}

var (
	flagApihubDepLocation    string
	flagApihubDepFormat      string
	flagApihubDepDestination string
	flagApihubDepSource      string
	flagApihubDepPageSize    int64
)

var (
	apihubDepDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete a dependency",
		Args: cobra.ExactArgs(1), RunE: runApihubDepDelete,
	}
	apihubDepDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe a dependency",
		Args: cobra.ExactArgs(1), RunE: runApihubDepDescribe,
	}
	apihubDepExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export a dependency to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runApihubDepExport,
	}
	apihubDepImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import a dependency from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runApihubDepImport,
	}
	apihubDepListCmd = &cobra.Command{
		Use: "list", Short: "List dependencies in a location",
		Args: cobra.NoArgs, RunE: runApihubDepList,
	}
)

func init() {
	all := []*cobra.Command{
		apihubDepDeleteCmd, apihubDepDescribeCmd,
		apihubDepExportCmd, apihubDepImportCmd,
		apihubDepListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagApihubDepLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagApihubDepFormat, "format", "", "Output format")
	}
	nsBindExportFlags(apihubDepExportCmd, &flagApihubDepDestination)
	nsBindImportFlags(apihubDepImportCmd, &flagApihubDepSource)
	apihubDepListCmd.Flags().Int64Var(&flagApihubDepPageSize, "page-size", 0, "Maximum results per page")

	apihubDepCmd.AddCommand(all...)
	apihubCmd.AddCommand(apihubDepCmd)
}

func apihubDepName(id string) (string, error) {
	return apihubResource(flagApihubDepLocation, "dependencies", id)
}

func runApihubDepDelete(cmd *cobra.Command, args []string) error {
	name, err := apihubDepName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Dependencies.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting dependency: %w", err)
	}
	fmt.Printf("Deleted dependency [%s].\n", args[0])
	return nil
}

func runApihubDepDescribe(cmd *cobra.Command, args []string) error {
	name, err := apihubDepName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Dependencies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing dependency: %w", err)
	}
	return emitFormatted(got, flagApihubDepFormat)
}

func runApihubDepExport(cmd *cobra.Command, args []string) error {
	name, err := apihubDepName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Dependencies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting dependency: %w", err)
	}
	return saveAsYAML(flagApihubDepDestination, got)
}

func runApihubDepImport(cmd *cobra.Command, args []string) error {
	parent, err := apihubLocationParent(flagApihubDepLocation)
	if err != nil {
		return err
	}
	name, err := apihubDepName(args[0])
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1Dependency{}
	if err := loadYAMLOrJSONInto(flagApihubDepSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Dependencies.Get(name).Context(ctx).Do(); err != nil {
		got, err := svc.Projects.Locations.Dependencies.Create(parent, body).DependencyId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating dependency: %w", err)
		}
		return emitFormatted(got, flagApihubDepFormat)
	}
	got, err := svc.Projects.Locations.Dependencies.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating dependency: %w", err)
	}
	return emitFormatted(got, flagApihubDepFormat)
}

func runApihubDepList(cmd *cobra.Command, args []string) error {
	parent, err := apihubLocationParent(flagApihubDepLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1Dependency
	pageToken := ""
	for {
		call := svc.Projects.Locations.Dependencies.List(parent).Context(ctx)
		if flagApihubDepPageSize > 0 {
			call = call.PageSize(flagApihubDepPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing dependencies: %w", err)
		}
		all = append(all, resp.Dependencies...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagApihubDepFormat)
}
