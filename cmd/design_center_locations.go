package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

// --- gcloud design-center locations (#1533) ---

func (s *designCenterCmds) initDesignCenterLocations() {
	for _, c := range []*cobra.Command{s.dcLocDescribeCmd, s.dcLocListCmd} {
		c.Flags().StringVar(&s.flagDCLocFormat, "format", "", "Output format")
	}
	s.dcLocListCmd.Flags().StringVar(&s.flagDCLocFilter, "filter", "", "Server-side filter expression")
	s.dcLocListCmd.Flags().Int64Var(&s.flagDCLocPageSize, "page-size", 0, "Maximum results per page")

	s.dcLocCmd.AddCommand(s.dcLocDescribeCmd, s.dcLocListCmd)
	s.designCenterCmd.AddCommand(s.dcLocCmd)
}

func (s *designCenterCmds) runDCLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(designCenterRest).do(ctx, http.MethodGet, "/"+dcLocationName(project, args[0]), nil, nil, &got); err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDCLocFormat)
}

func (s *designCenterCmds) runDCLocList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	base := url.Values{}
	if s.flagDCLocFilter != "" {
		base.Set("filter", s.flagDCLocFilter)
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(designCenterRest).paginate(ctx, fmt.Sprintf("/projects/%s/locations", project), base, "locations", s.flagDCLocPageSize)
	if err != nil {
		return fmt.Errorf("listing locations: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagDCLocFormat)
}
