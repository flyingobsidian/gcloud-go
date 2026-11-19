package cli

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	metastore "google.golang.org/api/metastore/v1"
)

// --- gcloud metastore services migrations (#1952) ---
//
// GA in gcloud-python 584.0.0: start, describe, list and delete for one-shot
// (mode BACKFILL) migrations to BigLake Metastore. cancel and complete, and
// the incremental-sync mode they serve, are alpha/beta only.

// msMigListFormat is gcloud-python's default `migrations list` table; the
// rows come from msMigTableRows.
const msMigListFormat = "table(name,state,createTime,endTime,message)"

// msBigLakeCatalogPattern is the BigLake catalog resource name format that
// gcloud-python's ValidateBigLakeCatalogResourceName accepts.
const msBigLakeCatalogPattern = `^projects/[^/]+/catalogs/[^/]+$`

func (g *metastoreCmds) newMetastoreMigrationsCmds() {
	g.metastoreServicesMigrationsCmd = &cobra.Command{
		Use: "migrations", Short: "Manage the migration to a Dataproc Metastore service",
	}
	g.msMigStartCmd = &cobra.Command{
		Use:   "start SERVICE",
		Short: "Start the migration from a Dataproc Metastore service to Lakehouse runtime catalog(s)",
		Args:  cobra.ExactArgs(1), RunE: g.runMSMigStart,
	}
	g.msMigDeleteCmd = &cobra.Command{
		Use: "delete MIGRATION", Short: "Delete a migration",
		Args: cobra.ExactArgs(1), RunE: g.runMSMigDelete,
	}
	g.msMigDescribeCmd = &cobra.Command{
		Use: "describe MIGRATION", Short: "Describe a migration",
		Args: cobra.ExactArgs(1), RunE: g.runMSMigDescribe,
	}
	g.msMigListCmd = &cobra.Command{
		Use: "list", Short: "List migrations for a Dataproc Metastore service",
		Args: cobra.NoArgs, RunE: g.runMSMigList,
	}
}

