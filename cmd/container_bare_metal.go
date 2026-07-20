package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	gkeonprem "google.golang.org/api/gkeonprem/v1"
)

// --- gcloud container bare-metal (#1134) ---
//
// Manages Anthos-on-bare-metal user clusters via the gkeonprem v1 client.

var containerBareMetalCmd = &cobra.Command{Use: "bare-metal", Short: "Manage Anthos on bare metal"}
var containerBareMetalClustersCmd = &cobra.Command{Use: "clusters", Short: "Manage bare-metal user clusters"}

var (
	flagCbmLocation   string
	flagCbmFormat     string
	flagCbmConfigFile string
	flagCbmUpdateMask string
	flagCbmPageSize   int64
)

var (
	containerBmClCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a bare-metal user cluster (loads body from --config-file)",
		Args: cobra.ExactArgs(1), RunE: runCbmClCreate,
	}
	containerBmClDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a bare-metal user cluster",
		Args: cobra.ExactArgs(1), RunE: runCbmClDelete,
	}
	containerBmClDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a bare-metal user cluster",
		Args: cobra.ExactArgs(1), RunE: runCbmClDescribe,
	}
	containerBmClListCmd = &cobra.Command{
		Use: "list", Short: "List bare-metal user clusters in a location",
		Args: cobra.NoArgs, RunE: runCbmClList,
	}
	containerBmClUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update a bare-metal user cluster (loads body from --config-file)",
		Args: cobra.ExactArgs(1), RunE: runCbmClUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		containerBmClCreateCmd, containerBmClDeleteCmd, containerBmClDescribeCmd,
		containerBmClListCmd, containerBmClUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagCbmLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagCbmFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{containerBmClCreateCmd, containerBmClUpdateCmd} {
		c.Flags().StringVar(&flagCbmConfigFile, "config-file", "", "YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	containerBmClUpdateCmd.Flags().StringVar(&flagCbmUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	containerBmClListCmd.Flags().Int64Var(&flagCbmPageSize, "page-size", 0, "Maximum results per page")

	containerBareMetalClustersCmd.AddCommand(all...)
	containerBareMetalCmd.AddCommand(containerBareMetalClustersCmd)
	containerCmd.AddCommand(containerBareMetalCmd)
}

func cbmLocationParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if flagCbmLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagCbmLocation), nil
}

func cbmClusterName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := cbmLocationParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/bareMetalClusters/%s", parent, id), nil
}

func runCbmClCreate(cmd *cobra.Command, args []string) error {
	parent, err := cbmLocationParent()
	if err != nil {
		return err
	}
	body := &gkeonprem.BareMetalCluster{}
	if err := loadYAMLOrJSONInto(flagCbmConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.GKEOnPremService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.BareMetalClusters.Create(parent, body).BareMetalClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating bare-metal cluster: %w", err)
	}
	fmt.Printf("Create bare-metal cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCbmFormat)
}

func runCbmClDelete(cmd *cobra.Command, args []string) error {
	name, err := cbmClusterName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.GKEOnPremService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.BareMetalClusters.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting bare-metal cluster: %w", err)
	}
	fmt.Printf("Delete bare-metal cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCbmFormat)
}

func runCbmClDescribe(cmd *cobra.Command, args []string) error {
	name, err := cbmClusterName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.GKEOnPremService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.BareMetalClusters.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing bare-metal cluster: %w", err)
	}
	return emitFormatted(got, flagCbmFormat)
}

func runCbmClList(cmd *cobra.Command, args []string) error {
	parent, err := cbmLocationParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.GKEOnPremService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*gkeonprem.BareMetalCluster
	pageToken := ""
	for {
		call := svc.Projects.Locations.BareMetalClusters.List(parent).Context(ctx)
		if flagCbmPageSize > 0 {
			call = call.PageSize(flagCbmPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing bare-metal clusters: %w", err)
		}
		all = append(all, resp.BareMetalClusters...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagCbmFormat)
}

func runCbmClUpdate(cmd *cobra.Command, args []string) error {
	name, err := cbmClusterName(args[0])
	if err != nil {
		return err
	}
	body := &gkeonprem.BareMetalCluster{}
	if err := loadYAMLOrJSONInto(flagCbmConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagCbmUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.GKEOnPremService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.BareMetalClusters.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating bare-metal cluster: %w", err)
	}
	fmt.Printf("Update bare-metal cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCbmFormat)
}
