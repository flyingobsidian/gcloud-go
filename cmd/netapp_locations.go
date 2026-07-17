package cmd

import (
	"context"
	"fmt"
	"path"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	netapp "google.golang.org/api/netapp/v1"
)

// --- gcloud netapp locations (#1202) ---

var netappLocCmd = &cobra.Command{Use: "locations", Short: "Get and list Cloud NetApp Files locations"}

var (
	flagNetAppLocFormat   string
	flagNetAppLocFilter   string
	flagNetAppLocPageSize int64
)

var (
	netappLocListCmd = &cobra.Command{
		Use: "list", Short: "List NetApp locations",
		Args: cobra.NoArgs, RunE: runNetAppLocList,
	}
	netappLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a NetApp location",
		Args: cobra.ExactArgs(1), RunE: runNetAppLocDescribe,
	}
)

func init() {
	netappLocListCmd.Flags().StringVar(&flagNetAppLocFormat, "format", "", "Output format")
	netappLocListCmd.Flags().StringVar(&flagNetAppLocFilter, "filter", "", "Server-side filter expression")
	netappLocListCmd.Flags().Int64Var(&flagNetAppLocPageSize, "page-size", 0, "Maximum number of results per page")
	netappLocDescribeCmd.Flags().StringVar(&flagNetAppLocFormat, "format", "", "Output format")

	netappLocCmd.AddCommand(netappLocListCmd, netappLocDescribeCmd)
	netappCmd.AddCommand(netappLocCmd)
}

func runNetAppLocList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*netapp.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if flagNetAppLocFilter != "" {
			call = call.Filter(flagNetAppLocFilter)
		}
		if flagNetAppLocPageSize > 0 {
			call = call.PageSize(flagNetAppLocPageSize)
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
	if flagNetAppLocFormat != "" {
		return emitFormatted(all, flagNetAppLocFormat)
	}
	fmt.Printf("%-20s %s\n", "LOCATION", "DISPLAY_NAME")
	for _, l := range all {
		fmt.Printf("%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

func runNetAppLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	loc, err := svc.Projects.Locations.Get(netappLocationParent(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	if flagNetAppLocFormat != "" {
		return emitFormatted(loc, flagNetAppLocFormat)
	}
	fmt.Printf("%s (%s)\n", path.Base(loc.Name), loc.DisplayName)
	return nil
}