func (g *metastoreCmds) initMetastoreMigrations() {
	all := []*cobra.Command{g.msMigStartCmd, g.msMigDeleteCmd, g.msMigDescribeCmd, g.msMigListCmd}
	for _, c := range all {
		c.Flags().StringVar(&g.flagMSLocation, "location", "", "Location of the service (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{g.msMigDeleteCmd, g.msMigDescribeCmd, g.msMigListCmd} {
		c.Flags().StringVar(&g.flagMSService, "service", "", "Parent service (required)")
		_ = c.MarkFlagRequired("service")
	}
	for _, c := range []*cobra.Command{g.msMigStartCmd, g.msMigDeleteCmd} {
		c.Flags().BoolVar(&g.flagMSAsync, "async", false, "Return the LRO without waiting")
	}
	for _, c := range []*cobra.Command{g.msMigDescribeCmd, g.msMigListCmd} {
		c.Flags().StringVar(&g.flagMSFormat, "format", "", "Output format")
	}

	f := g.msMigStartCmd.Flags()
	f.StringVar(&g.flagMSMigExecutionID, "migration-execution-id", "",
		"ID to use for the migration execution; a UUID is generated if omitted")
	f.StringVar(&g.flagMSMigMode, "mode", "backfill", "Execution mode of the migration: backfill")
	f.StringVar(&g.flagMSMigConflictPolicy, "conflict-policy", "",
		"Policy for resources that already exist in the target catalog: skip (default) or overwrite")
	f.BoolVar(&g.flagMSMigDryRun, "dry-run", false,
		"Plan the migration of the requested resources without migrating any metadata")
	f.StringVar(&g.flagMSMigReportPath, "report-path", "",
		"gs:// path where the backfill / dry run report is written (default: the service's artifacts bucket)")
	f.StringVar(&g.flagMSMigHiveCatalog, "hive-catalog", "",
		"Target Hive catalog, projects/{project_id_or_number}/catalogs/{catalog_id}")
	f.StringSliceVar(&g.flagMSMigHiveDatabases, "hive-databases", nil,
		"Comma-separated databases to migrate to the Hive catalog (default: *)")
	f.StringVar(&g.flagMSMigIcebergCatalog, "iceberg-catalog", "",
		"Target Iceberg REST catalog, projects/{project_id_or_number}/catalogs/{catalog_id}")
	f.StringSliceVar(&g.flagMSMigIcebergNamespaces, "iceberg-namespaces", nil,
		"Comma-separated namespaces to migrate to the Iceberg REST catalog (default: *)")

	g.metastoreServicesMigrationsCmd.AddCommand(all...)
	g.metastoreServicesCmd.AddCommand(g.metastoreServicesMigrationsCmd)
}

func msMigName(id, project, location, service string) string {
	return msChild("migrationExecutions", id, msSvcName(service, project, location))
}

// msMigChoice maps a choice flag value onto its API enum value.
func msMigChoice(flag, value string, choices map[string]string, order []string) (string, error) {
	if enum, ok := choices[value]; ok {
		return enum, nil
	}
	return "", fmt.Errorf("argument --%s: Invalid choice: '%s'.\n\nValid choices are [%s].",
		flag, value, strings.Join(order, ", "))
}

// msMigValidateList applies gcloud-python's ValidateDatabaseOrNamespaceList.
func msMigValidateList(flag string, items []string) error {
	if len(items) == 0 {
		return fmt.Errorf("Invalid value for [--%s]: The list cannot be empty.", flag)
	}
	for _, it := range items {
		if it == "*" && len(items) > 1 {
			return fmt.Errorf("Invalid value for [--%s]: When \"*\" is specified to include all items, it must be the only element in the list.", flag)
		}
	}
	for _, it := range items {
		if strings.TrimSpace(it) == "" {
			return fmt.Errorf("Invalid value for [--%s]: Item in list cannot be empty or whitespace.", flag)
		}
	}
	return nil
}

// msMigCatalogConfig validates one catalog/list flag group and returns the
// catalog and list to send, defaulting the list to ["*"] as gcloud-python's
// PopulateBigLakeMigrationDefaults does. ok is false when the group is unset.
func msMigCatalogConfig(cmd *cobra.Command, catalogFlag, catalog, listFlag string, list []string) (string, []string, bool, error) {
	catalogSet := cmd.Flags().Changed(catalogFlag)
	listSet := cmd.Flags().Changed(listFlag)
	if !catalogSet {
		if listSet {
			return "", nil, false, fmt.Errorf("argument --%s: --%s must be specified.", listFlag, catalogFlag)
		}
		return "", nil, false, nil
	}
	if ok, _ := regexp.MatchString(msBigLakeCatalogPattern, catalog); !ok {
		return "", nil, false, fmt.Errorf("Invalid value for [--%s]: The BigLake catalog resource name should be in the format projects/<project_id>/catalogs/<catalog_id>", catalogFlag)
	}
	if !listSet {
		return catalog, []string{"*"}, true, nil
	}
	if err := msMigValidateList(listFlag, list); err != nil {
		return "", nil, false, err
	}
	return catalog, list, true, nil
}

// msMigStartRequest builds the StartMigrationRequest from the start flags.
func (g *metastoreCmds) msMigStartRequest(cmd *cobra.Command) (*metastore.StartMigrationRequest, error) {
	mode, err := msMigChoice("mode", g.flagMSMigMode, map[string]string{"backfill": "BACKFILL"}, []string{"backfill"})
	if err != nil {
		return nil, err
	}
	cfg := &metastore.BigLakeMetastoreMigrationConfig{Mode: mode, DryRun: g.flagMSMigDryRun}
	if cmd.Flags().Changed("conflict-policy") {
		cfg.ConflictPolicy, err = msMigChoice("conflict-policy", g.flagMSMigConflictPolicy,
			map[string]string{"skip": "SKIP", "overwrite": "OVERWRITE"}, []string{"overwrite", "skip"})
		if err != nil {
			return nil, err
		}
	}
	if cmd.Flags().Changed("report-path") {
		if !strings.HasPrefix(g.flagMSMigReportPath, "gs://") {
			return nil, fmt.Errorf("Invalid value for [--report-path]: Expected URI %s to start with `gs://`.", g.flagMSMigReportPath)
		}
		cfg.ReportPath = g.flagMSMigReportPath
	}
	if catalog, dbs, ok, err := msMigCatalogConfig(cmd, "hive-catalog", g.flagMSMigHiveCatalog,
		"hive-databases", g.flagMSMigHiveDatabases); err != nil {
		return nil, err
	} else if ok {
		cfg.HiveConfig = &metastore.HiveConfig{Catalog: catalog, Databases: dbs}
	}
	if catalog, nss, ok, err := msMigCatalogConfig(cmd, "iceberg-catalog", g.flagMSMigIcebergCatalog,
		"iceberg-namespaces", g.flagMSMigIcebergNamespaces); err != nil {
		return nil, err
	} else if ok {
		cfg.IcebergConfig = &metastore.IcebergConfig{Catalog: catalog, Namespaces: nss}
	}
	return &metastore.StartMigrationRequest{
		MigrationExecutionId: g.flagMSMigExecutionID,
		MigrationExecution:   &metastore.MigrationExecution{BiglakeMetastoreMigrationConfig: cfg},
	}, nil
}

func (g *metastoreCmds) runMSMigStart(cmd *cobra.Command, args []string) error {
	req, err := g.msMigStartRequest(cmd)
	if err != nil {
		return err
	}
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.StartMigration(
		msSvcName(args[0], project, g.flagMSLocation), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting migration: %w", err)
	}
	return g.msFinishOp(ctx, svc, op, "Start migration", args[0], g.flagMSAsync)
}

func (g *metastoreCmds) runMSMigDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Services.MigrationExecutions.Delete(
		msMigName(args[0], project, g.flagMSLocation, g.flagMSService)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting migration: %w", err)
	}
	return g.msFinishOp(ctx, svc, op, "Delete migration", args[0], g.flagMSAsync)
}

