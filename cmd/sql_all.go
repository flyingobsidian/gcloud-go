package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	sqladmin "google.golang.org/api/sqladmin/v1"
)

// --- Shared helpers for the Cloud SQL surface (#875-#886) ---

// sqlWaitOp polls a Cloud SQL operation until Status == "DONE".
func sqlWaitOp(ctx context.Context, svc *sqladmin.Service, project string, op *sqladmin.Operation) (*sqladmin.Operation, error) {
	for op.Status != "DONE" {
		time.Sleep(2 * time.Second)
		got, err := svc.Operations.Get(project, op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil && len(op.Error.Errors) > 0 {
		return op, fmt.Errorf("operation %s failed: %s", op.Name, op.Error.Errors[0].Message)
	}
	return op, nil
}

// sqlFinishOp waits (or returns immediately if --async) for a SQL operation
// and prints the resulting operation resource.
func (s *sqlCmds) sqlFinishOp(ctx context.Context, svc *sqladmin.Service, project string, op *sqladmin.Operation, verb, target string) error {
	if s.flagSQLAsync {
		fmt.Fprintf(s.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagSQLFormat)
	}
	final, err := sqlWaitOp(ctx, svc, project, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "%s [%s] completed.\n", verb, target)
	return emitFormattedTo(s.inv.stdout, final, s.flagSQLFormat)
}

// sqlService returns an authenticated Cloud SQL Admin API client.
func (s *sqlCmds) sqlService(ctx context.Context) (*sqladmin.Service, error) {
	return s.inv.clients().SQLAdmin(ctx)
}

// requireInstance returns the resolved --instance flag or errors if unset.
func (s *sqlCmds) requireInstance() (string, error) {
	if s.flagSQLInstance == "" {
		return "", fmt.Errorf("--instance is required")
	}
	return s.flagSQLInstance, nil
}

// --- Top-level subgroups ---

// --- backups ---

func (s *sqlCmds) runSQLBackupCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	run := &sqladmin.BackupRun{
		Description: s.flagSQLDescription,
		Instance:    instance,
		Location:    s.flagSQLLocation,
		Kind:        "sql#backupRun",
	}
	op, err := svc.BackupRuns.Insert(project, instance, run).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Create backup", instance)
}

func (s *sqlCmds) runSQLBackupDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("backup ID must be an integer: %w", err)
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.BackupRuns.Delete(project, instance, id).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Delete backup", args[0])
}

func (s *sqlCmds) runSQLBackupDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("backup ID must be an integer: %w", err)
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.BackupRuns.Get(project, instance, id).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSQLFormat)
}

func (s *sqlCmds) runSQLBackupList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	call := svc.BackupRuns.List(project, instance)
	if s.flagSQLLimit > 0 {
		call = call.MaxResults(s.flagSQLLimit)
	}
	resp, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing backups: %w", err)
	}
	if s.flagSQLFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Items, s.flagSQLFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Items), "%-20s %-24s %-10s %s\n", "ID", "WINDOW_START_TIME", "STATUS", "TYPE")
	for _, b := range resp.Items {
		fmt.Fprintf(s.inv.stdout, "%-20d %-24s %-10s %s\n", b.Id, b.WindowStartTime, b.Status, b.Type)
	}
	return nil
}

func (s *sqlCmds) runSQLBackupRestore(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("backup ID must be an integer: %w", err)
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	req := &sqladmin.InstancesRestoreBackupRequest{
		RestoreBackupContext: &sqladmin.RestoreBackupContext{
			BackupRunId: id,
			Kind:        "sql#restoreBackupContext",
			InstanceId:  instance,
			Project:     project,
		},
	}
	op, err := svc.Instances.RestoreBackup(project, instance, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring backup: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Restore backup", instance)
}

// --- connect ---

// dbClientBinary returns the local client binary name (psql, mysql, mssql-cli)
// used to connect to a Cloud SQL instance's public IP.
func dbClientBinary(dbVersion string) (bin string, defaultUser string, err error) {
	prefix := strings.SplitN(dbVersion, "_", 2)[0]
	switch strings.ToUpper(prefix) {
	case "POSTGRES":
		return "psql", "postgres", nil
	case "MYSQL":
		return "mysql", "root", nil
	case "SQLSERVER":
		return "mssql-cli", "sqlserver", nil
	default:
		return "", "", fmt.Errorf("unsupported database version %q", dbVersion)
	}
}

// instancePublicIP returns the primary primary IP the local client should use
// to reach the instance (public, then private).
func instancePublicIP(inst *sqladmin.DatabaseInstance) string {
	for _, ip := range inst.IpAddresses {
		if ip.Type == "PRIMARY" && ip.IpAddress != "" {
			return ip.IpAddress
		}
	}
	for _, ip := range inst.IpAddresses {
		if ip.IpAddress != "" {
			return ip.IpAddress
		}
	}
	return ""
}

// runSQLConnect fetches the instance, verifies the local client is on PATH and
// execs it against the resolved IP.
func (s *sqlCmds) runSQLConnect(bin string, cmd *cobra.Command, args []string) error {
	instance := args[0]
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	inst, err := svc.Instances.Get(project, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("looking up instance: %w", err)
	}
	ip := instancePublicIP(inst)
	if ip == "" {
		return fmt.Errorf("instance %s has no reachable IP address", instance)
	}
	dbBin, defaultUser, err := dbClientBinary(inst.DatabaseVersion)
	if err != nil {
		return err
	}
	if bin != "" && bin != dbBin {
		fmt.Fprintf(s.inv.stderr, "warning: instance database is %s but connecting with %s\n", inst.DatabaseVersion, bin)
		dbBin = bin
	}
	if _, err := exec.LookPath(dbBin); err != nil {
		return fmt.Errorf("%s client not found in PATH: install a %s client and retry", dbBin, dbBin)
	}
	user := s.flagSQLUser
	if user == "" {
		user = defaultUser
	}
	var execArgs []string
	switch dbBin {
	case "psql":
		execArgs = []string{"-h", ip, "-U", user}
		if s.flagSQLDatabase != "" {
			execArgs = append(execArgs, "-d", s.flagSQLDatabase)
		}
	case "mysql":
		execArgs = []string{"-h", ip, "-u", user, "-p"}
		if s.flagSQLDatabase != "" {
			execArgs = append(execArgs, s.flagSQLDatabase)
		}
	case "mssql-cli":
		execArgs = []string{"-S", ip, "-U", user}
		if s.flagSQLDatabase != "" {
			execArgs = append(execArgs, "-d", s.flagSQLDatabase)
		}
	}
	c := exec.CommandContext(ctx, dbBin, execArgs...)
	c.Stdin = s.inv.stdin
	c.Stdout = s.inv.stdout
	c.Stderr = s.inv.stderr
	fmt.Fprintf(s.inv.stderr, "Connecting to %s at %s using %s...\n", instance, ip, dbBin)
	return c.Run()
}

// --- databases ---

func (s *sqlCmds) runSQLDatabaseCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	db := &sqladmin.Database{
		Name:      args[0],
		Instance:  instance,
		Project:   project,
		Charset:   s.flagSQLCharset,
		Collation: s.flagSQLCollation,
		Kind:      "sql#database",
	}
	op, err := svc.Databases.Insert(project, instance, db).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating database: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Create database", args[0])
}

