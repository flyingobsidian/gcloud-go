package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms locations (#1108) ---

var kmsLocationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "View Cloud KMS locations",
}

var (
	flagKmsLocFormat   string
	flagKmsLocFilter   string
	flagKmsLocPageSize int64
)

var kmsLocListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available Cloud KMS locations",
	Args:  cobra.NoArgs,
	RunE:  runKmsLocList,
}

func init() {
	kmsLocListCmd.Flags().StringVar(&flagKmsLocFormat, "format", "", "Output format")
	kmsLocListCmd.Flags().StringVar(&flagKmsLocFilter, "filter", "", "Filter expression")
	kmsLocListCmd.Flags().Int64Var(&flagKmsLocPageSize, "page-size", 0, "Page size")
	kmsLocationsCmd.AddCommand(kmsLocListCmd)
	kmsCmd.AddCommand(kmsLocationsCmd)
}

func runKmsLocList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.KMSService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := "projects/" + project
	var all []*cloudkms.Location
	token := ""
	for {
		call := svc.Projects.Locations.List(name).Context(ctx)
		if flagKmsLocFilter != "" {
			call = call.Filter(flagKmsLocFilter)
		}
		if flagKmsLocPageSize > 0 {
			call = call.PageSize(flagKmsLocPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing locations: %w", err)
		}
		all = append(all, resp.Locations...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormatted(all, flagKmsLocFormat)
}
