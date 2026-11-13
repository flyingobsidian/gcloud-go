package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms locations (#1108) ---

func (s *kmsCmds) initKmsLocations() {
	s.kmsLocListCmd.Flags().StringVar(&s.flagKmsLocFormat, "format", "", "Output format")
	s.kmsLocListCmd.Flags().StringVar(&s.flagKmsLocFilter, "filter", "", "Filter expression")
	s.kmsLocListCmd.Flags().Int64Var(&s.flagKmsLocPageSize, "page-size", 0, "Page size")
	s.kmsLocationsCmd.AddCommand(s.kmsLocListCmd)
	s.kmsCmd.AddCommand(s.kmsLocationsCmd)
}

func (s *kmsCmds) runKmsLocList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := "projects/" + project
	var all []*cloudkms.Location
	token := ""
	for {
		call := svc.Projects.Locations.List(name).Context(ctx)
		if s.flagKmsLocFilter != "" {
			call = call.Filter(s.flagKmsLocFilter)
		}
		if s.flagKmsLocPageSize > 0 {
			call = call.PageSize(s.flagKmsLocPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing locations: %w", err)
		}
		all = append(all, resp.Locations...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagKmsLocFormat)
}
