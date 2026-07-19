package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub discovered-api-observations (#1160) ---

var apihubDiscoveredApiObservationsCmd = &cobra.Command{Use: "discovered-api-observations", Short: "Manage API Hub discovered API observations"}

var (
	flagApihubDaoLocation string
	flagApihubDaoFormat   string
	flagApihubDaoPageSize int64
)

var (
	apihubDaoDescribeCmd = &cobra.Command{
		Use: "describe OBSERVATION", Short: "Describe a discovered API observation",
		Args: cobra.ExactArgs(1), RunE: runApihubDaoDescribe,
	}
	apihubDaoListCmd = &cobra.Command{
		Use: "list", Short: "List discovered API observations in a location",
		Args: cobra.NoArgs, RunE: runApihubDaoList,
	}
)

func init() {
	all := []*cobra.Command{apihubDaoDescribeCmd, apihubDaoListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagApihubDaoLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagApihubDaoFormat, "format", "", "Output format")
	}
	apihubDaoListCmd.Flags().Int64Var(&flagApihubDaoPageSize, "page-size", 0, "Maximum results per page")

	apihubDiscoveredApiObservationsCmd.AddCommand(all...)
	apihubCmd.AddCommand(apihubDiscoveredApiObservationsCmd)
}

func apihubDaoName(id string) (string, error) {
	return apihubResource(flagApihubDaoLocation, "discoveredApiObservations", id)
}

func runApihubDaoDescribe(cmd *cobra.Command, args []string) error {
	name, err := apihubDaoName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DiscoveredApiObservations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing discovered API observation: %w", err)
	}
	return emitFormatted(got, flagApihubDaoFormat)
}

func runApihubDaoList(cmd *cobra.Command, args []string) error {
	parent, err := apihubLocationParent(flagApihubDaoLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1DiscoveredApiObservation
	pageToken := ""
	for {
		call := svc.Projects.Locations.DiscoveredApiObservations.List(parent).Context(ctx)
		if flagApihubDaoPageSize > 0 {
			call = call.PageSize(flagApihubDaoPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing discovered API observations: %w", err)
		}
		all = append(all, resp.DiscoveredApiObservations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagApihubDaoFormat)
}
