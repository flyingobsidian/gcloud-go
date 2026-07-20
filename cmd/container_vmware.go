package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	gkeonprem "google.golang.org/api/gkeonprem/v1"
)

// --- gcloud container vmware (#1143) ---
//
// Manages Anthos-on-VMware user clusters via the gkeonprem v1 client.

var containerVmwareCmd = &cobra.Command{Use: "vmware", Short: "Manage Anthos on VMware"}
var containerVmwareClustersCmd = &cobra.Command{Use: "clusters", Short: "Manage VMware user clusters"}

var (
	flagCvmLocation   string
	flagCvmFormat     string
	flagCvmConfigFile string
	flagCvmUpdateMask string
	flagCvmPageSize   int64
)

var (
	containerVmClCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a VMware user cluster (loads body from --config-file)",
		Args: cobra.ExactArgs(1), RunE: runCvmClCreate,
	}
	containerVmClDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a VMware user cluster",
		Args: cobra.ExactArgs(1), RunE: runCvmClDelete,
	}
	containerVmClDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a VMware user cluster",
		Args: cobra.ExactArgs(1), RunE: runCvmClDescribe,
	}
	containerVmClListCmd = &cobra.Command{
		Use: "list", Short: "List VMware user clusters in a location",
		Args: cobra.NoArgs, RunE: runCvmClList,
	}
	containerVmClUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update a VMware user cluster (loads body from --config-file)",
		Args: cobra.ExactArgs(1), RunE: runCvmClUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		containerVmClCreateCmd, containerVmClDeleteCmd, containerVmClDescribeCmd,
		containerVmClListCmd, containerVmClUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagCvmLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagCvmFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{containerVmClCreateCmd, containerVmClUpdateCmd} {
		c.Flags().StringVar(&flagCvmConfigFile, "config-file", "", "YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	containerVmClUpdateCmd.Flags().StringVar(&flagCvmUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	containerVmClListCmd.Flags().Int64Var(&flagCvmPageSize, "page-size", 0, "Maximum results per page")

	containerVmwareClustersCmd.AddCommand(all...)
	containerVmwareCmd.AddCommand(containerVmwareClustersCmd)
	containerCmd.AddCommand(containerVmwareCmd)
}

func cvmLocationParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if flagCvmLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagCvmLocation), nil
}

func cvmClusterName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := cvmLocationParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/vmwareClusters/%s", parent, id), nil
}

func runCvmClCreate(cmd *cobra.Command, args []string) error {
	parent, err := cvmLocationParent()
	if err != nil {
		return err
	}
	body := &gkeonprem.VmwareCluster{}
	if err := loadYAMLOrJSONInto(flagCvmConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.GKEOnPremService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.VmwareClusters.Create(parent, body).VmwareClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating VMware cluster: %w", err)
	}
	fmt.Printf("Create VMware cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCvmFormat)
}

func runCvmClDelete(cmd *cobra.Command, args []string) error {
	name, err := cvmClusterName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.GKEOnPremService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.VmwareClusters.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting VMware cluster: %w", err)
	}
	fmt.Printf("Delete VMware cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCvmFormat)
}

func runCvmClDescribe(cmd *cobra.Command, args []string) error {
	name, err := cvmClusterName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.GKEOnPremService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.VmwareClusters.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing VMware cluster: %w", err)
	}
	return emitFormatted(got, flagCvmFormat)
}

func runCvmClList(cmd *cobra.Command, args []string) error {
	parent, err := cvmLocationParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.GKEOnPremService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*gkeonprem.VmwareCluster
	pageToken := ""
	for {
		call := svc.Projects.Locations.VmwareClusters.List(parent).Context(ctx)
		if flagCvmPageSize > 0 {
			call = call.PageSize(flagCvmPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing VMware clusters: %w", err)
		}
		all = append(all, resp.VmwareClusters...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagCvmFormat)
}

func runCvmClUpdate(cmd *cobra.Command, args []string) error {
	name, err := cvmClusterName(args[0])
	if err != nil {
		return err
	}
	body := &gkeonprem.VmwareCluster{}
	if err := loadYAMLOrJSONInto(flagCvmConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagCvmUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.GKEOnPremService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.VmwareClusters.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating VMware cluster: %w", err)
	}
	fmt.Printf("Update VMware cluster [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCvmFormat)
}
