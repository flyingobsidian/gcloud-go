package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud bigtable (#308) ---

func (s *bigtableCmds) initBigtable() {

}

// btProjectName returns "projects/PROJECT" for the resolved project.
func (s *bigtableCmds) btProjectName() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return "projects/" + project, nil
}

// btInstanceName returns projects/.../instances/INSTANCE.
func (s *bigtableCmds) btInstanceName(instance string) (string, error) {
	if instance == "" {
		return "", fmt.Errorf("--instance is required")
	}
	if strings.HasPrefix(instance, "projects/") {
		return instance, nil
	}
	project, err := s.btProjectName()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/instances/%s", project, instance), nil
}

// btClusterName returns projects/.../instances/INSTANCE/clusters/CLUSTER.
func (s *bigtableCmds) btClusterName(instance, cluster string) (string, error) {
	if cluster == "" {
		return "", fmt.Errorf("--cluster is required")
	}
	if strings.HasPrefix(cluster, "projects/") {
		return cluster, nil
	}
	inst, err := s.btInstanceName(instance)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/clusters/%s", inst, cluster), nil
}

// btTableName returns projects/.../instances/INSTANCE/tables/TABLE.
func (s *bigtableCmds) btTableName(instance, table string) (string, error) {
	if table == "" {
		return "", fmt.Errorf("--table is required")
	}
	if strings.HasPrefix(table, "projects/") {
		return table, nil
	}
	inst, err := s.btInstanceName(instance)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/tables/%s", inst, table), nil
}

