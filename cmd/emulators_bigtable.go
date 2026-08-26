package cmd

import (
	"github.com/spf13/cobra"
)

// --- gcloud emulators bigtable (#1709) ---

var (
	flagEmuBigtableHostPort string
)

var emuBigtableCmd = &cobra.Command{Use: "bigtable", Short: "Manage local Bigtable emulator"}

var emuBigtableStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a local Bigtable emulator (wraps cbtemulator)",
	Args:  cobra.NoArgs,
	RunE:  runEmuBigtableStart,
}

var emuBigtableEnvInitCmd = &cobra.Command{
	Use:   "env-init",
	Short: "Print shell exports for BIGTABLE_EMULATOR_HOST",
	Args:  cobra.NoArgs,
	RunE:  runEmuBigtableEnvInit,
}

func init() {
	emuBigtableStartCmd.Flags().StringVar(&flagEmuBigtableHostPort, "host-port", "localhost:8086",
		"The host:port to which the emulator should be bound")
	emuBigtableEnvInitCmd.Flags().StringVar(&flagEmuBigtableHostPort, "host-port", "localhost:8086",
		"The host:port the running emulator is bound to; used to shape the exported BIGTABLE_EMULATOR_HOST")

	emuBigtableCmd.AddCommand(emuBigtableStartCmd, emuBigtableEnvInitCmd)
	registerEmulatorGroup(emuBigtableCmd)
}

func runEmuBigtableStart(cmd *cobra.Command, args []string) error {
	return runEmulator("cbtemulator", "bigtable-emulator", []string{"--host=" + hostFrom(flagEmuBigtableHostPort), "--port=" + portFrom(flagEmuBigtableHostPort)})
}

func runEmuBigtableEnvInit(cmd *cobra.Command, args []string) error {
	return printEnvInit(map[string]string{
		"BIGTABLE_EMULATOR_HOST": flagEmuBigtableHostPort,
	})
}
