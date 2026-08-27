package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// --- gcloud emulators datastore (#1710) ---

var (
	flagEmuDatastoreHostPort    string
	flagEmuDatastoreDataDir     string
	flagEmuDatastoreStoreOnDisk bool
	flagEmuDatastoreConsistency string
	flagEmuDatastoreFirestore   bool
)

var emuDatastoreCmd = &cobra.Command{Use: "datastore", Short: "Manage local Cloud Datastore emulator"}

var emuDatastoreStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a local Datastore emulator (wraps cloud_datastore_emulator)",
	Args:  cobra.NoArgs,
	RunE:  runEmuDatastoreStart,
}

var emuDatastoreEnvInitCmd = &cobra.Command{
	Use:   "env-init",
	Short: "Print shell exports for DATASTORE_* env variables",
	Args:  cobra.NoArgs,
	RunE:  runEmuDatastoreEnvInit,
}

var emuDatastoreEnvUnsetCmd = &cobra.Command{
	Use:   "env-unset",
	Short: "Print shell unsets for DATASTORE_* env variables",
	Args:  cobra.NoArgs,
	RunE:  runEmuDatastoreEnvUnset,
}

func init() {
	for _, c := range []*cobra.Command{emuDatastoreStartCmd, emuDatastoreEnvInitCmd, emuDatastoreEnvUnsetCmd} {
		c.Flags().StringVar(&flagEmuDatastoreHostPort, "host-port", "localhost:8081",
			"The host:port to which the emulator is (or will be) bound")
		c.Flags().StringVar(&flagEmuDatastoreDataDir, "data-dir", "",
			"Directory used to store/retrieve data (defaults to gcloud's user config dir)")
	}
	emuDatastoreStartCmd.Flags().BoolVar(&flagEmuDatastoreStoreOnDisk, "store-on-disk", true,
		"Persist data to disk; use --no-store-on-disk to disable")
	emuDatastoreStartCmd.Flags().StringVar(&flagEmuDatastoreConsistency, "consistency", "",
		"Fraction of eventually consistent operations that succeed immediately (default 0.9)")
	emuDatastoreStartCmd.Flags().BoolVar(&flagEmuDatastoreFirestore, "use-firestore-in-datastore-mode", false,
		"Run the emulator in Firestore-in-Datastore-Mode; incompatible with --consistency")

	emuDatastoreCmd.AddCommand(emuDatastoreStartCmd, emuDatastoreEnvInitCmd, emuDatastoreEnvUnsetCmd)
	registerEmulatorGroup(emuDatastoreCmd)
}

func runEmuDatastoreStart(cmd *cobra.Command, args []string) error {
	if flagEmuDatastoreFirestore && flagEmuDatastoreConsistency != "" {
		return fmt.Errorf("--use-firestore-in-datastore-mode and --consistency are mutually exclusive")
	}
	argv := []string{"start", "--host=" + hostFrom(flagEmuDatastoreHostPort), "--port=" + portFrom(flagEmuDatastoreHostPort)}
	if flagEmuDatastoreDataDir != "" {
		argv = append(argv, "--data_dir="+flagEmuDatastoreDataDir)
	}
	if !flagEmuDatastoreStoreOnDisk {
		argv = append(argv, "--store_on_disk=false")
	}
	if flagEmuDatastoreConsistency != "" {
		argv = append(argv, "--consistency="+flagEmuDatastoreConsistency)
	}
	if flagEmuDatastoreFirestore {
		argv = append(argv, "--use_firestore_in_datastore_mode=true")
	}
	return runEmulator("cloud_datastore_emulator", "cloud-datastore-emulator", argv)
}

func runEmuDatastoreEnvInit(cmd *cobra.Command, args []string) error {
	// Match gcloud-python's env exports: DATASTORE_EMULATOR_HOST plus the
	// derived DATASTORE_PROJECT_ID / DATASTORE_HOST / DATASTORE_DATASET.
	host := flagEmuDatastoreHostPort
	project := resolveProjectOr("")
	vars := map[string]string{
		"DATASTORE_EMULATOR_HOST":      host,
		"DATASTORE_EMULATOR_HOST_PATH": host + "/datastore",
		"DATASTORE_HOST":               "http://" + host,
	}
	if project != "" {
		vars["DATASTORE_PROJECT_ID"] = project
		vars["DATASTORE_DATASET"] = project
	}
	return printEnvInit(vars)
}

func runEmuDatastoreEnvUnset(cmd *cobra.Command, args []string) error {
	return printEnvUnset([]string{
		"DATASTORE_EMULATOR_HOST",
		"DATASTORE_EMULATOR_HOST_PATH",
		"DATASTORE_HOST",
		"DATASTORE_PROJECT_ID",
		"DATASTORE_DATASET",
	})
}

// resolveProjectOr returns the active project (via resolveProject) or the
// given default if resolution fails. Used by env-init so a missing project
// doesn't turn into a full error.
func resolveProjectOr(dflt string) string {
	p, err := resolveProject()
	if err != nil {
		return dflt
	}
	return p
}
