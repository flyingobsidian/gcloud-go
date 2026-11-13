package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	datacatalog "google.golang.org/api/datacatalog/v1"
)

// --- gcloud data-catalog entry-groups (#1505) ---

func (s *dataCatalogCmds) initDataCatalogEntryGroups() {
	all := []*cobra.Command{
		s.dcEGCreateCmd, s.dcEGDeleteCmd, s.dcEGDescribeCmd, s.dcEGListCmd, s.dcEGUpdateCmd,
		s.dcEGGetIamCmd, s.dcEGSetIamCmd, s.dcEGAddIamCmd, s.dcEGRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDCEGLocation, "location", "", "Location for the entry group (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagDCEGFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.dcEGCreateCmd, s.dcEGUpdateCmd} {
		c.Flags().StringVar(&s.flagDCEGConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the EntryGroup body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.dcEGUpdateCmd.Flags().StringVar(&s.flagDCEGUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.dcEGListCmd.Flags().Int64Var(&s.flagDCEGPageSize, "page-size", 0, "Maximum results per page")

	for _, c := range []*cobra.Command{s.dcEGAddIamCmd, s.dcEGRemoveIamCmd} {
		dcIamMemberFlags(c, &s.flagDCEGIamMember, &s.flagDCEGIamRole, &s.flagDCEGIamCondExpr, &s.flagDCEGIamCondT, &s.flagDCEGIamCondD)
	}
	s.dcEGRemoveIamCmd.Flags().BoolVar(&s.flagDCEGIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.dcEntryGroupsCmd.AddCommand(all...)
	s.dataCatalogCmd.AddCommand(s.dcEntryGroupsCmd)
}

func (s *dataCatalogCmds) dcEGParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return dcLocationParent(project, s.flagDCEGLocation), nil
}

func (s *dataCatalogCmds) dcEGName(id string) (string, error) {
	parent, err := s.dcEGParent()
	if err != nil {
		return "", err
	}
	return dcChild("entryGroups", id, parent), nil
}

func (s *dataCatalogCmds) runDCEGCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.dcEGParent()
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1EntryGroup{}
	if err := loadYAMLOrJSONInto(s.flagDCEGConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.EntryGroups.Create(parent, body).EntryGroupId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating entry group: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created entry group [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDCEGFormat)
}

func (s *dataCatalogCmds) runDCEGDelete(cmd *cobra.Command, args []string) error {
	name, err := s.dcEGName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.EntryGroups.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting entry group: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted entry group [%s].\n", args[0])
	return nil
}

func (s *dataCatalogCmds) runDCEGDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.dcEGName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.EntryGroups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing entry group: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDCEGFormat)
}

func (s *dataCatalogCmds) runDCEGList(cmd *cobra.Command, args []string) error {
	parent, err := s.dcEGParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	var all []*datacatalog.GoogleCloudDatacatalogV1EntryGroup
	pageToken := ""
	for {
		call := svc.Projects.Locations.EntryGroups.List(parent).Context(ctx)
		if s.flagDCEGPageSize > 0 {
			call = call.PageSize(s.flagDCEGPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing entry groups: %w", err)
		}
		all = append(all, resp.EntryGroups...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagDCEGFormat)
}

func (s *dataCatalogCmds) runDCEGUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.dcEGName(args[0])
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1EntryGroup{}
	if err := loadYAMLOrJSONInto(s.flagDCEGConfigFile, body); err != nil {
		return err
	}
	mask := s.flagDCEGUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.EntryGroups.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating entry group: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated entry group [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDCEGFormat)
}

func (s *dataCatalogCmds) runDCEGGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.dcEGName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.EntryGroups.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagDCEGFormat)
}

func (s *dataCatalogCmds) runDCEGSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.dcEGName(args[0])
	if err != nil {
		return err
	}
	policy := &datacatalog.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.EntryGroups.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.dcUpdatedIam(fmt.Sprintf("entry group [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDCEGFormat)
}

func (s *dataCatalogCmds) runDCEGAddIam(cmd *cobra.Command, args []string) error {
	name, err := s.dcEGName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.EntryGroups.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	dcAddBinding(policy, s.flagDCEGIamRole, s.flagDCEGIamMember,
		dcBuildCondition(s.flagDCEGIamCondExpr, s.flagDCEGIamCondT, s.flagDCEGIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.EntryGroups.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.dcUpdatedIam(fmt.Sprintf("entry group [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDCEGFormat)
}

func (s *dataCatalogCmds) runDCEGRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := s.dcEGName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.EntryGroups.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !dcRemoveBinding(policy, s.flagDCEGIamRole, s.flagDCEGIamMember,
		dcBuildCondition(s.flagDCEGIamCondExpr, s.flagDCEGIamCondT, s.flagDCEGIamCondD), s.flagDCEGIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", s.flagDCEGIamRole, s.flagDCEGIamMember)
	}
	updated, err := svc.Projects.Locations.EntryGroups.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.dcUpdatedIam(fmt.Sprintf("entry group [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDCEGFormat)
}
