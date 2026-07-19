package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable hot-tablets (#1305) ---

var bigtableHotTabletsCmd = &cobra.Command{Use: "hot-tablets", Short: "Manage Bigtable hot tablets"}

var (
	flagBtHtInstance  string
	flagBtHtCluster   string
	flagBtHtFormat    string
	flagBtHtPageSize  int64
	flagBtHtStartTime string
	flagBtHtEndTime   string
)

var bigtableHtListCmd = &cobra.Command{
	Use: "list", Short: "List Bigtable hot tablets on a cluster",
	Args: cobra.NoArgs, RunE: runBtHtList,
}

func init() {
	bigtableHtListCmd.Flags().StringVar(&flagBtHtInstance, "instance", "", "Bigtable instance ID (required)")
	_ = bigtableHtListCmd.MarkFlagRequired("instance")
	bigtableHtListCmd.Flags().StringVar(&flagBtHtCluster, "cluster", "", "Bigtable cluster ID (required)")
	_ = bigtableHtListCmd.MarkFlagRequired("cluster")
	bigtableHtListCmd.Flags().StringVar(&flagBtHtFormat, "format", "", "Output format")
	bigtableHtListCmd.Flags().Int64Var(&flagBtHtPageSize, "page-size", 0, "Maximum results per page")
	bigtableHtListCmd.Flags().StringVar(&flagBtHtStartTime, "start-time", "", "RFC3339 timestamp for the start of the window")
	bigtableHtListCmd.Flags().StringVar(&flagBtHtEndTime, "end-time", "", "RFC3339 timestamp for the end of the window")

	bigtableHotTabletsCmd.AddCommand(bigtableHtListCmd)
	bigtableCmd.AddCommand(bigtableHotTabletsCmd)
}

func runBtHtList(cmd *cobra.Command, args []string) error {
	parent, err := btClusterName(flagBtHtInstance, flagBtHtCluster)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BigtableAdminService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*bigtableadmin.HotTablet
	pageToken := ""
	for {
		call := svc.Projects.Instances.Clusters.HotTablets.List(parent).Context(ctx)
		if flagBtHtPageSize > 0 {
			call = call.PageSize(flagBtHtPageSize)
		}
		if flagBtHtStartTime != "" {
			call = call.StartTime(flagBtHtStartTime)
		}
		if flagBtHtEndTime != "" {
			call = call.EndTime(flagBtHtEndTime)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing hot tablets: %w", err)
		}
		all = append(all, resp.HotTablets...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagBtHtFormat)
}
