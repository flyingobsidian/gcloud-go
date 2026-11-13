package cli

import (
	"github.com/spf13/cobra"
)

// --- gcloud emulators pubsub (#1711) ---

func (s *emulatorsCmds) initEmulatorsPubsub() {
	for _, c := range []*cobra.Command{s.emuPubsubStartCmd, s.emuPubsubEnvInitCmd} {
		c.Flags().StringVar(&s.flagEmuPubsubHostPort, "host-port", "localhost:8085",
			"The host:port to which the emulator is (or will be) bound")
	}
	s.emuPubsubStartCmd.Flags().StringVar(&s.flagEmuPubsubDataDir, "data-dir", "",
		"Directory used to store/retrieve data (defaults to gcloud's user config dir)")

	s.emuPubsubCmd.AddCommand(s.emuPubsubStartCmd, s.emuPubsubEnvInitCmd)
	s.registerEmulatorGroup(s.emuPubsubCmd)
}

func (s *emulatorsCmds) runEmuPubsubStart(cmd *cobra.Command, args []string) error {
	argv := []string{"--host=" + hostFrom(s.flagEmuPubsubHostPort), "--port=" + portFrom(s.flagEmuPubsubHostPort)}
	if s.flagEmuPubsubDataDir != "" {
		argv = append(argv, "--data-dir="+s.flagEmuPubsubDataDir)
	}
	return s.runEmulator("pubsub-emulator", "pubsub-emulator/bin", argv)
}

func (s *emulatorsCmds) runEmuPubsubEnvInit(cmd *cobra.Command, args []string) error {
	vars := map[string]string{
		"PUBSUB_EMULATOR_HOST": s.flagEmuPubsubHostPort,
	}
	if project := s.resolveProjectOr(""); project != "" {
		vars["PUBSUB_PROJECT_ID"] = project
	}
	return s.printEnvInit(vars)
}
