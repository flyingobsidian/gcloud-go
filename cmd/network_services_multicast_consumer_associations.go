package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services multicast-consumer-associations (#993) ---

var networkServicesMulticastConsumerAssociationsCmd = &cobra.Command{Use: "multicast-consumer-associations", Short: "Manage multicast consumer associations"}

var (
	flagNsMcaLocation    string
	flagNsMcaFormat      string
	flagNsMcaDestination string
	flagNsMcaSource      string
	flagNsMcaPageSize    int64
)

var (
	networkServicesMcaDeleteCmd = &cobra.Command{
		Use: "delete ASSOCIATION", Short: "Delete a multicast consumer association",
		Args: cobra.ExactArgs(1), RunE: runNsMcaDelete,
	}
	networkServicesMcaDescribeCmd = &cobra.Command{
		Use: "describe ASSOCIATION", Short: "Describe a multicast consumer association",
		Args: cobra.ExactArgs(1), RunE: runNsMcaDescribe,
	}
	networkServicesMcaExportCmd = &cobra.Command{
		Use: "export ASSOCIATION", Short: "Export a multicast consumer association to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsMcaExport,
	}
	networkServicesMcaImportCmd = &cobra.Command{
		Use: "import ASSOCIATION", Short: "Import a multicast consumer association from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsMcaImport,
	}
	networkServicesMcaListCmd = &cobra.Command{
		Use: "list", Short: "List multicast consumer associations in a location",
		Args: cobra.NoArgs, RunE: runNsMcaList,
	}
)

func init() {
	all := []*cobra.Command{
		networkServicesMcaDeleteCmd, networkServicesMcaDescribeCmd,
		networkServicesMcaExportCmd, networkServicesMcaImportCmd,
		networkServicesMcaListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagNsMcaLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNsMcaFormat, "format", "", "Output format")
	}
	nsBindExportFlags(networkServicesMcaExportCmd, &flagNsMcaDestination)
	nsBindImportFlags(networkServicesMcaImportCmd, &flagNsMcaSource)
	networkServicesMcaListCmd.Flags().Int64Var(&flagNsMcaPageSize, "page-size", 0, "Maximum results per page")

	networkServicesMulticastConsumerAssociationsCmd.AddCommand(all...)
	networkServicesCmd.AddCommand(networkServicesMulticastConsumerAssociationsCmd)
}

func nsMcaName(id string) (string, error) {
	return nsResourceName(flagNsMcaLocation, "multicastConsumerAssociations", id)
}

func runNsMcaDelete(cmd *cobra.Command, args []string) error {
	name, err := nsMcaName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.MulticastConsumerAssociations.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting multicast consumer association: %w", err)
	}
	fmt.Printf("Delete multicast consumer association [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsMcaFormat)
}

func runNsMcaDescribe(cmd *cobra.Command, args []string) error {
	name, err := nsMcaName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.MulticastConsumerAssociations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing multicast consumer association: %w", err)
	}
	return emitFormatted(got, flagNsMcaFormat)
}

func runNsMcaExport(cmd *cobra.Command, args []string) error {
	name, err := nsMcaName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.MulticastConsumerAssociations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting multicast consumer association: %w", err)
	}
	return saveAsYAML(flagNsMcaDestination, got)
}

func runNsMcaImport(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsMcaLocation)
	if err != nil {
		return err
	}
	name, err := nsMcaName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.MulticastConsumerAssociation{}
	if err := loadYAMLOrJSONInto(flagNsMcaSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.MulticastConsumerAssociations.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.MulticastConsumerAssociations.Create(parent, body).MulticastConsumerAssociationId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating multicast consumer association: %w", err)
		}
		fmt.Printf("Create multicast consumer association [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormatted(op, flagNsMcaFormat)
	}
	op, err := svc.Projects.Locations.MulticastConsumerAssociations.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating multicast consumer association: %w", err)
	}
	fmt.Printf("Update multicast consumer association [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsMcaFormat)
}

func runNsMcaList(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsMcaLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkservices.MulticastConsumerAssociation
	pageToken := ""
	for {
		call := svc.Projects.Locations.MulticastConsumerAssociations.List(parent).Context(ctx)
		if flagNsMcaPageSize > 0 {
			call = call.PageSize(flagNsMcaPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing multicast consumer associations: %w", err)
		}
		all = append(all, resp.MulticastConsumerAssociations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNsMcaFormat)
}
