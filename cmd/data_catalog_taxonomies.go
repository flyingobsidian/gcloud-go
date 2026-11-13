package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	datacatalog "google.golang.org/api/datacatalog/v1"
)

// --- gcloud data-catalog taxonomies (#1509) ---

func (s *dataCatalogCmds) initDataCatalogTaxonomies() {
	all := []*cobra.Command{
		s.dcTaxDescribeCmd, s.dcTaxListCmd, s.dcTaxExportCmd, s.dcTaxImportCmd,
		s.dcTaxGetIamCmd, s.dcTaxSetIamCmd, s.dcTaxAddIamCmd, s.dcTaxRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDCTaxLocation, "location", "", "Location for the taxonomy (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagDCTaxFormat, "format", "", "Output format")
	}
	s.dcTaxListCmd.Flags().StringVar(&s.flagDCTaxFilter, "filter", "", "Server-side filter expression")
	s.dcTaxListCmd.Flags().Int64Var(&s.flagDCTaxPageSize, "page-size", 0, "Maximum results per page")

	s.dcTaxExportCmd.Flags().StringSliceVar(&s.flagDCTaxTaxIds, "taxonomies", nil,
		"Taxonomy IDs to export (required, comma-separated or repeated)")
	_ = s.dcTaxExportCmd.MarkFlagRequired("taxonomies")
	s.dcTaxExportCmd.Flags().BoolVar(&s.flagDCTaxSerialized, "serialized", true,
		"Return serialized taxonomies (default true)")

	s.dcTaxImportCmd.Flags().StringVar(&s.flagDCTaxImportFile, "source", "",
		"Path to a JSON file with a SerializedTaxonomy list (mutually exclusive with --cross-regional-source)")
	s.dcTaxImportCmd.Flags().StringVar(&s.flagDCTaxCrossRegional, "cross-regional-source", "",
		"Fully-qualified source taxonomy resource name to import cross-regionally")

	for _, c := range []*cobra.Command{s.dcTaxAddIamCmd, s.dcTaxRemoveIamCmd} {
		dcIamMemberFlags(c, &s.flagDCTaxIamMember, &s.flagDCTaxIamRole, &s.flagDCTaxIamCondExpr, &s.flagDCTaxIamCondT, &s.flagDCTaxIamCondD)
	}
	s.dcTaxRemoveIamCmd.Flags().BoolVar(&s.flagDCTaxIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.dcTaxCmd.AddCommand(all...)
	s.dcTaxCmd.AddCommand(s.dcPTCmd)
	s.dataCatalogCmd.AddCommand(s.dcTaxCmd)
}

func (s *dataCatalogCmds) dcTaxParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return dcLocationParent(project, s.flagDCTaxLocation), nil
}

func (s *dataCatalogCmds) dcTaxName(id string) (string, error) {
	parent, err := s.dcTaxParent()
	if err != nil {
		return "", err
	}
	return dcChild("taxonomies", id, parent), nil
}

func (s *dataCatalogCmds) runDCTaxDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.dcTaxName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Taxonomies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing taxonomy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDCTaxFormat)
}