// bigtableCmds holds the commands and flag values of one invocation's "bigtable"
// command group, so invocations don't share state (#2070).
type bigtableCmds struct {
	inv *invocation

	bigtableCmd                          *cobra.Command
	bigtableAppProfilesCmd               *cobra.Command
	flagBtAppInstance                    string
	flagBtAppFormat                      string
	flagBtAppConfigFile                  string
	flagBtAppUpdateMask                  string
	flagBtAppPageSize                    int64
	bigtableAppCreateCmd                 *cobra.Command
	bigtableAppDeleteCmd                 *cobra.Command
	bigtableAppDescribeCmd               *cobra.Command
	bigtableAppListCmd                   *cobra.Command
	bigtableAppUpdateCmd                 *cobra.Command
	bigtableAuthorizedViewsCmd           *cobra.Command
	flagBtAvInstance                     string
	flagBtAvTable                        string
	flagBtAvFormat                       string
	flagBtAvConfigFile                   string
	flagBtAvUpdateMask                   string
	flagBtAvPageSize                     int64
	bigtableAvCreateCmd                  *cobra.Command
	bigtableAvDeleteCmd                  *cobra.Command
	bigtableAvDescribeCmd                *cobra.Command
	bigtableAvListCmd                    *cobra.Command
	bigtableAvUpdateCmd                  *cobra.Command
	bigtableBackupsCmd                   *cobra.Command
	flagBtBkInstance                     string
	flagBtBkCluster                      string
	flagBtBkFormat                       string
	flagBtBkConfigFile                   string
	flagBtBkUpdateMask                   string
	flagBtBkPageSize                     int64
	flagBtBkFilter                       string
	flagBtBkOrderBy                      string
	flagBtBkDestTable                    string
	bigtableBkCreateCmd                  *cobra.Command
	bigtableBkDeleteCmd                  *cobra.Command
	bigtableBkDescribeCmd                *cobra.Command
	bigtableBkListCmd                    *cobra.Command
	bigtableBkUpdateCmd                  *cobra.Command
	bigtableBkRestoreCmd                 *cobra.Command
	bigtableClustersCmd                  *cobra.Command
	flagBtClInstance                     string
	flagBtClFormat                       string
	flagBtClConfigFile                   string
	bigtableClCreateCmd                  *cobra.Command
	bigtableClDeleteCmd                  *cobra.Command
	bigtableClDescribeCmd                *cobra.Command
	bigtableClListCmd                    *cobra.Command
	bigtableClUpdateCmd                  *cobra.Command
	bigtableHotTabletsCmd                *cobra.Command
	flagBtHtInstance                     string
	flagBtHtCluster                      string
	flagBtHtFormat                       string
	flagBtHtPageSize                     int64
	flagBtHtStartTime                    string
	flagBtHtEndTime                      string
	bigtableHtListCmd                    *cobra.Command
	bigtableInstancesCmd                 *cobra.Command
	flagBtInstFormat                     string
	flagBtInstConfigFile                 string
	flagBtInstTags                       map[string]string
	flagBtInstIamMember                  string
	flagBtInstIamRole                    string
	flagBtInstIamCondE                   string
	flagBtInstIamCondT                   string
	flagBtInstIamCondD                   string
	flagBtInstIamAllCond                 bool
	bigtableInstCreateCmd                *cobra.Command
	bigtableInstDeleteCmd                *cobra.Command
	bigtableInstDescribeCmd              *cobra.Command
	bigtableInstListCmd                  *cobra.Command
	bigtableInstUpdateCmd                *cobra.Command
	bigtableInstGetIamCmd                *cobra.Command
	bigtableInstSetIamCmd                *cobra.Command
	bigtableInstAddIamCmd                *cobra.Command
	bigtableInstRemoveIamCmd             *cobra.Command
	bigtableLogicalViewsCmd              *cobra.Command
	flagBtLvInstance                     string
	flagBtLvFormat                       string
	flagBtLvConfigFile                   string
	flagBtLvUpdateMask                   string
	flagBtLvPageSize                     int64
	bigtableLvCreateCmd                  *cobra.Command
	bigtableLvDeleteCmd                  *cobra.Command
	bigtableLvDescribeCmd                *cobra.Command
	bigtableLvListCmd                    *cobra.Command
	bigtableLvUpdateCmd                  *cobra.Command
	bigtableMVCmd                        *cobra.Command
	flagBTMVInstance                     string
	flagBTMVFormat                       string
	flagBTMVConfigFile                   string
	flagBTMVUpdateMask                   string
	flagBTMVIgnoreWarn                   bool
	flagBTMVPageSize                     int64
	bigtableMVCreateCmd                  *cobra.Command
	bigtableMVDeleteCmd                  *cobra.Command
	bigtableMVDescribeCmd                *cobra.Command
	bigtableMVListCmd                    *cobra.Command
	bigtableMVUpdateCmd                  *cobra.Command
	bigtableOpCmd                        *cobra.Command
	flagBTOpFormat                       string
	flagBTOpFilter                       string
	flagBTOpPageSize                     int64
	bigtableOpDescribeCmd                *cobra.Command
	bigtableOpListCmd                    *cobra.Command
	bigtableSBCmd                        *cobra.Command
	flagBTSBInstance                     string
	flagBTSBTable                        string
	flagBTSBFormat                       string
	flagBTSBConfigFile                   string
	flagBTSBUpdateMask                   string
	flagBTSBPageSize                     int64
	flagBTSBIamMember                    string
	flagBTSBIamRole                      string
	flagBTSBIamCondExpr                  string
	flagBTSBIamCondT                     string
	flagBTSBIamCondD                     string
	flagBTSBIamAllCond                   bool
	bigtableSBCreateCmd                  *cobra.Command
	bigtableSBDeleteCmd                  *cobra.Command
	bigtableSBDescribeCmd                *cobra.Command
	bigtableSBListCmd                    *cobra.Command
	bigtableSBUpdateCmd                  *cobra.Command
	bigtableSBGetIamCmd                  *cobra.Command
	bigtableSBSetIamCmd                  *cobra.Command
	bigtableSBAddIamCmd                  *cobra.Command
	bigtableSBRemoveIamCmd               *cobra.Command
	bigtableTablesCmd                    *cobra.Command
	flagBTTblInstance                    string
	flagBTTblFormat                      string
	flagBTTblConfigFile                  string
	flagBTTblUpdateMask                  string
	flagBTTblIgnoreWarn                  bool
	flagBTTblFilter                      string
	flagBTTblPageSize                    int64
	flagBTTblView                        string
	flagBTTblBackup                      string
	flagBTTblIamMember                   string
	flagBTTblIamRole                     string
	flagBTTblIamCondExpr                 string
	flagBTTblIamCondT                    string
	flagBTTblIamCondD                    string
	flagBTTblIamAllCond                  bool
	bigtableTablesCreateCmd              *cobra.Command
	bigtableTablesDeleteCmd              *cobra.Command
	bigtableTablesDescribeCmd            *cobra.Command
	bigtableTablesListCmd                *cobra.Command
	bigtableTablesRestoreCmd             *cobra.Command
	bigtableTablesUndeleteCmd            *cobra.Command
	bigtableTablesUpdateCmd              *cobra.Command
	bigtableTablesGetIamCmd              *cobra.Command
	bigtableTablesSetIamCmd              *cobra.Command
	bigtableTablesAddIamCmd              *cobra.Command
	bigtableTablesRemoveIamCmd           *cobra.Command
	flagBTTblRestoreDeletionProtection   bool
	flagBTTblRestoreNoDeletionProtection bool
	flagBTTblRestoreEnableAutoBackup     bool
	flagBTTblRestoreAutoBackupRetention  string
	flagBTTblRestoreAutoBackupLocations  []string
}

