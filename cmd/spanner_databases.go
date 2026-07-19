package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	spanner "google.golang.org/api/spanner/v1"
)

// --- gcloud spanner databases (#1207) ---

var spannerDatabasesCmd = &cobra.Command{Use: "databases", Short: "Manage Cloud Spanner databases"}
var spannerDbDdlCmd = &cobra.Command{Use: "ddl", Short: "Manage database DDL"}
var spannerDbRolesCmd = &cobra.Command{Use: "roles", Short: "Manage database roles"}
var spannerDbSessionsCmd = &cobra.Command{Use: "sessions", Short: "Manage database sessions"}
var spannerDbSplitsCmd = &cobra.Command{Use: "splits", Short: "Manage database split points"}

var (
	flagSpDbInstance        string
	flagSpDbFormat          string
	flagSpDbConfigFile      string
	flagSpDbUpdateMask      string
	flagSpDbExtraStatements []string
	flagSpDbFilter          string
	flagSpDbPageSize        int64
	flagSpDbStatements      []string
	flagSpDbDdlFile         string
	flagSpDbSQL             string
	flagSpDbSourceBackup    string
	flagSpDbIamMember       string
	flagSpDbIamRole         string
	flagSpDbIamCondExpr     string
	flagSpDbIamCondT        string
	flagSpDbIamCondD        string
	flagSpDbIamAllCond      bool
)

var (
	spannerDbCreateCmd = &cobra.Command{
		Use: "create DATABASE", Short: "Create a Cloud Spanner database",
		Args: cobra.ExactArgs(1), RunE: runSpDbCreate,
	}
	spannerDbDeleteCmd = &cobra.Command{
		Use: "delete DATABASE", Short: "Drop a Cloud Spanner database",
		Args: cobra.ExactArgs(1), RunE: runSpDbDelete,
	}
	spannerDbDescribeCmd = &cobra.Command{
		Use: "describe DATABASE", Short: "Describe a Cloud Spanner database",
		Args: cobra.ExactArgs(1), RunE: runSpDbDescribe,
	}
	spannerDbListCmd = &cobra.Command{
		Use: "list", Short: "List databases in an instance",
		Args: cobra.NoArgs, RunE: runSpDbList,
	}
	spannerDbUpdateCmd = &cobra.Command{
		Use: "update DATABASE", Short: "Update database metadata",
		Args: cobra.ExactArgs(1), RunE: runSpDbUpdate,
	}
	spannerDbChangeQuorumCmd = &cobra.Command{
		Use: "change-quorum DATABASE", Short: "Change the quorum of a Cloud Spanner database",
		Args: cobra.ExactArgs(1), RunE: runSpDbChangeQuorum,
	}
	spannerDbDdlDescribeCmd = &cobra.Command{
		Use: "describe DATABASE", Short: "Show the DDL statements for a database",
		Args: cobra.ExactArgs(1), RunE: runSpDbDdlDescribe,
	}
	spannerDbDdlUpdateCmd = &cobra.Command{
		Use: "update DATABASE", Short: "Apply DDL statements to a database",
		Args: cobra.ExactArgs(1), RunE: runSpDbDdlUpdate,
	}
	spannerDbExecuteSQLCmd = &cobra.Command{
		Use: "execute-sql DATABASE", Short: "Execute a SQL statement against a database",
		Args: cobra.ExactArgs(1), RunE: runSpDbExecuteSQL,
	}
	spannerDbRestoreCmd = &cobra.Command{
		Use: "restore DATABASE", Short: "Restore a database from a backup",
		Args: cobra.ExactArgs(1), RunE: runSpDbRestore,
	}
	spannerDbGetIamCmd = &cobra.Command{
		Use: "get-iam-policy DATABASE", Short: "Get the IAM policy for a database",
		Args: cobra.ExactArgs(1), RunE: runSpDbGetIam,
	}
	spannerDbSetIamCmd = &cobra.Command{
		Use: "set-iam-policy DATABASE POLICY_FILE", Short: "Set the IAM policy for a database",
		Args: cobra.ExactArgs(2), RunE: runSpDbSetIam,
	}
	spannerDbAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding DATABASE", Short: "Add an IAM policy binding on a database",
		Args: cobra.ExactArgs(1), RunE: runSpDbAddIam,
	}
	spannerDbRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding DATABASE", Short: "Remove an IAM policy binding from a database",
		Args: cobra.ExactArgs(1), RunE: runSpDbRemoveIam,
	}
	spannerDbRolesListCmd = &cobra.Command{
		Use: "list", Short: "List roles defined in a database",
		Args: cobra.NoArgs, RunE: runSpDbRolesList,
	}
	spannerDbSessionsListCmd = &cobra.Command{
		Use: "list", Short: "List sessions for a database",
		Args: cobra.NoArgs, RunE: runSpDbSessionsList,
	}
	spannerDbSessionsDeleteCmd = &cobra.Command{
		Use: "delete SESSION", Short: "Delete a session",
		Args: cobra.ExactArgs(1), RunE: runSpDbSessionsDelete,
	}
	spannerDbSplitsAddCmd = &cobra.Command{
		Use: "add DATABASE", Short: "Add split points to a database",
		Args: cobra.ExactArgs(1), RunE: runSpDbSplitsAdd,
	}
)

