package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	sqladmin "google.golang.org/api/sqladmin/v1"
)

// --- Shared helpers for the Cloud SQL surface (#875-#886) ---

var (
	flagSQLInstance      string
	flagSQLConfigFile    string
	flagSQLFormat        string
	flagSQLAsync         bool
	flagSQLFilter        string
	flagSQLLimit         int64
	flagSQLPageSize      int64
	flagSQLDescription   string
	flagSQLLocation      string
	flagSQLProjectLevel  bool
	flagSQLBackup        string
	flagSQLURI           string
	flagSQLDatabase      string
	flagSQLDatabases     []string
	flagSQLTable         []string
	flagSQLColumns       []string
	flagSQLQuery         string
	flagSQLOffload       bool
	flagSQLUser          string
	flagSQLHost          string
	flagSQLPassword      string
	flagSQLType          string
	flagSQLCharset       string
	flagSQLCollation     string
	flagSQLDBVersion     string
	flagSQLTier          string
	flagSQLRegion        string
	flagSQLZone          string
	flagSQLPort          int
	flagSQLCommonName    string
	flagSQLCertKeyOutput string
	flagSQLSha1          string
	flagSQLVersion       string
	flagSQLNextVersion   string
	flagSQLReschedule    string
	flagSQLScheduleTime  string
	flagSQLPrivateIP     bool
	flagSQLPSC           bool
	flagSQLAutoIP        bool
	flagSQLSSHUser       string
	flagSQLSettingsVer   int64
	flagSQLStripeCount   int64
	flagSQLBAKType       string
	flagSQLExportUsers   []string
	flagSQLExportDBs     []string
	flagSQLImportUser    string
	flagSQLStartTime     string
	flagSQLEncryptionKey string
	flagSQLCertPath      string
	flagSQLPvKPath       string
	flagSQLCACertPath    string
)

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
func sqlFinishOp(ctx context.Context, svc *sqladmin.Service, project string, op *sqladmin.Operation, verb, target string) error {
	if flagSQLAsync {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, flagSQLFormat)
	}
	final, err := sqlWaitOp(ctx, svc, project, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, target)
	return emitFormatted(final, flagSQLFormat)
}

// sqlService returns an authenticated Cloud SQL Admin API client.
func sqlService(ctx context.Context) (*sqladmin.Service, error) {
	return gcp.SQLAdminService(ctx, flagAccount)
}

// requireInstance returns the resolved --instance flag or errors if unset.
func requireInstance() (string, error) {
	if flagSQLInstance == "" {
		return "", fmt.Errorf("--instance is required")
	}
	return flagSQLInstance, nil
}

// --- Top-level subgroups ---

var (
	sqlBackupsCmd     = &cobra.Command{Use: "backups", Short: "Manage Cloud SQL backups"}
	sqlConnectCmd     = &cobra.Command{Use: "connect", Short: "Connect to a Cloud SQL instance"}
	sqlDatabasesCmd   = &cobra.Command{Use: "databases", Short: "Manage Cloud SQL databases"}
	sqlExportCmd      = &cobra.Command{Use: "export", Short: "Export a Cloud SQL instance"}
	sqlFlagsCmd       = &cobra.Command{Use: "flags", Short: "List Cloud SQL flags"}
	sqlImportCmd      = &cobra.Command{Use: "import", Short: "Import into a Cloud SQL instance"}
	sqlInstancesCmd   = &cobra.Command{Use: "instances", Short: "Manage Cloud SQL instances"}
	sqlOperationsCmd  = &cobra.Command{Use: "operations", Short: "Manage Cloud SQL operations"}
	sqlSSLCmd         = &cobra.Command{Use: "ssl", Short: "Manage Cloud SQL SSL"}
	sqlSSLServerCACmd = &cobra.Command{Use: "server-ca-certs", Short: "Manage instance server CA certificates"}
	sqlSSLCertsCmd    = &cobra.Command{Use: "ssl-certs", Short: "(DEPRECATED) Manage Cloud SQL client SSL certificates"}
	sqlTiersCmd       = &cobra.Command{Use: "tiers", Short: "List Cloud SQL machine tiers"}
	sqlUsersCmd       = &cobra.Command{Use: "users", Short: "Manage Cloud SQL users"}
)

// --- backups ---

var (
	sqlBackupCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a Cloud SQL backup",
		Args: cobra.NoArgs, RunE: runSQLBackupCreate,
	}
	sqlBackupDeleteCmd = &cobra.Command{
		Use: "delete ID", Short: "Delete a Cloud SQL backup",
		Args: cobra.ExactArgs(1), RunE: runSQLBackupDelete,
	}
	sqlBackupDescribeCmd = &cobra.Command{
		Use: "describe ID", Short: "Describe a Cloud SQL backup",
		Args: cobra.ExactArgs(1), RunE: runSQLBackupDescribe,
	}
	sqlBackupListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud SQL backups",
		Args: cobra.NoArgs, RunE: runSQLBackupList,
	}
	sqlBackupRestoreCmd = &cobra.Command{
		Use: "restore BACKUP_ID", Short: "Restore a Cloud SQL instance from a backup",
		Args: cobra.ExactArgs(1), RunE: runSQLBackupRestore,
	}
)

func runSQLBackupCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	run := &sqladmin.BackupRun{
		Description: flagSQLDescription,
		Instance:    instance,
		Location:    flagSQLLocation,
		Kind:        "sql#backupRun",
	}
	op, err := svc.BackupRuns.Insert(project, instance, run).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Create backup", instance)
}

func runSQLBackupDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("backup ID must be an integer: %w", err)
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.BackupRuns.Delete(project, instance, id).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Delete backup", args[0])
}

func runSQLBackupDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("backup ID must be an integer: %w", err)
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.BackupRuns.Get(project, instance, id).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormatted(got, flagSQLFormat)
}

func runSQLBackupList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	call := svc.BackupRuns.List(project, instance)
	if flagSQLLimit > 0 {
		call = call.MaxResults(flagSQLLimit)
	}
	resp, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing backups: %w", err)
	}
	if flagSQLFormat != "" {
		return emitFormatted(resp.Items, flagSQLFormat)
	}
	fmt.Printf("%-20s %-24s %-10s %s\n", "ID", "WINDOW_START_TIME", "STATUS", "TYPE")
	for _, b := range resp.Items {
		fmt.Printf("%-20d %-24s %-10s %s\n", b.Id, b.WindowStartTime, b.Status, b.Type)
	}
	return nil
}