func (s *dataCatalogCmds) runDCTaxList(cmd *cobra.Command, args []string) error {
	parent, err := s.dcTaxParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	var all []*datacatalog.GoogleCloudDatacatalogV1Taxonomy
	pageToken := ""
	for {
		call := svc.Projects.Locations.Taxonomies.List(parent).Context(ctx)
		if s.flagDCTaxFilter != "" {
			call = call.Filter(s.flagDCTaxFilter)
		}
		if s.flagDCTaxPageSize > 0 {
			call = call.PageSize(s.flagDCTaxPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing taxonomies: %w", err)
		}
		all = append(all, resp.Taxonomies...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagDCTaxFormat)
}

func (s *dataCatalogCmds) runDCTaxExport(cmd *cobra.Command, args []string) error {
	parent, err := s.dcTaxParent()
	if err != nil {
		return err
	}
	// Build fully-qualified taxonomy names.
	names := make([]string, 0, len(s.flagDCTaxTaxIds))
	for _, id := range s.flagDCTaxTaxIds {
		if strings.HasPrefix(id, "projects/") {
			names = append(names, id)
		} else {
			names = append(names, dcChild("taxonomies", id, parent))
		}
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Taxonomies.Export(parent).
		Taxonomies(names...).SerializedTaxonomies(s.flagDCTaxSerialized).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting taxonomies: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagDCTaxFormat)
}

func (s *dataCatalogCmds) runDCTaxImport(cmd *cobra.Command, args []string) error {
	parent, err := s.dcTaxParent()
	if err != nil {
		return err
	}
	if (s.flagDCTaxImportFile == "" && s.flagDCTaxCrossRegional == "") ||
		(s.flagDCTaxImportFile != "" && s.flagDCTaxCrossRegional != "") {
		return fmt.Errorf("exactly one of --source or --cross-regional-source is required")
	}
	req := &datacatalog.GoogleCloudDatacatalogV1ImportTaxonomiesRequest{}
	if s.flagDCTaxImportFile != "" {
		var payload struct {
			Taxonomies []*datacatalog.GoogleCloudDatacatalogV1SerializedTaxonomy `json:"taxonomies"`
		}
		if err := loadYAMLOrJSONInto(s.flagDCTaxImportFile, &payload); err != nil {
			return err
		}
		if len(payload.Taxonomies) == 0 {
			return fmt.Errorf("no taxonomies in source file")
		}
		req.InlineSource = &datacatalog.GoogleCloudDatacatalogV1InlineSource{Taxonomies: payload.Taxonomies}
	} else {
		req.CrossRegionalSource = &datacatalog.GoogleCloudDatacatalogV1CrossRegionalSource{
			Taxonomy: s.flagDCTaxCrossRegional,
		}
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Taxonomies.Import(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing taxonomies: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagDCTaxFormat)
}

func (s *dataCatalogCmds) runDCTaxGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.dcTaxName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Taxonomies.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagDCTaxFormat)
}

func (s *dataCatalogCmds) runDCTaxSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.dcTaxName(args[0])
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
	updated, err := svc.Projects.Locations.Taxonomies.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.dcUpdatedIam(fmt.Sprintf("taxonomy [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDCTaxFormat)
}

func (s *dataCatalogCmds) runDCTaxAddIam(cmd *cobra.Command, args []string) error {
	name, err := s.dcTaxName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Taxonomies.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	dcAddBinding(policy, s.flagDCTaxIamRole, s.flagDCTaxIamMember,
		dcBuildCondition(s.flagDCTaxIamCondExpr, s.flagDCTaxIamCondT, s.flagDCTaxIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.Taxonomies.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.dcUpdatedIam(fmt.Sprintf("taxonomy [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDCTaxFormat)
}

func (s *dataCatalogCmds) runDCTaxRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := s.dcTaxName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Taxonomies.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !dcRemoveBinding(policy, s.flagDCTaxIamRole, s.flagDCTaxIamMember,
		dcBuildCondition(s.flagDCTaxIamCondExpr, s.flagDCTaxIamCondT, s.flagDCTaxIamCondD), s.flagDCTaxIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", s.flagDCTaxIamRole, s.flagDCTaxIamMember)
	}
	updated, err := svc.Projects.Locations.Taxonomies.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.dcUpdatedIam(fmt.Sprintf("taxonomy [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDCTaxFormat)
}

// --- policy-tags subgroup ---

func (s *dataCatalogCmds) initDataCatalogTaxonomies8() {
	all := []*cobra.Command{
		s.dcPTCreateCmd, s.dcPTDeleteCmd, s.dcPTDescribeCmd, s.dcPTListCmd, s.dcPTUpdateCmd,
		s.dcPTGetIamCmd, s.dcPTSetIamCmd, s.dcPTAddIamCmd, s.dcPTRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDCTaxLocation, "location", "", "Location for the taxonomy (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagDCPTTaxonomy, "taxonomy", "", "Taxonomy ID that owns the policy tag (required)")
		_ = c.MarkFlagRequired("taxonomy")
		c.Flags().StringVar(&s.flagDCTaxFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.dcPTCreateCmd, s.dcPTUpdateCmd} {
		c.Flags().StringVar(&s.flagDCPTConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the PolicyTag body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.dcPTUpdateCmd.Flags().StringVar(&s.flagDCPTUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.dcPTListCmd.Flags().Int64Var(&s.flagDCPTPageSize, "page-size", 0, "Maximum results per page")

	for _, c := range []*cobra.Command{s.dcPTAddIamCmd, s.dcPTRemoveIamCmd} {
		dcIamMemberFlags(c, &s.flagDCPTIamMember, &s.flagDCPTIamRole, &s.flagDCPTIamCondExpr, &s.flagDCPTIamCondT, &s.flagDCPTIamCondD)
	}
	s.dcPTRemoveIamCmd.Flags().BoolVar(&s.flagDCPTIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.dcPTCmd.AddCommand(all...)
}

func (s *dataCatalogCmds) dcPTParent() (string, error) {
	return s.dcTaxName(s.flagDCPTTaxonomy)
}

func (s *dataCatalogCmds) dcPTName(id string) (string, error) {
	parent, err := s.dcPTParent()
	if err != nil {
		return "", err
	}
	return dcChild("policyTags", id, parent), nil
}

func (s *dataCatalogCmds) runDCPTCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.dcPTParent()
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1PolicyTag{}
	if err := loadYAMLOrJSONInto(s.flagDCPTConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Taxonomies.PolicyTags.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating policy tag: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created policy tag [%s].\n", got.Name)
	return emitFormattedTo(s.inv.stdout, got, s.flagDCTaxFormat)
}

func (s *dataCatalogCmds) runDCPTDelete(cmd *cobra.Command, args []string) error {
	name, err := s.dcPTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Taxonomies.PolicyTags.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting policy tag: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted policy tag [%s].\n", args[0])
	return nil
}

func (s *dataCatalogCmds) runDCPTDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.dcPTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Taxonomies.PolicyTags.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing policy tag: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDCTaxFormat)
}

func (s *dataCatalogCmds) runDCPTList(cmd *cobra.Command, args []string) error {
	parent, err := s.dcPTParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	var all []*datacatalog.GoogleCloudDatacatalogV1PolicyTag
	pageToken := ""
	for {
		call := svc.Projects.Locations.Taxonomies.PolicyTags.List(parent).Context(ctx)
		if s.flagDCPTPageSize > 0 {
			call = call.PageSize(s.flagDCPTPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing policy tags: %w", err)
		}
		all = append(all, resp.PolicyTags...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagDCTaxFormat)
}

func (s *dataCatalogCmds) runDCPTUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.dcPTName(args[0])
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1PolicyTag{}
	if err := loadYAMLOrJSONInto(s.flagDCPTConfigFile, body); err != nil {
		return err
	}
	mask := s.flagDCPTUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Taxonomies.PolicyTags.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating policy tag: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated policy tag [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDCTaxFormat)
}

func (s *dataCatalogCmds) runDCPTGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.dcPTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Taxonomies.PolicyTags.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagDCTaxFormat)
}

func (s *dataCatalogCmds) runDCPTSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.dcPTName(args[0])
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
	updated, err := svc.Projects.Locations.Taxonomies.PolicyTags.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.dcUpdatedIam(fmt.Sprintf("policy tag [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDCTaxFormat)
}

func (s *dataCatalogCmds) runDCPTAddIam(cmd *cobra.Command, args []string) error {
	name, err := s.dcPTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Taxonomies.PolicyTags.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	dcAddBinding(policy, s.flagDCPTIamRole, s.flagDCPTIamMember,
		dcBuildCondition(s.flagDCPTIamCondExpr, s.flagDCPTIamCondT, s.flagDCPTIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.Taxonomies.PolicyTags.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.dcUpdatedIam(fmt.Sprintf("policy tag [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDCTaxFormat)
}

func (s *dataCatalogCmds) runDCPTRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := s.dcPTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Taxonomies.PolicyTags.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !dcRemoveBinding(policy, s.flagDCPTIamRole, s.flagDCPTIamMember,
		dcBuildCondition(s.flagDCPTIamCondExpr, s.flagDCPTIamCondT, s.flagDCPTIamCondD), s.flagDCPTIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", s.flagDCPTIamRole, s.flagDCPTIamMember)
	}
	updated, err := svc.Projects.Locations.Taxonomies.PolicyTags.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.dcUpdatedIam(fmt.Sprintf("policy tag [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDCTaxFormat)
}
