package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware locations (#1117) ---

var vmwareLocationsCmd = &cobra.Command{Use: "locations", Short: "List VMware Engine locations"}

var (
	flagVmwareLocFormat   string
	flagVmwareLocFilter   string
	flagVmwareLocPageSize int64
)

var (
	vmwareLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a VMware Engine location",
		Args: cobra.ExactArgs(1), RunE: runVmwareLocDescribe,
	}
	vmwareLocListCmd = &cobra.Command{
		Use: "list", Short: "List VMware Engine locations",
		Args: cobra.NoArgs, RunE: runVmwareLocList,
	}
)

func init() {
	all := []*cobra.Command{vmwareLocDescribeCmd, vmwareLocListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagVmwareLocFormat, "format", "", "Output format")
	}
	vmwareLocListCmd.Flags().StringVar(&flagVmwareLocFilter, "filter", "", "Server-side filter expression")
	vmwareLocListCmd.Flags().Int64Var(&flagVmwareLocPageSize, "page-size", 0, "Maximum results per page")

	vmwareLocationsCmd.AddCommand(all...)
	vmwareCmd.AddCommand(vmwareLocationsCmd)
}

func runVmwareLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	name := fmt.Sprintf("projects/%s/locations/%s", project, args[0])
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormatted(got, flagVmwareLocFormat)
}

func runVmwareLocList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("projects/%s", project)
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*vmwareengine.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List(parent).Context(ctx)
		if flagVmwareLocFilter != "" {
			call = call.Filter(flagVmwareLocFilter)
		}
		if flagVmwareLocPageSize > 0 {
			call = call.PageSize(flagVmwareLocPageSize)
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
	return emitFormatted(all, flagVmwareLocFormat)
}