func runSQLBackupRestore(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("backup ID must be an integer: %w", err)
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	req := &sqladmin.InstancesRestoreBackupRequest{
		RestoreBackupContext: &sqladmin.RestoreBackupContext{
			BackupRunId:      id,
			Kind:             "sql#restoreBackupContext",
			InstanceId:       instance,
			Project:          project,
		},
	}
	op, err := svc.Instances.RestoreBackup(project, instance, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring backup: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Restore backup", instance)
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
func runSQLConnect(bin string, cmd *cobra.Command, args []string) error {
	instance := args[0]
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
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
		fmt.Fprintf(os.Stderr, "warning: instance database is %s but connecting with %s\n", inst.DatabaseVersion, bin)
		dbBin = bin
	}
	if _, err := exec.LookPath(dbBin); err != nil {
		return fmt.Errorf("%s client not found in PATH: install a %s client and retry", dbBin, dbBin)
	}
	user := flagSQLUser
	if user == "" {
		user = defaultUser
	}
	var execArgs []string
	switch dbBin {
	case "psql":
		execArgs = []string{"-h", ip, "-U", user}
		if flagSQLDatabase != "" {
			execArgs = append(execArgs, "-d", flagSQLDatabase)
		}
	case "mysql":
		execArgs = []string{"-h", ip, "-u", user, "-p"}
		if flagSQLDatabase != "" {
			execArgs = append(execArgs, flagSQLDatabase)
		}
	case "mssql-cli":
		execArgs = []string{"-S", ip, "-U", user}
		if flagSQLDatabase != "" {
			execArgs = append(execArgs, "-d", flagSQLDatabase)
		}
	}
	c := exec.CommandContext(ctx, dbBin, execArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	fmt.Fprintf(os.Stderr, "Connecting to %s at %s using %s...\n", instance, ip, dbBin)
	return c.Run()
}

var (
	sqlConnectPsqlCmd = &cobra.Command{
		Use: "psql INSTANCE", Short: "Connect to a PostgreSQL Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
			return runSQLConnect("psql", cmd, args)
		},
	}
	sqlConnectMysqlCmd = &cobra.Command{
		Use: "mysql INSTANCE", Short: "Connect to a MySQL Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
			return runSQLConnect("mysql", cmd, args)
		},
	}
	sqlConnectSqlserverCmd = &cobra.Command{
		Use: "sqlserver INSTANCE", Short: "Connect to a SQL Server Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
			return runSQLConnect("mssql-cli", cmd, args)
		},
	}
)

// --- databases ---

var (
	sqlDatabaseCreateCmd = &cobra.Command{
		Use: "create DATABASE", Short: "Create a database on a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: runSQLDatabaseCreate,
	}
	sqlDatabaseDeleteCmd = &cobra.Command{
		Use: "delete DATABASE", Short: "Delete a database from a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: runSQLDatabaseDelete,
	}
	sqlDatabaseDescribeCmd = &cobra.Command{
		Use: "describe DATABASE", Short: "Describe a database",
		Args: cobra.ExactArgs(1), RunE: runSQLDatabaseDescribe,
	}
	sqlDatabaseListCmd = &cobra.Command{
		Use: "list", Short: "List databases on a Cloud SQL instance",
		Args: cobra.NoArgs, RunE: runSQLDatabaseList,
	}
	sqlDatabasePatchCmd = &cobra.Command{
		Use: "patch DATABASE", Short: "Patch a database on a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: runSQLDatabasePatch,
	}
)

func runSQLDatabaseCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	db := &sqladmin.Database{
		Name:      args[0],
		Instance:  instance,
		Project:   project,
		Charset:   flagSQLCharset,
		Collation: flagSQLCollation,
		Kind:      "sql#database",
	}
	op, err := svc.Databases.Insert(project, instance, db).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating database: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Create database", args[0])
}

func runSQLDatabaseDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Databases.Delete(project, instance, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting database: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Delete database", args[0])
}

func runSQLDatabaseDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Databases.Get(project, instance, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing database: %w", err)
	}
	return emitFormatted(got, flagSQLFormat)
}

func runSQLDatabaseList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Databases.List(project, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing databases: %w", err)
	}
	if flagSQLFormat != "" {
		return emitFormatted(resp.Items, flagSQLFormat)
	}
	fmt.Printf("%-32s %-16s %s\n", "NAME", "CHARSET", "COLLATION")
	for _, d := range resp.Items {
		fmt.Printf("%-32s %-16s %s\n", d.Name, d.Charset, d.Collation)
	}
	return nil
}

func runSQLDatabasePatch(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	db := &sqladmin.Database{}
	if flagSQLConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagSQLConfigFile, db); err != nil {
			return err
		}
	}
	if flagSQLCharset != "" {
		db.Charset = flagSQLCharset
	}
	if flagSQLCollation != "" {
		db.Collation = flagSQLCollation
	}
	db.Name = args[0]
	db.Instance = instance
	db.Project = project
	op, err := svc.Databases.Patch(project, instance, args[0], db).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("patching database: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Patch database", args[0])
}

// --- export ---

var (
	sqlExportSQLCmd = &cobra.Command{
		Use: "sql INSTANCE URI", Short: "Export a Cloud SQL instance as SQL statements to Cloud Storage",
		Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
			return runSQLExport("SQL", cmd, args)
		},
	}
	sqlExportCSVCmd = &cobra.Command{
		Use: "csv INSTANCE URI --query=QUERY --database=DB", Short: "Export a Cloud SQL instance as CSV to Cloud Storage",
		Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
			return runSQLExport("CSV", cmd, args)
		},
	}
	sqlExportBAKCmd = &cobra.Command{
		Use: "bak INSTANCE URI --database=DB", Short: "Export a SQL Server instance as a BAK to Cloud Storage",
		Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
			return runSQLExport("BAK", cmd, args)
		},
	}
)

func runSQLExport(fileType string, cmd *cobra.Command, args []string) error {
	instance := args[0]
	uri := args[1]
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	ec := &sqladmin.ExportContext{
		FileType:  fileType,
		Uri:       uri,
		Kind:      "sql#exportContext",
		Databases: flagSQLExportDBs,
		Offload:   flagSQLOffload,
	}
	if len(ec.Databases) == 0 && flagSQLDatabase != "" {
		ec.Databases = []string{flagSQLDatabase}
	}
	switch fileType {
	case "SQL":
		if len(flagSQLTable) > 0 {
			ec.SqlExportOptions = &sqladmin.ExportContextSqlExportOptions{Tables: flagSQLTable}
		}
	case "CSV":
		ec.CsvExportOptions = &sqladmin.ExportContextCsvExportOptions{SelectQuery: flagSQLQuery}
	case "BAK":
		if flagSQLBAKType != "" {
			ec.BakExportOptions = &sqladmin.ExportContextBakExportOptions{BakType: flagSQLBAKType}
		}
	}
	op, err := svc.Instances.Export(project, instance, &sqladmin.InstancesExportRequest{ExportContext: ec}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting instance: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Export instance", instance)
}

// --- flags ---

var sqlFlagsListCmd = &cobra.Command{
	Use: "list", Short: "List available database flags",
	Args: cobra.NoArgs, RunE: runSQLFlagsList,
}

func runSQLFlagsList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	call := svc.Flags.List()
	if flagSQLDBVersion != "" {
		call = call.DatabaseVersion(flagSQLDBVersion)
	}
	resp, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing flags: %w", err)
	}
	if flagSQLFormat != "" {
		return emitFormatted(resp.Items, flagSQLFormat)
	}
	fmt.Printf("%-40s %-12s %s\n", "NAME", "TYPE", "APPLIES_TO")
	for _, f := range resp.Items {
		fmt.Printf("%-40s %-12s %s\n", f.Name, f.Type, strings.Join(f.AppliesTo, ","))
	}
	return nil
}

