package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	managedidentities "google.golang.org/api/managedidentities/v1"
)

// --- gcloud active-directory peerings (#1450) ---

var adPeeringsCmd = &cobra.Command{Use: "peerings", Short: "Manage Managed AD domain peerings"}

var (
	flagADPeeringFormat     string
	flagADPeeringConfigFile string
	flagADPeeringUpdateMask string
	flagADPeeringPageSize   int64
	flagADPeeringFilter     string
	flagADPeeringOrderBy    string
)

var (
	adPeeringCreateCmd = &cobra.Command{
		Use: "create PEERING", Short: "Create a Managed AD peering",
		Args: cobra.ExactArgs(1), RunE: runADPeeringCreate,
	}
	adPeeringDeleteCmd = &cobra.Command{
		Use: "delete PEERING", Short: "Delete a Managed AD peering",
		Args: cobra.ExactArgs(1), RunE: runADPeeringDelete,
	}
	adPeeringDescribeCmd = &cobra.Command{
		Use: "describe PEERING", Short: "Describe a Managed AD peering",
		Args: cobra.ExactArgs(1), RunE: runADPeeringDescribe,
	}
	adPeeringListCmd = &cobra.Command{
		Use: "list", Short: "List Managed AD peerings",
		Args: cobra.NoArgs, RunE: runADPeeringList,
	}
	adPeeringUpdateCmd = &cobra.Command{
		Use: "update PEERING", Short: "Update a Managed AD peering",
		Args: cobra.ExactArgs(1), RunE: runADPeeringUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		adPeeringCreateCmd, adPeeringDeleteCmd, adPeeringDescribeCmd,
		adPeeringListCmd, adPeeringUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagADPeeringFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{adPeeringCreateCmd, adPeeringUpdateCmd} {
		c.Flags().StringVar(&flagADPeeringConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Peering body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	adPeeringUpdateCmd.Flags().StringVar(&flagADPeeringUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	adPeeringListCmd.Flags().Int64Var(&flagADPeeringPageSize, "page-size", 0, "Maximum results per page")
	adPeeringListCmd.Flags().StringVar(&flagADPeeringFilter, "filter", "", "Server-side list filter")
	adPeeringListCmd.Flags().StringVar(&flagADPeeringOrderBy, "order-by", "", "Server-side ordering")

	adPeeringsCmd.AddCommand(all...)
	activeDirectoryCmd.AddCommand(adPeeringsCmd)
}

func adPeeringParent(project string) string {
	return fmt.Sprintf("projects/%s/locations/global", project)
}

func adPeeringResource(project, peering string) string {
	return fmt.Sprintf("projects/%s/locations/global/peerings/%s", project, peering)
}

func runADPeeringCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &managedidentities.Peering{}
	if err := loadYAMLOrJSONInto(flagADPeeringConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Peerings.
		Create(adPeeringParent(project), body).
		PeeringId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating peering: %w", err)
	}
	fmt.Printf("Create request issued for peering [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagADPeeringFormat)
}

func runADPeeringDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Peerings.
		Delete(adPeeringResource(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting peering: %w", err)
	}
	fmt.Printf("Delete request issued for peering [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagADPeeringFormat)
}

func runADPeeringDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Global.Peerings.
		Get(adPeeringResource(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing peering: %w", err)
	}
	return emitFormatted(got, flagADPeeringFormat)
}

func runADPeeringList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*managedidentities.Peering
	pageToken := ""
	for {
		call := svc.Projects.Locations.Global.Peerings.List(adPeeringParent(project)).Context(ctx)
		if flagADPeeringPageSize > 0 {
			call = call.PageSize(flagADPeeringPageSize)
		}
		if flagADPeeringFilter != "" {
			call = call.Filter(flagADPeeringFilter)
		}
		if flagADPeeringOrderBy != "" {
			call = call.OrderBy(flagADPeeringOrderBy)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing peerings: %w", err)
		}
		all = append(all, resp.Peerings...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagADPeeringFormat)
}

func runADPeeringUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &managedidentities.Peering{}
	if err := loadYAMLOrJSONInto(flagADPeeringConfigFile, body); err != nil {
		return err
	}
	mask := flagADPeeringUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Global.Peerings.
		Patch(adPeeringResource(project, args[0]), body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating peering: %w", err)
	}
	fmt.Printf("Update request issued for peering [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagADPeeringFormat)
}
