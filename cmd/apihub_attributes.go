package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub attributes (#1156) ---

var apihubAttrsCmd = &cobra.Command{Use: "attributes", Short: "Manage API Hub attributes"}

var (
	flagApihubAttrsLocation    string
	flagApihubAttrsFormat      string
	flagApihubAttrsDestination string
	flagApihubAttrsSource      string
	flagApihubAttrsPageSize    int64
)

var (
	apihubAttrsDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an attribute",
		Args: cobra.ExactArgs(1), RunE: runApihubAttrsDelete,
	}
	apihubAttrsDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an attribute",
		Args: cobra.ExactArgs(1), RunE: runApihubAttrsDescribe,
	}
	apihubAttrsExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export an attribute to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runApihubAttrsExport,
	}
	apihubAttrsImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import an attribute from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runApihubAttrsImport,
	}
	apihubAttrsListCmd = &cobra.Command{
		Use: "list", Short: "List attributes in a location",
		Args: cobra.NoArgs, RunE: runApihubAttrsList,
	}
)

func init() {
	all := []*cobra.Command{
		apihubAttrsDeleteCmd, apihubAttrsDescribeCmd,
		apihubAttrsExportCmd, apihubAttrsImportCmd,
		apihubAttrsListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagApihubAttrsLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagApihubAttrsFormat, "format", "", "Output format")
	}
	nsBindExportFlags(apihubAttrsExportCmd, &flagApihubAttrsDestination)
	nsBindImportFlags(apihubAttrsImportCmd, &flagApihubAttrsSource)
	apihubAttrsListCmd.Flags().Int64Var(&flagApihubAttrsPageSize, "page-size", 0, "Maximum results per page")

	apihubAttrsCmd.AddCommand(all...)
	apihubCmd.AddCommand(apihubAttrsCmd)
}

func apihubAttrsName(id string) (string, error) {
	return apihubResource(flagApihubAttrsLocation, "attributes", id)
}

func runApihubAttrsDelete(cmd *cobra.Command, args []string) error {
	name, err := apihubAttrsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Attributes.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting attribute: %w", err)
	}
	fmt.Printf("Deleted attribute [%s].\n", args[0])
	return nil
}

func runApihubAttrsDescribe(cmd *cobra.Command, args []string) error {
	name, err := apihubAttrsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Attributes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing attribute: %w", err)
	}
	return emitFormatted(got, flagApihubAttrsFormat)
}

func runApihubAttrsExport(cmd *cobra.Command, args []string) error {
	name, err := apihubAttrsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Attributes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting attribute: %w", err)
	}
	return saveAsYAML(flagApihubAttrsDestination, got)
}

func runApihubAttrsImport(cmd *cobra.Command, args []string) error {
	parent, err := apihubLocationParent(flagApihubAttrsLocation)
	if err != nil {
		return err
	}
	name, err := apihubAttrsName(args[0])
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1Attribute{}
	if err := loadYAMLOrJSONInto(flagApihubAttrsSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Attributes.Get(name).Context(ctx).Do(); err != nil {
		got, err := svc.Projects.Locations.Attributes.Create(parent, body).AttributeId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating attribute: %w", err)
		}
		return emitFormatted(got, flagApihubAttrsFormat)
	}
	got, err := svc.Projects.Locations.Attributes.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating attribute: %w", err)
	}
	return emitFormatted(got, flagApihubAttrsFormat)
}

func runApihubAttrsList(cmd *cobra.Command, args []string) error {
	parent, err := apihubLocationParent(flagApihubAttrsLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1Attribute
	pageToken := ""
	for {
		call := svc.Projects.Locations.Attributes.List(parent).Context(ctx)
		if flagApihubAttrsPageSize > 0 {
			call = call.PageSize(flagApihubAttrsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing attributes: %w", err)
		}
		all = append(all, resp.Attributes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagApihubAttrsFormat)
}
