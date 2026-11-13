package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable hot-tablets (#1305) ---

func (s *bigtableCmds) initBigtableHotTablets() {
	s.bigtableHtListCmd.Flags().StringVar(&s.flagBtHtInstance, "instance", "", "Bigtable instance ID (required)")
	_ = s.bigtableHtListCmd.MarkFlagRequired("instance")
	s.bigtableHtListCmd.Flags().StringVar(&s.flagBtHtCluster, "cluster", "", "Bigtable cluster ID (required)")
	_ = s.bigtableHtListCmd.MarkFlagRequired("cluster")
	s.bigtableHtListCmd.Flags().StringVar(&s.flagBtHtFormat, "format", "", "Output format")
	s.bigtableHtListCmd.Flags().Int64Var(&s.flagBtHtPageSize, "page-size", 0, "Maximum results per page")
	s.bigtableHtListCmd.Flags().StringVar(&s.flagBtHtStartTime, "start-time", "", "RFC3339 timestamp for the start of the window")
	s.bigtableHtListCmd.Flags().StringVar(&s.flagBtHtEndTime, "end-time", "", "RFC3339 timestamp for the end of the window")

	s.bigtableHotTabletsCmd.AddCommand(s.bigtableHtListCmd)
	s.bigtableCmd.AddCommand(s.bigtableHotTabletsCmd)
}

func (s *bigtableCmds) runBtHtList(cmd *cobra.Command, args []string) error {
	parent, err := s.btClusterName(s.flagBtHtInstance, s.flagBtHtCluster)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	var all []*bigtableadmin.HotTablet
	pageToken := ""
	for {
		call := svc.Projects.Instances.Clusters.HotTablets.List(parent).Context(ctx)
		if s.flagBtHtPageSize > 0 {
			call = call.PageSize(s.flagBtHtPageSize)
		}
		if s.flagBtHtStartTime != "" {
			call = call.StartTime(s.flagBtHtStartTime)
		}
		if s.flagBtHtEndTime != "" {
			call = call.EndTime(s.flagBtHtEndTime)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagBtHtFormat)
}
