package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware locations (#1117) ---

func (s *vmwareCmds) initVmwareLocations() {
	all := []*cobra.Command{s.vmwareLocDescribeCmd, s.vmwareLocListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagVmwareLocFormat, "format", "", "Output format")
	}
	s.vmwareLocListCmd.Flags().StringVar(&s.flagVmwareLocFilter, "filter", "", "Server-side filter expression")
	s.vmwareLocListCmd.Flags().Int64Var(&s.flagVmwareLocPageSize, "page-size", 0, "Maximum results per page")

	s.vmwareLocationsCmd.AddCommand(all...)
	s.vmwareCmd.AddCommand(s.vmwareLocationsCmd)
}

func (s *vmwareCmds) runVmwareLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	name := fmt.Sprintf("projects/%s/locations/%s", project, args[0])
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagVmwareLocFormat)
}

func (s *vmwareCmds) runVmwareLocList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("projects/%s", project)
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	var all []*vmwareengine.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List(parent).Context(ctx)
		if s.flagVmwareLocFilter != "" {
			call = call.Filter(s.flagVmwareLocFilter)
		}
		if s.flagVmwareLocPageSize > 0 {
			call = call.PageSize(s.flagVmwareLocPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagVmwareLocFormat)
}
