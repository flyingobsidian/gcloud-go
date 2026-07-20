package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	gkehub "google.golang.org/api/gkehub/v1"
)

// --- gcloud container fleet (#1137) and gcloud container hub (#1138) ---
//
// `hub` is registered as a compatibility alias that delegates to the same
// implementations as `fleet`.

var containerFleetCmd = &cobra.Command{Use: "fleet", Short: "Manage GKE Hub fleets, features, and memberships"}
var containerHubCmd = &cobra.Command{Use: "hub", Short: "Alias for gcloud container fleet"}

var containerFleetFleetsCmd = &cobra.Command{Use: "fleets", Short: "Manage fleets"}
var containerFleetFeaturesCmd = &cobra.Command{Use: "features", Short: "Manage fleet features"}
var containerFleetMembershipsCmd = &cobra.Command{Use: "memberships", Short: "Manage fleet memberships"}

var (
	flagCfLocation    string
	flagCfFormat      string
	flagCfConfigFile  string
	flagCfUpdateMask  string
	flagCfPageSize    int64
)

var (
	// fleets: create/delete/describe/list/update
	containerFleetsCreateCmd = &cobra.Command{
		Use: "create FLEET", Short: "Create a fleet",
		Args: cobra.ExactArgs(1), RunE: runCfFleetsCreate,
	}
	containerFleetsDeleteCmd = &cobra.Command{
		Use: "delete FLEET", Short: "Delete a fleet",
		Args: cobra.ExactArgs(1), RunE: runCfFleetsDelete,
	}
	containerFleetsDescribeCmd = &cobra.Command{
		Use: "describe FLEET", Short: "Describe a fleet",
		Args: cobra.ExactArgs(1), RunE: runCfFleetsDescribe,
	}
	containerFleetsListCmd = &cobra.Command{
		Use: "list", Short: "List fleets in the organization",
		Args: cobra.NoArgs, RunE: runCfFleetsList,
	}
	containerFleetsUpdateCmd = &cobra.Command{
		Use: "update FLEET", Short: "Update a fleet",
		Args: cobra.ExactArgs(1), RunE: runCfFleetsUpdate,
	}

	// features: create/delete/describe/list/update
	containerFeaturesCreateCmd = &cobra.Command{
		Use: "create FEATURE", Short: "Create a fleet feature",
		Args: cobra.ExactArgs(1), RunE: runCfFeaturesCreate,
	}
	containerFeaturesDeleteCmd = &cobra.Command{
		Use: "delete FEATURE", Short: "Delete a fleet feature",
		Args: cobra.ExactArgs(1), RunE: runCfFeaturesDelete,
	}
	containerFeaturesDescribeCmd = &cobra.Command{
		Use: "describe FEATURE", Short: "Describe a fleet feature",
		Args: cobra.ExactArgs(1), RunE: runCfFeaturesDescribe,
	}
	containerFeaturesListCmd = &cobra.Command{
		Use: "list", Short: "List fleet features in a location",
		Args: cobra.NoArgs, RunE: runCfFeaturesList,
	}
	containerFeaturesUpdateCmd = &cobra.Command{
		Use: "update FEATURE", Short: "Update a fleet feature",
		Args: cobra.ExactArgs(1), RunE: runCfFeaturesUpdate,
	}

	// memberships: create/delete/describe/list/update
	containerMembershipsCreateCmd = &cobra.Command{
		Use: "create MEMBERSHIP", Short: "Create a fleet membership",
		Args: cobra.ExactArgs(1), RunE: runCfMembershipsCreate,
	}
	containerMembershipsDeleteCmd = &cobra.Command{
		Use: "delete MEMBERSHIP", Short: "Delete a fleet membership",
		Args: cobra.ExactArgs(1), RunE: runCfMembershipsDelete,
	}
	containerMembershipsDescribeCmd = &cobra.Command{
		Use: "describe MEMBERSHIP", Short: "Describe a fleet membership",
		Args: cobra.ExactArgs(1), RunE: runCfMembershipsDescribe,
	}
	containerMembershipsListCmd = &cobra.Command{
		Use: "list", Short: "List fleet memberships in a location",
		Args: cobra.NoArgs, RunE: runCfMembershipsList,
	}
	containerMembershipsUpdateCmd = &cobra.Command{
		Use: "update MEMBERSHIP", Short: "Update a fleet membership",
		Args: cobra.ExactArgs(1), RunE: runCfMembershipsUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		containerFleetsCreateCmd, containerFleetsDeleteCmd, containerFleetsDescribeCmd,
		containerFleetsListCmd, containerFleetsUpdateCmd,
		containerFeaturesCreateCmd, containerFeaturesDeleteCmd, containerFeaturesDescribeCmd,
		containerFeaturesListCmd, containerFeaturesUpdateCmd,
		containerMembershipsCreateCmd, containerMembershipsDeleteCmd, containerMembershipsDescribeCmd,
		containerMembershipsListCmd, containerMembershipsUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagCfLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagCfFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{
		containerFleetsCreateCmd, containerFleetsUpdateCmd,
		containerFeaturesCreateCmd, containerFeaturesUpdateCmd,
		containerMembershipsCreateCmd, containerMembershipsUpdateCmd,
	} {
		c.Flags().StringVar(&flagCfConfigFile, "config-file", "", "YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	for _, c := range []*cobra.Command{
		containerFleetsUpdateCmd, containerFeaturesUpdateCmd, containerMembershipsUpdateCmd,
	} {
		c.Flags().StringVar(&flagCfUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	}
	for _, c := range []*cobra.Command{
		containerFleetsListCmd, containerFeaturesListCmd, containerMembershipsListCmd,
	} {
		c.Flags().Int64Var(&flagCfPageSize, "page-size", 0, "Maximum results per page")
	}

	containerFleetFleetsCmd.AddCommand(
		containerFleetsCreateCmd, containerFleetsDeleteCmd, containerFleetsDescribeCmd,
		containerFleetsListCmd, containerFleetsUpdateCmd,
	)
	containerFleetFeaturesCmd.AddCommand(
		containerFeaturesCreateCmd, containerFeaturesDeleteCmd, containerFeaturesDescribeCmd,
		containerFeaturesListCmd, containerFeaturesUpdateCmd,
	)
	containerFleetMembershipsCmd.AddCommand(
		containerMembershipsCreateCmd, containerMembershipsDeleteCmd, containerMembershipsDescribeCmd,
		containerMembershipsListCmd, containerMembershipsUpdateCmd,
	)
	containerFleetCmd.AddCommand(containerFleetFleetsCmd, containerFleetFeaturesCmd, containerFleetMembershipsCmd)
	containerHubCmd.AddCommand(containerFleetFleetsCmd, containerFleetFeaturesCmd, containerFleetMembershipsCmd)
	containerCmd.AddCommand(containerFleetCmd, containerHubCmd)
}

func cfLocationParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if flagCfLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagCfLocation), nil
}

func cfResource(collection, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := cfLocationParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id), nil
}

