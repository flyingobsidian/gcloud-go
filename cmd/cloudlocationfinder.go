package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudlocationfinder "google.golang.org/api/cloudlocationfinder/v1"
)

// --- gcloud cloudlocationfinder cloud-locations (#968) ---

var cloudLocationFinderCmd = &cobra.Command{
	Use:   "cloudlocationfinder",
	Short: "Manage Cloud Location Finder",
}

var (
	flagCLFLocation            string
	flagCLFFilter              string
	flagCLFQuery               string
	flagCLFSourceCloudLocation string
	flagCLFPageSize            int64
	flagCLFFormat              string
)

var cloudLocationsCmd = &cobra.Command{
	Use:   "cloud-locations",
	Short: "Manage cloud locations",
}

var (
	cloudLocationsDescribeCmd = &cobra.Command{
		Use:   "describe CLOUD_LOCATION",
		Short: "Describe a cloud location",
		Args:  cobra.ExactArgs(1),
		RunE:  runCLFDescribe,
	}
	cloudLocationsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List cloud locations",
		Args:  cobra.NoArgs,
		RunE:  runCLFList,
	}
	cloudLocationsSearchCmd = &cobra.Command{
		Use:   "search",
		Short: "Search for cloud locations from a source cloud location",
		Args:  cobra.NoArgs,
		RunE:  runCLFSearch,
	}
)

func init() {
	all := []*cobra.Command{cloudLocationsDescribeCmd, cloudLocationsListCmd, cloudLocationsSearchCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagCLFLocation, "location", "global", "Location that owns the cloud locations collection (e.g. \"global\")")
		c.Flags().StringVar(&flagCLFFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{cloudLocationsListCmd, cloudLocationsSearchCmd} {
		c.Flags().Int64Var(&flagCLFPageSize, "page-size", 0, "Maximum number of results per page")
	}
	cloudLocationsListCmd.Flags().StringVar(&flagCLFFilter, "filter", "", "Filter expression (e.g. cloud_location_type=CLOUD_LOCATION_TYPE_REGION)")
	cloudLocationsSearchCmd.Flags().StringVar(&flagCLFQuery, "query", "", "Search query string")
	cloudLocationsSearchCmd.Flags().StringVar(&flagCLFSourceCloudLocation, "source-cloud-location", "", "Source cloud location to search from (required)")
	_ = cloudLocationsSearchCmd.MarkFlagRequired("source-cloud-location")

	cloudLocationsCmd.AddCommand(all...)
	cloudLocationFinderCmd.AddCommand(cloudLocationsCmd)
	rootCmd.AddCommand(cloudLocationFinderCmd)
}

func clfParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	loc := flagCLFLocation
	if loc == "" {
		loc = "global"
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, loc), nil
}

func runCLFDescribe(cmd *cobra.Command, args []string) error {
	parent, err := clfParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudLocationFinderService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s/cloudLocations/%s", parent, args[0])
	loc, err := svc.Projects.Locations.CloudLocations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cloud location: %w", err)
	}
	return emitFormatted(loc, flagCLFFormat)
}

func runCLFList(cmd *cobra.Command, args []string) error {
	parent, err := clfParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudLocationFinderService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*cloudlocationfinder.CloudLocation
	pageToken := ""
	for {
		call := svc.Projects.Locations.CloudLocations.List(parent).Context(ctx)
		if flagCLFFilter != "" {
			call = call.Filter(flagCLFFilter)
		}
		if flagCLFPageSize > 0 {
			call = call.PageSize(flagCLFPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing cloud locations: %w", err)
		}
		all = append(all, resp.CloudLocations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagCLFFormat)
}

func runCLFSearch(cmd *cobra.Command, args []string) error {
	parent, err := clfParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudLocationFinderService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*cloudlocationfinder.CloudLocation
	pageToken := ""
	for {
		call := svc.Projects.Locations.CloudLocations.Search(parent).
			SourceCloudLocation(flagCLFSourceCloudLocation).Context(ctx)
		if flagCLFQuery != "" {
			call = call.Query(flagCLFQuery)
		}
		if flagCLFPageSize > 0 {
			call = call.PageSize(flagCLFPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("searching cloud locations: %w", err)
		}
		all = append(all, resp.CloudLocations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagCLFFormat)
}
