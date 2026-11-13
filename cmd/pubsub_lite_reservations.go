package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	pubsublite "google.golang.org/api/pubsublite/v1"
)

// --- gcloud pubsub lite-reservations (#1172) ---

func (g *pubsubCmds) initPubsubLiteReservations() {
	all := []*cobra.Command{
		g.pubsubLiteResCreateCmd, g.pubsubLiteResDeleteCmd, g.pubsubLiteResDescribeCmd,
		g.pubsubLiteResListCmd, g.pubsubLiteResListTopicsCmd, g.pubsubLiteResUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&g.flagPSLResLocation, "location", "",
			"Regional location containing the reservation, e.g. us-central1 (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagPSLResFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{g.pubsubLiteResListCmd, g.pubsubLiteResListTopicsCmd} {
		c.Flags().Int64Var(&g.flagPSLResPageSize, "page-size", 0, "Maximum number of results per page")
	}
	for _, c := range []*cobra.Command{g.pubsubLiteResCreateCmd, g.pubsubLiteResUpdateCmd} {
		c.Flags().Int64Var(&g.flagPSLResThroughput, "throughput-capacity", 0,
			"Reserved throughput capacity in MiB/s (create: required; update: required)")
	}
	_ = g.pubsubLiteResCreateCmd.MarkFlagRequired("throughput-capacity")
	_ = g.pubsubLiteResUpdateCmd.MarkFlagRequired("throughput-capacity")

	g.pubsubLiteReservationsCmd.AddCommand(all...)
	g.pubsubCmd.AddCommand(g.pubsubLiteReservationsCmd)
}

func pslResName(id, project, location string) string {
	return pubsubLiteChild("reservations", id, pubsubLiteLocationParent(project, location))
}

func (g *pubsubCmds) pslResService(ctx context.Context) (*pubsublite.Service, string, error) {
	project, err := g.inv.project()
	if err != nil {
		return nil, "", err
	}
	region, err := pubsubLiteRegion(g.flagPSLResLocation)
	if err != nil {
		return nil, "", err
	}
	svc, err := g.inv.clients().PubSubLite(ctx, region)
	if err != nil {
		return nil, "", err
	}
	return svc, project, nil
}

func (g *pubsubCmds) runPSLResCreate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, project, err := g.pslResService(ctx)
	if err != nil {
		return err
	}
	body := &pubsublite.Reservation{ThroughputCapacity: g.flagPSLResThroughput}
	got, err := svc.Admin.Projects.Locations.Reservations.
		Create(pubsubLiteLocationParent(project, g.flagPSLResLocation), body).
		ReservationId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating reservation: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPSLResFormat)
}

func (g *pubsubCmds) runPSLResDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, project, err := g.pslResService(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Admin.Projects.Locations.Reservations.
		Delete(pslResName(args[0], project, g.flagPSLResLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting reservation: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted reservation [%s].\n", args[0])
	return nil
}

func (g *pubsubCmds) runPSLResDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, project, err := g.pslResService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Admin.Projects.Locations.Reservations.
		Get(pslResName(args[0], project, g.flagPSLResLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing reservation: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPSLResFormat)
}

func (g *pubsubCmds) runPSLResList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, project, err := g.pslResService(ctx)
	if err != nil {
		return err
	}
	parent := pubsubLiteLocationParent(project, g.flagPSLResLocation)
	var all []*pubsublite.Reservation
	pageToken := ""
	for {
		call := svc.Admin.Projects.Locations.Reservations.List(parent).Context(ctx)
		if g.flagPSLResPageSize > 0 {
			call = call.PageSize(g.flagPSLResPageSize)
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
	return emitFormattedTo(g.inv.stdout, all, g.flagPSLResFormat)
}

func (g *pubsubCmds) runPSLResListTopics(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, project, err := g.pslResService(ctx)
	if err != nil {
		return err
	}
	name := pslResName(args[0], project, g.flagPSLResLocation)
	var all []string
	pageToken := ""
	for {
		call := svc.Admin.Projects.Locations.Reservations.Topics.List(name).Context(ctx)
		if g.flagPSLResPageSize > 0 {
			call = call.PageSize(g.flagPSLResPageSize)
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
	return emitFormattedTo(g.inv.stdout, all, g.flagPSLResFormat)
}

func (g *pubsubCmds) runPSLResUpdate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, project, err := g.pslResService(ctx)
	if err != nil {
		return err
	}
	body := &pubsublite.Reservation{ThroughputCapacity: g.flagPSLResThroughput}
	got, err := svc.Admin.Projects.Locations.Reservations.
		Patch(pslResName(args[0], project, g.flagPSLResLocation), body).
		UpdateMask("throughputCapacity").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating reservation: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagPSLResFormat)
}
