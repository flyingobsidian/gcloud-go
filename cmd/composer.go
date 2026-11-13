package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// --- gcloud composer (#319) ---

func (s *composerCmds) initComposer() {

}

func composerLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func composerEnvName(project, location, env string) string {
	if strings.HasPrefix(env, "projects/") {
		return env
	}
	return fmt.Sprintf("%s/environments/%s", composerLocationParent(project, location), env)
}

// composerCmds holds the commands and flag values of one invocation's "composer"
// command group, so invocations don't share state (#2070).
type composerCmds struct {
	inv *invocation

	composerCmd                           *cobra.Command
	composerEnvCmd                        *cobra.Command
	flagComposerEnvLocation               string
	flagComposerEnvFormat                 string
	flagComposerEnvConfigFile             string
	flagComposerEnvUpdateMask             string
	flagComposerEnvFilter                 string
	flagComposerEnvListLocations          []string
	flagComposerEnvListLimit              int64
	flagComposerEnvPageSize               int64
	flagComposerEnvImageVersion           string
	flagComposerEnvSnapLocation           string
	flagComposerEnvSnapPath               string
	flagComposerEnvSkipOverrides          bool
	flagComposerEnvSkipEnv                bool
	flagComposerEnvSkipData               bool
	flagComposerEnvSkipPypi               bool
	flagComposerEnvRunParams              []string
	flagComposerEnvRunTimeout             time.Duration
	flagComposerEnvRunTree                bool
	composerEnvCreateCmd                  *cobra.Command
	composerEnvDeleteCmd                  *cobra.Command
	composerEnvDescribeCmd                *cobra.Command
	composerEnvListCmd                    *cobra.Command
	composerEnvUpdateCmd                  *cobra.Command
	composerEnvCheckUpgradeCmd            *cobra.Command
	composerEnvDatabaseFailoverCmd        *cobra.Command
	composerEnvFetchDatabasePropertiesCmd *cobra.Command
	composerEnvRestartWebServerCmd        *cobra.Command
	composerEnvListWorkloadsCmd           *cobra.Command
	composerEnvListUpgradesCmd            *cobra.Command
	composerEnvListPackagesCmd            *cobra.Command
	// composerEnvRunCmd forwards the caller's subcommand to Composer's
	// ExecuteAirflowCommand API without client-side allowlisting. That means
	// Airflow CLI support tracked by gcloud-python's SUBCOMMAND_ALLOWLIST
	// (backfill on Airflow 3, `config lint` on Airflow >=2.11, and any commands
	// enabled for the Airflow 3.2/3.3.x tracks introduced in gcloud-python
	// 578.0.0–582.0.0) is naturally available here as soon as the server
	// exposes it — no additional gating needed on our side.
	composerEnvRunCmd *cobra.Command
	// flagComposerEnvDevelopmentMode is --enable-development-mode on `create`
	// (gcloud-python 583.0.0), which sets Environment.mode=DEVELOPMENT.
	flagComposerEnvDevelopmentMode bool
	composerEnvHibernateCmd        *cobra.Command
	composerEnvResumeCmd           *cobra.Command
	flagComposerEnvPowerAsync      bool
	composerEnvSnapshotsCmd        *cobra.Command
	composerEnvSnapSaveCmd         *cobra.Command
	composerEnvSnapLoadCmd         *cobra.Command
	composerEnvStorageCmd          *cobra.Command
	composerEnvStorageDagsCmd      *cobra.Command
	composerEnvStorageDataCmd      *cobra.Command
	composerEnvStoragePluginsCmd   *cobra.Command
	flagComposerStEnvironment      string
	flagComposerStSource           string
	flagComposerStDestination      string
	flagComposerStSubdir           string
	composerEnvCMCmd               *cobra.Command
	composerEnvSecretsCmd          *cobra.Command
	flagComposerUWEnv              string
	flagComposerUWConfigFile       string
	flagComposerUWPageSize         int64
	composerCMCreateCmd            *cobra.Command
	composerCMDeleteCmd            *cobra.Command
	composerCMDescribeCmd          *cobra.Command
	composerCMListCmd              *cobra.Command
	composerCMUpdateCmd            *cobra.Command
	composerSecCreateCmd           *cobra.Command
	composerSecDeleteCmd           *cobra.Command
	composerSecDescribeCmd         *cobra.Command
	composerSecListCmd             *cobra.Command
	composerSecUpdateCmd           *cobra.Command
	composerOpCmd                  *cobra.Command
	flagComposerOpLocation         string
	flagComposerOpFormat           string
	flagComposerOpFilter           string
	flagComposerOpPageSize         int64
	flagComposerOpTimeout          time.Duration
	composerOpDeleteCmd            *cobra.Command
	composerOpDescribeCmd          *cobra.Command
	composerOpListCmd              *cobra.Command
	composerOpWaitCmd              *cobra.Command
	composerStorageKinds           []*composerStorageOp
}

