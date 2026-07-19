package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable authorized-views (#1302) ---

var bigtableAuthorizedViewsCmd = &cobra.Command{Use: "authorized-views", Short: "Manage Bigtable authorized views"}

var (
	flagBtAvInstance   string
	flagBtAvTable      string
	flagBtAvFormat     string
	flagBtAvConfigFile string
	flagBtAvUpdateMask string
	flagBtAvPageSize   int64
)

var (
	bigtableAvCreateCmd = &cobra.Command{
		Use: "create AUTHORIZED_VIEW", Short: "Create a Bigtable authorized view",
		Args: cobra.ExactArgs(1), RunE: runBtAvCreate,
	}
	bigtableAvDeleteCmd = &cobra.Command{
		Use: "delete AUTHORIZED_VIEW", Short: "Delete a Bigtable authorized view",
		Args: cobra.ExactArgs(1), RunE: runBtAvDelete,
	}
	bigtableAvDescribeCmd = &cobra.Command{
		Use: "describe AUTHORIZED_VIEW", Short: "Describe a Bigtable authorized view",
		Args: cobra.ExactArgs(1), RunE: runBtAvDescribe,
	}
	bigtableAvListCmd = &cobra.Command{
		Use: "list", Short: "List Bigtable authorized views on a table",
		Args: cobra.NoArgs, RunE: runBtAvList,
	}
	bigtableAvUpdateCmd = &cobra.Command{
		Use: "update AUTHORIZED_VIEW", Short: "Update a Bigtable authorized view",
		Args: cobra.ExactArgs(1), RunE: runBtAvUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		bigtableAvCreateCmd, bigtableAvDeleteCmd, bigtableAvDescribeCmd,
		bigtableAvListCmd, bigtableAvUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagBtAvInstance, "instance", "", "Bigtable instance ID (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&flagBtAvTable, "table", "", "Bigtable table ID (required)")
		_ = c.MarkFlagRequired("table")
		c.Flags().StringVar(&flagBtAvFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{bigtableAvCreateCmd, bigtableAvUpdateCmd} {
		c.Flags().StringVar(&flagBtAvConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the AuthorizedView body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	bigtableAvUpdateCmd.Flags().StringVar(&flagBtAvUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	bigtableAvListCmd.Flags().Int64Var(&flagBtAvPageSize, "page-size", 0, "Maximum results per page")

	bigtableAuthorizedViewsCmd.AddCommand(all...)
	bigtableCmd.AddCommand(bigtableAuthorizedViewsCmd)
}

func btAvName(id string) (string, error) {
	parent, err := btTableName(flagBtAvInstance, flagBtAvTable)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/authorizedViews/%s", parent, id), nil
}

func runBtAvCreate(cmd *cobra.Command, args []string) error {
	parent, err := btTableName(flagBtAvInstance, flagBtAvTable)
	if err != nil {
		return err
	}
	body := &bigtableadmin.AuthorizedView{}
	if err := loadYAMLOrJSONInto(flagBtAvConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Tables.AuthorizedViews.Create(parent, body).AuthorizedViewId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating authorized view: %w", err)
	}
	fmt.Printf("Create request issued for authorized view [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBtAvFormat)
}

func runBtAvDelete(cmd *cobra.Command, args []string) error {
	name, err := btAvName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Tables.AuthorizedViews.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting authorized view: %w", err)
	}
	fmt.Printf("Deleted authorized view [%s].\n", args[0])
	return nil
}

func runBtAvDescribe(cmd *cobra.Command, args []string) error {
	name, err := btAvName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Tables.AuthorizedViews.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing authorized view: %w", err)
	}
	return emitFormatted(got, flagBtAvFormat)
}

func runBtAvList(cmd *cobra.Command, args []string) error {
	parent, err := btTableName(flagBtAvInstance, flagBtAvTable)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*bigtableadmin.AuthorizedView
	pageToken := ""
	for {
		call := svc.Projects.Instances.Tables.AuthorizedViews.List(parent).Context(ctx)
		if flagBtAvPageSize > 0 {
			call = call.PageSize(flagBtAvPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing authorized views: %w", err)
		}
		all = append(all, resp.AuthorizedViews...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagBtAvFormat)
}

func runBtAvUpdate(cmd *cobra.Command, args []string) error {
	name, err := btAvName(args[0])
	if err != nil {
		return err
	}
	body := &bigtableadmin.AuthorizedView{}
	if err := loadYAMLOrJSONInto(flagBtAvConfigFile, body); err != nil {
		return err
	}
	mask := flagBtAvUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Instances.Tables.AuthorizedViews.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating authorized view: %w", err)
	}
	fmt.Printf("Update request issued for authorized view [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBtAvFormat)
}
