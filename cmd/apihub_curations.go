package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub curations (#1157) ---

var apihubCurCmd = &cobra.Command{Use: "curations", Short: "Manage API Hub curations"}

var (
	flagApihubCurLocation    string
	flagApihubCurFormat      string
	flagApihubCurDestination string
	flagApihubCurSource      string
	flagApihubCurPageSize    int64
)

var (
	apihubCurDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete a curation",
		Args: cobra.ExactArgs(1), RunE: runApihubCurDelete,
	}
	apihubCurDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe a curation",
		Args: cobra.ExactArgs(1), RunE: runApihubCurDescribe,
	}
	apihubCurExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export a curation to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runApihubCurExport,
	}
	apihubCurImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import a curation from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runApihubCurImport,
	}
	apihubCurListCmd = &cobra.Command{
		Use: "list", Short: "List curations in a location",
		Args: cobra.NoArgs, RunE: runApihubCurList,
	}
)

func init() {
	all := []*cobra.Command{
		apihubCurDeleteCmd, apihubCurDescribeCmd,
		apihubCurExportCmd, apihubCurImportCmd,
		apihubCurListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagApihubCurLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagApihubCurFormat, "format", "", "Output format")
	}
	nsBindExportFlags(apihubCurExportCmd, &flagApihubCurDestination)
	nsBindImportFlags(apihubCurImportCmd, &flagApihubCurSource)
	apihubCurListCmd.Flags().Int64Var(&flagApihubCurPageSize, "page-size", 0, "Maximum results per page")

	apihubCurCmd.AddCommand(all...)
	apihubCmd.AddCommand(apihubCurCmd)
}

func apihubCurName(id string) (string, error) {
	return apihubResource(flagApihubCurLocation, "curations", id)
}

func runApihubCurDelete(cmd *cobra.Command, args []string) error {
	name, err := apihubCurName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Curations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting curation: %w", err)
	}
	fmt.Printf("Deleted curation [%s].\n", args[0])
	return nil
}

func runApihubCurDescribe(cmd *cobra.Command, args []string) error {
	name, err := apihubCurName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Curations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing curation: %w", err)
	}
	return emitFormatted(got, flagApihubCurFormat)
}

func runApihubCurExport(cmd *cobra.Command, args []string) error {
	name, err := apihubCurName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Curations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting curation: %w", err)
	}
	return saveAsYAML(flagApihubCurDestination, got)
}

func runApihubCurImport(cmd *cobra.Command, args []string) error {
	parent, err := apihubLocationParent(flagApihubCurLocation)
	if err != nil {
		return err
	}
	name, err := apihubCurName(args[0])
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1Curation{}
	if err := loadYAMLOrJSONInto(flagApihubCurSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Curations.Get(name).Context(ctx).Do(); err != nil {
		got, err := svc.Projects.Locations.Curations.Create(parent, body).CurationId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating curation: %w", err)
		}
		return emitFormatted(got, flagApihubCurFormat)
	}
	got, err := svc.Projects.Locations.Curations.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating curation: %w", err)
	}
	return emitFormatted(got, flagApihubCurFormat)
}

func runApihubCurList(cmd *cobra.Command, args []string) error {
	parent, err := apihubLocationParent(flagApihubCurLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1Curation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Curations.List(parent).Context(ctx)
		if flagApihubCurPageSize > 0 {
			call = call.PageSize(flagApihubCurPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing curations: %w", err)
		}
		all = append(all, resp.Curations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagApihubCurFormat)
}
