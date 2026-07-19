package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkservices "google.golang.org/api/networkservices/v1"
)

// --- gcloud network-services endpoint-policies (#988) ---

var networkServicesEndpointPoliciesCmd = &cobra.Command{Use: "endpoint-policies", Short: "Manage endpoint policies"}

var (
	flagNsEpLocation    string
	flagNsEpFormat      string
	flagNsEpDestination string
	flagNsEpSource      string
	flagNsEpPageSize    int64
)

var (
	networkServicesEpDeleteCmd = &cobra.Command{
		Use: "delete POLICY", Short: "Delete an endpoint policy",
		Args: cobra.ExactArgs(1), RunE: runNsEpDelete,
	}
	networkServicesEpDescribeCmd = &cobra.Command{
		Use: "describe POLICY", Short: "Describe an endpoint policy",
		Args: cobra.ExactArgs(1), RunE: runNsEpDescribe,
	}
	networkServicesEpExportCmd = &cobra.Command{
		Use: "export POLICY", Short: "Export an endpoint policy to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsEpExport,
	}
	networkServicesEpImportCmd = &cobra.Command{
		Use: "import POLICY", Short: "Import an endpoint policy from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsEpImport,
	}
	networkServicesEpListCmd = &cobra.Command{
		Use: "list", Short: "List endpoint policies in a location",
		Args: cobra.NoArgs, RunE: runNsEpList,
	}
)

func init() {
	all := []*cobra.Command{
		networkServicesEpDeleteCmd, networkServicesEpDescribeCmd,
		networkServicesEpExportCmd, networkServicesEpImportCmd,
		networkServicesEpListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagNsEpLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNsEpFormat, "format", "", "Output format")
	}
	nsBindExportFlags(networkServicesEpExportCmd, &flagNsEpDestination)
	nsBindImportFlags(networkServicesEpImportCmd, &flagNsEpSource)
	networkServicesEpListCmd.Flags().Int64Var(&flagNsEpPageSize, "page-size", 0, "Maximum results per page")

	networkServicesEndpointPoliciesCmd.AddCommand(all...)
	networkServicesCmd.AddCommand(networkServicesEndpointPoliciesCmd)
}

func nsEpName(id string) (string, error) {
	return nsResourceName(flagNsEpLocation, "endpointPolicies", id)
}

func runNsEpDelete(cmd *cobra.Command, args []string) error {
	name, err := nsEpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.EndpointPolicies.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting endpoint policy: %w", err)
	}
	fmt.Printf("Delete endpoint policy [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsEpFormat)
}

func runNsEpDescribe(cmd *cobra.Command, args []string) error {
	name, err := nsEpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.EndpointPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing endpoint policy: %w", err)
	}
	return emitFormatted(got, flagNsEpFormat)
}

func runNsEpExport(cmd *cobra.Command, args []string) error {
	name, err := nsEpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.EndpointPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting endpoint policy: %w", err)
	}
	return saveAsYAML(flagNsEpDestination, got)
}

func runNsEpImport(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsEpLocation)
	if err != nil {
		return err
	}
	name, err := nsEpName(args[0])
	if err != nil {
		return err
	}
	body := &networkservices.EndpointPolicy{}
	if err := loadYAMLOrJSONInto(flagNsEpSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.EndpointPolicies.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.EndpointPolicies.Create(parent, body).EndpointPolicyId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating endpoint policy: %w", err)
		}
		fmt.Printf("Create endpoint policy [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormatted(op, flagNsEpFormat)
	}
	op, err := svc.Projects.Locations.EndpointPolicies.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating endpoint policy: %w", err)
	}
	fmt.Printf("Update endpoint policy [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsEpFormat)
}

func runNsEpList(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsEpLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkservices.EndpointPolicy
	pageToken := ""
	for {
		call := svc.Projects.Locations.EndpointPolicies.List(parent).Context(ctx)
		if flagNsEpPageSize > 0 {
			call = call.PageSize(flagNsEpPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing endpoint policies: %w", err)
		}
		all = append(all, resp.EndpointPolicies...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNsEpFormat)
}