func (s *sqlCmds) runSQLDatabaseDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Databases.Delete(project, instance, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting database: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Delete database", args[0])
}

func (s *sqlCmds) runSQLDatabaseDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Databases.Get(project, instance, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing database: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSQLFormat)
}

func (s *sqlCmds) runSQLDatabaseList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Databases.List(project, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing databases: %w", err)
	}
	if s.flagSQLFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Items, s.flagSQLFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Items), "%-32s %-16s %s\n", "NAME", "CHARSET", "COLLATION")
	for _, d := range resp.Items {
		fmt.Fprintf(s.inv.stdout, "%-32s %-16s %s\n", d.Name, d.Charset, d.Collation)
	}
	return nil
}

func (s *sqlCmds) runSQLDatabasePatch(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	db := &sqladmin.Database{}
	if s.flagSQLConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagSQLConfigFile, db); err != nil {
			return err
		}
	}
	if s.flagSQLCharset != "" {
		db.Charset = s.flagSQLCharset
	}
	if s.flagSQLCollation != "" {
		db.Collation = s.flagSQLCollation
	}
	db.Name = args[0]
	db.Instance = instance
	db.Project = project
	op, err := svc.Databases.Patch(project, instance, args[0], db).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("patching database: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Patch database", args[0])
}

// --- export ---

func (s *sqlCmds) runSQLExport(fileType string, cmd *cobra.Command, args []string) error {
	instance := args[0]
	uri := args[1]
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	ec := &sqladmin.ExportContext{
		FileType:  fileType,
		Uri:       uri,
		Kind:      "sql#exportContext",
		Databases: s.flagSQLExportDBs,
		Offload:   s.flagSQLOffload,
	}
	if len(ec.Databases) == 0 && s.flagSQLDatabase != "" {
		ec.Databases = []string{s.flagSQLDatabase}
	}
	switch fileType {
	case "SQL":
		if len(s.flagSQLTable) > 0 {
			ec.SqlExportOptions = &sqladmin.ExportContextSqlExportOptions{Tables: s.flagSQLTable}
		}
	case "CSV":
		ec.CsvExportOptions = &sqladmin.ExportContextCsvExportOptions{SelectQuery: s.flagSQLQuery}
	case "BAK":
		if s.flagSQLBAKType != "" {
			ec.BakExportOptions = &sqladmin.ExportContextBakExportOptions{BakType: s.flagSQLBAKType}
		}
	}
	op, err := svc.Instances.Export(project, instance, &sqladmin.InstancesExportRequest{ExportContext: ec}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting instance: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Export instance", instance)
}

// --- flags ---

func (s *sqlCmds) runSQLFlagsList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	call := svc.Flags.List()
	if s.flagSQLDBVersion != "" {
		call = call.DatabaseVersion(s.flagSQLDBVersion)
	}
	resp, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing flags: %w", err)
	}
	if s.flagSQLFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Items, s.flagSQLFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Items), "%-40s %-12s %s\n", "NAME", "TYPE", "APPLIES_TO")
	for _, f := range resp.Items {
		fmt.Fprintf(s.inv.stdout, "%-40s %-12s %s\n", f.Name, f.Type, strings.Join(f.AppliesTo, ","))
	}
	return nil
}

// --- import ---