// newComposerCmd returns the "composer" command group for inv.
func newComposerCmd(inv *invocation) *cobra.Command {
	return newComposerCmds(inv).composerCmd
}

// newComposerCmds builds the "composer" command group for inv and returns its state.
func newComposerCmds(inv *invocation) *composerCmds {
	s := &composerCmds{inv: inv}
	s.composerCmd = &cobra.Command{Use: "composer", Short: "Manage Cloud Composer"}
	s.composerEnvCmd = &cobra.Command{Use: "environments", Short: "Manage Composer environments"}
	s.composerEnvCreateCmd = &cobra.Command{
		Use: "create ENVIRONMENT", Short: "Create a Composer environment",
		Args: cobra.ExactArgs(1), RunE: s.runComposerEnvCreate,
	}
	s.composerEnvDeleteCmd = &cobra.Command{
		Use: "delete ENVIRONMENT", Short: "Delete a Composer environment",
		Args: cobra.ExactArgs(1), RunE: s.runComposerEnvDelete,
	}
	s.composerEnvDescribeCmd = &cobra.Command{
		Use: "describe ENVIRONMENT", Short: "Describe a Composer environment",
		Args: cobra.ExactArgs(1), RunE: s.runComposerEnvDescribe,
	}
	s.composerEnvListCmd = &cobra.Command{
		Use: "list", Short: "List Composer environments",
		Args: cobra.NoArgs, RunE: s.runComposerEnvList,
	}
	s.composerEnvUpdateCmd = &cobra.Command{
		Use: "update ENVIRONMENT", Short: "Update a Composer environment",
		Args: cobra.ExactArgs(1), RunE: s.runComposerEnvUpdate,
	}
	s.composerEnvCheckUpgradeCmd = &cobra.Command{
		Use: "check-upgrade ENVIRONMENT", Short: "Check whether the environment can be upgraded to an image version",
		Args: cobra.ExactArgs(1), RunE: s.runComposerEnvCheckUpgrade,
	}
	s.composerEnvDatabaseFailoverCmd = &cobra.Command{
		Use: "database-failover ENVIRONMENT", Short: "Trigger a Cloud SQL failover for a Composer environment",
		Args: cobra.ExactArgs(1), RunE: s.runComposerEnvDatabaseFailover,
	}
	s.composerEnvFetchDatabasePropertiesCmd = &cobra.Command{
		Use: "fetch-database-properties ENVIRONMENT", Short: "Fetch Cloud SQL database properties for a Composer environment",
		Args: cobra.ExactArgs(1), RunE: s.runComposerEnvFetchDatabaseProperties,
	}
	s.composerEnvRestartWebServerCmd = &cobra.Command{
		Use: "restart-web-server ENVIRONMENT", Short: "Restart the Composer web server",
		Args: cobra.ExactArgs(1), RunE: s.runComposerEnvRestartWebServer,
	}
	s.composerEnvListWorkloadsCmd = &cobra.Command{
		Use: "list-workloads ENVIRONMENT", Short: "List Airflow workloads in a Composer environment",
		Args: cobra.ExactArgs(1), RunE: s.runComposerEnvListWorkloads,
	}
	s.composerEnvListUpgradesCmd = &cobra.Command{
		Use:   "list-upgrades ENVIRONMENT",
		Short: "List the image versions this environment can upgrade to",
		Args:  cobra.ExactArgs(1), RunE: s.runComposerEnvListUpgrades,
	}
	s.composerEnvListPackagesCmd = &cobra.Command{
		Use:   "list-packages ENVIRONMENT",
		Short: "List Python packages installed in the Airflow worker",
		Args:  cobra.ExactArgs(1), RunE: s.runComposerEnvListPackages,
	}
	s.composerEnvRunCmd = &cobra.Command{
		Use:   "run ENVIRONMENT SUBCOMMAND [-- ARG ...]",
		Short: "Run an Airflow CLI subcommand in the environment (via composer's ExecuteAirflowCommand API)",
		Args:  cobra.MinimumNArgs(2), RunE: s.runComposerEnvRun,
	}
	s.composerEnvHibernateCmd = &cobra.Command{
		Use: "hibernate ENVIRONMENT", Short: "Hibernate a Cloud Composer environment",
		Args: cobra.ExactArgs(1), RunE: s.composerEnvPowerFn(composerHibernate),
	}
	s.composerEnvResumeCmd = &cobra.Command{
		Use: "resume ENVIRONMENT", Short: "Resume a Cloud Composer environment",
		Args: cobra.ExactArgs(1), RunE: s.composerEnvPowerFn(composerResume),
	}
	s.composerEnvSnapshotsCmd = &cobra.Command{Use: "snapshots", Short: "Save and load Composer environment snapshots"}
	s.composerEnvSnapSaveCmd = &cobra.Command{
		Use: "save ENVIRONMENT", Short: "Save a snapshot of a Composer environment",
		Args: cobra.ExactArgs(1), RunE: s.runComposerEnvSnapSave,
	}
	s.composerEnvSnapLoadCmd = &cobra.Command{
		Use: "load ENVIRONMENT", Short: "Load a snapshot into a Composer environment",
		Args: cobra.ExactArgs(1), RunE: s.runComposerEnvSnapLoad,
	}
	s.composerEnvStorageCmd = &cobra.Command{Use: "storage", Short: "Manage files in a Composer environment's Cloud Storage bucket"}
	s.composerEnvStorageDagsCmd = &cobra.Command{Use: "dags", Short: "Manage the DAGs folder"}
	s.composerEnvStorageDataCmd = &cobra.Command{Use: "data", Short: "Manage the data folder"}
	s.composerEnvStoragePluginsCmd = &cobra.Command{Use: "plugins", Short: "Manage the plugins folder"}
	s.composerEnvCMCmd = &cobra.Command{Use: "user-workloads-config-maps", Short: "Manage user workload ConfigMaps"}
	s.composerEnvSecretsCmd = &cobra.Command{Use: "user-workloads-secrets", Short: "Manage user workload Secrets"}
	s.composerCMCreateCmd = &cobra.Command{
		Use: "create NAME", Short: "Create a user workload ConfigMap",
		Args: cobra.ExactArgs(1), RunE: s.runComposerCMCreate,
	}
	s.composerCMDeleteCmd = &cobra.Command{
		Use: "delete NAME", Short: "Delete a user workload ConfigMap",
		Args: cobra.ExactArgs(1), RunE: s.runComposerCMDelete,
	}
	s.composerCMDescribeCmd = &cobra.Command{
		Use: "describe NAME", Short: "Describe a user workload ConfigMap",
		Args: cobra.ExactArgs(1), RunE: s.runComposerCMDescribe,
	}
	s.composerCMListCmd = &cobra.Command{
		Use: "list", Short: "List user workload ConfigMaps",
		Args: cobra.NoArgs, RunE: s.runComposerCMList,
	}
	s.composerCMUpdateCmd = &cobra.Command{
		Use: "update NAME", Short: "Update a user workload ConfigMap",
		Args: cobra.ExactArgs(1), RunE: s.runComposerCMUpdate,
	}
	s.composerSecCreateCmd = &cobra.Command{
		Use: "create NAME", Short: "Create a user workload Secret",
		Args: cobra.ExactArgs(1), RunE: s.runComposerSecCreate,
	}
	s.composerSecDeleteCmd = &cobra.Command{
		Use: "delete NAME", Short: "Delete a user workload Secret",
		Args: cobra.ExactArgs(1), RunE: s.runComposerSecDelete,
	}
	s.composerSecDescribeCmd = &cobra.Command{
		Use: "describe NAME", Short: "Describe a user workload Secret",
		Args: cobra.ExactArgs(1), RunE: s.runComposerSecDescribe,
	}
	s.composerSecListCmd = &cobra.Command{
		Use: "list", Short: "List user workload Secrets",
		Args: cobra.NoArgs, RunE: s.runComposerSecList,
	}
	s.composerSecUpdateCmd = &cobra.Command{
		Use: "update NAME", Short: "Update a user workload Secret",
		Args: cobra.ExactArgs(1), RunE: s.runComposerSecUpdate,
	}
	s.composerOpCmd = &cobra.Command{Use: "operations", Short: "Manage Cloud Composer operations"}
	s.composerOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a Composer operation record",
		Args: cobra.ExactArgs(1), RunE: s.runComposerOpDelete,
	}
	s.composerOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Composer operation",
		Args: cobra.ExactArgs(1), RunE: s.runComposerOpDescribe,
	}
	s.composerOpListCmd = &cobra.Command{
		Use: "list", Short: "List Composer operations",
		Args: cobra.NoArgs, RunE: s.runComposerOpList,
	}
	s.composerOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for a Composer operation to complete",
		Args: cobra.ExactArgs(1), RunE: s.runComposerOpWait,
	}
	s.composerStorageKinds = []*composerStorageOp{
		{kind: "dags", parent: s.composerEnvStorageDagsCmd},
		{kind: "data", parent: s.composerEnvStorageDataCmd},
		{kind: "plugins", parent: s.composerEnvStoragePluginsCmd},
	}
	s.initComposer()
	s.initComposerEnvironments()
	s.initComposerEnvironments2()
	s.initComposerEnvironmentsHibernate()
	s.initComposerEnvironmentsSnapshots()
	s.initComposerEnvironmentsStorage()
	s.initComposerEnvironmentsUserWorkloads()
	s.initComposerOperations()
	return s
}
