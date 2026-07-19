package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud service-extensions authz-extensions (#1042) ---

var seAuthzExtensionsCmd = &cobra.Command{Use: "authz-extensions", Short: "Manage AuthzExtension resources"}

var (
	flagSeAuthzLocation    string
	flagSeAuthzFormat      string
	flagSeAuthzDestination string
	flagSeAuthzSource      string
	flagSeAuthzPageSize    int64
)

var (
	seAuthzDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an AuthzExtension",
		Args: cobra.ExactArgs(1), RunE: runSeAuthzDelete,
	}
	seAuthzDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an AuthzExtension",
		Args: cobra.ExactArgs(1), RunE: runSeAuthzDescribe,
	}
	seAuthzExportCmd = &cobra.Command{
		Use: "export RESOURCE", Short: "Export an AuthzExtension to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runSeAuthzExport,
	}
	seAuthzImportCmd = &cobra.Command{
		Use: "import RESOURCE", Short: "Import an AuthzExtension from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runSeAuthzImport,
	}
	seAuthzListCmd = &cobra.Command{
		Use: "list", Short: "List AuthzExtensions in a location",
		Args: cobra.NoArgs, RunE: runSeAuthzList,
	}
)

func init() {
	all := []*cobra.Command{
		seAuthzDeleteCmd, seAuthzDescribeCmd,
		seAuthzExportCmd, seAuthzImportCmd,
		seAuthzListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagSeAuthzLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagSeAuthzFormat, "format", "", "Output format")
	}
	nsBindExportFlags(seAuthzExportCmd, &flagSeAuthzDestination)
	nsBindImportFlags(seAuthzImportCmd, &flagSeAuthzSource)
	seAuthzListCmd.Flags().Int64Var(&flagSeAuthzPageSize, "page-size", 0, "Maximum results per page")

	seAuthzExtensionsCmd.AddCommand(all...)
	serviceExtensionsCmd.AddCommand(seAuthzExtensionsCmd)
}

func seAuthzName(id string) (string, error) {
	return seResource(flagSeAuthzLocation, "authzExtensions", id)
}

func runSeAuthzDelete(cmd *cobra.Command, args []string) error {
	name, err := seAuthzName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AuthzExtensions.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting authz extension: %w", err)
	}
	fmt.Printf("Delete authz extension [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagSeAuthzFormat)
}

func runSeAuthzDescribe(cmd *cobra.Command, args []string) error {
	name, err := seAuthzName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AuthzExtensions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing authz extension: %w", err)
	}
	return emitFormatted(got, flagSeAuthzFormat)
}

func runSeAuthzExport(cmd *cobra.Command, args []string) error {
	name, err := seAuthzName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AuthzExtensions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting authz extension: %w", err)
	}
	return saveAsYAML(flagSeAuthzDestination, got)
}

func runSeAuthzImport(cmd *cobra.Command, args []string) error {
	parent, err := seLocationParent(flagSeAuthzLocation)
	if err != nil {
		return err
	}
	name, err := seAuthzName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.AuthzExtension{}
	if err := loadYAMLOrJSONInto(flagSeAuthzSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.AuthzExtensions.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.AuthzExtensions.Create(parent, body).AuthzExtensionId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating authz extension: %w", err)
		}
		fmt.Printf("Create authz extension [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormatted(op, flagSeAuthzFormat)
	}
	op, err := svc.Projects.Locations.AuthzExtensions.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating authz extension: %w", err)
	}
	fmt.Printf("Update authz extension [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagSeAuthzFormat)
}

func runSeAuthzList(cmd *cobra.Command, args []string) error {
	parent, err := seLocationParent(flagSeAuthzLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkservices.AuthzExtension
	pageToken := ""
	for {
		call := svc.Projects.Locations.AuthzExtensions.List(parent).Context(ctx)
		if flagSeAuthzPageSize > 0 {
			call = call.PageSize(flagSeAuthzPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing authz extensions: %w", err)
		}
		all = append(all, resp.AuthzExtensions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagSeAuthzFormat)
}