// --- import ---

var (
	sqlImportSQLCmd = &cobra.Command{
		Use: "sql INSTANCE URI", Short: "Import a SQL dump from Cloud Storage into a Cloud SQL instance",
		Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
			return runSQLImport("SQL", cmd, args)
		},
	}
	sqlImportCSVCmd = &cobra.Command{
		Use: "csv INSTANCE URI --database=DB --table=TABLE", Short: "Import CSV data from Cloud Storage into a Cloud SQL database",
		Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
			return runSQLImport("CSV", cmd, args)
		},
	}
	sqlImportBAKCmd = &cobra.Command{
		Use: "bak INSTANCE URI --database=DB", Short: "Import a SQL Server .BAK from Cloud Storage",
		Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
			return runSQLImport("BAK", cmd, args)
		},
	}
)

func runSQLImport(fileType string, cmd *cobra.Command, args []string) error {
	instance := args[0]
	uri := args[1]
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	ic := &sqladmin.ImportContext{
		FileType:   fileType,
		Uri:        uri,
		Kind:       "sql#importContext",
		Database:   flagSQLDatabase,
		ImportUser: flagSQLImportUser,
	}
	switch fileType {
	case "CSV":
		ic.CsvImportOptions = &sqladmin.ImportContextCsvImportOptions{
			Table:   firstNonEmpty(flagSQLTable),
			Columns: flagSQLColumns,
		}
	case "BAK":
		bak := &sqladmin.ImportContextBakImportOptions{BakType: flagSQLBAKType}
		if flagSQLEncryptionKey != "" {
			bak.EncryptionOptions = &sqladmin.ImportContextBakImportOptionsEncryptionOptions{
				PvkPath:  flagSQLPvKPath,
				CertPath: flagSQLCertPath,
				PvkPassword: flagSQLEncryptionKey,
			}
		}
		ic.BakImportOptions = bak
	}
	op, err := svc.Instances.Import(project, instance, &sqladmin.InstancesImportRequest{ImportContext: ic}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing to instance: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Import into instance", instance)
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

var (
	sqlInstanceCreateCmd = &cobra.Command{
		Use: "create INSTANCE", Short: "Create a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: runSQLInstanceCreate,
	}
	sqlInstanceDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: runSQLInstanceDelete,
	}
	sqlInstanceDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: runSQLInstanceDescribe,
	}
	sqlInstanceListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud SQL instances in the project",
		Args: cobra.NoArgs, RunE: runSQLInstanceList,
	}
	sqlInstancePatchCmd = &cobra.Command{
		Use: "patch INSTANCE", Short: "Patch a Cloud SQL instance from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSQLInstancePatch,
	}
	sqlInstanceCloneCmd = &cobra.Command{
		Use: "clone SOURCE DESTINATION", Short: "Clone a Cloud SQL instance",
		Args: cobra.ExactArgs(2), RunE: runSQLInstanceClone,
	}
	sqlInstanceFailoverCmd = &cobra.Command{
		Use: "failover INSTANCE", Short: "Failover a Cloud SQL instance to its replica",
		Args: cobra.ExactArgs(1), RunE: runSQLInstanceFailover,
	}
	sqlInstanceRestartCmd = &cobra.Command{
		Use: "restart INSTANCE", Short: "Restart a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: runSQLInstanceRestart,
	}
	sqlInstanceRestoreBackupCmd = &cobra.Command{
		Use: "restore-backup INSTANCE", Short: "Restore a Cloud SQL instance from a backup",
		Args: cobra.ExactArgs(1), RunE: runSQLInstanceRestoreBackup,
	}
	sqlInstancePromoteReplicaCmd = &cobra.Command{
		Use: "promote-replica INSTANCE", Short: "Promote a Cloud SQL replica to a standalone instance",
		Args: cobra.ExactArgs(1), RunE: runSQLInstancePromoteReplica,
	}
	sqlInstanceStartReplicaCmd = &cobra.Command{
		Use: "start-replica INSTANCE", Short: "Start replication on a Cloud SQL read replica",
		Args: cobra.ExactArgs(1), RunE: runSQLInstanceStartReplica,
	}
	sqlInstanceStopReplicaCmd = &cobra.Command{
		Use: "stop-replica INSTANCE", Short: "Stop replication on a Cloud SQL read replica",
		Args: cobra.ExactArgs(1), RunE: runSQLInstanceStopReplica,
	}
	sqlInstanceReencryptCmd = &cobra.Command{
		Use: "reencrypt INSTANCE", Short: "Re-encrypt a CMEK-enabled Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: runSQLInstanceReencrypt,
	}
	sqlInstanceResetSSLConfigCmd = &cobra.Command{
		Use: "reset-ssl-config INSTANCE", Short: "Reset the SSL configuration of a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: runSQLInstanceResetSSLConfig,
	}
	sqlInstanceImportCmd = &cobra.Command{
		Use: "import INSTANCE URI", Short: "Import a SQL dump into a Cloud SQL instance",
		Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
			return runSQLImport("SQL", cmd, args)
		},
	}
	sqlInstanceExportCmd = &cobra.Command{
		Use: "export INSTANCE URI", Short: "Export a Cloud SQL instance to Cloud Storage as SQL",
		Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
			return runSQLExport("SQL", cmd, args)
		},
	}
	sqlInstanceListServerCasCmd = &cobra.Command{
		Use: "list-server-cas INSTANCE", Short: "List server CA certificates for a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: runSQLInstanceListServerCas,
	}
	sqlInstanceRescheduleMaintenanceCmd = &cobra.Command{
		Use: "reschedule-maintenance INSTANCE", Short: "Reschedule a Cloud SQL instance's maintenance",
		Args: cobra.ExactArgs(1), RunE: runSQLInstanceRescheduleMaintenance,
	}
	sqlInstanceAcquireSsrsLeaseCmd = &cobra.Command{
		Use: "acquire-ssrs-lease INSTANCE", Short: "Acquire an SSRS lease on a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: runSQLInstanceAcquireSsrsLease,
	}
	sqlInstanceReleaseSsrsLeaseCmd = &cobra.Command{
		Use: "release-ssrs-lease INSTANCE", Short: "Release an SSRS lease on a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: runSQLInstanceReleaseSsrsLease,
	}
	sqlInstanceResetAsyncReplicaLagCmd = &cobra.Command{
		Use: "reset-async-replica-lag INSTANCE", Short: "Reset asynchronous replica lag (alias of reset-replica-size)",
		Args: cobra.ExactArgs(1), RunE: runSQLInstanceResetAsyncReplicaLag,
	}
	sqlInstanceVerifyExternalSyncSettingsCmd = &cobra.Command{
		Use: "verify-external-sync-settings INSTANCE", Short: "Verify external sync settings for a Cloud SQL instance",
		Args: cobra.ExactArgs(1), RunE: runSQLInstanceVerifyExternalSyncSettings,
	}
)

func runSQLInstanceCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	inst := &sqladmin.DatabaseInstance{Kind: "sql#instance"}
	if flagSQLConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagSQLConfigFile, inst); err != nil {
			return err
		}
	}
	inst.Name = args[0]
	if flagSQLDBVersion != "" {
		inst.DatabaseVersion = flagSQLDBVersion
	}
	if flagSQLRegion != "" {
		inst.Region = flagSQLRegion
	}
	if flagSQLTier != "" {
		if inst.Settings == nil {
			inst.Settings = &sqladmin.Settings{}
		}
		inst.Settings.Tier = flagSQLTier
	}
	if flagSQLZone != "" {
		if inst.Settings == nil {
			inst.Settings = &sqladmin.Settings{}
		}
		if inst.Settings.LocationPreference == nil {
			inst.Settings.LocationPreference = &sqladmin.LocationPreference{}
		}
		inst.Settings.LocationPreference.Zone = flagSQLZone
	}
	if flagSQLPassword != "" {
		inst.RootPassword = flagSQLPassword
	}
	op, err := svc.Instances.Insert(project, inst).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Create instance", args[0])
}

func runSQLInstanceDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Instances.Delete(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Delete instance", args[0])
}

func runSQLInstanceDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Instances.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormatted(got, flagSQLFormat)
}

func runSQLInstanceList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	call := svc.Instances.List(project)
	if flagSQLFilter != "" {
		call = call.Filter(flagSQLFilter)
	}
	resp, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing instances: %w", err)
	}
	if flagSQLFormat != "" {
		return emitFormatted(resp.Items, flagSQLFormat)
	}
	fmt.Printf("%-32s %-16s %-14s %-14s %s\n", "NAME", "DATABASE_VERSION", "REGION", "TIER", "STATE")
	for _, i := range resp.Items {
		tier := ""
		if i.Settings != nil {
			tier = i.Settings.Tier
		}
		fmt.Printf("%-32s %-16s %-14s %-14s %s\n", i.Name, i.DatabaseVersion, i.Region, tier, i.State)
	}
	return nil
}

func runSQLInstancePatch(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	if flagSQLConfigFile == "" {
		return fmt.Errorf("--config-file is required")
	}
	inst := &sqladmin.DatabaseInstance{}
	if err := loadYAMLOrJSONInto(flagSQLConfigFile, inst); err != nil {
		return err
	}
	inst.Name = args[0]
	op, err := svc.Instances.Patch(project, args[0], inst).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("patching instance: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Patch instance", args[0])
}

func runSQLInstanceClone(cmd *cobra.Command, args []string) error {
	source, dest := args[0], args[1]
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
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
	return sqlFinishOp(ctx, svc, project, op, "Clone instance", dest)
}

func runSQLInstanceFailover(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	req := &sqladmin.InstancesFailoverRequest{FailoverContext: &sqladmin.FailoverContext{
		Kind:            "sql#failoverContext",
		SettingsVersion: flagSQLSettingsVer,
	}}
	op, err := svc.Instances.Failover(project, args[0], req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failing over instance: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Failover instance", args[0])
}

func runSQLInstanceRestart(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Instances.Restart(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restarting instance: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Restart instance", args[0])
}

func runSQLInstanceRestoreBackup(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	if flagSQLBackup == "" {
		return fmt.Errorf("--backup-id is required")
	}
	id, err := strconv.ParseInt(flagSQLBackup, 10, 64)
	if err != nil {
		return fmt.Errorf("--backup-id must be an integer: %w", err)
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
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
	return sqlFinishOp(ctx, svc, project, op, "Restore backup", args[0])
}

func runSQLInstancePromoteReplica(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Instances.PromoteReplica(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("promoting replica: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Promote replica", args[0])
}

func runSQLInstanceStartReplica(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Instances.StartReplica(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting replica: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Start replica", args[0])
}

func runSQLInstanceStopReplica(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Instances.StopReplica(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping replica: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Stop replica", args[0])
}

func runSQLInstanceReencrypt(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	req := &sqladmin.InstancesReencryptRequest{}
	op, err := svc.Instances.Reencrypt(project, args[0], req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("re-encrypting instance: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Re-encrypt instance", args[0])
}

func runSQLInstanceResetSSLConfig(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Instances.ResetSslConfig(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resetting SSL config: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Reset SSL config", args[0])
}

func runSQLInstanceListServerCas(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Instances.ListServerCas(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing server CAs: %w", err)
	}
	return emitFormatted(resp, flagSQLFormat)
}

func runSQLInstanceRescheduleMaintenance(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	body := &sqladmin.SqlInstancesRescheduleMaintenanceRequestBody{
		Reschedule: &sqladmin.Reschedule{
			RescheduleType: flagSQLReschedule,
			ScheduleTime:   flagSQLScheduleTime,
		},
	}
	op, err := svc.Projects.Instances.RescheduleMaintenance(project, args[0], body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rescheduling maintenance: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Reschedule maintenance", args[0])
}

func runSQLInstanceAcquireSsrsLease(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	req := &sqladmin.InstancesAcquireSsrsLeaseRequest{}
	if flagSQLConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagSQLConfigFile, req); err != nil {
			return err
		}
	}
	resp, err := svc.Instances.AcquireSsrsLease(project, args[0], req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("acquiring SSRS lease: %w", err)
	}
	return emitFormatted(resp, flagSQLFormat)
}

func runSQLInstanceReleaseSsrsLease(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Instances.ReleaseSsrsLease(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("releasing SSRS lease: %w", err)
	}
	return emitFormatted(resp, flagSQLFormat)
}

func runSQLInstanceResetAsyncReplicaLag(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	req := &sqladmin.SqlInstancesResetReplicaSizeRequest{}
	op, err := svc.Projects.Instances.ResetReplicaSize(project, args[0], req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resetting replica size: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Reset replica size", args[0])
}

func runSQLInstanceVerifyExternalSyncSettings(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	req := &sqladmin.SqlInstancesVerifyExternalSyncSettingsRequest{}
	if flagSQLConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagSQLConfigFile, req); err != nil {
			return err
		}
	}
	resp, err := svc.Projects.Instances.VerifyExternalSyncSettings(project, args[0], req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("verifying external sync settings: %w", err)
	}
	return emitFormatted(resp, flagSQLFormat)
}

// --- operations ---

var (
	sqlOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Cloud SQL operation",
		Args: cobra.ExactArgs(1), RunE: runSQLOpCancel,
	}
	sqlOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Cloud SQL operation",
		Args: cobra.ExactArgs(1), RunE: runSQLOpDescribe,
	}
	sqlOpListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud SQL operations",
		Args: cobra.NoArgs, RunE: runSQLOpList,
	}
	sqlOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for a Cloud SQL operation to finish",
		Args: cobra.ExactArgs(1), RunE: runSQLOpWait,
	}
)

func runSQLOpCancel(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Operations.Cancel(project, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Cancelled operation [%s].\n", args[0])
	return nil
}

func runSQLOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Operations.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagSQLFormat)
}

func runSQLOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	call := svc.Operations.List(project)
	if flagSQLInstance != "" {
		call = call.Instance(flagSQLInstance)
	}
	if flagSQLLimit > 0 {
		call = call.MaxResults(flagSQLLimit)
	}
	resp, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if flagSQLFormat != "" {
		return emitFormatted(resp.Items, flagSQLFormat)
	}
	fmt.Printf("%-40s %-14s %-24s %s\n", "NAME", "TYPE", "START", "STATUS")
	for _, o := range resp.Items {
		fmt.Printf("%-40s %-14s %-24s %s\n", o.Name, o.OperationType, o.StartTime, o.Status)
	}
	return nil
}