// databases-level flags (need --database on some sub-groups)
var (
	flagSpDbRolesDatabase    string
	flagSpDbSessionsDatabase string
)

func init() {
	main := []*cobra.Command{
		spannerDbCreateCmd, spannerDbDeleteCmd, spannerDbDescribeCmd, spannerDbListCmd,
		spannerDbUpdateCmd, spannerDbChangeQuorumCmd, spannerDbExecuteSQLCmd, spannerDbRestoreCmd,
		spannerDbGetIamCmd, spannerDbSetIamCmd, spannerDbAddIamCmd, spannerDbRemoveIamCmd,
	}
	ddl := []*cobra.Command{spannerDbDdlDescribeCmd, spannerDbDdlUpdateCmd}
	roles := []*cobra.Command{spannerDbRolesListCmd}
	sessions := []*cobra.Command{spannerDbSessionsListCmd, spannerDbSessionsDeleteCmd}
	splits := []*cobra.Command{spannerDbSplitsAddCmd}

	all := append([]*cobra.Command{}, main...)
	all = append(all, ddl...)
	all = append(all, roles...)
	all = append(all, sessions...)
	all = append(all, splits...)

	for _, c := range all {
		c.Flags().StringVar(&flagSpDbInstance, "instance", "", "Spanner instance (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&flagSpDbFormat, "format", "", "Output format")
	}

	spannerDbCreateCmd.Flags().StringSliceVar(&flagSpDbExtraStatements, "extra-statements", nil,
		"Additional DDL statements to run inside the newly created database (repeatable or comma-separated)")

	spannerDbUpdateCmd.Flags().StringVar(&flagSpDbConfigFile, "config-file", "", "YAML/JSON file with the Database body (required)")
	_ = spannerDbUpdateCmd.MarkFlagRequired("config-file")
	spannerDbUpdateCmd.Flags().StringVar(&flagSpDbUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	spannerDbChangeQuorumCmd.Flags().StringVar(&flagSpDbConfigFile, "config-file", "", "YAML/JSON file with the ChangeQuorumRequest body (required)")
	_ = spannerDbChangeQuorumCmd.MarkFlagRequired("config-file")

	spannerDbListCmd.Flags().Int64Var(&flagSpDbPageSize, "page-size", 0, "Maximum results per page")

	spannerDbDdlUpdateCmd.Flags().StringSliceVar(&flagSpDbStatements, "statement", nil,
		"DDL statement to apply (repeatable). Mutually exclusive with --ddl-file")
	spannerDbDdlUpdateCmd.Flags().StringVar(&flagSpDbDdlFile, "ddl-file", "",
		"Path to a file with DDL statements, one per line. Mutually exclusive with --statement")

	spannerDbExecuteSQLCmd.Flags().StringVar(&flagSpDbSQL, "sql", "", "SQL statement to execute (required)")
	_ = spannerDbExecuteSQLCmd.MarkFlagRequired("sql")

	spannerDbRestoreCmd.Flags().StringVar(&flagSpDbSourceBackup, "source-backup", "",
		"Backup to restore from; ID or fully qualified name (required)")
	_ = spannerDbRestoreCmd.MarkFlagRequired("source-backup")

	spannerDbSplitsAddCmd.Flags().StringVar(&flagSpDbConfigFile, "config-file", "",
		"YAML/JSON file with the AddSplitPointsRequest body (required)")
	_ = spannerDbSplitsAddCmd.MarkFlagRequired("config-file")

	for _, c := range []*cobra.Command{spannerDbAddIamCmd, spannerDbRemoveIamCmd} {
		spIamMemberFlags(c, &flagSpDbIamMember, &flagSpDbIamRole,
			&flagSpDbIamCondExpr, &flagSpDbIamCondT, &flagSpDbIamCondD)
	}
	spannerDbRemoveIamCmd.Flags().BoolVar(&flagSpDbIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	// roles: --database
	for _, c := range roles {
		c.Flags().StringVar(&flagSpDbRolesDatabase, "database", "", "Database (required)")
		_ = c.MarkFlagRequired("database")
	}
	spannerDbRolesListCmd.Flags().Int64Var(&flagSpDbPageSize, "page-size", 0, "Maximum results per page")

	// sessions: --database
	for _, c := range sessions {
		c.Flags().StringVar(&flagSpDbSessionsDatabase, "database", "", "Database (required)")
		_ = c.MarkFlagRequired("database")
	}
	spannerDbSessionsListCmd.Flags().StringVar(&flagSpDbFilter, "filter", "", "Server-side filter expression")
	spannerDbSessionsListCmd.Flags().Int64Var(&flagSpDbPageSize, "page-size", 0, "Maximum results per page")

	spannerDbCmd := spannerDatabasesCmd
	spannerDbCmd.AddCommand(main...)
	spannerDbDdlCmd.AddCommand(ddl...)
	spannerDbCmd.AddCommand(spannerDbDdlCmd)
	spannerDbRolesCmd.AddCommand(roles...)
	spannerDbCmd.AddCommand(spannerDbRolesCmd)
	spannerDbSessionsCmd.AddCommand(sessions...)
	spannerDbCmd.AddCommand(spannerDbSessionsCmd)
	spannerDbSplitsCmd.AddCommand(splits...)
	spannerDbCmd.AddCommand(spannerDbSplitsCmd)
	spannerCmd.AddCommand(spannerDbCmd)
}

func runSpDbCreate(cmd *cobra.Command, args []string) error {
	parent, err := spannerInstance(flagSpDbInstance)
	if err != nil {
		return err
	}
	body := &spanner.CreateDatabaseRequest{
		CreateStatement: fmt.Sprintf("CREATE DATABASE `%s`", args[0]),
		ExtraStatements: flagSpDbExtraStatements,
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Databases.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating database: %w", err)
	}
	fmt.Printf("Create database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagSpDbFormat)
}

func runSpDbDelete(cmd *cobra.Command, args []string) error {
	name, err := spannerDatabase(flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Databases.DropDatabase(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("dropping database: %w", err)
	}
	fmt.Printf("Dropped database [%s].\n", args[0])
	return nil
}

func runSpDbDescribe(cmd *cobra.Command, args []string) error {
	name, err := spannerDatabase(flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Databases.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing database: %w", err)
	}
	return emitFormatted(got, flagSpDbFormat)
}

func runSpDbList(cmd *cobra.Command, args []string) error {
	parent, err := spannerInstance(flagSpDbInstance)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*spanner.Database
	pageToken := ""
	for {
		call := svc.Projects.Instances.Databases.List(parent).Context(ctx)
		if flagSpDbPageSize > 0 {
			call = call.PageSize(flagSpDbPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing databases: %w", err)
		}
		all = append(all, resp.Databases...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagSpDbFormat)
}

func runSpDbUpdate(cmd *cobra.Command, args []string) error {
	name, err := spannerDatabase(flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	body := &spanner.Database{}
	if err := loadYAMLOrJSONInto(flagSpDbConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagSpDbUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Instances.Databases.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating database: %w", err)
	}
	return emitFormatted(op, flagSpDbFormat)
}

func runSpDbChangeQuorum(cmd *cobra.Command, args []string) error {
	name, err := spannerDatabase(flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	body := &spanner.ChangeQuorumRequest{}
	if err := loadYAMLOrJSONInto(flagSpDbConfigFile, body); err != nil {
		return err
	}
	if body.Name == "" {
		body.Name = name
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Databases.Changequorum(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("changing quorum: %w", err)
	}
	return emitFormatted(op, flagSpDbFormat)
}

func runSpDbDdlDescribe(cmd *cobra.Command, args []string) error {
	name, err := spannerDatabase(flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Databases.GetDdl(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting DDL: %w", err)
	}
	return emitFormatted(got, flagSpDbFormat)
}

func runSpDbDdlUpdate(cmd *cobra.Command, args []string) error {
	name, err := spannerDatabase(flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	statements := flagSpDbStatements
	if flagSpDbDdlFile != "" {
		data, err := os.ReadFile(flagSpDbDdlFile)
		if err != nil {
			return fmt.Errorf("reading DDL file: %w", err)
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			statements = append(statements, line)
		}
	}
	if len(statements) == 0 {
		return fmt.Errorf("at least one --statement or a non-empty --ddl-file is required")
	}
	body := &spanner.UpdateDatabaseDdlRequest{Statements: statements}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Databases.UpdateDdl(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating DDL: %w", err)
	}
	return emitFormatted(op, flagSpDbFormat)
}

// spDbRunSingleUseSQL creates a session, runs an ExecuteSql with SingleUse
// transaction options, then deletes the session. It returns the ResultSet.
func spDbRunSingleUseSQL(ctx context.Context, svc *spanner.Service, database, sql string, readWrite bool) (*spanner.ResultSet, error) {
	session, err := svc.Projects.Instances.Databases.Sessions.Create(database, &spanner.CreateSessionRequest{}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}
	defer func() {
		_, _ = svc.Projects.Instances.Databases.Sessions.Delete(session.Name).Context(ctx).Do()
	}()
	req := &spanner.ExecuteSqlRequest{Sql: sql}
	if readWrite {
		req.Transaction = &spanner.TransactionSelector{Begin: &spanner.TransactionOptions{ReadWrite: &spanner.ReadWrite{}}}
	} else {
		req.Transaction = &spanner.TransactionSelector{SingleUse: &spanner.TransactionOptions{ReadOnly: &spanner.ReadOnly{Strong: true}}}
	}
	rs, err := svc.Projects.Instances.Databases.Sessions.ExecuteSql(session.Name, req).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("executing SQL: %w", err)
	}
	if readWrite && rs.Metadata != nil && rs.Metadata.Transaction != nil && rs.Metadata.Transaction.Id != "" {
		if _, err := svc.Projects.Instances.Databases.Sessions.Commit(session.Name,
			&spanner.CommitRequest{TransactionId: rs.Metadata.Transaction.Id}).Context(ctx).Do(); err != nil {
			return nil, fmt.Errorf("committing transaction: %w", err)
		}
	}
	return rs, nil
}

func runSpDbExecuteSQL(cmd *cobra.Command, args []string) error {
	database, err := spannerDatabase(flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	// Detect DML vs. read-only.
	trimmed := strings.ToUpper(strings.TrimSpace(flagSpDbSQL))
	rw := strings.HasPrefix(trimmed, "INSERT") || strings.HasPrefix(trimmed, "UPDATE") ||
		strings.HasPrefix(trimmed, "DELETE") || strings.HasPrefix(trimmed, "MERGE")
	rs, err := spDbRunSingleUseSQL(ctx, svc, database, flagSpDbSQL, rw)
	if err != nil {
		return err
	}
	return emitFormatted(rs, flagSpDbFormat)
}

func runSpDbRestore(cmd *cobra.Command, args []string) error {
	parent, err := spannerInstance(flagSpDbInstance)
	if err != nil {
		return err
	}
	backup, err := spannerBackup(flagSpDbInstance, flagSpDbSourceBackup)
	if err != nil {
		return err
	}
	body := &spanner.RestoreDatabaseRequest{
		DatabaseId: args[0],
		Backup:     backup,
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Databases.Restore(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring database: %w", err)
	}
	fmt.Printf("Restore database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagSpDbFormat)
}

func runSpDbGetIam(cmd *cobra.Command, args []string) error {
	name, err := spannerDatabase(flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Databases.GetIamPolicy(name, &spanner.GetIamPolicyRequest{
		Options: &spanner.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagSpDbFormat)
}

func runSpDbSetIam(cmd *cobra.Command, args []string) error {
	name, err := spannerDatabase(flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	policy := &spanner.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Instances.Databases.SetIamPolicy(name, &spanner.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	spUpdatedIam(fmt.Sprintf("database [%s]", args[0]))
	return emitFormatted(updated, flagSpDbFormat)
}

func runSpDbAddIam(cmd *cobra.Command, args []string) error {
	name, err := spannerDatabase(flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Databases.GetIamPolicy(name, &spanner.GetIamPolicyRequest{
		Options: &spanner.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	spIamAddBinding(policy, flagSpDbIamRole, flagSpDbIamMember,
		spIamBuildCondition(flagSpDbIamCondExpr, flagSpDbIamCondT, flagSpDbIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Instances.Databases.SetIamPolicy(name, &spanner.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	spUpdatedIam(fmt.Sprintf("database [%s]", args[0]))
	return emitFormatted(updated, flagSpDbFormat)
}

func runSpDbRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := spannerDatabase(flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Databases.GetIamPolicy(name, &spanner.GetIamPolicyRequest{
		Options: &spanner.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !spIamRemoveBinding(policy, flagSpDbIamRole, flagSpDbIamMember,
		spIamBuildCondition(flagSpDbIamCondExpr, flagSpDbIamCondT, flagSpDbIamCondD), flagSpDbIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagSpDbIamRole, flagSpDbIamMember)
	}
	updated, err := svc.Projects.Instances.Databases.SetIamPolicy(name, &spanner.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	spUpdatedIam(fmt.Sprintf("database [%s]", args[0]))
	return emitFormatted(updated, flagSpDbFormat)
}

func runSpDbRolesList(cmd *cobra.Command, args []string) error {
	parent, err := spannerDatabase(flagSpDbInstance, flagSpDbRolesDatabase)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*spanner.DatabaseRole
	pageToken := ""
	for {
		call := svc.Projects.Instances.Databases.DatabaseRoles.List(parent).Context(ctx)
		if flagSpDbPageSize > 0 {
			call = call.PageSize(flagSpDbPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing database roles: %w", err)
		}
		all = append(all, resp.DatabaseRoles...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagSpDbFormat)
}

func runSpDbSessionsList(cmd *cobra.Command, args []string) error {
	parent, err := spannerDatabase(flagSpDbInstance, flagSpDbSessionsDatabase)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*spanner.Session
	pageToken := ""
	for {
		call := svc.Projects.Instances.Databases.Sessions.List(parent).Context(ctx)
		if flagSpDbFilter != "" {
			call = call.Filter(flagSpDbFilter)
		}
		if flagSpDbPageSize > 0 {
			call = call.PageSize(flagSpDbPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing sessions: %w", err)
		}
		all = append(all, resp.Sessions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagSpDbFormat)
}

func runSpDbSessionsDelete(cmd *cobra.Command, args []string) error {
	sessionName := args[0]
	if !strings.HasPrefix(sessionName, "projects/") {
		db, err := spannerDatabase(flagSpDbInstance, flagSpDbSessionsDatabase)
		if err != nil {
			return err
		}
		sessionName = fmt.Sprintf("%s/sessions/%s", db, sessionName)
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Databases.Sessions.Delete(sessionName).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}
	fmt.Printf("Deleted session [%s].\n", args[0])
	return nil
}

func runSpDbSplitsAdd(cmd *cobra.Command, args []string) error {
	name, err := spannerDatabase(flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	body := &spanner.AddSplitPointsRequest{}
	if err := loadYAMLOrJSONInto(flagSpDbConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Instances.Databases.AddSplitPoints(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding split points: %w", err)
	}
	return emitFormatted(resp, flagSpDbFormat)
}

