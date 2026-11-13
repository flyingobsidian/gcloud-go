package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	cloudiam "google.golang.org/api/iam/v1"
)

// --- gcloud iam workforce-pools (#1014) ---
//
// Workforce pools live under organizations (not projects) at
// locations/LOCATION. This surface adds the top-level pool CRUD plus the
// providers subgroup; the subjects subgroup is in
// iam_workforce_pools_subjects.go. The operations sub-subgroups are not yet
// ported.

func (s *iamCmds) initIamWorkforcePools() {
	pools := []*cobra.Command{
		s.iamWpCreateCmd, s.iamWpDeleteCmd, s.iamWpDescribeCmd,
		s.iamWpListCmd, s.iamWpUpdateCmd, s.iamWpUndeleteCmd,
	}
	providers := []*cobra.Command{
		s.iamWpProvCreateCmd, s.iamWpProvDeleteCmd, s.iamWpProvDescribeCmd,
		s.iamWpProvListCmd, s.iamWpProvUpdateCmd, s.iamWpProvUndeleteCmd,
	}
	all := append(append([]*cobra.Command{}, pools...), providers...)
	for _, c := range all {
		c.Flags().StringVar(&s.flagIamWpOrganization, "organization", "", "Owning organization ID (required)")
		_ = c.MarkFlagRequired("organization")
		c.Flags().StringVar(&s.flagIamWpLocation, "location", "global", "Location (defaults to global)")
		c.Flags().StringVar(&s.flagIamWpFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.iamWpCreateCmd, s.iamWpUpdateCmd, s.iamWpProvCreateCmd, s.iamWpProvUpdateCmd} {
		c.Flags().StringVar(&s.flagIamWpConfigFile, "config-file", "", "YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	for _, c := range []*cobra.Command{s.iamWpUpdateCmd, s.iamWpProvUpdateCmd} {
		c.Flags().StringVar(&s.flagIamWpUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	}
	for _, c := range []*cobra.Command{s.iamWpListCmd, s.iamWpProvListCmd} {
		c.Flags().Int64Var(&s.flagIamWpPageSize, "page-size", 0, "Maximum results per page")
	}
	for _, c := range providers {
		c.Flags().StringVar(&s.flagIamWpPool, "workforce-pool", "", "Owning workforce pool (required)")
		_ = c.MarkFlagRequired("workforce-pool")
	}

	s.iamWorkforcePoolsProvidersCmd.AddCommand(providers...)
	s.iamWorkforcePoolsCmd.AddCommand(pools...)
	s.iamWorkforcePoolsCmd.AddCommand(s.iamWorkforcePoolsProvidersCmd)
	s.iamCmd.AddCommand(s.iamWorkforcePoolsCmd)
}

func (s *iamCmds) iamWpLocationParent() (string, error) {
	if s.flagIamWpOrganization == "" {
		return "", fmt.Errorf("--organization is required")
	}
	loc := s.flagIamWpLocation
	if loc == "" {
		loc = "global"
	}
	return fmt.Sprintf("locations/%s", loc), nil
}

func (s *iamCmds) iamWpPoolName(id string) (string, error) {
	if strings.HasPrefix(id, "locations/") {
		return id, nil
	}
	parent, err := s.iamWpLocationParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/workforcePools/%s", parent, id), nil
}

func (s *iamCmds) iamWpProviderName(id string) (string, error) {
	if strings.HasPrefix(id, "locations/") {
		return id, nil
	}
	if s.flagIamWpPool == "" {
		return "", fmt.Errorf("--workforce-pool is required")
	}
	parent, err := s.iamWpLocationParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/workforcePools/%s/providers/%s", parent, s.flagIamWpPool, id), nil
}

// pools

func (s *iamCmds) runIamWpCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.iamWpLocationParent()
	if err != nil {
		return err
	}
	body := &cloudiam.WorkforcePool{}
	if err := loadYAMLOrJSONInto(s.flagIamWpConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Locations.WorkforcePools.Create(parent, body).WorkforcePoolId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating workforce pool: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create workforce pool [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagIamWpFormat)
}

func (s *iamCmds) runIamWpDelete(cmd *cobra.Command, args []string) error {
	name, err := s.iamWpPoolName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Locations.WorkforcePools.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting workforce pool: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete workforce pool [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagIamWpFormat)
}

func (s *iamCmds) runIamWpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.iamWpPoolName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Locations.WorkforcePools.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing workforce pool: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIamWpFormat)
}

func (s *iamCmds) runIamWpList(cmd *cobra.Command, args []string) error {
	parent, err := s.iamWpLocationParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	var all []*cloudiam.WorkforcePool
	pageToken := ""
	for {
		call := svc.Locations.WorkforcePools.List(parent).Parent("organizations/" + s.flagIamWpOrganization).Context(ctx)
		if s.flagIamWpPageSize > 0 {
			call = call.PageSize(s.flagIamWpPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing workforce pools: %w", err)
		}
		all = append(all, resp.WorkforcePools...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagIamWpFormat)
}

func (s *iamCmds) runIamWpUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.iamWpPoolName(args[0])
	if err != nil {
		return err
	}
	body := &cloudiam.WorkforcePool{}
	if err := loadYAMLOrJSONInto(s.flagIamWpConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagIamWpUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	call := svc.Locations.WorkforcePools.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating workforce pool: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update workforce pool [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagIamWpFormat)
}

func (s *iamCmds) runIamWpUndelete(cmd *cobra.Command, args []string) error {
	name, err := s.iamWpPoolName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Locations.WorkforcePools.Undelete(name, &cloudiam.UndeleteWorkforcePoolRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeleting workforce pool: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Undelete workforce pool [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagIamWpFormat)
}

// providers

func (s *iamCmds) runIamWpProvCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.iamWpLocationParent()
	if err != nil {
		return err
	}
	poolParent := fmt.Sprintf("%s/workforcePools/%s", parent, s.flagIamWpPool)
	body := &cloudiam.WorkforcePoolProvider{}
	if err := loadYAMLOrJSONInto(s.flagIamWpConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Locations.WorkforcePools.Providers.Create(poolParent, body).WorkforcePoolProviderId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating provider: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create workforce-pool provider [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagIamWpFormat)
}

func (s *iamCmds) runIamWpProvDelete(cmd *cobra.Command, args []string) error {
	name, err := s.iamWpProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Locations.WorkforcePools.Providers.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting provider: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete workforce-pool provider [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagIamWpFormat)
}

func (s *iamCmds) runIamWpProvDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.iamWpProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Locations.WorkforcePools.Providers.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing provider: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIamWpFormat)
}

func (s *iamCmds) runIamWpProvList(cmd *cobra.Command, args []string) error {
	parent, err := s.iamWpLocationParent()
	if err != nil {
		return err
	}
	poolParent := fmt.Sprintf("%s/workforcePools/%s", parent, s.flagIamWpPool)
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	var all []*cloudiam.WorkforcePoolProvider
	pageToken := ""
	for {
		call := svc.Locations.WorkforcePools.Providers.List(poolParent).Context(ctx)
		if s.flagIamWpPageSize > 0 {
			call = call.PageSize(s.flagIamWpPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing providers: %w", err)
		}
		all = append(all, resp.WorkforcePoolProviders...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagIamWpFormat)
}

func (s *iamCmds) runIamWpProvUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.iamWpProviderName(args[0])
	if err != nil {
		return err
	}
	body := &cloudiam.WorkforcePoolProvider{}
	if err := loadYAMLOrJSONInto(s.flagIamWpConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagIamWpUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	call := svc.Locations.WorkforcePools.Providers.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating provider: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update workforce-pool provider [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagIamWpFormat)
}

func (s *iamCmds) runIamWpProvUndelete(cmd *cobra.Command, args []string) error {
	name, err := s.iamWpProviderName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Locations.WorkforcePools.Providers.Undelete(name, &cloudiam.UndeleteWorkforcePoolProviderRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeleting provider: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Undelete workforce-pool provider [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagIamWpFormat)
}
