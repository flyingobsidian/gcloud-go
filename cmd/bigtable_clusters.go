package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable clusters (#1304) ---

var bigtableClustersCmd = &cobra.Command{Use: "clusters", Short: "Manage Bigtable clusters"}

var (
	flagBtClInstance   string
	flagBtClFormat     string
	flagBtClConfigFile string
)

var (
	bigtableClCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a Bigtable cluster",
		Args: cobra.ExactArgs(1), RunE: runBtClCreate,
	}
	bigtableClDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a Bigtable cluster",
		Args: cobra.ExactArgs(1), RunE: runBtClDelete,
	}
	bigtableClDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a Bigtable cluster",
		Args: cobra.ExactArgs(1), RunE: runBtClDescribe,
	}
	bigtableClListCmd = &cobra.Command{
		Use: "list", Short: "List Bigtable clusters on an instance",
		Args: cobra.NoArgs, RunE: runBtClList,
	}
	bigtableClUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update a Bigtable cluster",
		Args: cobra.ExactArgs(1), RunE: runBtClUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		bigtableClCreateCmd, bigtableClDeleteCmd, bigtableClDescribeCmd,
		bigtableClListCmd, bigtableClUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagBtClInstance, "instance", "", "Bigtable instance ID (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&flagBtClFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{bigtableClCreateCmd, bigtableClUpdateCmd} {
		c.Flags().StringVar(&flagBtClConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Cluster body (required)")
		_ = c.MarkFlagRequired("config-file")
	}

	bigtableClustersCmd.AddCommand(all...)
	bigtableCmd.AddCommand(bigtableClustersCmd)
}

func runBtClCreate(cmd *cobra.Command, args []string) error {
	parent, err := btInstanceName(flagBtClInstance)
	if err != nil {
		return err
	}
	body := &bigtableadmin.Cluster{}
	if err := loadYAMLOrJSONInto(flagBtClConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Clusters.Create(parent, body).ClusterId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	fmt.Printf("Create request issued for cluster [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBtClFormat)
}

func runBtClDelete(cmd *cobra.Command, args []string) error {
	name, err := btClusterName(flagBtClInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Clusters.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting cluster: %w", err)
	}
	fmt.Printf("Deleted cluster [%s].\n", args[0])
	return nil
}

func runBtClDescribe(cmd *cobra.Command, args []string) error {
	name, err := btClusterName(flagBtClInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Clusters.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cluster: %w", err)
	}
	return emitFormatted(got, flagBtClFormat)
}

func runBtClList(cmd *cobra.Command, args []string) error {
	parent, err := btInstanceName(flagBtClInstance)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*bigtableadmin.Cluster
	pageToken := ""
	for {
		call := svc.Projects.Instances.Clusters.List(parent).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing clusters: %w", err)
		}
		all = append(all, resp.Clusters...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagBtClFormat)
}

func runBtClUpdate(cmd *cobra.Command, args []string) error {
	name, err := btClusterName(flagBtClInstance, args[0])
	if err != nil {
		return err
	}
	body := &bigtableadmin.Cluster{}
	if err := loadYAMLOrJSONInto(flagBtClConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Clusters.Update(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating cluster: %w", err)
	}
	fmt.Printf("Update request issued for cluster [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBtClFormat)
}
