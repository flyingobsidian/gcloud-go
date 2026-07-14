package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	datamigration "google.golang.org/api/datamigration/v1"
)

var dmObjectsCmd = &cobra.Command{
	Use:   "objects",
	Short: "Manage migration job objects",
}

var dmObjListCmd = &cobra.Command{
	Use:   "list",
	Short: "List objects belonging to a migration job",
	Args:  cobra.NoArgs,
	RunE:  runDMObjList,
}

var dmObjLookupCmd = &cobra.Command{
	Use:   "lookup",
	Short: "Look up a migration job object by its source identifier",
	Args:  cobra.NoArgs,
	RunE:  runDMObjLookup,
}

var (
	flagDMObjRegion         string
	flagDMObjMigrationJob   string
	flagDMObjFormat         string
	flagDMObjListPageSize   int64
	flagDMObjListLimit      int64
	flagDMObjListFilter     string
	flagDMObjListURI        bool
	flagDMObjLookupDatabase string
	flagDMObjLookupSchema   string
	flagDMObjLookupTable    string
	flagDMObjLookupType     string
)

func init() {
	for _, c := range []*cobra.Command{dmObjListCmd, dmObjLookupCmd} {
		c.Flags().StringVar(&flagDMObjRegion, "region", "", "Region containing the migration job (required)")
		c.Flags().StringVar(&flagDMObjMigrationJob, "migration-job", "", "Name of the parent migration job (required)")
		_ = c.MarkFlagRequired("region")
		_ = c.MarkFlagRequired("migration-job")
	}

	dmObjListCmd.Flags().StringVar(&flagDMObjFormat, "format", "", "Output format")
	dmObjListCmd.Flags().Int64Var(&flagDMObjListPageSize, "page-size", 0, "Page size for API pagination")
	dmObjListCmd.Flags().Int64Var(&flagDMObjListLimit, "limit", 0, "Cap total results (0 = no cap)")
	dmObjListCmd.Flags().StringVar(&flagDMObjListFilter, "filter", "", "Server-side filter expression")
	dmObjListCmd.Flags().BoolVar(&flagDMObjListURI, "uri", false, "Print resource names only")

	dmObjLookupCmd.Flags().StringVar(&flagDMObjLookupDatabase, "database", "", "Source database identifier")
	dmObjLookupCmd.Flags().StringVar(&flagDMObjLookupSchema, "schema", "", "Source schema identifier")
	dmObjLookupCmd.Flags().StringVar(&flagDMObjLookupTable, "table", "", "Source table identifier")
	dmObjLookupCmd.Flags().StringVar(&flagDMObjLookupType, "type", "DATABASE",
		"Kind of object to look up: DATABASE, SCHEMA, or TABLE")
	dmObjLookupCmd.Flags().StringVar(&flagDMObjFormat, "format", "", "Output format")

	dmObjectsCmd.AddCommand(dmObjListCmd, dmObjLookupCmd)
	databaseMigrationCmd.AddCommand(dmObjectsCmd)
}

func dmObjParent(project, region, mj string) string {
	if strings.HasPrefix(mj, "projects/") {
		return mj
	}
	return fmt.Sprintf("projects/%s/locations/%s/migrationJobs/%s", project, region, mj)
}

func runDMObjList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := dmObjParent(project, flagDMObjRegion, flagDMObjMigrationJob)
	var all []*datamigration.MigrationJobObject
	pageToken := ""
	for {
		call := svc.Projects.Locations.MigrationJobs.Objects.List(parent).Context(ctx)
		if flagDMObjListPageSize > 0 {
			call = call.PageSize(flagDMObjListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing migration job objects: %w", err)
		}
		all = append(all, resp.MigrationJobObjects...)
		if flagDMObjListLimit > 0 && int64(len(all)) >= flagDMObjListLimit {
			all = all[:flagDMObjListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagDMObjListURI {
		for _, o := range all {
			fmt.Println(o.Name)
		}
		return nil
	}
	if flagDMObjFormat != "" {
		return emitFormatted(all, flagDMObjFormat)
	}
	fmt.Printf("%-40s %-15s %s\n", "NAME", "PHASE", "STATE")
	for _, o := range all {
		fmt.Printf("%-40s %-15s %s\n", path.Base(o.Name), o.Phase, o.State)
	}
	return nil
}

func runDMObjLookup(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	if flagDMObjLookupDatabase == "" {
		return fmt.Errorf("--database is required")
	}
	ident := &datamigration.SourceObjectIdentifier{
		Database: flagDMObjLookupDatabase,
		Schema:   flagDMObjLookupSchema,
		Table:    flagDMObjLookupTable,
		Type:     strings.ToUpper(flagDMObjLookupType),
	}
	ctx := context.Background()
	svc, err := gcp.DataMigrationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := dmObjParent(project, flagDMObjRegion, flagDMObjMigrationJob)
	req := &datamigration.LookupMigrationJobObjectRequest{SourceObjectIdentifier: ident}
	obj, err := svc.Projects.Locations.MigrationJobs.Objects.Lookup(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("looking up migration job object: %w", err)
	}
	return emitFormatted(obj, flagDMObjFormat)
}
