package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	datamigration "google.golang.org/api/datamigration/v1"
)

var dmPrivateConnCmd = &cobra.Command{
	Use:   "private-connections",
	Short: "Manage Database Migration Service private connections",
}

var dmPCDescribeCmd = &cobra.Command{
	Use:   "describe NAME",
	Short: "Show details about a private connection",
	Args:  cobra.ExactArgs(1),
	RunE:  runDMPCDescribe,
}

var dmPCListCmd = &cobra.Command{
	Use:   "list",
	Short: "List private connections in a region",
	Args:  cobra.NoArgs,
	RunE:  runDMPCList,
}

var dmPCDeleteCmd = &cobra.Command{
	Use:   "delete NAME",
	Short: "Delete a private connection",
	Args:  cobra.ExactArgs(1),
	RunE:  runDMPCDelete,
}

var dmPCCreateCmd = &cobra.Command{
	Use:   "create NAME",
	Short: "Create a private connection from a --config-file",
	Args:  cobra.ExactArgs(1),
	RunE:  runDMPCCreate,
}

var (
	flagDMPCRegion       string
	flagDMPCFormat       string
	flagDMPCListPageSize int64
	flagDMPCListLimit    int64
	flagDMPCListFilter   string
	flagDMPCListURI      bool
	flagDMPCConfigFile   string
	flagDMPCSkipValidate bool
	flagDMPCAsync        bool
)

func init() {
	for _, c := range []*cobra.Command{dmPCDescribeCmd, dmPCListCmd, dmPCDeleteCmd, dmPCCreateCmd} {
		c.Flags().StringVar(&flagDMPCRegion, "region", "", "Region containing the resource (required)")
		_ = c.MarkFlagRequired("region")
	}
	dmPCDescribeCmd.Flags().StringVar(&flagDMPCFormat, "format", "", "Output format (yaml, json, table, ...)")

	dmPCListCmd.Flags().StringVar(&flagDMPCFormat, "format", "", "Output format (yaml, json, table, ...)")
	dmPCListCmd.Flags().Int64Var(&flagDMPCListPageSize, "page-size", 0, "Page size for API pagination")
	dmPCListCmd.Flags().Int64Var(&flagDMPCListLimit, "limit", 0, "Cap total results (0 = no cap)")
	dmPCListCmd.Flags().StringVar(&flagDMPCListFilter, "filter", "", "Server-side filter expression")
	dmPCListCmd.Flags().BoolVar(&flagDMPCListURI, "uri", false, "Print resource names only")

	dmPCCreateCmd.Flags().StringVar(&flagDMPCConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the PrivateConnection message body (required)")
	_ = dmPCCreateCmd.MarkFlagRequired("config-file")
	dmPCCreateCmd.Flags().BoolVar(&flagDMPCSkipValidate, "skip-validation", false,
		"Do not validate the private connection at creation time")
	dmPCCreateCmd.Flags().BoolVar(&flagDMPCAsync, "async", false,
		"Return the long-running operation without waiting")

	dmPCDeleteCmd.Flags().BoolVar(&flagDMPCAsync, "async", false,
		"Return the long-running operation without waiting")

	dmPrivateConnCmd.AddCommand(dmPCCreateCmd, dmPCDeleteCmd, dmPCDescribeCmd, dmPCListCmd)
	databaseMigrationCmd.AddCommand(dmPrivateConnCmd)
}

func dmPCResourceName(name, project, region string) string {
	if strings.HasPrefix(name, "projects/") {
		return name
	}
	return fmt.Sprintf("projects/%s/locations/%s/privateConnections/%s", project, region, name)
}

func runDMPCDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	pc, err := svc.Projects.Locations.PrivateConnections.Get(dmPCResourceName(args[0], project, flagDMPCRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing private connection: %w", err)
	}
	return emitFormatted(pc, flagDMPCFormat)
}

func runDMPCList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := dmParent(project, flagDMPCRegion)
	var all []*datamigration.PrivateConnection
	pageToken := ""
	for {
		call := svc.Projects.Locations.PrivateConnections.List(parent).Context(ctx)
		if flagDMPCListFilter != "" {
			call = call.Filter(flagDMPCListFilter)
		}
		if flagDMPCListPageSize > 0 {
			call = call.PageSize(flagDMPCListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing private connections: %w", err)
		}
		all = append(all, resp.PrivateConnections...)
		if flagDMPCListLimit > 0 && int64(len(all)) >= flagDMPCListLimit {
			all = all[:flagDMPCListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagDMPCListURI {
		for _, p := range all {
			fmt.Println(p.Name)
		}
		return nil
	}
	if flagDMPCFormat != "" {
		return emitFormatted(all, flagDMPCFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, p := range all {
		fmt.Printf("%-40s %s\n", path.Base(p.Name), p.State)
	}
	return nil
}

func runDMPCDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateConnections.Delete(dmPCResourceName(args[0], project, flagDMPCRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting private connection: %w", err)
	}
	if flagDMPCAsync {
		fmt.Fprintf(os.Stderr, "Delete in progress (operation: %s).\n", op.Name)
		return emitFormatted(op, "")
	}
	if _, err := waitForDMOperation(ctx, svc, op); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Deleted private connection [%s].\n", args[0])
	return nil
}

func runDMPCCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	pc := &datamigration.PrivateConnection{}
	if err := loadYAMLOrJSONInto(flagDMPCConfigFile, pc); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.PrivateConnections.Create(dmParent(project, flagDMPCRegion), pc).
		PrivateConnectionId(args[0]).Context(ctx)
	if flagDMPCSkipValidate {
		call = call.SkipValidation(true)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating private connection: %w", err)
	}
	if flagDMPCAsync {
		fmt.Fprintf(os.Stderr, "Create in progress (operation: %s).\n", op.Name)
		return emitFormatted(op, "")
	}
	op, err = waitForDMOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Created private connection [%s].\n", args[0])
	return emitFormatted(op.Response, "")
}
