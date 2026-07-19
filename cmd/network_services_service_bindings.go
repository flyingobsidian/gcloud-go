package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services service-bindings (#1002) ---

var networkServicesServiceBindingsCmd = &cobra.Command{Use: "service-bindings", Short: "Manage service bindings"}

var (
	flagNsSbLocation    string
	flagNsSbFormat      string
	flagNsSbDestination string
	flagNsSbSource      string
	flagNsSbConfigFile  string
	flagNsSbUpdateMask  string
	flagNsSbPageSize    int64
)

var (
	networkServicesSbCreateCmd = &cobra.Command{
		Use: "create BINDING", Short: "Create a service binding",
		Args: cobra.ExactArgs(1), RunE: runNsSbCreate,
	}
	networkServicesSbDeleteCmd = &cobra.Command{
		Use: "delete BINDING", Short: "Delete a service binding",
		Args: cobra.ExactArgs(1), RunE: runNsSbDelete,
	}
	networkServicesSbDescribeCmd = &cobra.Command{
		Use: "describe BINDING", Short: "Describe a service binding",
		Args: cobra.ExactArgs(1), RunE: runNsSbDescribe,
	}
	networkServicesSbExportCmd = &cobra.Command{
		Use: "export BINDING", Short: "Export a service binding to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsSbExport,
	}
	networkServicesSbImportCmd = &cobra.Command{
		Use: "import BINDING", Short: "Import a service binding from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsSbImport,
	}
	networkServicesSbListCmd = &cobra.Command{
		Use: "list", Short: "List service bindings in a location",
		Args: cobra.NoArgs, RunE: runNsSbList,
	}
	networkServicesSbUpdateCmd = &cobra.Command{
		Use: "update BINDING", Short: "Update a service binding",
		Args: cobra.ExactArgs(1), RunE: runNsSbUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		networkServicesSbCreateCmd, networkServicesSbDeleteCmd,
		networkServicesSbDescribeCmd, networkServicesSbExportCmd,
		networkServicesSbImportCmd, networkServicesSbListCmd,
		networkServicesSbUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagNsSbLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNsSbFormat, "format", "", "Output format")
	}
	networkServicesSbCreateCmd.Flags().StringVar(&flagNsSbConfigFile, "config-file", "", "YAML/JSON file with the ServiceBinding body (required)")
	_ = networkServicesSbCreateCmd.MarkFlagRequired("config-file")
	nsBindExportFlags(networkServicesSbExportCmd, &flagNsSbDestination)
	nsBindImportFlags(networkServicesSbImportCmd, &flagNsSbSource)
	networkServicesSbListCmd.Flags().Int64Var(&flagNsSbPageSize, "page-size", 0, "Maximum results per page")
	networkServicesSbUpdateCmd.Flags().StringVar(&flagNsSbConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = networkServicesSbUpdateCmd.MarkFlagRequired("config-file")
	networkServicesSbUpdateCmd.Flags().StringVar(&flagNsSbUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	networkServicesServiceBindingsCmd.AddCommand(all...)
	networkServicesCmd.AddCommand(networkServicesServiceBindingsCmd)
}

func nsSbName(id string) (string, error) {
	return nsResourceName(flagNsSbLocation, "serviceBindings", id)
}

func runNsSbCreate(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsSbLocation)
	if err != nil {
		return err
	}
	body := &networkservices.ServiceBinding{}
	if err := loadYAMLOrJSONInto(flagNsSbConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ServiceBindings.Create(parent, body).ServiceBindingId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating service binding: %w", err)
	}
	fmt.Printf("Create service binding [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsSbFormat)
}

func runNsSbDelete(cmd *cobra.Command, args []string) error {
	name, err := nsSbName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ServiceBindings.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting service binding: %w", err)
	}
	fmt.Printf("Delete service binding [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsSbFormat)
}

func runNsSbDescribe(cmd *cobra.Command, args []string) error {
	name, err := nsSbName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ServiceBindings.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service binding: %w", err)
	}
	return emitFormatted(got, flagNsSbFormat)
}

func runNsSbExport(cmd *cobra.Command, args []string) error {
	name, err := nsSbName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ServiceBindings.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting service binding: %w", err)
	}
	return saveAsYAML(flagNsSbDestination, got)
}

func runNsSbImport(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsSbLocation)
	if err != nil {
		return err
	}
	name, err := nsSbName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.ServiceBinding{}
	if err := loadYAMLOrJSONInto(flagNsSbSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.ServiceBindings.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.ServiceBindings.Create(parent, body).ServiceBindingId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating service binding: %w", err)
		}
		fmt.Printf("Create service binding [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormatted(op, flagNsSbFormat)
	}
	op, err := svc.Projects.Locations.ServiceBindings.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating service binding: %w", err)
	}
	fmt.Printf("Update service binding [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsSbFormat)
}

func runNsSbList(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsSbLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkservices.ServiceBinding
	pageToken := ""
	for {
		call := svc.Projects.Locations.ServiceBindings.List(parent).Context(ctx)
		if flagNsSbPageSize > 0 {
			call = call.PageSize(flagNsSbPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing service bindings: %w", err)
		}
		all = append(all, resp.ServiceBindings...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNsSbFormat)
}

func runNsSbUpdate(cmd *cobra.Command, args []string) error {
	name, err := nsSbName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.ServiceBinding{}
	if err := loadYAMLOrJSONInto(flagNsSbConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagNsSbUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.ServiceBindings.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating service binding: %w", err)
	}
	fmt.Printf("Update service binding [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsSbFormat)
}
