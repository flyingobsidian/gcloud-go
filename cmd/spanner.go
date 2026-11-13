package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	spanner "google.golang.org/api/spanner/v1"
)

// --- gcloud spanner (#387) ---

func (s *spannerCmds) initSpanner() {
	// backups (#1206), databases (#1207), instance-configs (#1208),
	// instance-partitions (#1209), instances (#1210), operations (#1211),
	// and rows (#1212) live in their own files. samples, cli, and
	// backup-schedules remain stubs pending their own issues.
	registerStubGroup(s.spannerCmd, "samples", "Sample apps", "list", "run")
	registerStubCommand(s.spannerCmd, "cli", "Interactive Spanner shell")

}

// spannerProject returns the `projects/PROJECT` prefix, resolving the project
// from the standard --project/env/config chain.
func (s *spannerCmds) spannerProject() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return "projects/" + project, nil
}

// spannerInstance returns a fully qualified instance name. If the input is
// already a full URI (starts with "projects/"), it is returned as-is.
func (s *spannerCmds) spannerInstance(instance string) (string, error) {
	if strings.HasPrefix(instance, "projects/") {
		return instance, nil
	}
	project, err := s.spannerProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/instances/%s", project, instance), nil
}

// spannerDatabase returns a fully qualified database name.
func (s *spannerCmds) spannerDatabase(instance, database string) (string, error) {
	if strings.HasPrefix(database, "projects/") {
		return database, nil
	}
	inst, err := s.spannerInstance(instance)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/databases/%s", inst, database), nil
}

// spannerBackup returns a fully qualified backup name.
func (s *spannerCmds) spannerBackup(instance, backup string) (string, error) {
	if strings.HasPrefix(backup, "projects/") {
		return backup, nil
	}
	inst, err := s.spannerInstance(instance)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/backups/%s", inst, backup), nil
}

// spannerInstanceConfig returns a fully qualified instance config name.
func (s *spannerCmds) spannerInstanceConfig(cfg string) (string, error) {
	if strings.HasPrefix(cfg, "projects/") {
		return cfg, nil
	}
	project, err := s.spannerProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/instanceConfigs/%s", project, cfg), nil
}

// spannerInstancePartition returns a fully qualified instance partition name.
func (s *spannerCmds) spannerInstancePartition(instance, partition string) (string, error) {
	if strings.HasPrefix(partition, "projects/") {
		return partition, nil
	}
	inst, err := s.spannerInstance(instance)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/instancePartitions/%s", inst, partition), nil
}

// spannerBackupSchedule returns a fully qualified backup schedule name.
func (s *spannerCmds) spannerBackupSchedule(instance, database, schedule string) (string, error) {
	if strings.HasPrefix(schedule, "projects/") {
		return schedule, nil
	}
	db, err := s.spannerDatabase(instance, database)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/backupSchedules/%s", db, schedule), nil
}

// spIamMemberFlags binds the standard IAM member/role/condition flags shared by
// all Cloud Spanner add/remove-iam-policy-binding commands.
func spIamMemberFlags(c *cobra.Command, member, role, condExpr, condTitle, condDesc *string) {
	c.Flags().StringVar(member, "member", "", "IAM member (required)")
	c.Flags().StringVar(role, "role", "", "IAM role to bind (required)")
	c.Flags().StringVar(condExpr, "condition-expression", "", "CEL expression for a conditional binding")
	c.Flags().StringVar(condTitle, "condition-title", "", "Title for a conditional binding")
	c.Flags().StringVar(condDesc, "condition-description", "", "Description for a conditional binding")
	_ = c.MarkFlagRequired("member")
	_ = c.MarkFlagRequired("role")
}

func spIamBuildCondition(expr, title, desc string) *spanner.Expr {
	if expr == "" && title == "" && desc == "" {
		return nil
	}
	return &spanner.Expr{Expression: expr, Title: title, Description: desc}
}

