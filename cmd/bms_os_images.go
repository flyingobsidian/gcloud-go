package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	baremetalsolution "google.golang.org/api/baremetalsolution/v2"
)

// --- gcloud bms os-images (#1230) ---

func (s *bmsCmds) initBmsOsImages() {
	all := []*cobra.Command{s.bmsOsListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagBmsOsLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagBmsOsFormat, "format", "", "Output format")
	}
	s.bmsOsListCmd.Flags().Int64Var(&s.flagBmsOsPageSize, "page-size", 0, "Maximum results per page")

	s.bmsOsImagesCmd.AddCommand(all...)
	s.bmsCmd.AddCommand(s.bmsOsImagesCmd)
}

func (s *bmsCmds) runBmsOsList(cmd *cobra.Command, args []string) error {
	parent, err := s.bmsLocationParent(s.flagBmsOsLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	var all []*baremetalsolution.OSImage
	pageToken := ""
	for {
		call := svc.Projects.Locations.OsImages.List(parent).Context(ctx)
		if s.flagBmsOsPageSize > 0 {
			call = call.PageSize(s.flagBmsOsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing os images: %w", err)
		}
		all = append(all, resp.OsImages...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagBmsOsFormat)
}
