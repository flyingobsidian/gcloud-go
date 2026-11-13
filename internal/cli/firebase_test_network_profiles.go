package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// --- gcloud firebase test network-profiles (#1235) ---

func (s *firebaseCmds) initFirebaseTestNetworkProfiles() {
	for _, c := range []*cobra.Command{s.firebaseTestNetworkProfilesListCmd, s.firebaseTestNetworkProfilesDescribeCmd} {
		c.Flags().StringVar(&s.flagFtnpFormat, "format", "", "Output format")
	}

	s.firebaseTestNetworkProfilesCmd.AddCommand(s.firebaseTestNetworkProfilesListCmd, s.firebaseTestNetworkProfilesDescribeCmd)
	s.firebaseTestCmd.AddCommand(s.firebaseTestNetworkProfilesCmd)
}

func (s *firebaseCmds) runFtnpList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Testing(ctx)
	if err != nil {
		return err
	}
	cat, err := svc.TestEnvironmentCatalog.Get("NETWORK_CONFIGURATION").ProjectId(project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching network configuration catalog: %w", err)
	}
	if cat.NetworkConfigurationCatalog == nil {
		return emitFormattedTo(s.inv.stdout, []any{}, s.flagFtnpFormat)
	}
	return emitFormattedTo(s.inv.stdout, cat.NetworkConfigurationCatalog.Configurations, s.flagFtnpFormat)
}

func (s *firebaseCmds) runFtnpDescribe(cmd *cobra.Command, args []string) error {
	s.flagFtnpID = args[0]
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Testing(ctx)
	if err != nil {
		return err
	}
	cat, err := svc.TestEnvironmentCatalog.Get("NETWORK_CONFIGURATION").ProjectId(project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("fetching network configuration catalog: %w", err)
	}
	if cat.NetworkConfigurationCatalog != nil {
		for _, nc := range cat.NetworkConfigurationCatalog.Configurations {
			if nc.Id == s.flagFtnpID {
				return emitFormattedTo(s.inv.stdout, nc, s.flagFtnpFormat)
			}
		}
	}
	return fmt.Errorf("network profile [%s] not found", s.flagFtnpID)
}
