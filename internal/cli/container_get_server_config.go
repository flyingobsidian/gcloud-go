package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// --- gcloud container get-server-config (#1145) ---

func (s *containerCmds) initContainerGetServerConfig() {
	s.containerGetServerConfigCmd.Flags().StringVar(&s.flagCtnCfgLocation, "location", "", "Location (region or zone) (required)")
	_ = s.containerGetServerConfigCmd.MarkFlagRequired("location")
	s.containerGetServerConfigCmd.Flags().StringVar(&s.flagCtnCfgFormat, "format", "", "Output format")

	s.containerCmd.AddCommand(s.containerGetServerConfigCmd)
}

func (s *containerCmds) runCtnGetServerConfig(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	name := fmt.Sprintf("projects/%s/locations/%s", project, s.flagCtnCfgLocation)
	ctx := cmd.Context()
	svc, err := s.inv.clients().Container(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.GetServerConfig(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting server config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagCtnCfgFormat)
}
