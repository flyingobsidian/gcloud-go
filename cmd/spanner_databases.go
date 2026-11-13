package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	spanner "google.golang.org/api/spanner/v1"
)

// --- gcloud spanner databases (#1207) ---

func (s *spannerCmds) initSpannerDatabases() {
	main := []*cobra.Command{
		s.spannerDbCreateCmd, s.spannerDbDeleteCmd, s.spannerDbDescribeCmd, s.spannerDbListCmd,
		s.spannerDbUpdateCmd, s.spannerDbChangeQuorumCmd, s.spannerDbExecuteSQLCmd, s.spannerDbRestoreCmd,
		s.spannerDbGetIamCmd, s.spannerDbSetIamCmd, s.spannerDbAddIamCmd, s.spannerDbRemoveIamCmd,
	}
	ddl := []*cobra.Command{s.spannerDbDdlDescribeCmd, s.spannerDbDdlUpdateCmd}
	roles := []*cobra.Command{s.spannerDbRolesListCmd}
	sessions := []*cobra.Command{s.spannerDbSessionsListCmd, s.spannerDbSessionsDeleteCmd}
	splits := []*cobra.Command{s.spannerDbSplitsAddCmd}

	all := append([]*cobra.Command{}, main...)
	all = append(all, ddl...)
	all = append(all, roles...)
	all = append(all, sessions...)
	all = append(all, splits...)

	for _, c := range all {
		c.Flags().StringVar(&s.flagSpDbInstance, "instance", "", "Spanner instance (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&s.flagSpDbFormat, "format", "", "Output format")
	}

	s.spannerDbCreateCmd.Flags().StringSliceVar(&s.flagSpDbExtraStatements, "extra-statements", nil,
		"Additional DDL statements to run inside the newly created database (repeatable or comma-separated)")

	s.spannerDbUpdateCmd.Flags().StringVar(&s.flagSpDbConfigFile, "config-file", "", "YAML/JSON file with the Database body (required)")
	_ = s.spannerDbUpdateCmd.MarkFlagRequired("config-file")
	s.spannerDbUpdateCmd.Flags().StringVar(&s.flagSpDbUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	s.spannerDbChangeQuorumCmd.Flags().StringVar(&s.flagSpDbConfigFile, "config-file", "", "YAML/JSON file with the ChangeQuorumRequest body (required)")
	_ = s.spannerDbChangeQuorumCmd.MarkFlagRequired("config-file")

	s.spannerDbListCmd.Flags().Int64Var(&s.flagSpDbPageSize, "page-size", 0, "Maximum results per page")

	s.spannerDbDdlUpdateCmd.Flags().StringSliceVar(&s.flagSpDbStatements, "statement", nil,
		"DDL statement to apply (repeatable). Mutually exclusive with --ddl-file")
	s.spannerDbDdlUpdateCmd.Flags().StringVar(&s.flagSpDbDdlFile, "ddl-file", "",
		"Path to a file with DDL statements, one per line. Mutually exclusive with --statement")

	s.spannerDbExecuteSQLCmd.Flags().StringVar(&s.flagSpDbSQL, "sql", "", "SQL statement to execute (required)")
	_ = s.spannerDbExecuteSQLCmd.MarkFlagRequired("sql")

	s.spannerDbRestoreCmd.Flags().StringVar(&s.flagSpDbSourceBackup, "source-backup", "",
		"Backup to restore from; ID or fully qualified name (required)")
	_ = s.spannerDbRestoreCmd.MarkFlagRequired("source-backup")

	s.spannerDbSplitsAddCmd.Flags().StringVar(&s.flagSpDbConfigFile, "config-file", "",
		"YAML/JSON file with the AddSplitPointsRequest body (required)")
	_ = s.spannerDbSplitsAddCmd.MarkFlagRequired("config-file")

	for _, c := range []*cobra.Command{s.spannerDbAddIamCmd, s.spannerDbRemoveIamCmd} {
		spIamMemberFlags(c, &s.flagSpDbIamMember, &s.flagSpDbIamRole,
			&s.flagSpDbIamCondExpr, &s.flagSpDbIamCondT, &s.flagSpDbIamCondD)
	}
	s.spannerDbRemoveIamCmd.Flags().BoolVar(&s.flagSpDbIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	// roles: --database
	for _, c := range roles {
		c.Flags().StringVar(&s.flagSpDbRolesDatabase, "database", "", "Database (required)")
		_ = c.MarkFlagRequired("database")
	}
	s.spannerDbRolesListCmd.Flags().Int64Var(&s.flagSpDbPageSize, "page-size", 0, "Maximum results per page")

	// sessions: --database
	for _, c := range sessions {
		c.Flags().StringVar(&s.flagSpDbSessionsDatabase, "database", "", "Database (required)")
		_ = c.MarkFlagRequired("database")
	}
	s.spannerDbSessionsListCmd.Flags().StringVar(&s.flagSpDbFilter, "filter", "", "Server-side filter expression")
	s.spannerDbSessionsListCmd.Flags().Int64Var(&s.flagSpDbPageSize, "page-size", 0, "Maximum results per page")

	spannerDbCmd := s.spannerDatabasesCmd
	spannerDbCmd.AddCommand(main...)
	s.spannerDbDdlCmd.AddCommand(ddl...)
	spannerDbCmd.AddCommand(s.spannerDbDdlCmd)
	s.spannerDbRolesCmd.AddCommand(roles...)
	spannerDbCmd.AddCommand(s.spannerDbRolesCmd)
	s.spannerDbSessionsCmd.AddCommand(sessions...)
	spannerDbCmd.AddCommand(s.spannerDbSessionsCmd)
	s.spannerDbSplitsCmd.AddCommand(splits...)
	spannerDbCmd.AddCommand(s.spannerDbSplitsCmd)
	s.spannerCmd.AddCommand(spannerDbCmd)
}

func (s *spannerCmds) runSpDbCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.spannerInstance(s.flagSpDbInstance)
	if err != nil {
		return err
	}
	body := &spanner.CreateDatabaseRequest{
		CreateStatement: fmt.Sprintf("CREATE DATABASE `%s`", args[0]),
		ExtraStatements: s.flagSpDbExtraStatements,
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Databases.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating database: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagSpDbFormat)
}

func (s *spannerCmds) runSpDbDelete(cmd *cobra.Command, args []string) error {
	name, err := s.spannerDatabase(s.flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Databases.DropDatabase(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("dropping database: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Dropped database [%s].\n", args[0])
	return nil
}

func (s *spannerCmds) runSpDbDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.spannerDatabase(s.flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Databases.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing database: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSpDbFormat)
}

func (s *spannerCmds) runSpDbList(cmd *cobra.Command, args []string) error {
	parent, err := s.spannerInstance(s.flagSpDbInstance)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	var all []*spanner.Database
	pageToken := ""
	for {
		call := svc.Projects.Instances.Databases.List(parent).Context(ctx)
		if s.flagSpDbPageSize > 0 {
			call = call.PageSize(s.flagSpDbPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagSpDbFormat)
}

func (s *spannerCmds) runSpDbUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.spannerDatabase(s.flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	body := &spanner.Database{}
	if err := loadYAMLOrJSONInto(s.flagSpDbConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagSpDbUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
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
	return emitFormattedTo(s.inv.stdout, op, s.flagSpDbFormat)
}

func (s *spannerCmds) runSpDbChangeQuorum(cmd *cobra.Command, args []string) error {
	name, err := s.spannerDatabase(s.flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	body := &spanner.ChangeQuorumRequest{}
	if err := loadYAMLOrJSONInto(s.flagSpDbConfigFile, body); err != nil {
		return err
	}
	if body.Name == "" {
		body.Name = name
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Databases.Changequorum(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("changing quorum: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagSpDbFormat)
}

func (s *spannerCmds) runSpDbDdlDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.spannerDatabase(s.flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.Databases.GetDdl(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting DDL: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSpDbFormat)
}

func (s *spannerCmds) runSpDbDdlUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.spannerDatabase(s.flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	statements := s.flagSpDbStatements
	if s.flagSpDbDdlFile != "" {
		data, err := os.ReadFile(s.flagSpDbDdlFile)
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
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Databases.UpdateDdl(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating DDL: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagSpDbFormat)
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

func (s *spannerCmds) runSpDbExecuteSQL(cmd *cobra.Command, args []string) error {
	database, err := s.spannerDatabase(s.flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	// Detect DML vs. read-only.
	trimmed := strings.ToUpper(strings.TrimSpace(s.flagSpDbSQL))
	rw := strings.HasPrefix(trimmed, "INSERT") || strings.HasPrefix(trimmed, "UPDATE") ||
		strings.HasPrefix(trimmed, "DELETE") || strings.HasPrefix(trimmed, "MERGE")
	rs, err := spDbRunSingleUseSQL(ctx, svc, database, s.flagSpDbSQL, rw)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, rs, s.flagSpDbFormat)
}

func (s *spannerCmds) runSpDbRestore(cmd *cobra.Command, args []string) error {
	parent, err := s.spannerInstance(s.flagSpDbInstance)
	if err != nil {
		return err
	}
	backup, err := s.spannerBackup(s.flagSpDbInstance, s.flagSpDbSourceBackup)
	if err != nil {
		return err
	}
	body := &spanner.RestoreDatabaseRequest{
		DatabaseId: args[0],
		Backup:     backup,
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.Databases.Restore(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring database: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Restore database [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagSpDbFormat)
}

func (s *spannerCmds) runSpDbGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.spannerDatabase(s.flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Databases.GetIamPolicy(name, &spanner.GetIamPolicyRequest{
		Options: &spanner.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagSpDbFormat)
}

func (s *spannerCmds) runSpDbSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.spannerDatabase(s.flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	policy := &spanner.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Instances.Databases.SetIamPolicy(name, &spanner.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.spUpdatedIam(fmt.Sprintf("database [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagSpDbFormat)
}

func (s *spannerCmds) runSpDbAddIam(cmd *cobra.Command, args []string) error {
	name, err := s.spannerDatabase(s.flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Databases.GetIamPolicy(name, &spanner.GetIamPolicyRequest{
		Options: &spanner.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	spIamAddBinding(policy, s.flagSpDbIamRole, s.flagSpDbIamMember,
		spIamBuildCondition(s.flagSpDbIamCondExpr, s.flagSpDbIamCondT, s.flagSpDbIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Instances.Databases.SetIamPolicy(name, &spanner.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.spUpdatedIam(fmt.Sprintf("database [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagSpDbFormat)
}

func (s *spannerCmds) runSpDbRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := s.spannerDatabase(s.flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Instances.Databases.GetIamPolicy(name, &spanner.GetIamPolicyRequest{
		Options: &spanner.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !spIamRemoveBinding(policy, s.flagSpDbIamRole, s.flagSpDbIamMember,
		spIamBuildCondition(s.flagSpDbIamCondExpr, s.flagSpDbIamCondT, s.flagSpDbIamCondD), s.flagSpDbIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", s.flagSpDbIamRole, s.flagSpDbIamMember)
	}
	updated, err := svc.Projects.Instances.Databases.SetIamPolicy(name, &spanner.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.spUpdatedIam(fmt.Sprintf("database [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagSpDbFormat)
}

func (s *spannerCmds) runSpDbRolesList(cmd *cobra.Command, args []string) error {
	parent, err := s.spannerDatabase(s.flagSpDbInstance, s.flagSpDbRolesDatabase)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	var all []*spanner.DatabaseRole
	pageToken := ""
	for {
		call := svc.Projects.Instances.Databases.DatabaseRoles.List(parent).Context(ctx)
		if s.flagSpDbPageSize > 0 {
			call = call.PageSize(s.flagSpDbPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagSpDbFormat)
}

func (s *spannerCmds) runSpDbSessionsList(cmd *cobra.Command, args []string) error {
	parent, err := s.spannerDatabase(s.flagSpDbInstance, s.flagSpDbSessionsDatabase)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	var all []*spanner.Session
	pageToken := ""
	for {
		call := svc.Projects.Instances.Databases.Sessions.List(parent).Context(ctx)
		if s.flagSpDbFilter != "" {
			call = call.Filter(s.flagSpDbFilter)
		}
		if s.flagSpDbPageSize > 0 {
			call = call.PageSize(s.flagSpDbPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagSpDbFormat)
}

func (s *spannerCmds) runSpDbSessionsDelete(cmd *cobra.Command, args []string) error {
	sessionName := args[0]
	if !strings.HasPrefix(sessionName, "projects/") {
		db, err := s.spannerDatabase(s.flagSpDbInstance, s.flagSpDbSessionsDatabase)
		if err != nil {
			return err
		}
		sessionName = fmt.Sprintf("%s/sessions/%s", db, sessionName)
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.Databases.Sessions.Delete(sessionName).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted session [%s].\n", args[0])
	return nil
}

func (s *spannerCmds) runSpDbSplitsAdd(cmd *cobra.Command, args []string) error {
	name, err := s.spannerDatabase(s.flagSpDbInstance, args[0])
	if err != nil {
		return err
	}
	body := &spanner.AddSplitPointsRequest{}
	if err := loadYAMLOrJSONInto(s.flagSpDbConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Instances.Databases.AddSplitPoints(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding split points: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagSpDbFormat)
}
