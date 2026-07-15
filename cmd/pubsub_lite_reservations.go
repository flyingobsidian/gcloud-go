package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	pubsublite "google.golang.org/api/pubsublite/v1"
)

// --- gcloud pubsub lite-reservations (#1172) ---

var pubsubLiteReservationsCmd = &cobra.Command{
	Use:   "lite-reservations",
	Short: "Manage Pub/Sub Lite reservations",
}

var (
	flagPSLResLocation   string
	flagPSLResThroughput int64
	flagPSLResFormat     string
	flagPSLResPageSize   int64
)

var (
	pubsubLiteResCreateCmd = &cobra.Command{
		Use: "create RESERVATION", Short: "Create a Pub/Sub Lite reservation",
		Args: cobra.ExactArgs(1), RunE: runPSLResCreate,
	}
	pubsubLiteResDeleteCmd = &cobra.Command{
		Use: "delete RESERVATION", Short: "Delete a Pub/Sub Lite reservation",
		Args: cobra.ExactArgs(1), RunE: runPSLResDelete,
	}
	pubsubLiteResDescribeCmd = &cobra.Command{
		Use: "describe RESERVATION", Short: "Describe a Pub/Sub Lite reservation",
		Args: cobra.ExactArgs(1), RunE: runPSLResDescribe,
	}
	pubsubLiteResListCmd = &cobra.Command{
		Use: "list", Short: "List Pub/Sub Lite reservations in a location",
		Args: cobra.NoArgs, RunE: runPSLResList,
	}
	pubsubLiteResListTopicsCmd = &cobra.Command{
		Use: "list-topics RESERVATION", Short: "List Pub/Sub Lite topics attached to a reservation",
		Args: cobra.ExactArgs(1), RunE: runPSLResListTopics,
	}
	pubsubLiteResUpdateCmd = &cobra.Command{
		Use: "update RESERVATION", Short: "Update a Pub/Sub Lite reservation",
		Args: cobra.ExactArgs(1), RunE: runPSLResUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		pubsubLiteResCreateCmd, pubsubLiteResDeleteCmd, pubsubLiteResDescribeCmd,
		pubsubLiteResListCmd, pubsubLiteResListTopicsCmd, pubsubLiteResUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagPSLResLocation, "location", "",
			"Regional location containing the reservation, e.g. us-central1 (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagPSLResFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{pubsubLiteResListCmd, pubsubLiteResListTopicsCmd} {
		c.Flags().Int64Var(&flagPSLResPageSize, "page-size", 0, "Maximum number of results per page")
	}
	for _, c := range []*cobra.Command{pubsubLiteResCreateCmd, pubsubLiteResUpdateCmd} {
		c.Flags().Int64Var(&flagPSLResThroughput, "throughput-capacity", 0,
			"Reserved throughput capacity in MiB/s (create: required; update: required)")
	}
	_ = pubsubLiteResCreateCmd.MarkFlagRequired("throughput-capacity")
	_ = pubsubLiteResUpdateCmd.MarkFlagRequired("throughput-capacity")

	pubsubLiteReservationsCmd.AddCommand(all...)
	pubsubCmd.AddCommand(pubsubLiteReservationsCmd)
}

func pslResName(id, project, location string) string {
	return pubsubLiteChild("reservations", id, pubsubLiteLocationParent(project, location))
}

func pslResService(ctx context.Context) (*pubsublite.Service, string, error) {
	project, err := resolveProject()
	if err != nil {
		return nil, "", err
	}
	region, err := pubsubLiteRegion(flagPSLResLocation)
	if err != nil {
		return nil, "", err
	}
	svc, err := gcp.PubSubLiteService(ctx, flagAccount, region)
	if err != nil {
		return nil, "", err
	}
	return svc, project, nil
}

func runPSLResCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, project, err := pslResService(ctx)
	if err != nil {
		return err
	}
	body := &pubsublite.Reservation{ThroughputCapacity: flagPSLResThroughput}
	got, err := svc.Admin.Projects.Locations.Reservations.
		Create(pubsubLiteLocationParent(project, flagPSLResLocation), body).
		ReservationId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating reservation: %w", err)
	}
	return emitFormatted(got, flagPSLResFormat)
}

func runPSLResDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, project, err := pslResService(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Admin.Projects.Locations.Reservations.
		Delete(pslResName(args[0], project, flagPSLResLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting reservation: %w", err)
	}
	fmt.Printf("Deleted reservation [%s].\n", args[0])
	return nil
}

func runPSLResDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, project, err := pslResService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Admin.Projects.Locations.Reservations.
		Get(pslResName(args[0], project, flagPSLResLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing reservation: %w", err)
	}
	return emitFormatted(got, flagPSLResFormat)
}

func runPSLResList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, project, err := pslResService(ctx)
	if err != nil {
		return err
	}
	parent := pubsubLiteLocationParent(project, flagPSLResLocation)
	var all []*pubsublite.Reservation
	pageToken := ""
	for {
		call := svc.Admin.Projects.Locations.Reservations.List(parent).Context(ctx)
		if flagPSLResPageSize > 0 {
			call = call.PageSize(flagPSLResPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing reservations: %w", err)
		}
		all = append(all, resp.Reservations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagPSLResFormat)
}

func runPSLResListTopics(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, project, err := pslResService(ctx)
	if err != nil {
		return err
	}
	name := pslResName(args[0], project, flagPSLResLocation)
	var all []string
	pageToken := ""
	for {
		call := svc.Admin.Projects.Locations.Reservations.Topics.List(name).Context(ctx)
		if flagPSLResPageSize > 0 {
			call = call.PageSize(flagPSLResPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing reservation topics: %w", err)
		}
		all = append(all, resp.Topics...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagPSLResFormat)
}

func runPSLResUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, project, err := pslResService(ctx)
	if err != nil {
		return err
	}
	body := &pubsublite.Reservation{ThroughputCapacity: flagPSLResThroughput}
	got, err := svc.Admin.Projects.Locations.Reservations.
		Patch(pslResName(args[0], project, flagPSLResLocation), body).
		UpdateMask("throughputCapacity").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating reservation: %w", err)
	}
	return emitFormatted(got, flagPSLResFormat)
}
