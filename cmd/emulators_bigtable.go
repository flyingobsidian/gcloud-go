package cmd

import (
	"github.com/spf13/cobra"
)

// --- gcloud emulators bigtable (#1709) ---

func (s *emulatorsCmds) initEmulatorsBigtable() {
	s.emuBigtableStartCmd.Flags().StringVar(&s.flagEmuBigtableHostPort, "host-port", "localhost:8086",
		"The host:port to which the emulator should be bound")
	s.emuBigtableEnvInitCmd.Flags().StringVar(&s.flagEmuBigtableHostPort, "host-port", "localhost:8086",
		"The host:port the running emulator is bound to; used to shape the exported BIGTABLE_EMULATOR_HOST")

	s.emuBigtableCmd.AddCommand(s.emuBigtableStartCmd, s.emuBigtableEnvInitCmd)
	s.registerEmulatorGroup(s.emuBigtableCmd)
}

func (s *emulatorsCmds) runEmuBigtableStart(cmd *cobra.Command, args []string) error {
	return s.runEmulator("cbtemulator", "bigtable-emulator", []string{"--host=" + hostFrom(s.flagEmuBigtableHostPort), "--port=" + portFrom(s.flagEmuBigtableHostPort)})
}

func (s *emulatorsCmds) runEmuBigtableEnvInit(cmd *cobra.Command, args []string) error {
	return s.printEnvInit(map[string]string{
		"BIGTABLE_EMULATOR_HOST": s.flagEmuBigtableHostPort,
	})
}
