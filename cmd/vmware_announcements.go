package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware announcements (#1114) ---

var vmwareAnnouncementsCmd = &cobra.Command{Use: "announcements", Short: "Manage VMware Engine announcements"}

var (
	flagVmwareAnnLocation string
	flagVmwareAnnFormat   string
	flagVmwareAnnPageSize int64
)

var (
	vmwareAnnDescribeCmd = &cobra.Command{
		Use: "describe ANNOUNCEMENT", Short: "Describe a VMware Engine announcement",
		Args: cobra.ExactArgs(1), RunE: runVmwareAnnDescribe,
	}
	vmwareAnnListCmd = &cobra.Command{
		Use: "list", Short: "List VMware Engine announcements in a location",
		Args: cobra.NoArgs, RunE: runVmwareAnnList,
	}
)

func init() {
	all := []*cobra.Command{vmwareAnnDescribeCmd, vmwareAnnListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagVmwareAnnLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagVmwareAnnFormat, "format", "", "Output format")
	}
	vmwareAnnListCmd.Flags().Int64Var(&flagVmwareAnnPageSize, "page-size", 0, "Maximum results per page")

	vmwareAnnouncementsCmd.AddCommand(all...)
	vmwareCmd.AddCommand(vmwareAnnouncementsCmd)
}

func vmwareAnnName(id string) (string, error) {
	return vmwareResource(flagVmwareAnnLocation, "announcements", id)
}

func runVmwareAnnDescribe(cmd *cobra.Command, args []string) error {
	name, err := vmwareAnnName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Announcements.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing announcement: %w", err)
	}
	return emitFormatted(got, flagVmwareAnnFormat)
}

func runVmwareAnnList(cmd *cobra.Command, args []string) error {
	parent, err := vmwareLocationParent(flagVmwareAnnLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*vmwareengine.Announcement
	pageToken := ""
	for {
		call := svc.Projects.Locations.Announcements.List(parent).Context(ctx)
		if flagVmwareAnnPageSize > 0 {
			call = call.PageSize(flagVmwareAnnPageSize)
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
	return emitFormatted(all, flagVmwareAnnFormat)
}
