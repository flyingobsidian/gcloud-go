package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable logical-views (#1307) ---

var bigtableLogicalViewsCmd = &cobra.Command{Use: "logical-views", Short: "Manage Bigtable logical views"}

var (
	flagBtLvInstance   string
	flagBtLvFormat     string
	flagBtLvConfigFile string
	flagBtLvUpdateMask string
	flagBtLvPageSize   int64
)

var (
	bigtableLvCreateCmd = &cobra.Command{
		Use: "create LOGICAL_VIEW", Short: "Create a Bigtable logical view",
		Args: cobra.ExactArgs(1), RunE: runBtLvCreate,
	}
	bigtableLvDeleteCmd = &cobra.Command{
		Use: "delete LOGICAL_VIEW", Short: "Delete a Bigtable logical view",
		Args: cobra.ExactArgs(1), RunE: runBtLvDelete,
	}
	bigtableLvDescribeCmd = &cobra.Command{
		Use: "describe LOGICAL_VIEW", Short: "Describe a Bigtable logical view",
		Args: cobra.ExactArgs(1), RunE: runBtLvDescribe,
	}
	bigtableLvListCmd = &cobra.Command{
		Use: "list", Short: "List Bigtable logical views on an instance",
		Args: cobra.NoArgs, RunE: runBtLvList,
	}
	bigtableLvUpdateCmd = &cobra.Command{
		Use: "update LOGICAL_VIEW", Short: "Update a Bigtable logical view",
		Args: cobra.ExactArgs(1), RunE: runBtLvUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		bigtableLvCreateCmd, bigtableLvDeleteCmd, bigtableLvDescribeCmd,
		bigtableLvListCmd, bigtableLvUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagBtLvInstance, "instance", "", "Bigtable instance ID (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&flagBtLvFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{bigtableLvCreateCmd, bigtableLvUpdateCmd} {
		c.Flags().StringVar(&flagBtLvConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the LogicalView body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	bigtableLvUpdateCmd.Flags().StringVar(&flagBtLvUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	bigtableLvListCmd.Flags().Int64Var(&flagBtLvPageSize, "page-size", 0, "Maximum results per page")

	bigtableLogicalViewsCmd.AddCommand(all...)
	bigtableCmd.AddCommand(bigtableLogicalViewsCmd)
}

func btLvName(id string) (string, error) {
	parent, err := btInstanceName(flagBtLvInstance)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/logicalViews/%s", parent, id), nil
}

func runBtLvCreate(cmd *cobra.Command, args []string) error {
	parent, err := btInstanceName(flagBtLvInstance)
	if err != nil {
		return err
	}
	body := &bigtableadmin.LogicalView{}
	if err := loadYAMLOrJSONInto(flagBtLvConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.LogicalViews.Create(parent, body).LogicalViewId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating logical view: %w", err)
	}
	fmt.Printf("Create request issued for logical view [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBtLvFormat)
}

func runBtLvDelete(cmd *cobra.Command, args []string) error {
	name, err := btLvName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.LogicalViews.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting logical view: %w", err)
	}
	fmt.Printf("Deleted logical view [%s].\n", args[0])
	return nil
}

func runBtLvDescribe(cmd *cobra.Command, args []string) error {
	name, err := btLvName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.LogicalViews.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing logical view: %w", err)
	}
	return emitFormatted(got, flagBtLvFormat)
}

func runBtLvList(cmd *cobra.Command, args []string) error {
	parent, err := btInstanceName(flagBtLvInstance)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*bigtableadmin.LogicalView
	pageToken := ""
	for {
		call := svc.Projects.Instances.LogicalViews.List(parent).Context(ctx)
		if flagBtLvPageSize > 0 {
			call = call.PageSize(flagBtLvPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing logical views: %w", err)
		}
		all = append(all, resp.LogicalViews...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagBtLvFormat)
}

func runBtLvUpdate(cmd *cobra.Command, args []string) error {
	name, err := btLvName(args[0])
	if err != nil {
		return err
	}
	body := &bigtableadmin.LogicalView{}
	if err := loadYAMLOrJSONInto(flagBtLvConfigFile, body); err != nil {
		return err
	}
	mask := flagBtLvUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Instances.LogicalViews.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating logical view: %w", err)
	}
	fmt.Printf("Update request issued for logical view [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBtLvFormat)
}