func (s *sqlCmds) runSQLImport(fileType string, cmd *cobra.Command, args []string) error {
	instance := args[0]
	uri := args[1]
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	ic := &sqladmin.ImportContext{
		FileType:   fileType,
		Uri:        uri,
		Kind:       "sql#importContext",
		Database:   s.flagSQLDatabase,
		ImportUser: s.flagSQLImportUser,
	}
	switch fileType {
	case "CSV":
		ic.CsvImportOptions = &sqladmin.ImportContextCsvImportOptions{
			Table:   firstNonEmpty(s.flagSQLTable),
			Columns: s.flagSQLColumns,
		}
	case "BAK":
		bak := &sqladmin.ImportContextBakImportOptions{BakType: s.flagSQLBAKType}
		if s.flagSQLEncryptionKey != "" {
			bak.EncryptionOptions = &sqladmin.ImportContextBakImportOptionsEncryptionOptions{
				PvkPath:     s.flagSQLPvKPath,
				CertPath:    s.flagSQLCertPath,
				PvkPassword: s.flagSQLEncryptionKey,
			}
		}
		ic.BakImportOptions = bak
	}
	op, err := svc.Instances.Import(project, instance, &sqladmin.InstancesImportRequest{ImportContext: ic}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing to instance: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Import into instance", instance)
}

func firstNonEmpty(ss []string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

// --- instances ---

func (s *sqlCmds) runSQLInstanceCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	inst := &sqladmin.DatabaseInstance{Kind: "sql#instance"}
	if s.flagSQLConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagSQLConfigFile, inst); err != nil {
			return err
		}
	}
	inst.Name = args[0]
	if s.flagSQLDBVersion != "" {
		inst.DatabaseVersion = s.flagSQLDBVersion
	}
	if s.flagSQLRegion != "" {
		inst.Region = s.flagSQLRegion
	}
	if s.flagSQLTier != "" {
		if inst.Settings == nil {
			inst.Settings = &sqladmin.Settings{}
		}
		inst.Settings.Tier = s.flagSQLTier
	}
	if s.flagSQLZone != "" {
		if inst.Settings == nil {
			inst.Settings = &sqladmin.Settings{}
		}
		if inst.Settings.LocationPreference == nil {
			inst.Settings.LocationPreference = &sqladmin.LocationPreference{}
		}
		inst.Settings.LocationPreference.Zone = s.flagSQLZone
	}
	if s.flagSQLPassword != "" {
		inst.RootPassword = s.flagSQLPassword
	}
	if cmd.Flags().Changed("storage-auto-increase-limit") {
		if inst.Settings == nil {
			inst.Settings = &sqladmin.Settings{}
		}
		inst.Settings.StorageAutoResizeLimit = s.flagSQLStorageAutoIncreaseLimit
		inst.Settings.ForceSendFields = append(inst.Settings.ForceSendFields, "StorageAutoResizeLimit")
	}
	call := svc.Instances.Insert(project, inst).Context(ctx)
	if params := sqlInstanceRoutingParams(inst); params != "" {
		call.Header().Set("x-goog-request-params", params)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Create instance", args[0])
}

// sqlInstanceRoutingParams builds the x-goog-request-params routing header
// that gcloud-python 588.0.0 sends on instance creation
// (_BuildRoutingParams in surface/sql/instances/create.py).
func sqlInstanceRoutingParams(inst *sqladmin.DatabaseInstance) string {
	var params []string
	if inst.Settings != nil && inst.Settings.Edition != "" {
		params = append(params, "edition="+inst.Settings.Edition)
	}
	if inst.Region != "" {
		params = append(params, "region="+inst.Region)
	}
	return strings.Join(params, "&")
}

func (s *sqlCmds) runSQLInstanceDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Instances.Delete(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Delete instance", args[0])
}

func (s *sqlCmds) runSQLInstanceDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Instances.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSQLFormat)
}

func (s *sqlCmds) runSQLInstanceList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	call := svc.Instances.List(project)
	if s.flagSQLFilter != "" {
		call = call.Filter(s.flagSQLFilter)
	}
	resp, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing instances: %w", err)
	}
	if s.flagSQLFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Items, s.flagSQLFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Items), "%-32s %-16s %-14s %-14s %s\n", "NAME", "DATABASE_VERSION", "REGION", "TIER", "STATE")
	for _, i := range resp.Items {
		tier := ""
		if i.Settings != nil {
			tier = i.Settings.Tier
		}
		fmt.Fprintf(s.inv.stdout, "%-32s %-16s %-14s %-14s %s\n", i.Name, i.DatabaseVersion, i.Region, tier, i.State)
	}
	return nil
}

func (s *sqlCmds) runSQLInstancePatch(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	inst := &sqladmin.DatabaseInstance{}
	if s.flagSQLConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagSQLConfigFile, inst); err != nil {
			return err
		}
	}
	inst.Name = args[0]
	if cmd.Flags().Changed("storage-auto-increase-limit") {
		if inst.Settings == nil {
			inst.Settings = &sqladmin.Settings{}
		}
		inst.Settings.StorageAutoResizeLimit = s.flagSQLStorageAutoIncreaseLimit
		inst.Settings.ForceSendFields = append(inst.Settings.ForceSendFields, "StorageAutoResizeLimit")
	}
	if s.flagSQLConfigFile == "" && !cmd.Flags().Changed("storage-auto-increase-limit") {
		return fmt.Errorf("--config-file or a supported patch flag (e.g. --storage-auto-increase-limit) is required")
	}
	op, err := svc.Instances.Patch(project, args[0], inst).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("patching instance: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Patch instance", args[0])
}