// newBigtableCmd returns the "bigtable" command group for inv.
func newBigtableCmd(inv *invocation) *cobra.Command {
	return newBigtableCmds(inv).bigtableCmd
}

// newBigtableCmds builds the "bigtable" command group for inv and returns its state.
func newBigtableCmds(inv *invocation) *bigtableCmds {
	s := &bigtableCmds{inv: inv}
	s.bigtableCmd = &cobra.Command{Use: "bigtable", Short: "Manage Cloud Bigtable"}
	s.bigtableAppProfilesCmd = &cobra.Command{Use: "app-profiles", Short: "Manage Bigtable app profiles"}
	s.bigtableAppCreateCmd = &cobra.Command{
		Use: "create APP_PROFILE", Short: "Create a Bigtable app profile",
		Args: cobra.ExactArgs(1), RunE: s.runBtAppCreate,
	}
	s.bigtableAppDeleteCmd = &cobra.Command{
		Use: "delete APP_PROFILE", Short: "Delete a Bigtable app profile",
		Args: cobra.ExactArgs(1), RunE: s.runBtAppDelete,
	}
	s.bigtableAppDescribeCmd = &cobra.Command{
		Use: "describe APP_PROFILE", Short: "Describe a Bigtable app profile",
		Args: cobra.ExactArgs(1), RunE: s.runBtAppDescribe,
	}
	s.bigtableAppListCmd = &cobra.Command{
		Use: "list", Short: "List Bigtable app profiles on an instance",
		Args: cobra.NoArgs, RunE: s.runBtAppList,
	}
	s.bigtableAppUpdateCmd = &cobra.Command{
		Use: "update APP_PROFILE", Short: "Update a Bigtable app profile",
		Args: cobra.ExactArgs(1), RunE: s.runBtAppUpdate,
	}
	s.bigtableAuthorizedViewsCmd = &cobra.Command{Use: "authorized-views", Short: "Manage Bigtable authorized views"}
	s.bigtableAvCreateCmd = &cobra.Command{
		Use: "create AUTHORIZED_VIEW", Short: "Create a Bigtable authorized view",
		Args: cobra.ExactArgs(1), RunE: s.runBtAvCreate,
	}
	s.bigtableAvDeleteCmd = &cobra.Command{
		Use: "delete AUTHORIZED_VIEW", Short: "Delete a Bigtable authorized view",
		Args: cobra.ExactArgs(1), RunE: s.runBtAvDelete,
	}
	s.bigtableAvDescribeCmd = &cobra.Command{
		Use: "describe AUTHORIZED_VIEW", Short: "Describe a Bigtable authorized view",
		Args: cobra.ExactArgs(1), RunE: s.runBtAvDescribe,
	}
	s.bigtableAvListCmd = &cobra.Command{
		Use: "list", Short: "List Bigtable authorized views on a table",
		Args: cobra.NoArgs, RunE: s.runBtAvList,
	}
	s.bigtableAvUpdateCmd = &cobra.Command{
		Use: "update AUTHORIZED_VIEW", Short: "Update a Bigtable authorized view",
		Args: cobra.ExactArgs(1), RunE: s.runBtAvUpdate,
	}
	s.bigtableBackupsCmd = &cobra.Command{Use: "backups", Short: "Manage Bigtable backups"}
	s.bigtableBkCreateCmd = &cobra.Command{
		Use: "create BACKUP", Short: "Create a Bigtable backup",
		Args: cobra.ExactArgs(1), RunE: s.runBtBkCreate,
	}
	s.bigtableBkDeleteCmd = &cobra.Command{
		Use: "delete BACKUP", Short: "Delete a Bigtable backup",
		Args: cobra.ExactArgs(1), RunE: s.runBtBkDelete,
	}
	s.bigtableBkDescribeCmd = &cobra.Command{
		Use: "describe BACKUP", Short: "Describe a Bigtable backup",
		Args: cobra.ExactArgs(1), RunE: s.runBtBkDescribe,
	}
	s.bigtableBkListCmd = &cobra.Command{
		Use: "list", Short: "List Bigtable backups on a cluster",
		Args: cobra.NoArgs, RunE: s.runBtBkList,
	}
	s.bigtableBkUpdateCmd = &cobra.Command{
		Use: "update BACKUP", Short: "Update a Bigtable backup",
		Args: cobra.ExactArgs(1), RunE: s.runBtBkUpdate,
	}
	s.bigtableBkRestoreCmd = &cobra.Command{
		Use: "restore BACKUP", Short: "Restore a Bigtable backup to a new table",
		Args: cobra.ExactArgs(1), RunE: s.runBtBkRestore,
	}
	s.bigtableClustersCmd = &cobra.Command{Use: "clusters", Short: "Manage Bigtable clusters"}
	s.bigtableClCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a Bigtable cluster",
		Args: cobra.ExactArgs(1), RunE: s.runBtClCreate,
	}
	s.bigtableClDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a Bigtable cluster",
		Args: cobra.ExactArgs(1), RunE: s.runBtClDelete,
	}
	s.bigtableClDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a Bigtable cluster",
		Args: cobra.ExactArgs(1), RunE: s.runBtClDescribe,
	}
	s.bigtableClListCmd = &cobra.Command{
		Use: "list", Short: "List Bigtable clusters on an instance",
		Args: cobra.NoArgs, RunE: s.runBtClList,
	}
	s.bigtableClUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update a Bigtable cluster",
		Args: cobra.ExactArgs(1), RunE: s.runBtClUpdate,
	}
	s.bigtableHotTabletsCmd = &cobra.Command{Use: "hot-tablets", Short: "Manage Bigtable hot tablets"}
	s.bigtableHtListCmd = &cobra.Command{
		Use: "list", Short: "List Bigtable hot tablets on a cluster",
		Args: cobra.NoArgs, RunE: s.runBtHtList,
	}
	s.bigtableInstancesCmd = &cobra.Command{Use: "instances", Short: "Manage Bigtable instances"}
	s.bigtableInstCreateCmd = &cobra.Command{
		Use: "create INSTANCE", Short: "Create a Bigtable instance",
		Args: cobra.ExactArgs(1), RunE: s.runBtInstCreate,
	}
	s.bigtableInstDeleteCmd = &cobra.Command{
		Use: "delete INSTANCE", Short: "Delete a Bigtable instance",
		Args: cobra.ExactArgs(1), RunE: s.runBtInstDelete,
	}
	s.bigtableInstDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe a Bigtable instance",
		Args: cobra.ExactArgs(1), RunE: s.runBtInstDescribe,
	}
	s.bigtableInstListCmd = &cobra.Command{
		Use: "list", Short: "List Bigtable instances in the project",
		Args: cobra.NoArgs, RunE: s.runBtInstList,
	}
	s.bigtableInstUpdateCmd = &cobra.Command{
		Use: "update INSTANCE", Short: "Update a Bigtable instance",
		Args: cobra.ExactArgs(1), RunE: s.runBtInstUpdate,
	}
	s.bigtableInstGetIamCmd = &cobra.Command{
		Use: "get-iam-policy INSTANCE", Short: "Get the IAM policy for a Bigtable instance",
		Args: cobra.ExactArgs(1), RunE: s.runBtInstGetIam,
	}
	s.bigtableInstSetIamCmd = &cobra.Command{
		Use: "set-iam-policy INSTANCE POLICY_FILE", Short: "Set the IAM policy for a Bigtable instance",
		Args: cobra.ExactArgs(2), RunE: s.runBtInstSetIam,
	}
	s.bigtableInstAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding INSTANCE", Short: "Add an IAM policy binding to a Bigtable instance",
		Args: cobra.ExactArgs(1), RunE: s.runBtInstAddIam,
	}
	s.bigtableInstRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding INSTANCE", Short: "Remove an IAM policy binding from a Bigtable instance",
		Args: cobra.ExactArgs(1), RunE: s.runBtInstRemoveIam,
	}
	s.bigtableLogicalViewsCmd = &cobra.Command{Use: "logical-views", Short: "Manage Bigtable logical views"}
	s.bigtableLvCreateCmd = &cobra.Command{
		Use: "create LOGICAL_VIEW", Short: "Create a Bigtable logical view",
		Args: cobra.ExactArgs(1), RunE: s.runBtLvCreate,
	}
	s.bigtableLvDeleteCmd = &cobra.Command{
		Use: "delete LOGICAL_VIEW", Short: "Delete a Bigtable logical view",
		Args: cobra.ExactArgs(1), RunE: s.runBtLvDelete,
	}
	s.bigtableLvDescribeCmd = &cobra.Command{
		Use: "describe LOGICAL_VIEW", Short: "Describe a Bigtable logical view",
		Args: cobra.ExactArgs(1), RunE: s.runBtLvDescribe,
	}
	s.bigtableLvListCmd = &cobra.Command{
		Use: "list", Short: "List Bigtable logical views on an instance",
		Args: cobra.NoArgs, RunE: s.runBtLvList,
	}
	s.bigtableLvUpdateCmd = &cobra.Command{
		Use: "update LOGICAL_VIEW", Short: "Update a Bigtable logical view",
		Args: cobra.ExactArgs(1), RunE: s.runBtLvUpdate,
	}
	s.bigtableMVCmd = &cobra.Command{Use: "materialized-views", Short: "Manage Cloud Bigtable materialized views"}
	s.bigtableMVCreateCmd = &cobra.Command{
		Use: "create MATERIALIZED_VIEW", Short: "Create a materialized view",
		Args: cobra.ExactArgs(1), RunE: s.runBTMVCreate,
	}
	s.bigtableMVDeleteCmd = &cobra.Command{
		Use: "delete MATERIALIZED_VIEW", Short: "Delete a materialized view",
		Args: cobra.ExactArgs(1), RunE: s.runBTMVDelete,
	}
	s.bigtableMVDescribeCmd = &cobra.Command{
		Use: "describe MATERIALIZED_VIEW", Short: "Describe a materialized view",
		Args: cobra.ExactArgs(1), RunE: s.runBTMVDescribe,
	}
	s.bigtableMVListCmd = &cobra.Command{
		Use: "list", Short: "List materialized views in an instance",
		Args: cobra.NoArgs, RunE: s.runBTMVList,
	}
	s.bigtableMVUpdateCmd = &cobra.Command{
		Use: "update MATERIALIZED_VIEW", Short: "Update a materialized view",
		Args: cobra.ExactArgs(1), RunE: s.runBTMVUpdate,
	}
	s.bigtableOpCmd = &cobra.Command{Use: "operations", Short: "Manage Cloud Bigtable operations"}
	s.bigtableOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Bigtable operation",
		Args: cobra.ExactArgs(1), RunE: s.runBTOpDescribe,
	}
	s.bigtableOpListCmd = &cobra.Command{
		Use: "list", Short: "List Bigtable operations for the current project",
		Args: cobra.NoArgs, RunE: s.runBTOpList,
	}
	s.bigtableSBCmd = &cobra.Command{Use: "schema-bundles", Short: "Manage Cloud Bigtable schema bundles"}
	s.bigtableSBCreateCmd = &cobra.Command{
		Use: "create SCHEMA_BUNDLE", Short: "Create a schema bundle",
		Args: cobra.ExactArgs(1), RunE: s.runBTSBCreate,
	}
	s.bigtableSBDeleteCmd = &cobra.Command{
		Use: "delete SCHEMA_BUNDLE", Short: "Delete a schema bundle",
		Args: cobra.ExactArgs(1), RunE: s.runBTSBDelete,
	}
	s.bigtableSBDescribeCmd = &cobra.Command{
		Use: "describe SCHEMA_BUNDLE", Short: "Describe a schema bundle",
		Args: cobra.ExactArgs(1), RunE: s.runBTSBDescribe,
	}
	s.bigtableSBListCmd = &cobra.Command{
		Use: "list", Short: "List schema bundles on a table",
		Args: cobra.NoArgs, RunE: s.runBTSBList,
	}
	s.bigtableSBUpdateCmd = &cobra.Command{
		Use: "update SCHEMA_BUNDLE", Short: "Update a schema bundle",
		Args: cobra.ExactArgs(1), RunE: s.runBTSBUpdate,
	}
	s.bigtableSBGetIamCmd = &cobra.Command{
		Use: "get-iam-policy SCHEMA_BUNDLE", Short: "Get the IAM policy for a schema bundle",
		Args: cobra.ExactArgs(1), RunE: s.runBTSBGetIam,
	}
	s.bigtableSBSetIamCmd = &cobra.Command{
		Use: "set-iam-policy SCHEMA_BUNDLE POLICY_FILE", Short: "Set the IAM policy for a schema bundle",
		Args: cobra.ExactArgs(2), RunE: s.runBTSBSetIam,
	}
	s.bigtableSBAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding SCHEMA_BUNDLE", Short: "Add an IAM policy binding to a schema bundle",
		Args: cobra.ExactArgs(1), RunE: s.runBTSBAddIam,
	}
	s.bigtableSBRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding SCHEMA_BUNDLE", Short: "Remove an IAM policy binding from a schema bundle",
		Args: cobra.ExactArgs(1), RunE: s.runBTSBRemoveIam,
	}
	s.bigtableTablesCmd = &cobra.Command{Use: "tables", Short: "Manage Cloud Bigtable tables"}
	s.bigtableTablesCreateCmd = &cobra.Command{
		Use: "create TABLE", Short: "Create a Bigtable table",
		Args: cobra.ExactArgs(1), RunE: s.runBTTblCreate,
	}
	s.bigtableTablesDeleteCmd = &cobra.Command{
		Use: "delete TABLE", Short: "Delete a Bigtable table",
		Args: cobra.ExactArgs(1), RunE: s.runBTTblDelete,
	}
	s.bigtableTablesDescribeCmd = &cobra.Command{
		Use: "describe TABLE", Short: "Describe a Bigtable table",
		Args: cobra.ExactArgs(1), RunE: s.runBTTblDescribe,
	}
	s.bigtableTablesListCmd = &cobra.Command{
		Use: "list", Short: "List Bigtable tables in an instance",
		Args: cobra.NoArgs, RunE: s.runBTTblList,
	}
	s.bigtableTablesRestoreCmd = &cobra.Command{
		Use: "restore TABLE", Short: "Restore a Bigtable table from a backup",
		Args: cobra.ExactArgs(1), RunE: s.runBTTblRestore,
	}
	s.bigtableTablesUndeleteCmd = &cobra.Command{
		Use: "undelete TABLE", Short: "Undelete a soft-deleted Bigtable table",
		Args: cobra.ExactArgs(1), RunE: s.runBTTblUndelete,
	}
	s.bigtableTablesUpdateCmd = &cobra.Command{
		Use: "update TABLE", Short: "Update a Bigtable table",
		Args: cobra.ExactArgs(1), RunE: s.runBTTblUpdate,
	}
	s.bigtableTablesGetIamCmd = &cobra.Command{
		Use: "get-iam-policy TABLE", Short: "Get the IAM policy for a Bigtable table",
		Args: cobra.ExactArgs(1), RunE: s.runBTTblGetIam,
	}
	s.bigtableTablesSetIamCmd = &cobra.Command{
		Use: "set-iam-policy TABLE POLICY_FILE", Short: "Set the IAM policy for a Bigtable table",
		Args: cobra.ExactArgs(2), RunE: s.runBTTblSetIam,
	}
	s.bigtableTablesAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding TABLE", Short: "Add an IAM policy binding to a Bigtable table",
		Args: cobra.ExactArgs(1), RunE: s.runBTTblAddIam,
	}
	s.bigtableTablesRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding TABLE", Short: "Remove an IAM policy binding from a Bigtable table",
		Args: cobra.ExactArgs(1), RunE: s.runBTTblRemoveIam,
	}
	s.initBigtable()
	s.initBigtableAppProfiles()
	s.initBigtableAuthorizedViews()
	s.initBigtableBackups()
	s.initBigtableClusters()
	s.initBigtableHotTablets()
	s.initBigtableInstances()
	s.initBigtableLogicalViews()
	s.initBigtableMaterializedViews()
	s.initBigtableOperations()
	s.initBigtableSchemaBundles()
	s.initBigtableTables()
	s.initBigtableTablesRestore()
	return s
}
