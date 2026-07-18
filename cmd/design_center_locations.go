package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

// --- gcloud design-center locations (#1533) ---

var dcLocCmd = &cobra.Command{Use: "locations", Short: "Manage Design Center locations"}

var (
	flagDCLocFormat   string
	flagDCLocFilter   string
	flagDCLocPageSize int64
)

var (
	dcLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a Design Center location",
		Args: cobra.ExactArgs(1), RunE: runDCLocDescribe,
	}
	dcLocListCmd = &cobra.Command{
		Use: "list", Short: "List Design Center locations for the current project",
		Args: cobra.NoArgs, RunE: runDCLocList,
	}
)

func init() {
	for _, c := range []*cobra.Command{dcLocDescribeCmd, dcLocListCmd} {
		c.Flags().StringVar(&flagDCLocFormat, "format", "", "Output format")
	}
	dcLocListCmd.Flags().StringVar(&flagDCLocFilter, "filter", "", "Server-side filter expression")
	dcLocListCmd.Flags().Int64Var(&flagDCLocPageSize, "page-size", 0, "Maximum results per page")

	dcLocCmd.AddCommand(dcLocDescribeCmd, dcLocListCmd)
	designCenterCmd.AddCommand(dcLocCmd)
}

func runDCLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := designCenterRest.do(ctx, http.MethodGet, "/"+dcLocationName(project, args[0]), nil, nil, &got); err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormatted(got, flagDCLocFormat)
}

func runDCLocList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	base := url.Values{}
	if flagDCLocFilter != "" {
		base.Set("filter", flagDCLocFilter)
	}
	ctx := context.Background()
	items, err := designCenterRest.paginate(ctx, fmt.Sprintf("/projects/%s/locations", project), base, "locations", flagDCLocPageSize)
	if err != nil {
		return fmt.Errorf("listing locations: %w", err)
	}
	return emitFormatted(items, flagDCLocFormat)
}
