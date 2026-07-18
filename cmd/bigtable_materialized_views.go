package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable materialized-views (#1487) ---

var bigtableMVCmd = &cobra.Command{Use: "materialized-views", Short: "Manage Cloud Bigtable materialized views"}

var (
	flagBTMVInstance   string
	flagBTMVFormat     string
	flagBTMVConfigFile string
	flagBTMVUpdateMask string
	flagBTMVPageSize   int64
)

var (
	bigtableMVCreateCmd = &cobra.Command{
		Use: "create MATERIALIZED_VIEW", Short: "Create a materialized view",
		Args: cobra.ExactArgs(1), RunE: runBTMVCreate,
	}
	bigtableMVDeleteCmd = &cobra.Command{
		Use: "delete MATERIALIZED_VIEW", Short: "Delete a materialized view",
		Args: cobra.ExactArgs(1), RunE: runBTMVDelete,
	}
	bigtableMVDescribeCmd = &cobra.Command{
		Use: "describe MATERIALIZED_VIEW", Short: "Describe a materialized view",
		Args: cobra.ExactArgs(1), RunE: runBTMVDescribe,
	}
	bigtableMVListCmd = &cobra.Command{
		Use: "list", Short: "List materialized views in an instance",
		Args: cobra.NoArgs, RunE: runBTMVList,
	}
	bigtableMVUpdateCmd = &cobra.Command{
		Use: "update MATERIALIZED_VIEW", Short: "Update a materialized view",
		Args: cobra.ExactArgs(1), RunE: runBTMVUpdate,
	}
)

func init() {
	all := []*cobra.Command{bigtableMVCreateCmd, bigtableMVDeleteCmd, bigtableMVDescribeCmd, bigtableMVListCmd, bigtableMVUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagBTMVInstance, "instance", "", "Bigtable instance ID (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&flagBTMVFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{bigtableMVCreateCmd, bigtableMVUpdateCmd} {
		c.Flags().StringVar(&flagBTMVConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the MaterializedView body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	bigtableMVUpdateCmd.Flags().StringVar(&flagBTMVUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	bigtableMVListCmd.Flags().Int64Var(&flagBTMVPageSize, "page-size", 0, "Maximum results per page")

	bigtableMVCmd.AddCommand(all...)
	bigtableCmd.AddCommand(bigtableMVCmd)
}

func btMVParent() (string, error) {
	return bigtableInstanceParent(flagBTMVInstance)
}

func btMVName(id string) (string, error) {
	parent, err := btMVParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/materializedViews/%s", parent, id), nil
}

func runBTMVCreate(cmd *cobra.Command, args []string) error {
	parent, err := btMVParent()
	if err != nil {
		return err
	}
	body := &bigtableadmin.MaterializedView{}
	if err := loadYAMLOrJSONInto(flagBTMVConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.MaterializedViews.Create(parent, body).MaterializedViewId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating materialized view: %w", err)
	}
	fmt.Printf("Create request issued for materialized view [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBTMVFormat)
}

func runBTMVDelete(cmd *cobra.Command, args []string) error {
	name, err := btMVName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.MaterializedViews.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting materialized view: %w", err)
	}
	fmt.Printf("Deleted materialized view [%s].\n", args[0])
	return nil
}

func runBTMVDescribe(cmd *cobra.Command, args []string) error {
	name, err := btMVName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.MaterializedViews.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing materialized view: %w", err)
	}
	return emitFormatted(got, flagBTMVFormat)
}

func runBTMVList(cmd *cobra.Command, args []string) error {
	parent, err := btMVParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*bigtableadmin.MaterializedView
	pageToken := ""
	for {
		call := svc.Projects.Instances.MaterializedViews.List(parent).Context(ctx)
		if flagBTMVPageSize > 0 {
			call = call.PageSize(flagBTMVPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing materialized views: %w", err)
		}
		all = append(all, resp.MaterializedViews...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagBTMVFormat)
}

func runBTMVUpdate(cmd *cobra.Command, args []string) error {
	name, err := btMVName(args[0])
	if err != nil {
		return err
	}
	body := &bigtableadmin.MaterializedView{}
	if err := loadYAMLOrJSONInto(flagBTMVConfigFile, body); err != nil {
		return err
	}
	mask := flagBTMVUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Instances.MaterializedViews.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating materialized view: %w", err)
	}
	fmt.Printf("Update request issued for materialized view [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBTMVFormat)
}
