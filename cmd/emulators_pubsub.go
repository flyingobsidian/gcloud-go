package cmd

import (
	"github.com/spf13/cobra"
)

// --- gcloud emulators pubsub (#1711) ---

var (
	flagEmuPubsubHostPort string
	flagEmuPubsubDataDir  string
)

var emuPubsubCmd = &cobra.Command{Use: "pubsub", Short: "Manage local Cloud Pub/Sub emulator"}

var emuPubsubStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a local Pub/Sub emulator (wraps pubsub-emulator)",
	Args:  cobra.NoArgs,
	RunE:  runEmuPubsubStart,
}

var emuPubsubEnvInitCmd = &cobra.Command{
	Use:   "env-init",
	Short: "Print shell exports for PUBSUB_EMULATOR_HOST and PUBSUB_PROJECT_ID",
	Args:  cobra.NoArgs,
	RunE:  runEmuPubsubEnvInit,
}

func init() {
	for _, c := range []*cobra.Command{emuPubsubStartCmd, emuPubsubEnvInitCmd} {
		c.Flags().StringVar(&flagEmuPubsubHostPort, "host-port", "localhost:8085",
			"The host:port to which the emulator is (or will be) bound")
	}
	emuPubsubStartCmd.Flags().StringVar(&flagEmuPubsubDataDir, "data-dir", "",
		"Directory used to store/retrieve data (defaults to gcloud's user config dir)")

	emuPubsubCmd.AddCommand(emuPubsubStartCmd, emuPubsubEnvInitCmd)
	registerEmulatorGroup(emuPubsubCmd)
}

func runEmuPubsubStart(cmd *cobra.Command, args []string) error {
	argv := []string{"--host=" + hostFrom(flagEmuPubsubHostPort), "--port=" + portFrom(flagEmuPubsubHostPort)}
	if flagEmuPubsubDataDir != "" {
		argv = append(argv, "--data-dir="+flagEmuPubsubDataDir)
	}
	return runEmulator("pubsub-emulator", "pubsub-emulator/bin", argv)
}

func runEmuPubsubEnvInit(cmd *cobra.Command, args []string) error {
	vars := map[string]string{
		"PUBSUB_EMULATOR_HOST": flagEmuPubsubHostPort,
	}
	if project := resolveProjectOr(""); project != "" {
		vars["PUBSUB_PROJECT_ID"] = project
	}
	return printEnvInit(vars)
}
