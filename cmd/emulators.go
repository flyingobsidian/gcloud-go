package cmd

import "github.com/spf13/cobra"

// --- gcloud emulators (#335) ---

func (s *emulatorsCmds) initEmulators() {
	registerStubGroup(s.emulatorsCmd, "firestore", "Manage local Firestore emulator", "start", "env-init")
	registerStubGroup(s.emulatorsCmd, "spanner", "Manage local Spanner emulator", "start", "env-init")
	registerStubGroup(s.emulatorsCmd, "bigtable", "Manage local Bigtable emulator", "start", "env-init")
	registerStubGroup(s.emulatorsCmd, "datastore", "Manage local Datastore emulator", "start", "env-init", "env-unset")
	registerStubGroup(s.emulatorsCmd, "pubsub", "Manage local Pub/Sub emulator", "start", "env-init")

}

// emulatorsCmds holds the commands and flag values of one invocation's "emulators"
// command group, so invocations don't share state (#2070).
type emulatorsCmds struct {
	inv *invocation

	emulatorsCmd                *cobra.Command
	flagEmuBigtableHostPort     string
	emuBigtableCmd              *cobra.Command
	emuBigtableStartCmd         *cobra.Command
	emuBigtableEnvInitCmd       *cobra.Command
	flagEmuDatastoreHostPort    string
	flagEmuDatastoreDataDir     string
	flagEmuDatastoreStoreOnDisk bool
	flagEmuDatastoreConsistency string
	flagEmuDatastoreFirestore   bool
	emuDatastoreCmd             *cobra.Command
	emuDatastoreStartCmd        *cobra.Command
	emuDatastoreEnvInitCmd      *cobra.Command
	emuDatastoreEnvUnsetCmd     *cobra.Command
	flagEmuPubsubHostPort       string
	flagEmuPubsubDataDir        string
	emuPubsubCmd                *cobra.Command
	emuPubsubStartCmd           *cobra.Command
	emuPubsubEnvInitCmd         *cobra.Command
}

// newEmulatorsCmd returns the "emulators" command group for inv.
func newEmulatorsCmd(inv *invocation) *cobra.Command {
	return newEmulatorsCmds(inv).emulatorsCmd
}

// newEmulatorsCmds builds the "emulators" command group for inv and returns its state.
func newEmulatorsCmds(inv *invocation) *emulatorsCmds {
	s := &emulatorsCmds{inv: inv}
	s.emulatorsCmd = &cobra.Command{Use: "emulators", Short: "Local emulators"}
	s.emuBigtableCmd = &cobra.Command{Use: "bigtable", Short: "Manage local Bigtable emulator"}
	s.emuBigtableStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Start a local Bigtable emulator (wraps cbtemulator)",
		Args:  cobra.NoArgs,
		RunE:  s.runEmuBigtableStart,
	}
	s.emuBigtableEnvInitCmd = &cobra.Command{
		Use:   "env-init",
		Short: "Print shell exports for BIGTABLE_EMULATOR_HOST",
		Args:  cobra.NoArgs,
		RunE:  s.runEmuBigtableEnvInit,
	}
	s.emuDatastoreCmd = &cobra.Command{Use: "datastore", Short: "Manage local Cloud Datastore emulator"}
	s.emuDatastoreStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Start a local Datastore emulator (wraps cloud_datastore_emulator)",
		Args:  cobra.NoArgs,
		RunE:  s.runEmuDatastoreStart,
	}
	s.emuDatastoreEnvInitCmd = &cobra.Command{
		Use:   "env-init",
		Short: "Print shell exports for DATASTORE_* env variables",
		Args:  cobra.NoArgs,
		RunE:  s.runEmuDatastoreEnvInit,
	}
	s.emuDatastoreEnvUnsetCmd = &cobra.Command{
		Use:   "env-unset",
		Short: "Print shell unsets for DATASTORE_* env variables",
		Args:  cobra.NoArgs,
		RunE:  s.runEmuDatastoreEnvUnset,
	}
	s.emuPubsubCmd = &cobra.Command{Use: "pubsub", Short: "Manage local Cloud Pub/Sub emulator"}
	s.emuPubsubStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Start a local Pub/Sub emulator (wraps pubsub-emulator)",
		Args:  cobra.NoArgs,
		RunE:  s.runEmuPubsubStart,
	}
	s.emuPubsubEnvInitCmd = &cobra.Command{
		Use:   "env-init",
		Short: "Print shell exports for PUBSUB_EMULATOR_HOST and PUBSUB_PROJECT_ID",
		Args:  cobra.NoArgs,
		RunE:  s.runEmuPubsubEnvInit,
	}
	s.initEmulators()
	s.initEmulatorsBigtable()
	s.initEmulatorsDatastore()
	s.initEmulatorsPubsub()
	return s
}