func runSQLOpWait(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
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
	return emitFormatted(final, flagSQLFormat)
}

// --- ssl server-ca-certs ---

var (
	sqlSSLServerCAListCmd = &cobra.Command{
		Use: "list", Short: "List server CA certificates for a Cloud SQL instance",
		Args: cobra.NoArgs, RunE: runSQLSSLServerCAList,
	}
	sqlSSLServerCAAddCmd = &cobra.Command{
		Use: "add", Short: "Add a new server CA certificate to a Cloud SQL instance",
		Args: cobra.NoArgs, RunE: runSQLSSLServerCAAdd,
	}
	sqlSSLServerCARemoveCmd = &cobra.Command{
		Use: "remove", Short: "Remove a server CA certificate from a Cloud SQL instance",
		Args: cobra.NoArgs, RunE: runSQLSSLServerCARemove,
	}
	sqlSSLServerCARotateCmd = &cobra.Command{
		Use: "rotate", Short: "Rotate the server CA certificate for a Cloud SQL instance",
		Args: cobra.NoArgs, RunE: runSQLSSLServerCARotate,
	}
)

func runSQLSSLServerCAList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Instances.ListServerCas(project, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing server CAs: %w", err)
	}
	return emitFormatted(resp, flagSQLFormat)
}

func runSQLSSLServerCAAdd(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Instances.AddServerCa(project, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding server CA: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Add server CA", instance)
}

// runSQLSSLServerCARemove maps gcloud's `remove SHA1_FINGERPRINT` semantics
// onto RotateServerCa, which retires the currently active CA in favor of the
// next-most-recent one.
func runSQLSSLServerCARemove(cmd *cobra.Command, args []string) error {
	// The upstream Cloud SQL Admin API does not expose an explicit
	// "remove server CA" endpoint; server CAs are retired by a rotate.
	// We mirror gcloud python's behavior by delegating to rotate with the
	// caller-supplied --next-version (or the most recently added CA).
	return runSQLSSLServerCARotate(cmd, args)
}

func runSQLSSLServerCARotate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	req := &sqladmin.InstancesRotateServerCaRequest{RotateServerCaContext: &sqladmin.RotateServerCaContext{
		Kind:        "sql#rotateServerCaContext",
		NextVersion: flagSQLNextVersion,
	}}
	op, err := svc.Instances.RotateServerCa(project, instance, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rotating server CA: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Rotate server CA", instance)
}

// --- ssl-certs (DEPRECATED) ---

var (
	sqlSSLCertCreateCmd = &cobra.Command{
		Use: "create COMMON_NAME CERT_FILE", Short: "(DEPRECATED) Create a Cloud SQL client SSL certificate",
		Args: cobra.ExactArgs(2), RunE: runSQLSSLCertCreate,
	}
	sqlSSLCertDeleteCmd = &cobra.Command{
		Use: "delete COMMON_NAME", Short: "(DEPRECATED) Delete a Cloud SQL client SSL certificate",
		Args: cobra.ExactArgs(1), RunE: runSQLSSLCertDelete,
	}
	sqlSSLCertDescribeCmd = &cobra.Command{
		Use: "describe COMMON_NAME", Short: "(DEPRECATED) Describe a Cloud SQL client SSL certificate",
		Args: cobra.ExactArgs(1), RunE: runSQLSSLCertDescribe,
	}
	sqlSSLCertListCmd = &cobra.Command{
		Use: "list", Short: "(DEPRECATED) List Cloud SQL client SSL certificates",
		Args: cobra.NoArgs, RunE: runSQLSSLCertList,
	}
)

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

func runSQLSSLCertCreate(cmd *cobra.Command, args []string) error {
	commonName, certFile := args[0], args[1]
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
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
		fmt.Fprintf(os.Stderr, "Wrote private key to %s.\n", certFile)
	}
	return emitFormatted(resp, flagSQLFormat)
}

func runSQLSSLCertDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
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
	return sqlFinishOp(ctx, svc, project, op, "Delete ssl-cert", args[0])
}

func runSQLSSLCertDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
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
	return emitFormatted(got, flagSQLFormat)
}

func runSQLSSLCertList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.SslCerts.List(project, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing ssl-certs: %w", err)
	}
	if flagSQLFormat != "" {
		return emitFormatted(resp.Items, flagSQLFormat)
	}
	fmt.Printf("%-30s %-40s %s\n", "COMMON_NAME", "SHA1_FINGERPRINT", "EXPIRATION_TIME")
	for _, c := range resp.Items {
		fmt.Printf("%-30s %-40s %s\n", c.CommonName, c.Sha1Fingerprint, c.ExpirationTime)
	}
	return nil
}

// --- tiers ---

var sqlTiersListCmd = &cobra.Command{
	Use: "list", Short: "List available Cloud SQL tiers",
	Args: cobra.NoArgs, RunE: runSQLTiersList,
}

func runSQLTiersList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Tiers.List(project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing tiers: %w", err)
	}
	if flagSQLFormat != "" {
		return emitFormatted(resp.Items, flagSQLFormat)
	}
	fmt.Printf("%-30s %-12s %-14s %s\n", "TIER", "AVAILABLE_REGIONS", "RAM_BYTES", "DISK_BYTES")
	for _, t := range resp.Items {
		fmt.Printf("%-30s %-12d %-14d %d\n", t.Tier, len(t.Region), t.RAM, t.DiskQuota)
	}
	return nil
}

