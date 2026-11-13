package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	gkehub "google.golang.org/api/gkehub/v1"
)

// --- gcloud container fleet (#1137) and gcloud container hub (#1138) ---
//
// `hub` is registered as a compatibility alias that delegates to the same
// implementations as `fleet`.

func (s *containerCmds) initContainerFleet() {
	all := []*cobra.Command{
		s.containerFleetsCreateCmd, s.containerFleetsDeleteCmd, s.containerFleetsDescribeCmd,
		s.containerFleetsListCmd, s.containerFleetsUpdateCmd,
		s.containerFeaturesCreateCmd, s.containerFeaturesDeleteCmd, s.containerFeaturesDescribeCmd,
		s.containerFeaturesListCmd, s.containerFeaturesUpdateCmd,
		s.containerMembershipsCreateCmd, s.containerMembershipsDeleteCmd, s.containerMembershipsDescribeCmd,
		s.containerMembershipsListCmd, s.containerMembershipsUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagCfLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagCfFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{
		s.containerFleetsCreateCmd, s.containerFleetsUpdateCmd,
		s.containerFeaturesCreateCmd, s.containerFeaturesUpdateCmd,
		s.containerMembershipsCreateCmd, s.containerMembershipsUpdateCmd,
	} {
		c.Flags().StringVar(&s.flagCfConfigFile, "config-file", "", "YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	for _, c := range []*cobra.Command{
		s.containerFleetsUpdateCmd, s.containerFeaturesUpdateCmd, s.containerMembershipsUpdateCmd,
	} {
		c.Flags().StringVar(&s.flagCfUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	}
	for _, c := range []*cobra.Command{
		s.containerFleetsListCmd, s.containerFeaturesListCmd, s.containerMembershipsListCmd,
	} {
		c.Flags().Int64Var(&s.flagCfPageSize, "page-size", 0, "Maximum results per page")
	}

	s.containerFleetFleetsCmd.AddCommand(
		s.containerFleetsCreateCmd, s.containerFleetsDeleteCmd, s.containerFleetsDescribeCmd,
		s.containerFleetsListCmd, s.containerFleetsUpdateCmd,
	)
	s.containerFleetFeaturesCmd.AddCommand(
		s.containerFeaturesCreateCmd, s.containerFeaturesDeleteCmd, s.containerFeaturesDescribeCmd,
		s.containerFeaturesListCmd, s.containerFeaturesUpdateCmd,
	)
	s.containerFleetMembershipsCmd.AddCommand(
		s.containerMembershipsCreateCmd, s.containerMembershipsDeleteCmd, s.containerMembershipsDescribeCmd,
		s.containerMembershipsListCmd, s.containerMembershipsUpdateCmd,
	)
	s.containerFleetCmd.AddCommand(s.containerFleetFleetsCmd, s.containerFleetFeaturesCmd, s.containerFleetMembershipsCmd)
	s.containerHubCmd.AddCommand(s.containerFleetFleetsCmd, s.containerFleetFeaturesCmd, s.containerFleetMembershipsCmd)

	// config-management surface added upstream through 580.0.0 (#1805). Register
	// the subcommands as stubs so `gcloud container fleet|hub config-management`
	// is discoverable; a follow-up will wire them up to the GKE Hub API.
	for _, parent := range []*cobra.Command{s.containerFleetCmd, s.containerHubCmd} {
		registerStubGroup(parent, "config-management",
			"Manage Config Management fleet feature",
			"apply", "delete", "describe", "disable", "enable",
			"fetch-for-apply", "status", "unmanage", "update", "upgrade", "version")
	}

	// rollouts and rolloutsequences promoted to GA in 575.0.0 (#1805). Register
	// as stub groups until the underlying GKE Hub rollout API is wired up.
	registerStubGroup(s.containerFleetCmd, "rollouts",
		"Manage GKE Hub fleet rollouts",
		"cancel", "create", "delete", "describe", "force-complete-stage",
		"list", "pause", "resume")
	registerStubGroup(s.containerFleetCmd, "rolloutsequences",
		"Manage GKE Hub fleet rollout sequences",
		"create", "delete", "describe", "list", "update", "upgrade")

	s.containerCmd.AddCommand(s.containerFleetCmd, s.containerHubCmd)
}

func (s *containerCmds) cfLocationParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	if s.flagCfLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagCfLocation), nil
}

func (s *containerCmds) cfResource(collection, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.cfLocationParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id), nil
}

// fleets