// fleets

func runCfFleetsCreate(cmd *cobra.Command, args []string) error {
	parent, err := cfLocationParent()
	if err != nil {
		return err
	}
	body := &gkehub.Fleet{}
	if err := loadYAMLOrJSONInto(flagCfConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.GKEHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Fleets.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating fleet: %w", err)
	}
	fmt.Printf("Create fleet [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCfFormat)
}

func runCfFleetsDelete(cmd *cobra.Command, args []string) error {
	name, err := cfResource("fleets", args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.GKEHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Fleets.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting fleet: %w", err)
	}
	fmt.Printf("Delete fleet [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCfFormat)
}

func runCfFleetsDescribe(cmd *cobra.Command, args []string) error {
	name, err := cfResource("fleets", args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.GKEHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Fleets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing fleet: %w", err)
	}
	return emitFormatted(got, flagCfFormat)
}

func runCfFleetsList(cmd *cobra.Command, args []string) error {
	parent, err := cfLocationParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.GKEHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*gkehub.Fleet
	pageToken := ""
	for {
		call := svc.Projects.Locations.Fleets.List(parent).Context(ctx)
		if flagCfPageSize > 0 {
			call = call.PageSize(flagCfPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing fleets: %w", err)
		}
		all = append(all, resp.Fleets...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagCfFormat)
}

func runCfFleetsUpdate(cmd *cobra.Command, args []string) error {
	name, err := cfResource("fleets", args[0])
	if err != nil {
		return err
	}
	body := &gkehub.Fleet{}
	if err := loadYAMLOrJSONInto(flagCfConfigFile, body); err != nil {
		return err
	}
	mask := flagCfUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.GKEHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Fleets.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating fleet: %w", err)
	}
	fmt.Printf("Update fleet [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCfFormat)
}

// features

func runCfFeaturesCreate(cmd *cobra.Command, args []string) error {
	parent, err := cfLocationParent()
	if err != nil {
		return err
	}
	body := &gkehub.Feature{}
	if err := loadYAMLOrJSONInto(flagCfConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.GKEHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Features.Create(parent, body).FeatureId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating feature: %w", err)
	}
	fmt.Printf("Create feature [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCfFormat)
}

func runCfFeaturesDelete(cmd *cobra.Command, args []string) error {
	name, err := cfResource("features", args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.GKEHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Features.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting feature: %w", err)
	}
	fmt.Printf("Delete feature [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCfFormat)
}

func runCfFeaturesDescribe(cmd *cobra.Command, args []string) error {
	name, err := cfResource("features", args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.GKEHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Features.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing feature: %w", err)
	}
	return emitFormatted(got, flagCfFormat)
}

func runCfFeaturesList(cmd *cobra.Command, args []string) error {
	parent, err := cfLocationParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.GKEHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*gkehub.Feature
	pageToken := ""
	for {
		call := svc.Projects.Locations.Features.List(parent).Context(ctx)
		if flagCfPageSize > 0 {
			call = call.PageSize(flagCfPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing features: %w", err)
		}
		all = append(all, resp.Resources...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagCfFormat)
}

func runCfFeaturesUpdate(cmd *cobra.Command, args []string) error {
	name, err := cfResource("features", args[0])
	if err != nil {
		return err
	}
	body := &gkehub.Feature{}
	if err := loadYAMLOrJSONInto(flagCfConfigFile, body); err != nil {
		return err
	}
	mask := flagCfUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.GKEHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Features.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating feature: %w", err)
	}
	fmt.Printf("Update feature [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCfFormat)
}

// memberships

func runCfMembershipsCreate(cmd *cobra.Command, args []string) error {
	parent, err := cfLocationParent()
	if err != nil {
		return err
	}
	body := &gkehub.Membership{}
	if err := loadYAMLOrJSONInto(flagCfConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.GKEHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Memberships.Create(parent, body).MembershipId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating membership: %w", err)
	}
	fmt.Printf("Create membership [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCfFormat)
}

func runCfMembershipsDelete(cmd *cobra.Command, args []string) error {
	name, err := cfResource("memberships", args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.GKEHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Memberships.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting membership: %w", err)
	}
	fmt.Printf("Delete membership [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCfFormat)
}

func runCfMembershipsDescribe(cmd *cobra.Command, args []string) error {
	name, err := cfResource("memberships", args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.GKEHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Memberships.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing membership: %w", err)
	}
	return emitFormatted(got, flagCfFormat)
}

func runCfMembershipsList(cmd *cobra.Command, args []string) error {
	parent, err := cfLocationParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.GKEHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*gkehub.Membership
	pageToken := ""
	for {
		call := svc.Projects.Locations.Memberships.List(parent).Context(ctx)
		if flagCfPageSize > 0 {
			call = call.PageSize(flagCfPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing memberships: %w", err)
		}
		all = append(all, resp.Resources...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagCfFormat)
}

func runCfMembershipsUpdate(cmd *cobra.Command, args []string) error {
	name, err := cfResource("memberships", args[0])
	if err != nil {
		return err
	}
	body := &gkehub.Membership{}
	if err := loadYAMLOrJSONInto(flagCfConfigFile, body); err != nil {
		return err
	}
	mask := flagCfUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.GKEHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Memberships.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating membership: %w", err)
	}
	fmt.Printf("Update membership [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagCfFormat)
}