// --- users ---

var (
	sqlUserCreateCmd = &cobra.Command{
		Use: "create USERNAME", Short: "Create a Cloud SQL user",
		Args: cobra.ExactArgs(1), RunE: runSQLUserCreate,
	}
	sqlUserDeleteCmd = &cobra.Command{
		Use: "delete USERNAME", Short: "Delete a Cloud SQL user",
		Args: cobra.ExactArgs(1), RunE: runSQLUserDelete,
	}
	sqlUserDescribeCmd = &cobra.Command{
		Use: "describe USERNAME", Short: "Describe a Cloud SQL user",
		Args: cobra.ExactArgs(1), RunE: runSQLUserDescribe,
	}
	sqlUserListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud SQL users on an instance",
		Args: cobra.NoArgs, RunE: runSQLUserList,
	}
	sqlUserSetPasswordCmd = &cobra.Command{
		Use: "set-password USERNAME", Short: "Set the password for a Cloud SQL user",
		Args: cobra.ExactArgs(1), RunE: runSQLUserSetPassword,
	}
	sqlUserPatchCmd = &cobra.Command{
		Use: "patch USERNAME", Short: "Patch a Cloud SQL user from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSQLUserPatch,
	}
)

func runSQLUserCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	u := &sqladmin.User{
		Name:     args[0],
		Instance: instance,
		Project:  project,
		Password: flagSQLPassword,
		Host:     flagSQLHost,
		Type:     flagSQLType,
		Kind:     "sql#user",
	}
	op, err := svc.Users.Insert(project, instance, u).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Create user", args[0])
}

func runSQLUserDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	call := svc.Users.Delete(project, instance).Name(args[0])
	if flagSQLHost != "" {
		call = call.Host(flagSQLHost)
	}
	op, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Delete user", args[0])
}

func runSQLUserDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	call := svc.Users.Get(project, instance, args[0])
	if flagSQLHost != "" {
		call = call.Host(flagSQLHost)
	}
	got, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing user: %w", err)
	}
	return emitFormatted(got, flagSQLFormat)
}

func runSQLUserList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Users.List(project, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing users: %w", err)
	}
	if flagSQLFormat != "" {
		return emitFormatted(resp.Items, flagSQLFormat)
	}
	fmt.Printf("%-30s %-16s %s\n", "NAME", "HOST", "TYPE")
	for _, u := range resp.Items {
		fmt.Printf("%-30s %-16s %s\n", u.Name, u.Host, u.Type)
	}
	return nil
}

func runSQLUserSetPassword(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	if flagSQLPassword == "" {
		return fmt.Errorf("--password is required")
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	body := &sqladmin.User{
		Name:     args[0],
		Instance: instance,
		Project:  project,
		Password: flagSQLPassword,
	}
	call := svc.Users.Update(project, instance, body).Name(args[0])
	if flagSQLHost != "" {
		call = call.Host(flagSQLHost)
	}
	op, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting password: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Set password", args[0])
}

func runSQLUserPatch(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	instance, err := requireInstance()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := sqlService(ctx)
	if err != nil {
		return err
	}
	body := &sqladmin.User{}
	if flagSQLConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagSQLConfigFile, body); err != nil {
			return err
		}
	}
	body.Name = args[0]
	body.Instance = instance
	body.Project = project
	if flagSQLPassword != "" {
		body.Password = flagSQLPassword
	}
	call := svc.Users.Update(project, instance, body).Name(args[0])
	if flagSQLHost != "" {
		call = call.Host(flagSQLHost)
	} else if body.Host != "" {
		call = call.Host(body.Host)
	}
	op, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("patching user: %w", err)
	}
	return sqlFinishOp(ctx, svc, project, op, "Patch user", args[0])
}

// --- flag registration ---

