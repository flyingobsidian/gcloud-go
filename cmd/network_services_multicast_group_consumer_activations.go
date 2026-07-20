package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services multicast-group-consumer-activations (#995) ---

var networkServicesMulticastGroupConsumerActivationsCmd = &cobra.Command{Use: "multicast-group-consumer-activations", Short: "Manage multicast group consumer activations"}

var (
	flagNsMgcaLocation    string
	flagNsMgcaFormat      string
	flagNsMgcaDestination string
	flagNsMgcaSource      string
	flagNsMgcaPageSize    int64
)

var (
	networkServicesMgcaDeleteCmd = &cobra.Command{
		Use: "delete ACTIVATION", Short: "Delete a multicast group consumer activation",
		Args: cobra.ExactArgs(1), RunE: runNsMgcaDelete,
	}
	networkServicesMgcaDescribeCmd = &cobra.Command{
		Use: "describe ACTIVATION", Short: "Describe a multicast group consumer activation",
		Args: cobra.ExactArgs(1), RunE: runNsMgcaDescribe,
	}
	networkServicesMgcaExportCmd = &cobra.Command{
		Use: "export ACTIVATION", Short: "Export a multicast group consumer activation to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsMgcaExport,
	}
	networkServicesMgcaImportCmd = &cobra.Command{
		Use: "import ACTIVATION", Short: "Import a multicast group consumer activation from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsMgcaImport,
	}
	networkServicesMgcaListCmd = &cobra.Command{
		Use: "list", Short: "List multicast group consumer activations in a location",
		Args: cobra.NoArgs, RunE: runNsMgcaList,
	}
)

func init() {
	all := []*cobra.Command{
		networkServicesMgcaDeleteCmd, networkServicesMgcaDescribeCmd,
		networkServicesMgcaExportCmd, networkServicesMgcaImportCmd,
		networkServicesMgcaListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagNsMgcaLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNsMgcaFormat, "format", "", "Output format")
	}
	nsBindExportFlags(networkServicesMgcaExportCmd, &flagNsMgcaDestination)
	nsBindImportFlags(networkServicesMgcaImportCmd, &flagNsMgcaSource)
	networkServicesMgcaListCmd.Flags().Int64Var(&flagNsMgcaPageSize, "page-size", 0, "Maximum results per page")

	networkServicesMulticastGroupConsumerActivationsCmd.AddCommand(all...)
	networkServicesCmd.AddCommand(networkServicesMulticastGroupConsumerActivationsCmd)
}

func nsMgcaName(id string) (string, error) {
	return nsResourceName(flagNsMgcaLocation, "multicastGroupConsumerActivations", id)
}

func runNsMgcaDelete(cmd *cobra.Command, args []string) error {
	name, err := nsMgcaName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.MulticastGroupConsumerActivations.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting multicast group consumer activation: %w", err)
	}
	fmt.Printf("Delete multicast group consumer activation [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsMgcaFormat)
}

func runNsMgcaDescribe(cmd *cobra.Command, args []string) error {
	name, err := nsMgcaName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.MulticastGroupConsumerActivations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing multicast group consumer activation: %w", err)
	}
	return emitFormatted(got, flagNsMgcaFormat)
}

func runNsMgcaExport(cmd *cobra.Command, args []string) error {
	name, err := nsMgcaName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.MulticastGroupConsumerActivations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting multicast group consumer activation: %w", err)
	}
	return saveAsYAML(flagNsMgcaDestination, got)
}

func runNsMgcaImport(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsMgcaLocation)
	if err != nil {
		return err
	}
	name, err := nsMgcaName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.MulticastGroupConsumerActivation{}
	if err := loadYAMLOrJSONInto(flagNsMgcaSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.MulticastGroupConsumerActivations.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.MulticastGroupConsumerActivations.Create(parent, body).MulticastGroupConsumerActivationId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating multicast group consumer activation: %w", err)
		}
		fmt.Printf("Create multicast group consumer activation [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormatted(op, flagNsMgcaFormat)
	}
	op, err := svc.Projects.Locations.MulticastGroupConsumerActivations.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating multicast group consumer activation: %w", err)
	}
	fmt.Printf("Update multicast group consumer activation [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsMgcaFormat)
}

func runNsMgcaList(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsMgcaLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkservices.MulticastGroupConsumerActivation
	pageToken := ""
	for {
		call := svc.Projects.Locations.MulticastGroupConsumerActivations.List(parent).Context(ctx)
		if flagNsMgcaPageSize > 0 {
			call = call.PageSize(flagNsMgcaPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing multicast group consumer activations: %w", err)
		}
		all = append(all, resp.MulticastGroupConsumerActivations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNsMgcaFormat)
}
