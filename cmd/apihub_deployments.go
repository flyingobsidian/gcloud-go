package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub deployments (#1159) ---

var apihubDeplCmd = &cobra.Command{Use: "deployments", Short: "Manage API Hub deployments"}

var (
	flagApihubDeplLocation    string
	flagApihubDeplFormat      string
	flagApihubDeplDestination string
	flagApihubDeplSource      string
	flagApihubDeplPageSize    int64
)

var (
	apihubDeplDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete a deployment",
		Args: cobra.ExactArgs(1), RunE: runApihubDeplDelete,
	}
	apihubDeplDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe a deployment",
		Args: cobra.ExactArgs(1), RunE: runApihubDeplDescribe,
	}
	apihubDeplExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export a deployment to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runApihubDeplExport,
	}
	apihubDeplImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import a deployment from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runApihubDeplImport,
	}
	apihubDeplListCmd = &cobra.Command{
		Use: "list", Short: "List deployments in a location",
		Args: cobra.NoArgs, RunE: runApihubDeplList,
	}
)

func init() {
	all := []*cobra.Command{
		apihubDeplDeleteCmd, apihubDeplDescribeCmd,
		apihubDeplExportCmd, apihubDeplImportCmd,
		apihubDeplListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagApihubDeplLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagApihubDeplFormat, "format", "", "Output format")
	}
	nsBindExportFlags(apihubDeplExportCmd, &flagApihubDeplDestination)
	nsBindImportFlags(apihubDeplImportCmd, &flagApihubDeplSource)
	apihubDeplListCmd.Flags().Int64Var(&flagApihubDeplPageSize, "page-size", 0, "Maximum results per page")

	apihubDeplCmd.AddCommand(all...)
	apihubCmd.AddCommand(apihubDeplCmd)
}

func apihubDeplName(id string) (string, error) {
	return apihubResource(flagApihubDeplLocation, "deployments", id)
}

func runApihubDeplDelete(cmd *cobra.Command, args []string) error {
	name, err := apihubDeplName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Deployments.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting deployment: %w", err)
	}
	fmt.Printf("Deleted deployment [%s].\n", args[0])
	return nil
}

func runApihubDeplDescribe(cmd *cobra.Command, args []string) error {
	name, err := apihubDeplName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Deployments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing deployment: %w", err)
	}
	return emitFormatted(got, flagApihubDeplFormat)
}

func runApihubDeplExport(cmd *cobra.Command, args []string) error {
	name, err := apihubDeplName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Deployments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting deployment: %w", err)
	}
	return saveAsYAML(flagApihubDeplDestination, got)
}

func runApihubDeplImport(cmd *cobra.Command, args []string) error {
	parent, err := apihubLocationParent(flagApihubDeplLocation)
	if err != nil {
		return err
	}
	name, err := apihubDeplName(args[0])
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1Deployment{}
	if err := loadYAMLOrJSONInto(flagApihubDeplSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Deployments.Get(name).Context(ctx).Do(); err != nil {
		got, err := svc.Projects.Locations.Deployments.Create(parent, body).DeploymentId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating deployment: %w", err)
		}
		return emitFormatted(got, flagApihubDeplFormat)
	}
	got, err := svc.Projects.Locations.Deployments.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating deployment: %w", err)
	}
	return emitFormatted(got, flagApihubDeplFormat)
}

func runApihubDeplList(cmd *cobra.Command, args []string) error {
	parent, err := apihubLocationParent(flagApihubDeplLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1Deployment
	pageToken := ""
	for {
		call := svc.Projects.Locations.Deployments.List(parent).Context(ctx)
		if flagApihubDeplPageSize > 0 {
			call = call.PageSize(flagApihubDeplPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing deployments: %w", err)
		}
		all = append(all, resp.Deployments...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagApihubDeplFormat)
}