func (s *containerCmds) runCfFleetsCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.cfLocationParent()
	if err != nil {
		return err
	}
	body := &gkehub.Fleet{}
	if err := loadYAMLOrJSONInto(s.flagCfConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Fleets.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating fleet: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create fleet [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCfFormat)
}

func (s *containerCmds) runCfFleetsDelete(cmd *cobra.Command, args []string) error {
	name, err := s.cfResource("fleets", args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Fleets.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting fleet: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete fleet [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCfFormat)
}

func (s *containerCmds) runCfFleetsDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.cfResource("fleets", args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Fleets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing fleet: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagCfFormat)
}

func (s *containerCmds) runCfFleetsList(cmd *cobra.Command, args []string) error {
	parent, err := s.cfLocationParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEHub(ctx)
	if err != nil {
		return err
	}
	var all []*gkehub.Fleet
	pageToken := ""
	for {
		call := svc.Projects.Locations.Fleets.List(parent).Context(ctx)
		if s.flagCfPageSize > 0 {
			call = call.PageSize(s.flagCfPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagCfFormat)
}

func (s *containerCmds) runCfFleetsUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.cfResource("fleets", args[0])
	if err != nil {
		return err
	}
	body := &gkehub.Fleet{}
	if err := loadYAMLOrJSONInto(s.flagCfConfigFile, body); err != nil {
		return err
	}
	mask := s.flagCfUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEHub(ctx)
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
	fmt.Fprintf(s.inv.stdout, "Update fleet [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCfFormat)
}

// features

func (s *containerCmds) runCfFeaturesCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.cfLocationParent()
	if err != nil {
		return err
	}
	body := &gkehub.Feature{}
	if err := loadYAMLOrJSONInto(s.flagCfConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Features.Create(parent, body).FeatureId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating feature: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create feature [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCfFormat)
}

func (s *containerCmds) runCfFeaturesDelete(cmd *cobra.Command, args []string) error {
	name, err := s.cfResource("features", args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Features.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting feature: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete feature [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCfFormat)
}

func (s *containerCmds) runCfFeaturesDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.cfResource("features", args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Features.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing feature: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagCfFormat)
}

func (s *containerCmds) runCfFeaturesList(cmd *cobra.Command, args []string) error {
	parent, err := s.cfLocationParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEHub(ctx)
	if err != nil {
		return err
	}
	var all []*gkehub.Feature
	pageToken := ""
	for {
		call := svc.Projects.Locations.Features.List(parent).Context(ctx)
		if s.flagCfPageSize > 0 {
			call = call.PageSize(s.flagCfPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagCfFormat)
}

func (s *containerCmds) runCfFeaturesUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.cfResource("features", args[0])
	if err != nil {
		return err
	}
	body := &gkehub.Feature{}
	if err := loadYAMLOrJSONInto(s.flagCfConfigFile, body); err != nil {
		return err
	}
	mask := s.flagCfUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEHub(ctx)
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
	fmt.Fprintf(s.inv.stdout, "Update feature [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCfFormat)
}

// memberships

func (s *containerCmds) runCfMembershipsCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.cfLocationParent()
	if err != nil {
		return err
	}
	body := &gkehub.Membership{}
	if err := loadYAMLOrJSONInto(s.flagCfConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Memberships.Create(parent, body).MembershipId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating membership: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create membership [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCfFormat)
}

func (s *containerCmds) runCfMembershipsDelete(cmd *cobra.Command, args []string) error {
	name, err := s.cfResource("memberships", args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEHub(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Memberships.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting membership: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete membership [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCfFormat)
}

func (s *containerCmds) runCfMembershipsDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.cfResource("memberships", args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Memberships.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing membership: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagCfFormat)
}

func (s *containerCmds) runCfMembershipsList(cmd *cobra.Command, args []string) error {
	parent, err := s.cfLocationParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEHub(ctx)
	if err != nil {
		return err
	}
	var all []*gkehub.Membership
	pageToken := ""
	for {
		call := svc.Projects.Locations.Memberships.List(parent).Context(ctx)
		if s.flagCfPageSize > 0 {
			call = call.PageSize(s.flagCfPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagCfFormat)
}

func (s *containerCmds) runCfMembershipsUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.cfResource("memberships", args[0])
	if err != nil {
		return err
	}
	body := &gkehub.Membership{}
	if err := loadYAMLOrJSONInto(s.flagCfConfigFile, body); err != nil {
		return err
	}
	mask := s.flagCfUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().GKEHub(ctx)
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
	fmt.Fprintf(s.inv.stdout, "Update membership [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagCfFormat)
}
