package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services service-lb-policies (#1003) ---

var networkServicesServiceLbPoliciesCmd = &cobra.Command{Use: "service-lb-policies", Short: "Manage service LB policies"}

var (
	flagNsSlLocation    string
	flagNsSlFormat      string
	flagNsSlDestination string
	flagNsSlSource      string
	flagNsSlConfigFile  string
	flagNsSlUpdateMask  string
	flagNsSlPageSize    int64
)

var (
	networkServicesSlCreateCmd = &cobra.Command{
		Use: "create POLICY", Short: "Create a service LB policy",
		Args: cobra.ExactArgs(1), RunE: runNsSlCreate,
	}
	networkServicesSlDeleteCmd = &cobra.Command{
		Use: "delete POLICY", Short: "Delete a service LB policy",
		Args: cobra.ExactArgs(1), RunE: runNsSlDelete,
	}
	networkServicesSlDescribeCmd = &cobra.Command{
		Use: "describe POLICY", Short: "Describe a service LB policy",
		Args: cobra.ExactArgs(1), RunE: runNsSlDescribe,
	}
	networkServicesSlExportCmd = &cobra.Command{
		Use: "export POLICY", Short: "Export a service LB policy to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsSlExport,
	}
	networkServicesSlImportCmd = &cobra.Command{
		Use: "import POLICY", Short: "Import a service LB policy from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsSlImport,
	}
	networkServicesSlListCmd = &cobra.Command{
		Use: "list", Short: "List service LB policies in a location",
		Args: cobra.NoArgs, RunE: runNsSlList,
	}
	networkServicesSlUpdateCmd = &cobra.Command{
		Use: "update POLICY", Short: "Update a service LB policy",
		Args: cobra.ExactArgs(1), RunE: runNsSlUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		networkServicesSlCreateCmd, networkServicesSlDeleteCmd,
		networkServicesSlDescribeCmd, networkServicesSlExportCmd,
		networkServicesSlImportCmd, networkServicesSlListCmd,
		networkServicesSlUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagNsSlLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNsSlFormat, "format", "", "Output format")
	}
	networkServicesSlCreateCmd.Flags().StringVar(&flagNsSlConfigFile, "config-file", "", "YAML/JSON file with the ServiceLbPolicy body (required)")
	_ = networkServicesSlCreateCmd.MarkFlagRequired("config-file")
	nsBindExportFlags(networkServicesSlExportCmd, &flagNsSlDestination)
	nsBindImportFlags(networkServicesSlImportCmd, &flagNsSlSource)
	networkServicesSlListCmd.Flags().Int64Var(&flagNsSlPageSize, "page-size", 0, "Maximum results per page")
	networkServicesSlUpdateCmd.Flags().StringVar(&flagNsSlConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = networkServicesSlUpdateCmd.MarkFlagRequired("config-file")
	networkServicesSlUpdateCmd.Flags().StringVar(&flagNsSlUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	networkServicesServiceLbPoliciesCmd.AddCommand(all...)
	networkServicesCmd.AddCommand(networkServicesServiceLbPoliciesCmd)
}

func nsSlName(id string) (string, error) {
	return nsResourceName(flagNsSlLocation, "serviceLbPolicies", id)
}

func runNsSlCreate(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsSlLocation)
	if err != nil {
		return err
	}
	body := &networkservices.ServiceLbPolicy{}
	if err := loadYAMLOrJSONInto(flagNsSlConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ServiceLbPolicies.Create(parent, body).ServiceLbPolicyId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating service LB policy: %w", err)
	}
	fmt.Printf("Create service LB policy [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsSlFormat)
}

func runNsSlDelete(cmd *cobra.Command, args []string) error {
	name, err := nsSlName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ServiceLbPolicies.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting service LB policy: %w", err)
	}
	fmt.Printf("Delete service LB policy [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsSlFormat)
}

func runNsSlDescribe(cmd *cobra.Command, args []string) error {
	name, err := nsSlName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ServiceLbPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service LB policy: %w", err)
	}
	return emitFormatted(got, flagNsSlFormat)
}

func runNsSlExport(cmd *cobra.Command, args []string) error {
	name, err := nsSlName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ServiceLbPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting service LB policy: %w", err)
	}
	return saveAsYAML(flagNsSlDestination, got)
}

func runNsSlImport(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsSlLocation)
	if err != nil {
		return err
	}
	name, err := nsSlName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.ServiceLbPolicy{}
	if err := loadYAMLOrJSONInto(flagNsSlSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.ServiceLbPolicies.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.ServiceLbPolicies.Create(parent, body).ServiceLbPolicyId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating service LB policy: %w", err)
		}
		fmt.Printf("Create service LB policy [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormatted(op, flagNsSlFormat)
	}
	op, err := svc.Projects.Locations.ServiceLbPolicies.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating service LB policy: %w", err)
	}
	fmt.Printf("Update service LB policy [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsSlFormat)
}

func runNsSlList(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsSlLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkservices.ServiceLbPolicy
	pageToken := ""
	for {
		call := svc.Projects.Locations.ServiceLbPolicies.List(parent).Context(ctx)
		if flagNsSlPageSize > 0 {
			call = call.PageSize(flagNsSlPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing service LB policies: %w", err)
		}
		all = append(all, resp.ServiceLbPolicies...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNsSlFormat)
}

func runNsSlUpdate(cmd *cobra.Command, args []string) error {
	name, err := nsSlName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.ServiceLbPolicy{}
	if err := loadYAMLOrJSONInto(flagNsSlConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagNsSlUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.ServiceLbPolicies.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating service LB policy: %w", err)
	}
	fmt.Printf("Update service LB policy [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsSlFormat)
}