func (g *metastoreCmds) runMSMigDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Services.MigrationExecutions.Get(
		msMigName(args[0], project, g.flagMSLocation, g.flagMSService)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing migration: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagMSFormat)
}

func (g *metastoreCmds) runMSMigList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Metastore(ctx)
	if err != nil {
		return err
	}
	var all []*metastore.MigrationExecution
	err = svc.Projects.Locations.Services.MigrationExecutions.List(
		msSvcName(g.flagMSService, project, g.flagMSLocation)).Context(ctx).
		Pages(ctx, func(resp *metastore.ListMigrationExecutionsResponse) error {
			all = append(all, resp.MigrationExecutions...)
			return nil
		})
	if err != nil {
		return fmt.Errorf("listing migrations: %w", err)
	}
	if g.flagMSFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagMSFormat)
	}
	if len(all) == 0 {
		fmt.Fprintln(g.inv.stderr, "Listed 0 items.")
		return nil
	}
	return emitFormattedTo(g.inv.stdout, msMigTableRows(all), msMigListFormat)
}

// msMigTableRows projects migrations onto gcloud-python's default list
// columns: name.basename(), state, createTime.date(), endTime.date(),
// stateMessage.
func msMigTableRows(migs []*metastore.MigrationExecution) []map[string]any {
	rows := make([]map[string]any, 0, len(migs))
	for _, m := range migs {
		rows = append(rows, map[string]any{
			"name":       path.Base(m.Name),
			"state":      m.State,
			"createTime": msMigDate(m.CreateTime),
			"endTime":    msMigDate(m.EndTime),
			"message":    m.StateMessage,
		})
	}
	return rows
}

// msMigDate is gcloud's date() transform with its default format: the time in
// the timestamp's own zone as %Y-%m-%dT%H:%M:%S, or "" when unparseable.
func msMigDate(ts string) string {
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return ""
	}
	return t.Format("2006-01-02T15:04:05")
}
