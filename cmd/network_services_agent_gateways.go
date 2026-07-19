package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkservicesbeta "google.golang.org/api/networkservices/v1beta1"
)

// --- gcloud network-services agent-gateways (#1006) ---
//
// agent-gateways is only exposed on the v1beta1 surface as of this build,
// so this file uses gcp.NetworkServicesBetaService rather than the v1
// client used by its siblings.

var networkServicesAgentGatewaysCmd = &cobra.Command{Use: "agent-gateways", Short: "Manage agent gateways"}

var (
	flagNsAgLocation    string
	flagNsAgFormat      string
	flagNsAgDestination string
	flagNsAgSource      string
	flagNsAgPageSize    int64
)

var (
	networkServicesAgDeleteCmd = &cobra.Command{
		Use: "delete GATEWAY", Short: "Delete an agent gateway",
		Args: cobra.ExactArgs(1), RunE: runNsAgDelete,
	}
	networkServicesAgDescribeCmd = &cobra.Command{
		Use: "describe GATEWAY", Short: "Describe an agent gateway",
		Args: cobra.ExactArgs(1), RunE: runNsAgDescribe,
	}
	networkServicesAgExportCmd = &cobra.Command{
		Use: "export GATEWAY", Short: "Export an agent gateway to a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsAgExport,
	}
	networkServicesAgImportCmd = &cobra.Command{
		Use: "import GATEWAY", Short: "Import an agent gateway from a YAML file",
		Args: cobra.ExactArgs(1), RunE: runNsAgImport,
	}
	networkServicesAgListCmd = &cobra.Command{
		Use: "list", Short: "List agent gateways in a location",
		Args: cobra.NoArgs, RunE: runNsAgList,
	}
)

func init() {
	all := []*cobra.Command{
		networkServicesAgDeleteCmd, networkServicesAgDescribeCmd,
		networkServicesAgExportCmd, networkServicesAgImportCmd,
		networkServicesAgListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagNsAgLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNsAgFormat, "format", "", "Output format")
	}
	nsBindExportFlags(networkServicesAgExportCmd, &flagNsAgDestination)
	nsBindImportFlags(networkServicesAgImportCmd, &flagNsAgSource)
	networkServicesAgListCmd.Flags().Int64Var(&flagNsAgPageSize, "page-size", 0, "Maximum results per page")

	networkServicesAgentGatewaysCmd.AddCommand(all...)
	networkServicesCmd.AddCommand(networkServicesAgentGatewaysCmd)
}

func nsAgName(id string) (string, error) {
	return nsResourceName(flagNsAgLocation, "agentGateways", id)
}

func runNsAgDelete(cmd *cobra.Command, args []string) error {
	name, err := nsAgName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesBetaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.AgentGateways.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting agent gateway: %w", err)
	}
	fmt.Printf("Delete agent gateway [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsAgFormat)
}

func runNsAgDescribe(cmd *cobra.Command, args []string) error {
	name, err := nsAgName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesBetaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AgentGateways.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing agent gateway: %w", err)
	}
	return emitFormatted(got, flagNsAgFormat)
}

func runNsAgExport(cmd *cobra.Command, args []string) error {
	name, err := nsAgName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesBetaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AgentGateways.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting agent gateway: %w", err)
	}
	return saveAsYAML(flagNsAgDestination, got)
}

func runNsAgImport(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsAgLocation)
	if err != nil {
		return err
	}
	name, err := nsAgName(args[0])
	if err != nil {
		return err
	}
	body := &networkservicesbeta.AgentGateway{}
	if err := loadYAMLOrJSONInto(flagNsAgSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.NetworkServicesBetaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.AgentGateways.Get(name).Context(ctx).Do(); err != nil {
		op, err := svc.Projects.Locations.AgentGateways.Create(parent, body).AgentGatewayId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating agent gateway: %w", err)
		}
		fmt.Printf("Create agent gateway [%s] initiated (operation: %s).\n", args[0], op.Name)
		return emitFormatted(op, flagNsAgFormat)
	}
	op, err := svc.Projects.Locations.AgentGateways.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating agent gateway: %w", err)
	}
	fmt.Printf("Update agent gateway [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNsAgFormat)
}

func runNsAgList(cmd *cobra.Command, args []string) error {
	parent, err := nsLocationParent(flagNsAgLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkServicesBetaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networkservicesbeta.AgentGateway
	pageToken := ""
	for {
		call := svc.Projects.Locations.AgentGateways.List(parent).Context(ctx)
		if flagNsAgPageSize > 0 {
			call = call.PageSize(flagNsAgPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing agent gateways: %w", err)
		}
		all = append(all, resp.AgentGateways...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNsAgFormat)
}