func (s *sqlCmds) runSQLInstanceClone(cmd *cobra.Command, args []string) error {
	source, dest := args[0], args[1]
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	req := &sqladmin.InstancesCloneRequest{CloneContext: &sqladmin.CloneContext{
		DestinationInstanceName: dest,
		Kind:                    "sql#cloneContext",
	}}
	op, err := svc.Instances.Clone(project, source, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cloning instance: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Clone instance", dest)
}

func (s *sqlCmds) runSQLInstanceFailover(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	req := &sqladmin.InstancesFailoverRequest{FailoverContext: &sqladmin.FailoverContext{
		Kind:            "sql#failoverContext",
		SettingsVersion: s.flagSQLSettingsVer,
	}}
	op, err := svc.Instances.Failover(project, args[0], req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failing over instance: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Failover instance", args[0])
}

func (s *sqlCmds) runSQLInstanceRestart(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Instances.Restart(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restarting instance: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Restart instance", args[0])
}

func (s *sqlCmds) runSQLInstanceRestoreBackup(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	if s.flagSQLBackup == "" {
		return fmt.Errorf("--backup-id is required")
	}
	id, err := strconv.ParseInt(s.flagSQLBackup, 10, 64)
	if err != nil {
		return fmt.Errorf("--backup-id must be an integer: %w", err)
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	req := &sqladmin.InstancesRestoreBackupRequest{
		RestoreBackupContext: &sqladmin.RestoreBackupContext{
			BackupRunId: id,
			InstanceId:  args[0],
			Project:     project,
			Kind:        "sql#restoreBackupContext",
		},
	}
	op, err := svc.Instances.RestoreBackup(project, args[0], req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring backup: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Restore backup", args[0])
}

func (s *sqlCmds) runSQLInstancePromoteReplica(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Instances.PromoteReplica(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("promoting replica: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Promote replica", args[0])
}

func (s *sqlCmds) runSQLInstanceStartReplica(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Instances.StartReplica(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting replica: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Start replica", args[0])
}

func (s *sqlCmds) runSQLInstanceStopReplica(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Instances.StopReplica(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping replica: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Stop replica", args[0])
}

func (s *sqlCmds) runSQLInstanceReencrypt(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	req := &sqladmin.InstancesReencryptRequest{}
	op, err := svc.Instances.Reencrypt(project, args[0], req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("re-encrypting instance: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Re-encrypt instance", args[0])
}

func (s *sqlCmds) runSQLInstanceResetSSLConfig(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Instances.ResetSslConfig(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resetting SSL config: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Reset SSL config", args[0])
}

func (s *sqlCmds) runSQLInstanceListServerCas(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Instances.ListServerCas(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing server CAs: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagSQLFormat)
}

func (s *sqlCmds) runSQLInstanceRescheduleMaintenance(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	body := &sqladmin.SqlInstancesRescheduleMaintenanceRequestBody{
		Reschedule: &sqladmin.Reschedule{
			RescheduleType: s.flagSQLReschedule,
			ScheduleTime:   s.flagSQLScheduleTime,
		},
	}
	op, err := svc.Projects.Instances.RescheduleMaintenance(project, args[0], body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rescheduling maintenance: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Reschedule maintenance", args[0])
}

func (s *sqlCmds) runSQLInstanceAcquireSsrsLease(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	req := &sqladmin.InstancesAcquireSsrsLeaseRequest{}
	if s.flagSQLConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagSQLConfigFile, req); err != nil {
			return err
		}
	}
	resp, err := svc.Instances.AcquireSsrsLease(project, args[0], req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("acquiring SSRS lease: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagSQLFormat)
}

func (s *sqlCmds) runSQLInstanceReleaseSsrsLease(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Instances.ReleaseSsrsLease(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("releasing SSRS lease: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagSQLFormat)
}

func (s *sqlCmds) runSQLInstanceResetAsyncReplicaLag(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	req := &sqladmin.SqlInstancesResetReplicaSizeRequest{}
	op, err := svc.Projects.Instances.ResetReplicaSize(project, args[0], req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resetting replica size: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Reset replica size", args[0])
}

func (s *sqlCmds) runSQLInstanceVerifyExternalSyncSettings(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	req := &sqladmin.SqlInstancesVerifyExternalSyncSettingsRequest{}
	if s.flagSQLConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagSQLConfigFile, req); err != nil {
			return err
		}
	}
	resp, err := svc.Projects.Instances.VerifyExternalSyncSettings(project, args[0], req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("verifying external sync settings: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagSQLFormat)
}

// --- operations ---

func (s *sqlCmds) runSQLOpCancel(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Operations.Cancel(project, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (s *sqlCmds) runSQLOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Operations.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSQLFormat)
}

func (s *sqlCmds) runSQLOpList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	call := svc.Operations.List(project)
	if s.flagSQLInstance != "" {
		call = call.Instance(s.flagSQLInstance)
	}
	if s.flagSQLLimit > 0 {
		call = call.MaxResults(s.flagSQLLimit)
	}
	resp, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if s.flagSQLFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Items, s.flagSQLFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Items), "%-40s %-14s %-24s %s\n", "NAME", "TYPE", "START", "STATUS")
	for _, o := range resp.Items {
		fmt.Fprintf(s.inv.stdout, "%-40s %-14s %-24s %s\n", o.Name, o.OperationType, o.StartTime, o.Status)
	}
	return nil
}

func (s *sqlCmds) runSQLOpWait(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Operations.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching operation: %w", err)
	}
	final, err := sqlWaitOp(ctx, svc, project, op)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, final, s.flagSQLFormat)
}

// --- ssl server-ca-certs ---

func (s *sqlCmds) runSQLSSLServerCAList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Instances.ListServerCas(project, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing server CAs: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagSQLFormat)
}

func (s *sqlCmds) runSQLSSLServerCAAdd(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Instances.AddServerCa(project, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding server CA: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Add server CA", instance)
}

// runSQLSSLServerCARemove maps gcloud's `remove SHA1_FINGERPRINT` semantics
// onto RotateServerCa, which retires the currently active CA in favor of the
// next-most-recent one.
func (s *sqlCmds) runSQLSSLServerCARemove(cmd *cobra.Command, args []string) error {
	// The upstream Cloud SQL Admin API does not expose an explicit
	// "remove server CA" endpoint; server CAs are retired by a rotate.
	// We mirror gcloud python's behavior by delegating to rotate with the
	// caller-supplied --next-version (or the most recently added CA).
	return s.runSQLSSLServerCARotate(cmd, args)
}

func (s *sqlCmds) runSQLSSLServerCARotate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	req := &sqladmin.InstancesRotateServerCaRequest{RotateServerCaContext: &sqladmin.RotateServerCaContext{
		Kind:        "sql#rotateServerCaContext",
		NextVersion: s.flagSQLNextVersion,
	}}
	op, err := svc.Instances.RotateServerCa(project, instance, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rotating server CA: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Rotate server CA", instance)
}

// --- ssl-certs (DEPRECATED) ---

// findSSLCertBySha1 resolves a common name to a sha1 fingerprint by listing.
func findSSLCertBySha1(ctx context.Context, svc *sqladmin.Service, project, instance, commonName string) (string, error) {
	resp, err := svc.SslCerts.List(project, instance).Context(ctx).Do()
	if err != nil {
		return "", err
	}
	for _, c := range resp.Items {
		if c.CommonName == commonName || c.Sha1Fingerprint == commonName {
			return c.Sha1Fingerprint, nil
		}
	}
	return "", fmt.Errorf("no ssl-cert found with common name %q", commonName)
}

func (s *sqlCmds) runSQLSSLCertCreate(cmd *cobra.Command, args []string) error {
	commonName, certFile := args[0], args[1]
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.SslCerts.Insert(project, instance, &sqladmin.SslCertsInsertRequest{CommonName: commonName}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating ssl-cert: %w", err)
	}
	if resp.ClientCert != nil && resp.ClientCert.CertPrivateKey != "" {
		if err := os.WriteFile(certFile, []byte(resp.ClientCert.CertPrivateKey), 0600); err != nil {
			return fmt.Errorf("writing private key to %s: %w", certFile, err)
		}
		fmt.Fprintf(s.inv.stderr, "Wrote private key to %s.\n", certFile)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagSQLFormat)
}

func (s *sqlCmds) runSQLSSLCertDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	sha, err := findSSLCertBySha1(ctx, svc, project, instance, args[0])
	if err != nil {
		return err
	}
	op, err := svc.SslCerts.Delete(project, instance, sha).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting ssl-cert: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Delete ssl-cert", args[0])
}

func (s *sqlCmds) runSQLSSLCertDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	sha, err := findSSLCertBySha1(ctx, svc, project, instance, args[0])
	if err != nil {
		return err
	}
	got, err := svc.SslCerts.Get(project, instance, sha).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing ssl-cert: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSQLFormat)
}

func (s *sqlCmds) runSQLSSLCertList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.SslCerts.List(project, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing ssl-certs: %w", err)
	}
	if s.flagSQLFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Items, s.flagSQLFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Items), "%-30s %-40s %s\n", "COMMON_NAME", "SHA1_FINGERPRINT", "EXPIRATION_TIME")
	for _, c := range resp.Items {
		fmt.Fprintf(s.inv.stdout, "%-30s %-40s %s\n", c.CommonName, c.Sha1Fingerprint, c.ExpirationTime)
	}
	return nil
}

// --- tiers ---

func (s *sqlCmds) runSQLTiersList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Tiers.List(project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing tiers: %w", err)
	}
	if s.flagSQLFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Items, s.flagSQLFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Items), "%-30s %-12s %-14s %s\n", "TIER", "AVAILABLE_REGIONS", "RAM_BYTES", "DISK_BYTES")
	for _, t := range resp.Items {
		fmt.Fprintf(s.inv.stdout, "%-30s %-12d %-14d %d\n", t.Tier, len(t.Region), t.RAM, t.DiskQuota)
	}
	return nil
}