func spIamCondsEqual(a, b *spanner.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

func spIamAddBinding(policy *spanner.Policy, role, member string, cond *spanner.Expr) {
	for _, b := range policy.Bindings {
		if b.Role != role || !spIamCondsEqual(b.Condition, cond) {
			continue
		}
		for _, m := range b.Members {
			if m == member {
				return
			}
		}
		b.Members = append(b.Members, member)
		return
	}
	policy.Bindings = append(policy.Bindings, &spanner.Binding{
		Role: role, Members: []string{member}, Condition: cond,
	})
}

func spIamRemoveBinding(policy *spanner.Policy, role, member string, cond *spanner.Expr, allConds bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConds || spIamCondsEqual(b.Condition, cond))
		if !match {
			kept = append(kept, b)
			continue
		}
		newMembers := b.Members[:0]
		for _, m := range b.Members {
			if m == member {
				continue
			}
			newMembers = append(newMembers, m)
		}
		if len(newMembers) != len(b.Members) {
			changed = true
		}
		b.Members = newMembers
		if len(b.Members) > 0 {
			kept = append(kept, b)
		} else {
			changed = true
		}
	}
	policy.Bindings = kept
	return changed
}

func (s *spannerCmds) spUpdatedIam(who string) {
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for %s.\n", who)
}

