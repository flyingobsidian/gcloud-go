package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	workstations "google.golang.org/api/workstations/v1"
)

// --- gcloud workstations clusters (#1259) ---

var workstationsClustersCmd = &cobra.Command{Use: "clusters", Short: "Manage workstation clusters"}

var (
	flagWsClLocation   string
	flagWsClFormat     string
	flagWsClConfigFile string
	flagWsClUpdateMask string
	flagWsClPageSize   int64
)

var (
	workstationsClCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a workstation cluster",
		Args: cobra.ExactArgs(1), RunE: runWsClCreate,
	}
	workstationsClDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a workstation cluster",
		Args: cobra.ExactArgs(1), RunE: runWsClDelete,
	}
	workstationsClDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a workstation cluster",
		Args: cobra.ExactArgs(1), RunE: runWsClDescribe,
	}
	workstationsClListCmd = &cobra.Command{
		Use: "list", Short: "List workstation clusters in a location",
		Args: cobra.NoArgs, RunE: runWsClList,
	}
	workstationsClUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update a workstation cluster",
		Args: cobra.ExactArgs(1), RunE: runWsClUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		workstationsClCreateCmd, workstationsClDeleteCmd,
		workstationsClDescribeCmd, workstationsClListCmd,
		workstationsClUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagWsClLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagWsClFormat, "format", "", "Output format")
	}
	workstationsClCreateCmd.Flags().StringVar(&flagWsClConfigFile, "config-file", "", "YAML/JSON file with the WorkstationCluster body (required)")
	_ = workstationsClCreateCmd.MarkFlagRequired("config-file")
	workstationsClListCmd.Flags().Int64Var(&flagWsClPageSize, "page-size", 0, "Maximum results per page")
	workstationsClUpdateCmd.Flags().StringVar(&flagWsClConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = workstationsClUpdateCmd.MarkFlagRequired("config-file")
	workstationsClUpdateCmd.Flags().StringVar(&flagWsClUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	workstationsClustersCmd.AddCommand(all...)
	workstationsCmd.AddCommand(workstationsClustersCmd)
}

func runWsClCreate(cmd *cobra.Command, args []string) error {
	parent, err := wsLocationParent(flagWsClLocation)
	if err != nil {
		return err
	}
	body := &workstations.WorkstationCluster{}
	if err := loadYAMLOrJSONInto(flagWsClConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.WorkstationsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WorkstationClusters.Create(parent, body).WorkstationClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating workstation cluster: %w", err)
	}
	fmt.Printf("Create workstation cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagWsClFormat)
}

func runWsClDelete(cmd *cobra.Command, args []string) error {
	name, err := wsClusterName(flagWsClLocation, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.WorkstationsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WorkstationClusters.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting workstation cluster: %w", err)
	}
	fmt.Printf("Delete workstation cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagWsClFormat)
}

func runWsClDescribe(cmd *cobra.Command, args []string) error {
	name, err := wsClusterName(flagWsClLocation, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.WorkstationsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.WorkstationClusters.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing workstation cluster: %w", err)
	}
	return emitFormatted(got, flagWsClFormat)
}

func runWsClList(cmd *cobra.Command, args []string) error {
	parent, err := wsLocationParent(flagWsClLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.WorkstationsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*workstations.WorkstationCluster
	pageToken := ""
	for {
		call := svc.Projects.Locations.WorkstationClusters.List(parent).Context(ctx)
		if flagWsClPageSize > 0 {
			call = call.PageSize(flagWsClPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing workstation clusters: %w", err)
		}
		all = append(all, resp.WorkstationClusters...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagWsClFormat)
}

func runWsClUpdate(cmd *cobra.Command, args []string) error {
	name, err := wsClusterName(flagWsClLocation, args[0])
	if err != nil {
		return err
	}
	body := &workstations.WorkstationCluster{}
	if err := loadYAMLOrJSONInto(flagWsClConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagWsClUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.WorkstationsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.WorkstationClusters.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating workstation cluster: %w", err)
	}
	fmt.Printf("Update workstation cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagWsClFormat)
}