// --- users ---

func (s *sqlCmds) runSQLUserCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	u := &sqladmin.User{
		Name:     args[0],
		Instance: instance,
		Project:  project,
		Password: s.flagSQLPassword,
		Host:     s.flagSQLHost,
		Type:     s.flagSQLType,
		Kind:     "sql#user",
	}
	op, err := svc.Users.Insert(project, instance, u).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Create user", args[0])
}

func (s *sqlCmds) runSQLUserDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	call := svc.Users.Delete(project, instance).Name(args[0])
	if s.flagSQLHost != "" {
		call = call.Host(s.flagSQLHost)
	}
	op, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Delete user", args[0])
}

func (s *sqlCmds) runSQLUserDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	call := svc.Users.Get(project, instance, args[0])
	if s.flagSQLHost != "" {
		call = call.Host(s.flagSQLHost)
	}
	got, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing user: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSQLFormat)
}

func (s *sqlCmds) runSQLUserList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Users.List(project, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing users: %w", err)
	}
	if s.flagSQLFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Items, s.flagSQLFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Items), "%-30s %-16s %s\n", "NAME", "HOST", "TYPE")
	for _, u := range resp.Items {
		fmt.Fprintf(s.inv.stdout, "%-30s %-16s %s\n", u.Name, u.Host, u.Type)
	}
	return nil
}

func (s *sqlCmds) runSQLUserSetPassword(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	if s.flagSQLPassword == "" {
		return fmt.Errorf("--password is required")
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	body := &sqladmin.User{
		Name:     args[0],
		Instance: instance,
		Project:  project,
		Password: s.flagSQLPassword,
	}
	call := svc.Users.Update(project, instance, body).Name(args[0])
	if s.flagSQLHost != "" {
		call = call.Host(s.flagSQLHost)
	}
	op, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting password: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Set password", args[0])
}

func (s *sqlCmds) runSQLUserPatch(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	instance, err := s.requireInstance()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.sqlService(ctx)
	if err != nil {
		return err
	}
	body := &sqladmin.User{}
	if s.flagSQLConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagSQLConfigFile, body); err != nil {
			return err
		}
	}
	body.Name = args[0]
	body.Instance = instance
	body.Project = project
	if s.flagSQLPassword != "" {
		body.Password = s.flagSQLPassword
	}
	call := svc.Users.Update(project, instance, body).Name(args[0])
	if s.flagSQLHost != "" {
		call = call.Host(s.flagSQLHost)
	} else if body.Host != "" {
		call = call.Host(body.Host)
	}
	op, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("patching user: %w", err)
	}
	return s.sqlFinishOp(ctx, svc, project, op, "Patch user", args[0])
}

// --- flag registration ---

