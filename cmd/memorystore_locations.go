package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud memorystore locations (#979) ---

var memstoreLocCmd = &cobra.Command{Use: "locations", Short: "Manage Memorystore locations"}

var (
	flagMemstoreLocFormat   string
	flagMemstoreLocPageSize int64
)

var (
	memstoreLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a Memorystore location",
		Args: cobra.ExactArgs(1), RunE: runMemstoreLocDescribe,
	}
	memstoreLocListCmd = &cobra.Command{
		Use: "list", Short: "List Memorystore locations",
		Args: cobra.NoArgs, RunE: runMemstoreLocList,
	}
)

func init() {
	all := []*cobra.Command{memstoreLocDescribeCmd, memstoreLocListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagMemstoreLocFormat, "format", "", "Output format")
	}
	memstoreLocListCmd.Flags().Int64Var(&flagMemstoreLocPageSize, "page-size", 0, "Maximum results per page")

	memstoreLocCmd.AddCommand(all...)
	memorystoreCmd.AddCommand(memstoreLocCmd)
}

func memstoreLocationName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, id), nil
}

func runMemstoreLocDescribe(cmd *cobra.Command, args []string) error {
	name, err := memstoreLocationName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := memorystoreRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormatted(got, flagMemstoreLocFormat)
}

func runMemstoreLocList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	items, err := memorystoreRest.paginate(ctx, fmt.Sprintf("/projects/%s/locations", project), nil, "locations", flagMemstoreLocPageSize)
	if err != nil {
		return fmt.Errorf("listing locations: %w", err)
	}
	return emitFormatted(items, flagMemstoreLocFormat)
}