// spannerCmds holds the commands and flag values of one invocation's "spanner"
// command group, so invocations don't share state (#2070).
type spannerCmds struct {
	inv *invocation

	spannerCmd                   *cobra.Command
	spannerBackupSchedulesCmd    *cobra.Command
	flagSpBsInstance             string
	flagSpBsDatabase             string
	flagSpBsFormat               string
	flagSpBsConfigFile           string
	flagSpBsUpdateMask           string
	flagSpBsPageSize             int64
	spannerBsCreateCmd           *cobra.Command
	spannerBsDeleteCmd           *cobra.Command
	spannerBsDescribeCmd         *cobra.Command
	spannerBsListCmd             *cobra.Command
	spannerBsUpdateCmd           *cobra.Command
	spannerBsGetIamCmd           *cobra.Command
	spannerBsSetIamCmd           *cobra.Command
	spannerBackupsCmd            *cobra.Command
	flagSpBackupsInstance        string
	flagSpBackupsFormat          string
	flagSpBackupsDatabase        string
	flagSpBackupsConfigFile      string
	flagSpBackupsUpdateMask      string
	flagSpBackupsFilter          string
	flagSpBackupsPageSize        int64
	flagSpBackupsDestBackupID    string
	flagSpBackupsSourceBackup    string
	flagSpBackupsExpireTime      string
	flagSpBackupsIamMember       string
	flagSpBackupsIamRole         string
	flagSpBackupsIamCondExpr     string
	flagSpBackupsIamCondT        string
	flagSpBackupsIamCondD        string
	flagSpBackupsIamAllCond      bool
	spannerBackupsCreateCmd      *cobra.Command
	spannerBackupsDeleteCmd      *cobra.Command
	spannerBackupsDescribeCmd    *cobra.Command
	spannerBackupsListCmd        *cobra.Command
	spannerBackupsCopyCmd        *cobra.Command
	spannerBackupsUpdateMetaCmd  *cobra.Command
	spannerBackupsGetIamCmd      *cobra.Command
	spannerBackupsSetIamCmd      *cobra.Command
	spannerBackupsAddIamCmd      *cobra.Command
	spannerBackupsRemoveIamCmd   *cobra.Command
	spannerDatabasesCmd          *cobra.Command
	spannerDbDdlCmd              *cobra.Command
	spannerDbRolesCmd            *cobra.Command
	spannerDbSessionsCmd         *cobra.Command
	spannerDbSplitsCmd           *cobra.Command
	flagSpDbInstance             string
	flagSpDbFormat               string
	flagSpDbConfigFile           string
	flagSpDbUpdateMask           string
	flagSpDbExtraStatements      []string
	flagSpDbFilter               string
	flagSpDbPageSize             int64
	flagSpDbStatements           []string
	flagSpDbDdlFile              string
	flagSpDbSQL                  string
	flagSpDbSourceBackup         string
	flagSpDbIamMember            string
	flagSpDbIamRole              string
	flagSpDbIamCondExpr          string
	flagSpDbIamCondT             string
	flagSpDbIamCondD             string
	flagSpDbIamAllCond           bool
	spannerDbCreateCmd           *cobra.Command
	spannerDbDeleteCmd           *cobra.Command
	spannerDbDescribeCmd         *cobra.Command
	spannerDbListCmd             *cobra.Command
	spannerDbUpdateCmd           *cobra.Command
	spannerDbChangeQuorumCmd     *cobra.Command
	spannerDbDdlDescribeCmd      *cobra.Command
	spannerDbDdlUpdateCmd        *cobra.Command
	spannerDbExecuteSQLCmd       *cobra.Command
	spannerDbRestoreCmd          *cobra.Command
	spannerDbGetIamCmd           *cobra.Command
	spannerDbSetIamCmd           *cobra.Command
	spannerDbAddIamCmd           *cobra.Command
	spannerDbRemoveIamCmd        *cobra.Command
	spannerDbRolesListCmd        *cobra.Command
	spannerDbSessionsListCmd     *cobra.Command
	spannerDbSessionsDeleteCmd   *cobra.Command
	spannerDbSplitsAddCmd        *cobra.Command
	flagSpDbRolesDatabase        string
	flagSpDbSessionsDatabase     string
	spannerInstanceConfigsCmd    *cobra.Command
	flagSpICFormat               string
	flagSpICConfigFile           string
	flagSpICUpdateMask           string
	flagSpICPageSize             int64
	spannerICCreateCmd           *cobra.Command
	spannerICDeleteCmd           *cobra.Command
	spannerICDescribeCmd         *cobra.Command
	spannerICListCmd             *cobra.Command
	spannerICUpdateCmd           *cobra.Command
	spannerInstancePartitionsCmd *cobra.Command
	flagSpIPInstance             string
	flagSpIPFormat               string
	flagSpIPConfigFile           string
	flagSpIPUpdateMask           string
	flagSpIPPageSize             int64
	spannerIPCreateCmd           *cobra.Command
	spannerIPDeleteCmd           *cobra.Command
	spannerIPDescribeCmd         *cobra.Command
	spannerIPListCmd             *cobra.Command
	spannerIPUpdateCmd           *cobra.Command
	spannerInstancesCmd          *cobra.Command
	flagSpInstFormat             string
	flagSpInstConfigFile         string
	flagSpInstUpdateMask         string
	flagSpInstConfig             string
	flagSpInstNodes              int64
	flagSpInstPU                 int64
	flagSpInstDisplayName        string
	flagSpInstFilter             string
	flagSpInstPageSize           int64
	flagSpInstTargetCfg          string
	flagSpInstIamMember          string
	flagSpInstIamRole            string
	flagSpInstIamCondExpr        string
	flagSpInstIamCondT           string
	flagSpInstIamCondD           string
	flagSpInstIamAllCond         bool
	flagSpInstLabels             map[string]string
	flagSpInstUpdateLabels       map[string]string
	flagSpInstRemoveLabels       []string
	flagSpInstClearLabels        bool
	spannerInstCreateCmd         *cobra.Command
	spannerInstDeleteCmd         *cobra.Command
	spannerInstDescribeCmd       *cobra.Command
	spannerInstListCmd           *cobra.Command
	spannerInstUpdateCmd         *cobra.Command
	spannerInstMoveCmd           *cobra.Command
	spannerInstGetLocsCmd        *cobra.Command
	spannerInstGetIamCmd         *cobra.Command
	spannerInstSetIamCmd         *cobra.Command
	spannerInstAddIamCmd         *cobra.Command
	spannerInstRemoveIamCmd      *cobra.Command
	spannerOperationsCmd         *cobra.Command
	flagSpOpInstance             string
	flagSpOpDatabase             string
	flagSpOpBackup               string
	flagSpOpFormat               string
	flagSpOpFilter               string
	flagSpOpPageSize             int64
	spannerOpCancelCmd           *cobra.Command
	spannerOpDescribeCmd         *cobra.Command
	spannerOpListCmd             *cobra.Command
	spannerRowsCmd               *cobra.Command
	flagSpRowsInstance           string
	flagSpRowsDatabase           string
	flagSpRowsTable              string
	flagSpRowsFormat             string
	flagSpRowsKeys               []string
	flagSpRowsData               map[string]string
	spannerRowsInsertCmd         *cobra.Command
	spannerRowsUpdateCmd         *cobra.Command
	spannerRowsDeleteCmd         *cobra.Command
}

