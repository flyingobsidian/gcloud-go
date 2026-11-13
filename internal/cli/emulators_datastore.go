package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// --- gcloud emulators datastore (#1710) ---

func (s *emulatorsCmds) initEmulatorsDatastore() {
	for _, c := range []*cobra.Command{s.emuDatastoreStartCmd, s.emuDatastoreEnvInitCmd, s.emuDatastoreEnvUnsetCmd} {
		c.Flags().StringVar(&s.flagEmuDatastoreHostPort, "host-port", "localhost:8081",
			"The host:port to which the emulator is (or will be) bound")
		c.Flags().StringVar(&s.flagEmuDatastoreDataDir, "data-dir", "",
			"Directory used to store/retrieve data (defaults to gcloud's user config dir)")
	}
	s.emuDatastoreStartCmd.Flags().BoolVar(&s.flagEmuDatastoreStoreOnDisk, "store-on-disk", true,
		"Persist data to disk; use --no-store-on-disk to disable")
	s.emuDatastoreStartCmd.Flags().StringVar(&s.flagEmuDatastoreConsistency, "consistency", "",
		"Fraction of eventually consistent operations that succeed immediately (default 0.9)")
	s.emuDatastoreStartCmd.Flags().BoolVar(&s.flagEmuDatastoreFirestore, "use-firestore-in-datastore-mode", false,
		"Run the emulator in Firestore-in-Datastore-Mode; incompatible with --consistency")

	s.emuDatastoreCmd.AddCommand(s.emuDatastoreStartCmd, s.emuDatastoreEnvInitCmd, s.emuDatastoreEnvUnsetCmd)
	s.registerEmulatorGroup(s.emuDatastoreCmd)
}

func (s *emulatorsCmds) runEmuDatastoreStart(cmd *cobra.Command, args []string) error {
	if s.flagEmuDatastoreFirestore && s.flagEmuDatastoreConsistency != "" {
		return fmt.Errorf("--use-firestore-in-datastore-mode and --consistency are mutually exclusive")
	}
	argv := []string{"start", "--host=" + hostFrom(s.flagEmuDatastoreHostPort), "--port=" + portFrom(s.flagEmuDatastoreHostPort)}
	if s.flagEmuDatastoreDataDir != "" {
		argv = append(argv, "--data_dir="+s.flagEmuDatastoreDataDir)
	}
	if !s.flagEmuDatastoreStoreOnDisk {
		argv = append(argv, "--store_on_disk=false")
	}
	if s.flagEmuDatastoreConsistency != "" {
		argv = append(argv, "--consistency="+s.flagEmuDatastoreConsistency)
	}
	if s.flagEmuDatastoreFirestore {
		argv = append(argv, "--use_firestore_in_datastore_mode=true")
	}
	return s.runEmulator("cloud_datastore_emulator", "cloud-datastore-emulator", argv)
}

func (s *emulatorsCmds) runEmuDatastoreEnvInit(cmd *cobra.Command, args []string) error {
	// Match gcloud-python's env exports: DATASTORE_EMULATOR_HOST plus the
	// derived DATASTORE_PROJECT_ID / DATASTORE_HOST / DATASTORE_DATASET.
	host := s.flagEmuDatastoreHostPort
	project := s.resolveProjectOr("")
	vars := map[string]string{
		"DATASTORE_EMULATOR_HOST":      host,
		"DATASTORE_EMULATOR_HOST_PATH": host + "/datastore",
		"DATASTORE_HOST":               "http://" + host,
	}
	if project != "" {
		vars["DATASTORE_PROJECT_ID"] = project
		vars["DATASTORE_DATASET"] = project
	}
	return s.printEnvInit(vars)
}

func (s *emulatorsCmds) runEmuDatastoreEnvUnset(cmd *cobra.Command, args []string) error {
	return s.printEnvUnset([]string{
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
func (s *emulatorsCmds) resolveProjectOr(dflt string) string {
	p, err := s.inv.project()
	if err != nil {
		return dflt
	}
	return p
}
