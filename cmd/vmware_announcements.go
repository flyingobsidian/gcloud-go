package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware announcements (#1114) ---

func (s *vmwareCmds) initVmwareAnnouncements() {
	all := []*cobra.Command{s.vmwareAnnDescribeCmd, s.vmwareAnnListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagVmwareAnnLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagVmwareAnnFormat, "format", "", "Output format")
	}
	s.vmwareAnnListCmd.Flags().Int64Var(&s.flagVmwareAnnPageSize, "page-size", 0, "Maximum results per page")

	s.vmwareAnnouncementsCmd.AddCommand(all...)
	s.vmwareCmd.AddCommand(s.vmwareAnnouncementsCmd)
}

func (s *vmwareCmds) vmwareAnnName(id string) (string, error) {
	return s.vmwareResource(s.flagVmwareAnnLocation, "announcements", id)
}

func (s *vmwareCmds) runVmwareAnnDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.vmwareAnnName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Announcements.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing announcement: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagVmwareAnnFormat)
}

func (s *vmwareCmds) runVmwareAnnList(cmd *cobra.Command, args []string) error {
	parent, err := s.vmwareLocationParent(s.flagVmwareAnnLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	var all []*vmwareengine.Announcement
	pageToken := ""
	for {
		call := svc.Projects.Locations.Announcements.List(parent).Context(ctx)
		if s.flagVmwareAnnPageSize > 0 {
			call = call.PageSize(s.flagVmwareAnnPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing announcements: %w", err)
		}
		all = append(all, resp.Announcements...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagVmwareAnnFormat)
}