// newSpannerCmd returns the "spanner" command group for inv.
func newSpannerCmd(inv *invocation) *cobra.Command {
	return newSpannerCmds(inv).spannerCmd
}

// newSpannerCmds builds the "spanner" command group for inv and returns its state.
func newSpannerCmds(inv *invocation) *spannerCmds {
	s := &spannerCmds{inv: inv}
	s.spannerCmd = &cobra.Command{Use: "spanner", Short: "Manage Cloud Spanner"}
	s.spannerBackupSchedulesCmd = &cobra.Command{Use: "backup-schedules", Short: "Manage Spanner backup schedules"}
	s.spannerBsCreateCmd = &cobra.Command{
		Use: "create SCHEDULE", Short: "Create a Spanner backup schedule",
		Args: cobra.ExactArgs(1), RunE: s.runSpBsCreate,
	}
	s.spannerBsDeleteCmd = &cobra.Command{
		Use: "delete SCHEDULE", Short: "Delete a Spanner backup schedule",
		Args: cobra.ExactArgs(1), RunE: s.runSpBsDelete,
	}
	s.spannerBsDescribeCmd = &cobra.Command{
		Use: "describe SCHEDULE", Short: "Describe a Spanner backup schedule",
		Args: cobra.ExactArgs(1), RunE: s.runSpBsDescribe,
	}
	s.spannerBsListCmd = &cobra.Command{
		Use: "list", Short: "List Spanner backup schedules on a database",
		Args: cobra.NoArgs, RunE: s.runSpBsList,
	}
	s.spannerBsUpdateCmd = &cobra.Command{
		Use: "update SCHEDULE", Short: "Update a Spanner backup schedule",
		Args: cobra.ExactArgs(1), RunE: s.runSpBsUpdate,
	}
	s.spannerBsGetIamCmd = &cobra.Command{
		Use: "get-iam-policy SCHEDULE", Short: "Get the IAM policy for a Spanner backup schedule",
		Args: cobra.ExactArgs(1), RunE: s.runSpBsGetIam,
	}
	s.spannerBsSetIamCmd = &cobra.Command{
		Use: "set-iam-policy SCHEDULE POLICY_FILE", Short: "Set the IAM policy for a Spanner backup schedule",
		Args: cobra.ExactArgs(2), RunE: s.runSpBsSetIam,
	}
	s.spannerBackupsCmd = &cobra.Command{Use: "backups", Short: "Manage Cloud Spanner backups"}
	s.spannerBackupsCreateCmd = &cobra.Command{
		Use: "create BACKUP", Short: "Create a Spanner backup",
		Args: cobra.ExactArgs(1), RunE: s.runSpBackupCreate,
	}
	s.spannerBackupsDeleteCmd = &cobra.Command{
		Use: "delete BACKUP", Short: "Delete a Spanner backup",
		Args: cobra.ExactArgs(1), RunE: s.runSpBackupDelete,
	}
	s.spannerBackupsDescribeCmd = &cobra.Command{
		Use: "describe BACKUP", Short: "Describe a Spanner backup",
		Args: cobra.ExactArgs(1), RunE: s.runSpBackupDescribe,
	}
	s.spannerBackupsListCmd = &cobra.Command{
		Use: "list", Short: "List Spanner backups",
		Args: cobra.NoArgs, RunE: s.runSpBackupList,
	}
	s.spannerBackupsCopyCmd = &cobra.Command{
		Use: "copy", Short: "Copy a Spanner backup",
		Args: cobra.NoArgs, RunE: s.runSpBackupCopy,
	}
	s.spannerBackupsUpdateMetaCmd = &cobra.Command{
		Use: "update-metadata BACKUP", Short: "Update metadata (expire-time, version-time) of a Spanner backup",
		Args: cobra.ExactArgs(1), RunE: s.runSpBackupUpdateMeta,
	}
	s.spannerBackupsGetIamCmd = &cobra.Command{
		Use: "get-iam-policy BACKUP", Short: "Get the IAM policy for a Spanner backup",
		Args: cobra.ExactArgs(1), RunE: s.runSpBackupGetIam,
	}
	s.spannerBackupsSetIamCmd = &cobra.Command{
		Use: "set-iam-policy BACKUP POLICY_FILE", Short: "Set the IAM policy for a Spanner backup",
		Args: cobra.ExactArgs(2), RunE: s.runSpBackupSetIam,
	}
	s.spannerBackupsAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding BACKUP", Short: "Add an IAM policy binding to a Spanner backup",
		Args: cobra.ExactArgs(1), RunE: s.runSpBackupAddIam,
	}
	s.spannerBackupsRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding BACKUP", Short: "Remove an IAM policy binding from a Spanner backup",
		Args: cobra.ExactArgs(1), RunE: s.runSpBackupRemoveIam,
	}
	s.spannerDatabasesCmd = &cobra.Command{Use: "databases", Short: "Manage Cloud Spanner databases"}
	s.spannerDbDdlCmd = &cobra.Command{Use: "ddl", Short: "Manage database DDL"}
	s.spannerDbRolesCmd = &cobra.Command{Use: "roles", Short: "Manage database roles"}
	s.spannerDbSessionsCmd = &cobra.Command{Use: "sessions", Short: "Manage database sessions"}
	s.spannerDbSplitsCmd = &cobra.Command{Use: "splits", Short: "Manage database split points"}
	s.spannerDbCreateCmd = &cobra.Command{
		Use: "create DATABASE", Short: "Create a Cloud Spanner database",
		Args: cobra.ExactArgs(1), RunE: s.runSpDbCreate,
	}
	s.spannerDbDeleteCmd = &cobra.Command{
		Use: "delete DATABASE", Short: "Drop a Cloud Spanner database",
		Args: cobra.ExactArgs(1), RunE: s.runSpDbDelete,
	}
	s.spannerDbDescribeCmd = &cobra.Command{
		Use: "describe DATABASE", Short: "Describe a Cloud Spanner database",
		Args: cobra.ExactArgs(1), RunE: s.runSpDbDescribe,
	}
	s.spannerDbListCmd = &cobra.Command{
		Use: "list", Short: "List databases in an instance",
		Args: cobra.NoArgs, RunE: s.runSpDbList,
	}
	s.spannerDbUpdateCmd = &cobra.Command{
		Use: "update DATABASE", Short: "Update database metadata",
		Args: cobra.ExactArgs(1), RunE: s.runSpDbUpdate,
	}
	s.spannerDbChangeQuorumCmd = &cobra.Command{
		Use: "change-quorum DATABASE", Short: "Change the quorum of a Cloud Spanner database",
		Args: cobra.ExactArgs(1), RunE: s.runSpDbChangeQuorum,
	}
	s.spannerDbDdlDescribeCmd = &cobra.Command{
		Use: "describe DATABASE", Short: "Show the DDL statements for a database",
		Args: cobra.ExactArgs(1), RunE: s.runSpDbDdlDescribe,
	}
	s.spannerDbDdlUpdateCmd = &cobra.Command{
		Use: "update DATABASE", Short: "Apply DDL statements to a database",
		Args: cobra.ExactArgs(1), RunE: s.runSpDbDdlUpdate,
	}
	s.spannerDbExecuteSQLCmd = &cobra.Command{
		Use: "execute-sql DATABASE", Short: "Execute a SQL statement against a database",
		Args: cobra.ExactArgs(1), RunE: s.runSpDbExecuteSQL,
	}
	s.spannerDbRestoreCmd = &cobra.Command{
		Use: "restore DATABASE", Short: "Restore a database from a backup",
		Args: cobra.ExactArgs(1), RunE: s.runSpDbRestore,
	}
	s.spannerDbGetIamCmd = &cobra.Command{
		Use: "get-iam-policy DATABASE", Short: "Get the IAM policy for a database",
		Args: cobra.ExactArgs(1), RunE: s.runSpDbGetIam,
	}
	s.spannerDbSetIamCmd = &cobra.Command{
		Use: "set-iam-policy DATABASE POLICY_FILE", Short: "Set the IAM policy for a database",
		Args: cobra.ExactArgs(2), RunE: s.runSpDbSetIam,
	}
	s.spannerDbAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding DATABASE", Short: "Add an IAM policy binding on a database",
		Args: cobra.ExactArgs(1), RunE: s.runSpDbAddIam,
	}
	s.spannerDbRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding DATABASE", Short: "Remove an IAM policy binding from a database",
		Args: cobra.ExactArgs(1), RunE: s.runSpDbRemoveIam,
	}
	s.spannerDbRolesListCmd = &cobra.Command{
		Use: "list", Short: "List roles defined in a database",
		Args: cobra.NoArgs, RunE: s.runSpDbRolesList,
	}
	s.spannerDbSessionsListCmd = &cobra.Command{
		Use: "list", Short: "List sessions for a database",
		Args: cobra.NoArgs, RunE: s.runSpDbSessionsList,
	}
	s.spannerDbSessionsDeleteCmd = &cobra.Command{
		Use: "delete SESSION", Short: "Delete a session",
		Args: cobra.ExactArgs(1), RunE: s.runSpDbSessionsDelete,
	}
	s.spannerDbSplitsAddCmd = &cobra.Command{
		Use: "add DATABASE", Short: "Add split points to a database",
		Args: cobra.ExactArgs(1), RunE: s.runSpDbSplitsAdd,
	}
	s.spannerInstanceConfigsCmd = &cobra.Command{Use: "instance-configs", Short: "Manage Cloud Spanner instance configurations"}
	s.spannerICCreateCmd = &cobra.Command{
		Use: "create CONFIG", Short: "Create a Cloud Spanner instance configuration",
		Args: cobra.ExactArgs(1), RunE: s.runSpICCreate,
	}
	s.spannerICDeleteCmd = &cobra.Command{
		Use: "delete CONFIG", Short: "Delete a Cloud Spanner instance configuration",
		Args: cobra.ExactArgs(1), RunE: s.runSpICDelete,
	}
	s.spannerICDescribeCmd = &cobra.Command{
		Use: "describe CONFIG", Short: "Describe a Cloud Spanner instance configuration",
		Args: cobra.ExactArgs(1), RunE: s.runSpICDescribe,
	}
	s.spannerICListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Spanner instance configurations",
		Args: cobra.NoArgs, RunE: s.runSpICList,
	}
	s.spannerICUpdateCmd = &cobra.Command{
		Use: "update CONFIG", Short: "Update a Cloud Spanner instance configuration",
		Args: cobra.ExactArgs(1), RunE: s.runSpICUpdate,
	}
	s.spannerInstancePartitionsCmd = &cobra.Command{Use: "instance-partitions", Short: "Manage Cloud Spanner instance partitions"}
	s.spannerIPCreateCmd = &cobra.Command{
		Use: "create PARTITION", Short: "Create a Cloud Spanner instance partition",
		Args: cobra.ExactArgs(1), RunE: s.runSpIPCreate,
	}
	s.spannerIPDeleteCmd = &cobra.Command{
		Use: "delete PARTITION", Short: "Delete a Cloud Spanner instance partition",
		Args: cobra.ExactArgs(1), RunE: s.runSpIPDelete,
	}
	s.spannerIPDescribeCmd = &cobra.Command{
		Use: "describe PARTITION", Short: "Describe a Cloud Spanner instance partition",
		Args: cobra.ExactArgs(1), RunE: s.runSpIPDescribe,
	}
	s.spannerIPListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Spanner instance partitions",
		Args: cobra.NoArgs, RunE: s.runSpIPList,
	}
	s.spannerIPUpdateCmd = &cobra.Command{
		Use: "update PARTITION", Short: "Update a Cloud Spanner instance partition",
		Args: cobra.ExactArgs(1), RunE: s.runSpIPUpdate,
	}
	s.spannerInstancesCmd = &cobra.Command{Use: "instances", Short: "Manage Cloud Spanner instances"}
	s.spannerInstCreateCmd = &cobra.Command{
		Use: "create INSTANCE", Short: "Create a Cloud Spanner instance",
		Args: cobra.ExactArgs(1), RunE: s.runSpInstCreate,
	}
	s.spannerInstDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete a Cloud Spanner instance",
		Args: cobra.ExactArgs(1), RunE: s.runSpInstDelete,
	}
	s.spannerInstDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe a Cloud Spanner instance",
		Args: cobra.ExactArgs(1), RunE: s.runSpInstDescribe,
	}
	s.spannerInstListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Spanner instances",
		Args: cobra.NoArgs, RunE: s.runSpInstList,
	}
	s.spannerInstUpdateCmd = &cobra.Command{
		Use: "update INSTANCE", Short: "Update a Cloud Spanner instance",
		Args: cobra.ExactArgs(1), RunE: s.runSpInstUpdate,
	}
	s.spannerInstMoveCmd = &cobra.Command{
		Use: "move INSTANCE", Short: "Move a Cloud Spanner instance to a new instance configuration",
		Args: cobra.ExactArgs(1), RunE: s.runSpInstMove,
	}
	s.spannerInstGetLocsCmd = &cobra.Command{
		Use: "get-locations INSTANCE", Short: "Print the locations (replicas) of the instance's config",
		Args: cobra.ExactArgs(1), RunE: s.runSpInstGetLocations,
	}
	s.spannerInstGetIamCmd = &cobra.Command{
		Use: "get-iam-policy INSTANCE", Short: "Get the IAM policy for an instance",
		Args: cobra.ExactArgs(1), RunE: s.runSpInstGetIam,
	}
	s.spannerInstSetIamCmd = &cobra.Command{
		Use: "set-iam-policy INSTANCE POLICY_FILE", Short: "Set the IAM policy for an instance",
		Args: cobra.ExactArgs(2), RunE: s.runSpInstSetIam,
	}
	s.spannerInstAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding INSTANCE", Short: "Add an IAM policy binding on an instance",
		Args: cobra.ExactArgs(1), RunE: s.runSpInstAddIam,
	}
	s.spannerInstRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding INSTANCE", Short: "Remove an IAM policy binding from an instance",
		Args: cobra.ExactArgs(1), RunE: s.runSpInstRemoveIam,
	}
	s.spannerOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Cloud Spanner operations"}
	s.spannerOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Spanner operation",
		Args: cobra.ExactArgs(1), RunE: s.runSpOpCancel,
	}
	s.spannerOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Spanner operation",
		Args: cobra.ExactArgs(1), RunE: s.runSpOpDescribe,
	}
	s.spannerOpListCmd = &cobra.Command{
		Use: "list", Short: "List Spanner operations at the requested scope",
		Args: cobra.NoArgs, RunE: s.runSpOpList,
	}
	s.spannerRowsCmd = &cobra.Command{Use: "rows", Short: "Manage Cloud Spanner rows via DML"}
	s.spannerRowsInsertCmd = &cobra.Command{
		Use: "insert", Short: "Insert a row via DML",
		Args: cobra.NoArgs, RunE: s.runSpRowsInsert,
	}
	s.spannerRowsUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update rows via DML",
		Args: cobra.NoArgs, RunE: s.runSpRowsUpdate,
	}
	s.spannerRowsDeleteCmd = &cobra.Command{
		Use: "delete", Short: "Delete rows via DML",
		Args: cobra.NoArgs, RunE: s.runSpRowsDelete,
	}
	s.initSpanner()
	s.initSpannerBackupSchedules()
	s.initSpannerBackups()
	s.initSpannerDatabases()
	s.initSpannerInstanceConfigs()
	s.initSpannerInstancePartitions()
	s.initSpannerInstances()
	s.initSpannerOperations()
	s.initSpannerRows()
	return s
}