// registerSQL wires all Cloud SQL subgroups, subcommands and shared flags.
// Called from init().
func registerSQL() {
	// Backups
	for _, c := range []*cobra.Command{sqlBackupCreateCmd, sqlBackupDeleteCmd, sqlBackupDescribeCmd, sqlBackupListCmd, sqlBackupRestoreCmd} {
		c.Flags().StringVar(&flagSQLInstance, "instance", "", "Cloud SQL instance ID (required)")
		c.Flags().StringVar(&flagSQLFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{sqlBackupCreateCmd, sqlBackupDeleteCmd, sqlBackupRestoreCmd} {
		c.Flags().BoolVar(&flagSQLAsync, "async", false, "Return without waiting for the operation to complete")
	}
	sqlBackupCreateCmd.Flags().StringVar(&flagSQLDescription, "description", "", "Description of the backup")
	sqlBackupCreateCmd.Flags().StringVar(&flagSQLLocation, "location", "", "Region for storing the backup")
	sqlBackupListCmd.Flags().Int64Var(&flagSQLLimit, "limit", 0, "Maximum number of backups to return")
	sqlBackupsCmd.AddCommand(sqlBackupCreateCmd, sqlBackupDeleteCmd, sqlBackupDescribeCmd, sqlBackupListCmd, sqlBackupRestoreCmd)
	sqlCmd.AddCommand(sqlBackupsCmd)

	// Connect
	for _, c := range []*cobra.Command{sqlConnectPsqlCmd, sqlConnectMysqlCmd, sqlConnectSqlserverCmd} {
		c.Flags().StringVarP(&flagSQLUser, "user", "u", "", "Database user to connect as")
		c.Flags().StringVar(&flagSQLDatabase, "database", "", "Database to connect to")
		c.Flags().IntVar(&flagSQLPort, "port", 0, "Port that gcloud uses locally to connect through Cloud SQL Proxy")
		c.Flags().BoolVar(&flagSQLPrivateIP, "private-ip", false, "Use the instance's private IP")
		c.Flags().BoolVar(&flagSQLPSC, "psc", false, "Use Private Service Connect (PSC) to reach the instance")
		c.Flags().BoolVar(&flagSQLAutoIP, "auto-ip", false, "Auto-select an IP address")
	}
	sqlConnectCmd.AddCommand(sqlConnectPsqlCmd, sqlConnectMysqlCmd, sqlConnectSqlserverCmd)
	sqlCmd.AddCommand(sqlConnectCmd)

	// Databases
	for _, c := range []*cobra.Command{sqlDatabaseCreateCmd, sqlDatabaseDeleteCmd, sqlDatabaseDescribeCmd, sqlDatabaseListCmd, sqlDatabasePatchCmd} {
		c.Flags().StringVar(&flagSQLInstance, "instance", "", "Cloud SQL instance ID (required)")
		c.Flags().StringVar(&flagSQLFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{sqlDatabaseCreateCmd, sqlDatabaseDeleteCmd, sqlDatabasePatchCmd} {
		c.Flags().BoolVar(&flagSQLAsync, "async", false, "Return without waiting for the operation to complete")
	}
	for _, c := range []*cobra.Command{sqlDatabaseCreateCmd, sqlDatabasePatchCmd} {
		c.Flags().StringVar(&flagSQLCharset, "charset", "", "Database character set")
		c.Flags().StringVar(&flagSQLCollation, "collation", "", "Database collation")
	}
	sqlDatabasePatchCmd.Flags().StringVar(&flagSQLConfigFile, "config-file", "", "YAML/JSON file with a Database body")
	sqlDatabasesCmd.AddCommand(sqlDatabaseCreateCmd, sqlDatabaseDeleteCmd, sqlDatabaseDescribeCmd, sqlDatabaseListCmd, sqlDatabasePatchCmd)
	sqlCmd.AddCommand(sqlDatabasesCmd)

	// Export
	for _, c := range []*cobra.Command{sqlExportSQLCmd, sqlExportCSVCmd, sqlExportBAKCmd} {
		c.Flags().StringVar(&flagSQLDatabase, "database", "", "Database to export from")
		c.Flags().StringSliceVar(&flagSQLExportDBs, "databases", nil, "Databases to export (repeat)")
		c.Flags().StringSliceVar(&flagSQLTable, "table", nil, "Tables to export (SQL only)")
		c.Flags().BoolVar(&flagSQLOffload, "offload", false, "Perform a serverless export")
		c.Flags().StringVar(&flagSQLFormat, "format", "", "Output format")
		c.Flags().BoolVar(&flagSQLAsync, "async", false, "Return without waiting for the operation to complete")
	}
	sqlExportCSVCmd.Flags().StringVar(&flagSQLQuery, "query", "", "SELECT query for CSV export")
	sqlExportBAKCmd.Flags().StringVar(&flagSQLBAKType, "bak-type", "", "BAK type (FULL, DIFF, TLOG)")
	sqlExportCmd.AddCommand(sqlExportSQLCmd, sqlExportCSVCmd, sqlExportBAKCmd)
	sqlCmd.AddCommand(sqlExportCmd)

	// Flags
	sqlFlagsListCmd.Flags().StringVar(&flagSQLDBVersion, "database-version", "", "Database version filter")
	sqlFlagsListCmd.Flags().StringVar(&flagSQLFormat, "format", "", "Output format")
	sqlFlagsCmd.AddCommand(sqlFlagsListCmd)
	sqlCmd.AddCommand(sqlFlagsCmd)

	// Import
	for _, c := range []*cobra.Command{sqlImportSQLCmd, sqlImportCSVCmd, sqlImportBAKCmd} {
		c.Flags().StringVar(&flagSQLDatabase, "database", "", "Database to import into")
		c.Flags().StringVar(&flagSQLImportUser, "user", "", "Postgres user for the import")
		c.Flags().StringVar(&flagSQLFormat, "format", "", "Output format")
		c.Flags().BoolVar(&flagSQLAsync, "async", false, "Return without waiting for the operation to complete")
	}
	sqlImportCSVCmd.Flags().StringSliceVar(&flagSQLTable, "table", nil, "Target table (CSV import)")
	sqlImportCSVCmd.Flags().StringSliceVar(&flagSQLColumns, "columns", nil, "Columns to import (CSV)")
	sqlImportBAKCmd.Flags().StringVar(&flagSQLBAKType, "bak-type", "", "BAK type (FULL, DIFF, TLOG)")
	sqlImportBAKCmd.Flags().StringVar(&flagSQLEncryptionKey, "encryption-password", "", "PVK password for encrypted BAKs")
	sqlImportBAKCmd.Flags().StringVar(&flagSQLPvKPath, "pvk-path", "", "GCS path to the PVK for encrypted BAKs")
	sqlImportBAKCmd.Flags().StringVar(&flagSQLCertPath, "cert-path", "", "GCS path to the encryption certificate")
	sqlImportCmd.AddCommand(sqlImportSQLCmd, sqlImportCSVCmd, sqlImportBAKCmd)
	sqlCmd.AddCommand(sqlImportCmd)

	// Instances
	allInst := []*cobra.Command{
		sqlInstanceCreateCmd, sqlInstanceDeleteCmd, sqlInstanceDescribeCmd, sqlInstanceListCmd,
		sqlInstancePatchCmd, sqlInstanceCloneCmd, sqlInstanceFailoverCmd, sqlInstanceRestartCmd,
		sqlInstanceRestoreBackupCmd, sqlInstancePromoteReplicaCmd, sqlInstanceStartReplicaCmd,
		sqlInstanceStopReplicaCmd, sqlInstanceReencryptCmd, sqlInstanceResetSSLConfigCmd,
		sqlInstanceImportCmd, sqlInstanceExportCmd, sqlInstanceListServerCasCmd,
		sqlInstanceRescheduleMaintenanceCmd, sqlInstanceAcquireSsrsLeaseCmd, sqlInstanceReleaseSsrsLeaseCmd,
		sqlInstanceResetAsyncReplicaLagCmd, sqlInstanceVerifyExternalSyncSettingsCmd,
	}
	for _, c := range allInst {
		c.Flags().StringVar(&flagSQLFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{
		sqlInstanceCreateCmd, sqlInstanceDeleteCmd, sqlInstancePatchCmd, sqlInstanceCloneCmd,
		sqlInstanceFailoverCmd, sqlInstanceRestartCmd, sqlInstanceRestoreBackupCmd,
		sqlInstancePromoteReplicaCmd, sqlInstanceStartReplicaCmd, sqlInstanceStopReplicaCmd,
		sqlInstanceReencryptCmd, sqlInstanceResetSSLConfigCmd, sqlInstanceImportCmd, sqlInstanceExportCmd,
		sqlInstanceRescheduleMaintenanceCmd, sqlInstanceResetAsyncReplicaLagCmd,
	} {
		c.Flags().BoolVar(&flagSQLAsync, "async", false, "Return without waiting for the operation to complete")
	}
	for _, c := range []*cobra.Command{sqlInstanceCreateCmd, sqlInstancePatchCmd, sqlInstanceAcquireSsrsLeaseCmd, sqlInstanceVerifyExternalSyncSettingsCmd} {
		c.Flags().StringVar(&flagSQLConfigFile, "config-file", "", "YAML/JSON file with a DatabaseInstance body")
	}
	sqlInstanceCreateCmd.Flags().StringVar(&flagSQLDBVersion, "database-version", "", "Database version, e.g. POSTGRES_15")
	sqlInstanceCreateCmd.Flags().StringVar(&flagSQLRegion, "region", "", "Region for the instance")
	sqlInstanceCreateCmd.Flags().StringVar(&flagSQLTier, "tier", "", "Machine tier, e.g. db-custom-2-4096")
	sqlInstanceCreateCmd.Flags().StringVar(&flagSQLZone, "zone", "", "Zone preference for the instance")
	sqlInstanceCreateCmd.Flags().StringVar(&flagSQLPassword, "root-password", "", "Root password for the instance")
	sqlInstanceListCmd.Flags().StringVar(&flagSQLFilter, "filter", "", "Server-side filter expression")
	sqlInstanceFailoverCmd.Flags().Int64Var(&flagSQLSettingsVer, "settings-version", 0, "Current settings version of this instance")
	sqlInstanceRestoreBackupCmd.Flags().StringVar(&flagSQLBackup, "backup-id", "", "Backup run ID to restore from (required)")
	sqlInstanceRescheduleMaintenanceCmd.Flags().StringVar(&flagSQLReschedule, "reschedule-type", "IMMEDIATE", "Reschedule type (IMMEDIATE, NEXT_AVAILABLE_WINDOW, SPECIFIC_TIME)")
	sqlInstanceRescheduleMaintenanceCmd.Flags().StringVar(&flagSQLScheduleTime, "schedule-time", "", "RFC3339 timestamp for SPECIFIC_TIME reschedule")
	// Instances subgroup import/export mirror flag sets
	for _, c := range []*cobra.Command{sqlInstanceImportCmd, sqlInstanceExportCmd} {
		c.Flags().StringVar(&flagSQLDatabase, "database", "", "Database name")
		c.Flags().StringSliceVar(&flagSQLExportDBs, "databases", nil, "Databases to export")
		c.Flags().BoolVar(&flagSQLOffload, "offload", false, "Perform a serverless export")
	}
	sqlInstancesCmd.AddCommand(allInst...)
	sqlCmd.AddCommand(sqlInstancesCmd)

	// Operations
	for _, c := range []*cobra.Command{sqlOpCancelCmd, sqlOpDescribeCmd, sqlOpListCmd, sqlOpWaitCmd} {
		c.Flags().StringVar(&flagSQLFormat, "format", "", "Output format")
	}
	sqlOpListCmd.Flags().StringVar(&flagSQLInstance, "instance", "", "Restrict to operations on this instance")
	sqlOpListCmd.Flags().Int64Var(&flagSQLLimit, "limit", 0, "Maximum number of operations to return")
	sqlOperationsCmd.AddCommand(sqlOpCancelCmd, sqlOpDescribeCmd, sqlOpListCmd, sqlOpWaitCmd)
	sqlCmd.AddCommand(sqlOperationsCmd)

	// SSL server-ca-certs (group under `sql ssl server-ca-certs`)
	for _, c := range []*cobra.Command{sqlSSLServerCAListCmd, sqlSSLServerCAAddCmd, sqlSSLServerCARemoveCmd, sqlSSLServerCARotateCmd} {
		c.Flags().StringVar(&flagSQLInstance, "instance", "", "Cloud SQL instance ID (required)")
		c.Flags().StringVar(&flagSQLFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{sqlSSLServerCAAddCmd, sqlSSLServerCARemoveCmd, sqlSSLServerCARotateCmd} {
		c.Flags().BoolVar(&flagSQLAsync, "async", false, "Return without waiting for the operation to complete")
	}
	for _, c := range []*cobra.Command{sqlSSLServerCARemoveCmd, sqlSSLServerCARotateCmd} {
		c.Flags().StringVar(&flagSQLNextVersion, "next-version", "", "Fingerprint of the next server CA version to rotate to")
	}
	sqlSSLServerCACmd.AddCommand(sqlSSLServerCAListCmd, sqlSSLServerCAAddCmd, sqlSSLServerCARemoveCmd, sqlSSLServerCARotateCmd)
	sqlSSLCmd.AddCommand(sqlSSLServerCACmd)
	sqlCmd.AddCommand(sqlSSLCmd)

	// SSL certs (DEPRECATED)
	for _, c := range []*cobra.Command{sqlSSLCertCreateCmd, sqlSSLCertDeleteCmd, sqlSSLCertDescribeCmd, sqlSSLCertListCmd} {
		c.Flags().StringVar(&flagSQLInstance, "instance", "", "Cloud SQL instance ID (required)")
		c.Flags().StringVar(&flagSQLFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{sqlSSLCertCreateCmd, sqlSSLCertDeleteCmd} {
		c.Flags().BoolVar(&flagSQLAsync, "async", false, "Return without waiting for the operation to complete")
	}
	sqlSSLCertsCmd.AddCommand(sqlSSLCertCreateCmd, sqlSSLCertDeleteCmd, sqlSSLCertDescribeCmd, sqlSSLCertListCmd)
	sqlCmd.AddCommand(sqlSSLCertsCmd)

	// Tiers
	sqlTiersListCmd.Flags().StringVar(&flagSQLFormat, "format", "", "Output format")
	sqlTiersCmd.AddCommand(sqlTiersListCmd)
	sqlCmd.AddCommand(sqlTiersCmd)

	// Users
	for _, c := range []*cobra.Command{sqlUserCreateCmd, sqlUserDeleteCmd, sqlUserDescribeCmd, sqlUserListCmd, sqlUserSetPasswordCmd, sqlUserPatchCmd} {
		c.Flags().StringVar(&flagSQLInstance, "instance", "", "Cloud SQL instance ID (required)")
		c.Flags().StringVar(&flagSQLFormat, "format", "", "Output format")
		c.Flags().StringVar(&flagSQLHost, "host", "", "Host for the user (MySQL only)")
	}
	for _, c := range []*cobra.Command{sqlUserCreateCmd, sqlUserDeleteCmd, sqlUserSetPasswordCmd, sqlUserPatchCmd} {
		c.Flags().BoolVar(&flagSQLAsync, "async", false, "Return without waiting for the operation to complete")
	}
	sqlUserCreateCmd.Flags().StringVar(&flagSQLPassword, "password", "", "Password for the user")
	sqlUserCreateCmd.Flags().StringVar(&flagSQLType, "type", "", "User type (BUILT_IN, CLOUD_IAM_USER, CLOUD_IAM_SERVICE_ACCOUNT, ...)")
	sqlUserSetPasswordCmd.Flags().StringVar(&flagSQLPassword, "password", "", "New password (required)")
	sqlUserPatchCmd.Flags().StringVar(&flagSQLConfigFile, "config-file", "", "YAML/JSON file with a User body")
	sqlUserPatchCmd.Flags().StringVar(&flagSQLPassword, "password", "", "Password for the user")
	sqlUsersCmd.AddCommand(sqlUserCreateCmd, sqlUserDeleteCmd, sqlUserDescribeCmd, sqlUserListCmd, sqlUserSetPasswordCmd, sqlUserPatchCmd)
	sqlCmd.AddCommand(sqlUsersCmd)
}

func init() {
	registerSQL()
}
