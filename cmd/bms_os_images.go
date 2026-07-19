package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	baremetalsolution "google.golang.org/api/baremetalsolution/v2"
)

// --- gcloud bms os-images (#1230) ---

var bmsOsImagesCmd = &cobra.Command{Use: "os-images", Short: "Manage bare metal OS images"}

var (
	flagBmsOsLocation string
	flagBmsOsFormat   string
	flagBmsOsPageSize int64
)

var (
	bmsOsListCmd = &cobra.Command{
		Use: "list", Short: "List available bare metal OS images",
		Args: cobra.NoArgs, RunE: runBmsOsList,
	}
)

func init() {
	all := []*cobra.Command{bmsOsListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagBmsOsLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagBmsOsFormat, "format", "", "Output format")
	}
	bmsOsListCmd.Flags().Int64Var(&flagBmsOsPageSize, "page-size", 0, "Maximum results per page")

	bmsOsImagesCmd.AddCommand(all...)
	bmsCmd.AddCommand(bmsOsImagesCmd)
}

func runBmsOsList(cmd *cobra.Command, args []string) error {
	parent, err := bmsLocationParent(flagBmsOsLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*baremetalsolution.OSImage
	pageToken := ""
	for {
		call := svc.Projects.Locations.OsImages.List(parent).Context(ctx)
		if flagBmsOsPageSize > 0 {
			call = call.PageSize(flagBmsOsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing os images: %w", err)
		}
		all = append(all, resp.OsImages...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagBmsOsFormat)
}
