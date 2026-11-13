package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	publicca "google.golang.org/api/publicca/v1"
)

// --- gcloud publicca (#376, #794) ---

func (s *publiccaCmds) initPublicca() {
	s.publiccaEakCreateCmd.Flags().StringVar(&s.flagPublicCALocation, "location", "global", "Location (defaults to global)")
	s.publiccaEakCreateCmd.Flags().StringVar(&s.flagPublicCAFormat, "format", "", "Output format")
	s.publiccaEakCmd.AddCommand(s.publiccaEakCreateCmd)
	s.publiccaCmd.AddCommand(s.publiccaEakCmd)

}

func (s *publiccaCmds) runPublicCAEakCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	location := s.flagPublicCALocation
	if location == "" {
		location = "global"
	}
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	ctx := cmd.Context()
	svc, err := s.inv.clients().PublicCA(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ExternalAccountKeys.Create(parent, &publicca.ExternalAccountKey{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating external account key: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagPublicCAFormat)
}

// publiccaCmds holds the commands and flag values of one invocation's "publicca"
// command group, so invocations don't share state (#2070).
type publiccaCmds struct {
	inv *invocation

	publiccaCmd          *cobra.Command
	flagPublicCALocation string
	flagPublicCAFormat   string
	publiccaEakCmd       *cobra.Command
	publiccaEakCreateCmd *cobra.Command
}

// newPubliccaCmd returns the "publicca" command group for inv.
func newPubliccaCmd(inv *invocation) *cobra.Command {
	return newPubliccaCmds(inv).publiccaCmd
}

// newPubliccaCmds builds the "publicca" command group for inv and returns its state.
func newPubliccaCmds(inv *invocation) *publiccaCmds {
	s := &publiccaCmds{inv: inv}
	s.publiccaCmd = &cobra.Command{Use: "publicca", Short: "Manage Google Trust Services PublicCA"}
	s.publiccaEakCmd = &cobra.Command{Use: "external-account-keys", Short: "Manage ACME external account keys"}
	s.publiccaEakCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a new ACME external account key",
		Args: cobra.NoArgs, RunE: s.runPublicCAEakCreate,
	}
	s.initPublicca()
	return s
}
