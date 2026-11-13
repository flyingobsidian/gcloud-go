package cmd

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	netapp "google.golang.org/api/netapp/v1"
)

// --- gcloud netapp locations (#1202) ---

func (s *netappCmds) initNetappLocations() {
	s.netappLocListCmd.Flags().StringVar(&s.flagNetAppLocFormat, "format", "", "Output format")
	s.netappLocListCmd.Flags().StringVar(&s.flagNetAppLocFilter, "filter", "", "Server-side filter expression")
	s.netappLocListCmd.Flags().Int64Var(&s.flagNetAppLocPageSize, "page-size", 0, "Maximum number of results per page")
	s.netappLocDescribeCmd.Flags().StringVar(&s.flagNetAppLocFormat, "format", "", "Output format")

	s.netappLocCmd.AddCommand(s.netappLocListCmd, s.netappLocDescribeCmd)
	s.netappCmd.AddCommand(s.netappLocCmd)
}

func (s *netappCmds) runNetAppLocList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	var all []*netapp.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if s.flagNetAppLocFilter != "" {
			call = call.Filter(s.flagNetAppLocFilter)
		}
		if s.flagNetAppLocPageSize > 0 {
			call = call.PageSize(s.flagNetAppLocPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing locations: %w", err)
		}
		all = append(all, resp.Locations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagNetAppLocFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagNetAppLocFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-20s %s\n", "LOCATION", "DISPLAY_NAME")
	for _, l := range all {
		fmt.Fprintf(s.inv.stdout, "%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

func (s *netappCmds) runNetAppLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	loc, err := svc.Projects.Locations.Get(netappLocationParent(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	if s.flagNetAppLocFormat != "" {
		return emitFormattedTo(s.inv.stdout, loc, s.flagNetAppLocFormat)
	}
	fmt.Fprintf(s.inv.stdout, "%s (%s)\n", path.Base(loc.Name), loc.DisplayName)
	return nil
}