// registerSQL wires all Cloud SQL subgroups, subcommands and shared flags.
// Called from init().
func (s *sqlCmds) registerSQL() {
	// Backups
	for _, c := range []*cobra.Command{s.sqlBackupCreateCmd, s.sqlBackupDeleteCmd, s.sqlBackupDescribeCmd, s.sqlBackupListCmd, s.sqlBackupRestoreCmd} {
		c.Flags().StringVar(&s.flagSQLInstance, "instance", "", "Cloud SQL instance ID (required)")
		c.Flags().StringVar(&s.flagSQLFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.sqlBackupCreateCmd, s.sqlBackupDeleteCmd, s.sqlBackupRestoreCmd} {
		c.Flags().BoolVar(&s.flagSQLAsync, "async", false, "Return without waiting for the operation to complete")
	}
	s.sqlBackupCreateCmd.Flags().StringVar(&s.flagSQLDescription, "description", "", "Description of the backup")
	s.sqlBackupCreateCmd.Flags().StringVar(&s.flagSQLLocation, "location", "", "Region for storing the backup")
	s.sqlBackupListCmd.Flags().Int64Var(&s.flagSQLLimit, "limit", 0, "Maximum number of backups to return")
	s.sqlBackupsCmd.AddCommand(s.sqlBackupCreateCmd, s.sqlBackupDeleteCmd, s.sqlBackupDescribeCmd, s.sqlBackupListCmd, s.sqlBackupRestoreCmd)
	s.sqlCmd.AddCommand(s.sqlBackupsCmd)

	// Connect
	for _, c := range []*cobra.Command{s.sqlConnectPsqlCmd, s.sqlConnectMysqlCmd, s.sqlConnectSqlserverCmd} {
		c.Flags().StringVarP(&s.flagSQLUser, "user", "u", "", "Database user to connect as")
		c.Flags().StringVar(&s.flagSQLDatabase, "database", "", "Database to connect to")
		c.Flags().IntVar(&s.flagSQLPort, "port", 0, "Port that gcloud uses locally to connect through Cloud SQL Proxy")
		c.Flags().BoolVar(&s.flagSQLPrivateIP, "private-ip", false, "Use the instance's private IP")
		c.Flags().BoolVar(&s.flagSQLPSC, "psc", false, "Use Private Service Connect (PSC) to reach the instance")
		c.Flags().BoolVar(&s.flagSQLAutoIP, "auto-ip", false, "Auto-select an IP address")
	}
	s.sqlConnectCmd.AddCommand(s.sqlConnectPsqlCmd, s.sqlConnectMysqlCmd, s.sqlConnectSqlserverCmd)
	s.sqlCmd.AddCommand(s.sqlConnectCmd)

	// Databases
	for _, c := range []*cobra.Command{s.sqlDatabaseCreateCmd, s.sqlDatabaseDeleteCmd, s.sqlDatabaseDescribeCmd, s.sqlDatabaseListCmd, s.sqlDatabasePatchCmd} {
		c.Flags().StringVar(&s.flagSQLInstance, "instance", "", "Cloud SQL instance ID (required)")
		c.Flags().StringVar(&s.flagSQLFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.sqlDatabaseCreateCmd, s.sqlDatabaseDeleteCmd, s.sqlDatabasePatchCmd} {
		c.Flags().BoolVar(&s.flagSQLAsync, "async", false, "Return without waiting for the operation to complete")
	}
	for _, c := range []*cobra.Command{s.sqlDatabaseCreateCmd, s.sqlDatabasePatchCmd} {
		c.Flags().StringVar(&s.flagSQLCharset, "charset", "", "Database character set")
		c.Flags().StringVar(&s.flagSQLCollation, "collation", "", "Database collation")
	}
	s.sqlDatabasePatchCmd.Flags().StringVar(&s.flagSQLConfigFile, "config-file", "", "YAML/JSON file with a Database body")
	s.sqlDatabasesCmd.AddCommand(s.sqlDatabaseCreateCmd, s.sqlDatabaseDeleteCmd, s.sqlDatabaseDescribeCmd, s.sqlDatabaseListCmd, s.sqlDatabasePatchCmd)
	s.sqlCmd.AddCommand(s.sqlDatabasesCmd)

	// Export
	for _, c := range []*cobra.Command{s.sqlExportSQLCmd, s.sqlExportCSVCmd, s.sqlExportBAKCmd} {
		c.Flags().StringVar(&s.flagSQLDatabase, "database", "", "Database to export from")
		c.Flags().StringSliceVar(&s.flagSQLExportDBs, "databases", nil, "Databases to export (repeat)")
		c.Flags().StringSliceVar(&s.flagSQLTable, "table", nil, "Tables to export (SQL only)")
		c.Flags().BoolVar(&s.flagSQLOffload, "offload", false, "Perform a serverless export")
		c.Flags().StringVar(&s.flagSQLFormat, "format", "", "Output format")
		c.Flags().BoolVar(&s.flagSQLAsync, "async", false, "Return without waiting for the operation to complete")
	}
	s.sqlExportCSVCmd.Flags().StringVar(&s.flagSQLQuery, "query", "", "SELECT query for CSV export")
	s.sqlExportBAKCmd.Flags().StringVar(&s.flagSQLBAKType, "bak-type", "", "BAK type (FULL, DIFF, TLOG)")
	s.sqlExportCmd.AddCommand(s.sqlExportSQLCmd, s.sqlExportCSVCmd, s.sqlExportBAKCmd)
	s.sqlCmd.AddCommand(s.sqlExportCmd)

	// Flags
	s.sqlFlagsListCmd.Flags().StringVar(&s.flagSQLDBVersion, "database-version", "", "Database version filter")
	s.sqlFlagsListCmd.Flags().StringVar(&s.flagSQLFormat, "format", "", "Output format")
	s.sqlFlagsCmd.AddCommand(s.sqlFlagsListCmd)
	s.sqlCmd.AddCommand(s.sqlFlagsCmd)

	// Import
	for _, c := range []*cobra.Command{s.sqlImportSQLCmd, s.sqlImportCSVCmd, s.sqlImportBAKCmd} {
		c.Flags().StringVar(&s.flagSQLDatabase, "database", "", "Database to import into")
		c.Flags().StringVar(&s.flagSQLImportUser, "user", "", "Postgres user for the import")
		c.Flags().StringVar(&s.flagSQLFormat, "format", "", "Output format")
		c.Flags().BoolVar(&s.flagSQLAsync, "async", false, "Return without waiting for the operation to complete")
	}
	s.sqlImportCSVCmd.Flags().StringSliceVar(&s.flagSQLTable, "table", nil, "Target table (CSV import)")
	s.sqlImportCSVCmd.Flags().StringSliceVar(&s.flagSQLColumns, "columns", nil, "Columns to import (CSV)")
	s.sqlImportBAKCmd.Flags().StringVar(&s.flagSQLBAKType, "bak-type", "", "BAK type (FULL, DIFF, TLOG)")
	s.sqlImportBAKCmd.Flags().StringVar(&s.flagSQLEncryptionKey, "encryption-password", "", "PVK password for encrypted BAKs")
	s.sqlImportBAKCmd.Flags().StringVar(&s.flagSQLPvKPath, "pvk-path", "", "GCS path to the PVK for encrypted BAKs")
	s.sqlImportBAKCmd.Flags().StringVar(&s.flagSQLCertPath, "cert-path", "", "GCS path to the encryption certificate")
	s.sqlImportCmd.AddCommand(s.sqlImportSQLCmd, s.sqlImportCSVCmd, s.sqlImportBAKCmd)
	s.sqlCmd.AddCommand(s.sqlImportCmd)

	// Instances
	allInst := []*cobra.Command{
		s.sqlInstanceCreateCmd, s.sqlInstanceDeleteCmd, s.sqlInstanceDescribeCmd, s.sqlInstanceListCmd,
		s.sqlInstancePatchCmd, s.sqlInstanceCloneCmd, s.sqlInstanceFailoverCmd, s.sqlInstanceRestartCmd,
		s.sqlInstanceRestoreBackupCmd, s.sqlInstancePromoteReplicaCmd, s.sqlInstanceStartReplicaCmd,
		s.sqlInstanceStopReplicaCmd, s.sqlInstanceReencryptCmd, s.sqlInstanceResetSSLConfigCmd,
		s.sqlInstanceImportCmd, s.sqlInstanceExportCmd, s.sqlInstanceListServerCasCmd,
		s.sqlInstanceRescheduleMaintenanceCmd, s.sqlInstanceAcquireSsrsLeaseCmd, s.sqlInstanceReleaseSsrsLeaseCmd,
		s.sqlInstanceResetAsyncReplicaLagCmd, s.sqlInstanceVerifyExternalSyncSettingsCmd,
	}
	for _, c := range allInst {
		c.Flags().StringVar(&s.flagSQLFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{
		s.sqlInstanceCreateCmd, s.sqlInstanceDeleteCmd, s.sqlInstancePatchCmd, s.sqlInstanceCloneCmd,
		s.sqlInstanceFailoverCmd, s.sqlInstanceRestartCmd, s.sqlInstanceRestoreBackupCmd,
		s.sqlInstancePromoteReplicaCmd, s.sqlInstanceStartReplicaCmd, s.sqlInstanceStopReplicaCmd,
		s.sqlInstanceReencryptCmd, s.sqlInstanceResetSSLConfigCmd, s.sqlInstanceImportCmd, s.sqlInstanceExportCmd,
		s.sqlInstanceRescheduleMaintenanceCmd, s.sqlInstanceResetAsyncReplicaLagCmd,
	} {
		c.Flags().BoolVar(&s.flagSQLAsync, "async", false, "Return without waiting for the operation to complete")
	}
	for _, c := range []*cobra.Command{s.sqlInstanceCreateCmd, s.sqlInstancePatchCmd, s.sqlInstanceAcquireSsrsLeaseCmd, s.sqlInstanceVerifyExternalSyncSettingsCmd} {
		c.Flags().StringVar(&s.flagSQLConfigFile, "config-file", "", "YAML/JSON file with a DatabaseInstance body")
	}
	s.sqlInstanceCreateCmd.Flags().StringVar(&s.flagSQLDBVersion, "database-version", "", "Database version, e.g. POSTGRES_15")
	s.sqlInstanceCreateCmd.Flags().StringVar(&s.flagSQLRegion, "region", "", "Region for the instance")
	s.sqlInstanceCreateCmd.Flags().StringVar(&s.flagSQLTier, "tier", "", "Machine tier, e.g. db-custom-2-4096")
	s.sqlInstanceCreateCmd.Flags().StringVar(&s.flagSQLZone, "zone", "", "Zone preference for the instance")
	s.sqlInstanceCreateCmd.Flags().StringVar(&s.flagSQLPassword, "root-password", "", "Root password for the instance")
	for _, c := range []*cobra.Command{s.sqlInstanceCreateCmd, s.sqlInstancePatchCmd} {
		c.Flags().Int64Var(&s.flagSQLStorageAutoIncreaseLimit, "storage-auto-increase-limit", 0,
			"Maximum size (in GB) to which storage capacity can be automatically increased. "+
				"Zero means no limit; requires storage auto-increase to be enabled on the instance.")
	}
	s.sqlInstanceListCmd.Flags().StringVar(&s.flagSQLFilter, "filter", "", "Server-side filter expression")
	s.sqlInstanceFailoverCmd.Flags().Int64Var(&s.flagSQLSettingsVer, "settings-version", 0, "Current settings version of this instance")
	s.sqlInstanceRestoreBackupCmd.Flags().StringVar(&s.flagSQLBackup, "backup-id", "", "Backup run ID to restore from (required)")
	s.sqlInstanceRescheduleMaintenanceCmd.Flags().StringVar(&s.flagSQLReschedule, "reschedule-type", "IMMEDIATE", "Reschedule type (IMMEDIATE, NEXT_AVAILABLE_WINDOW, SPECIFIC_TIME)")
	s.sqlInstanceRescheduleMaintenanceCmd.Flags().StringVar(&s.flagSQLScheduleTime, "schedule-time", "", "RFC3339 timestamp for SPECIFIC_TIME reschedule")
	// Instances subgroup import/export mirror flag sets
	for _, c := range []*cobra.Command{s.sqlInstanceImportCmd, s.sqlInstanceExportCmd} {
		c.Flags().StringVar(&s.flagSQLDatabase, "database", "", "Database name")
		c.Flags().StringSliceVar(&s.flagSQLExportDBs, "databases", nil, "Databases to export")
		c.Flags().BoolVar(&s.flagSQLOffload, "offload", false, "Perform a serverless export")
	}
	s.sqlInstancesCmd.AddCommand(allInst...)
	s.sqlCmd.AddCommand(s.sqlInstancesCmd)

	// Operations
	for _, c := range []*cobra.Command{s.sqlOpCancelCmd, s.sqlOpDescribeCmd, s.sqlOpListCmd, s.sqlOpWaitCmd} {
		c.Flags().StringVar(&s.flagSQLFormat, "format", "", "Output format")
	}
	s.sqlOpListCmd.Flags().StringVar(&s.flagSQLInstance, "instance", "", "Restrict to operations on this instance")
	s.sqlOpListCmd.Flags().Int64Var(&s.flagSQLLimit, "limit", 0, "Maximum number of operations to return")
	s.sqlOperationsCmd.AddCommand(s.sqlOpCancelCmd, s.sqlOpDescribeCmd, s.sqlOpListCmd, s.sqlOpWaitCmd)
	s.sqlCmd.AddCommand(s.sqlOperationsCmd)

	// SSL server-ca-certs (group under `sql ssl server-ca-certs`)
	for _, c := range []*cobra.Command{s.sqlSSLServerCAListCmd, s.sqlSSLServerCAAddCmd, s.sqlSSLServerCARemoveCmd, s.sqlSSLServerCARotateCmd} {
		c.Flags().StringVar(&s.flagSQLInstance, "instance", "", "Cloud SQL instance ID (required)")
		c.Flags().StringVar(&s.flagSQLFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.sqlSSLServerCAAddCmd, s.sqlSSLServerCARemoveCmd, s.sqlSSLServerCARotateCmd} {
		c.Flags().BoolVar(&s.flagSQLAsync, "async", false, "Return without waiting for the operation to complete")
	}
	for _, c := range []*cobra.Command{s.sqlSSLServerCARemoveCmd, s.sqlSSLServerCARotateCmd} {
		c.Flags().StringVar(&s.flagSQLNextVersion, "next-version", "", "Fingerprint of the next server CA version to rotate to")
	}
	s.sqlSSLServerCACmd.AddCommand(s.sqlSSLServerCAListCmd, s.sqlSSLServerCAAddCmd, s.sqlSSLServerCARemoveCmd, s.sqlSSLServerCARotateCmd)
	s.sqlSSLCmd.AddCommand(s.sqlSSLServerCACmd)
	s.sqlCmd.AddCommand(s.sqlSSLCmd)

	// SSL certs (DEPRECATED)
	for _, c := range []*cobra.Command{s.sqlSSLCertCreateCmd, s.sqlSSLCertDeleteCmd, s.sqlSSLCertDescribeCmd, s.sqlSSLCertListCmd} {
		c.Flags().StringVar(&s.flagSQLInstance, "instance", "", "Cloud SQL instance ID (required)")
		c.Flags().StringVar(&s.flagSQLFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.sqlSSLCertCreateCmd, s.sqlSSLCertDeleteCmd} {
		c.Flags().BoolVar(&s.flagSQLAsync, "async", false, "Return without waiting for the operation to complete")
	}
	s.sqlSSLCertsCmd.AddCommand(s.sqlSSLCertCreateCmd, s.sqlSSLCertDeleteCmd, s.sqlSSLCertDescribeCmd, s.sqlSSLCertListCmd)
	s.sqlCmd.AddCommand(s.sqlSSLCertsCmd)

	// Tiers
	s.sqlTiersListCmd.Flags().StringVar(&s.flagSQLFormat, "format", "", "Output format")
	s.sqlTiersCmd.AddCommand(s.sqlTiersListCmd)
	s.sqlCmd.AddCommand(s.sqlTiersCmd)

	// Users
	for _, c := range []*cobra.Command{s.sqlUserCreateCmd, s.sqlUserDeleteCmd, s.sqlUserDescribeCmd, s.sqlUserListCmd, s.sqlUserSetPasswordCmd, s.sqlUserPatchCmd} {
		c.Flags().StringVar(&s.flagSQLInstance, "instance", "", "Cloud SQL instance ID (required)")
		c.Flags().StringVar(&s.flagSQLFormat, "format", "", "Output format")
		c.Flags().StringVar(&s.flagSQLHost, "host", "", "Host for the user (MySQL only)")
	}
	for _, c := range []*cobra.Command{s.sqlUserCreateCmd, s.sqlUserDeleteCmd, s.sqlUserSetPasswordCmd, s.sqlUserPatchCmd} {
		c.Flags().BoolVar(&s.flagSQLAsync, "async", false, "Return without waiting for the operation to complete")
	}
	s.sqlUserCreateCmd.Flags().StringVar(&s.flagSQLPassword, "password", "", "Password for the user")
	s.sqlUserCreateCmd.Flags().StringVar(&s.flagSQLType, "type", "", "User type (BUILT_IN, CLOUD_IAM_USER, CLOUD_IAM_SERVICE_ACCOUNT, ...)")
	s.sqlUserSetPasswordCmd.Flags().StringVar(&s.flagSQLPassword, "password", "", "New password (required)")
	s.sqlUserPatchCmd.Flags().StringVar(&s.flagSQLConfigFile, "config-file", "", "YAML/JSON file with a User body")
	s.sqlUserPatchCmd.Flags().StringVar(&s.flagSQLPassword, "password", "", "Password for the user")
	s.sqlUsersCmd.AddCommand(s.sqlUserCreateCmd, s.sqlUserDeleteCmd, s.sqlUserDescribeCmd, s.sqlUserListCmd, s.sqlUserSetPasswordCmd, s.sqlUserPatchCmd)
	s.sqlCmd.AddCommand(s.sqlUsersCmd)
}

func (s *sqlCmds) initSqlAll() {
	s.registerSQL()
}
