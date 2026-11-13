package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	datacatalog "google.golang.org/api/datacatalog/v1"
)

// --- gcloud data-catalog tag-templates (#1508) ---

func (s *dataCatalogCmds) initDataCatalogTagTemplates() {
	all := []*cobra.Command{
		s.dcTTCreateCmd, s.dcTTDeleteCmd, s.dcTTDescribeCmd, s.dcTTUpdateCmd,
		s.dcTTGetIamCmd, s.dcTTSetIamCmd, s.dcTTAddIamCmd, s.dcTTRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDCTTLocation, "location", "", "Location for the tag template (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagDCTTFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.dcTTCreateCmd, s.dcTTUpdateCmd} {
		c.Flags().StringVar(&s.flagDCTTConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the TagTemplate body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.dcTTUpdateCmd.Flags().StringVar(&s.flagDCTTUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.dcTTDeleteCmd.Flags().BoolVar(&s.flagDCTTForce, "force", false, "Force delete even if the template has tags")

	for _, c := range []*cobra.Command{s.dcTTAddIamCmd, s.dcTTRemoveIamCmd} {
		dcIamMemberFlags(c, &s.flagDCTTIamMember, &s.flagDCTTIamRole, &s.flagDCTTIamCondExpr, &s.flagDCTTIamCondT, &s.flagDCTTIamCondD)
	}
	s.dcTTRemoveIamCmd.Flags().BoolVar(&s.flagDCTTIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.dcTTCmd.AddCommand(all...)
	s.dcTTCmd.AddCommand(s.dcTTFieldsCmd)
	s.dataCatalogCmd.AddCommand(s.dcTTCmd)
}

func (s *dataCatalogCmds) dcTTParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return dcLocationParent(project, s.flagDCTTLocation), nil
}

func (s *dataCatalogCmds) dcTTName(id string) (string, error) {
	parent, err := s.dcTTParent()
	if err != nil {
		return "", err
	}
	return dcChild("tagTemplates", id, parent), nil
}

func (s *dataCatalogCmds) runDCTTCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.dcTTParent()
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1TagTemplate{}
	if err := loadYAMLOrJSONInto(s.flagDCTTConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TagTemplates.Create(parent, body).TagTemplateId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating tag template: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created tag template [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDCTTFormat)
}

func (s *dataCatalogCmds) runDCTTDelete(cmd *cobra.Command, args []string) error {
	name, err := s.dcTTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.TagTemplates.Delete(name).Context(ctx)
	if s.flagDCTTForce {
		call = call.Force(true)
	}
	if _, err := call.Do(); err != nil {
		return fmt.Errorf("deleting tag template: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted tag template [%s].\n", args[0])
	return nil
}

func (s *dataCatalogCmds) runDCTTDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.dcTTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TagTemplates.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing tag template: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDCTTFormat)
}

func (s *dataCatalogCmds) runDCTTUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.dcTTName(args[0])
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1TagTemplate{}
	if err := loadYAMLOrJSONInto(s.flagDCTTConfigFile, body); err != nil {
		return err
	}
	mask := s.flagDCTTUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.TagTemplates.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating tag template: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated tag template [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDCTTFormat)
}

func (s *dataCatalogCmds) runDCTTGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.dcTTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.TagTemplates.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagDCTTFormat)
}

func (s *dataCatalogCmds) runDCTTSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.dcTTName(args[0])
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
	updated, err := svc.Projects.Locations.TagTemplates.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.dcUpdatedIam(fmt.Sprintf("tag template [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDCTTFormat)
}

func (s *dataCatalogCmds) runDCTTAddIam(cmd *cobra.Command, args []string) error {
	name, err := s.dcTTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.TagTemplates.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	dcAddBinding(policy, s.flagDCTTIamRole, s.flagDCTTIamMember,
		dcBuildCondition(s.flagDCTTIamCondExpr, s.flagDCTTIamCondT, s.flagDCTTIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.TagTemplates.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.dcUpdatedIam(fmt.Sprintf("tag template [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDCTTFormat)
}

func (s *dataCatalogCmds) runDCTTRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := s.dcTTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.TagTemplates.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !dcRemoveBinding(policy, s.flagDCTTIamRole, s.flagDCTTIamMember,
		dcBuildCondition(s.flagDCTTIamCondExpr, s.flagDCTTIamCondT, s.flagDCTTIamCondD), s.flagDCTTIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", s.flagDCTTIamRole, s.flagDCTTIamMember)
	}
	updated, err := svc.Projects.Locations.TagTemplates.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.dcUpdatedIam(fmt.Sprintf("tag template [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDCTTFormat)
}

// --- tag-templates fields subgroup ---

func (s *dataCatalogCmds) initDataCatalogTagTemplates5() {
	all := []*cobra.Command{s.dcTTFieldCreateCmd, s.dcTTFieldDeleteCmd, s.dcTTFieldUpdateCmd, s.dcTTFieldRenameCmd, s.dcTTFieldRenameEnumCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDCTTLocation, "location", "", "Location for the tag template (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagDCTTFieldTemplate, "template", "", "Tag template ID that owns the field (required)")
		_ = c.MarkFlagRequired("template")
		c.Flags().StringVar(&s.flagDCTTFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.dcTTFieldCreateCmd, s.dcTTFieldUpdateCmd} {
		c.Flags().StringVar(&s.flagDCTTFieldConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the TagTemplateField body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.dcTTFieldUpdateCmd.Flags().StringVar(&s.flagDCTTFieldUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.dcTTFieldRenameCmd.Flags().StringVar(&s.flagDCTTFieldNewID, "new-id", "", "New tag template field ID (required)")
	_ = s.dcTTFieldRenameCmd.MarkFlagRequired("new-id")
	s.dcTTFieldRenameEnumCmd.Flags().StringVar(&s.flagDCTTFieldEnumRename, "new-display-name", "",
		"New display name for the enum value (required)")
	_ = s.dcTTFieldRenameEnumCmd.MarkFlagRequired("new-display-name")

	s.dcTTFieldsCmd.AddCommand(all...)
}

func (s *dataCatalogCmds) dcTTFieldName(template, field string) (string, error) {
	parent, err := s.dcTTName(template)
	if err != nil {
		return "", err
	}
	return dcChild("fields", field, parent), nil
}

func (s *dataCatalogCmds) runDCTTFieldCreate(cmd *cobra.Command, args []string) error {
	// args[0] = FIELD id; require --template as flag
	template, err := s.requireTemplate()
	if err != nil {
		return err
	}
	parent, err := s.dcTTName(template)
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1TagTemplateField{}
	if err := loadYAMLOrJSONInto(s.flagDCTTFieldConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TagTemplates.Fields.Create(parent, body).TagTemplateFieldId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating tag template field: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created field [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDCTTFormat)
}

func (s *dataCatalogCmds) runDCTTFieldDelete(cmd *cobra.Command, args []string) error {
	template, err := s.requireTemplate()
	if err != nil {
		return err
	}
	name, err := s.dcTTFieldName(template, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.TagTemplates.Fields.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting field: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted field [%s].\n", args[0])
	return nil
}

func (s *dataCatalogCmds) runDCTTFieldUpdate(cmd *cobra.Command, args []string) error {
	template, err := s.requireTemplate()
	if err != nil {
		return err
	}
	name, err := s.dcTTFieldName(template, args[0])
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1TagTemplateField{}
	if err := loadYAMLOrJSONInto(s.flagDCTTFieldConfigFile, body); err != nil {
		return err
	}
	mask := s.flagDCTTFieldUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.TagTemplates.Fields.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating field: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated field [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDCTTFormat)
}

func (s *dataCatalogCmds) runDCTTFieldRename(cmd *cobra.Command, args []string) error {
	template, err := s.requireTemplate()
	if err != nil {
		return err
	}
	name, err := s.dcTTFieldName(template, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TagTemplates.Fields.Rename(name, &datacatalog.GoogleCloudDatacatalogV1RenameTagTemplateFieldRequest{
		NewTagTemplateFieldId: s.flagDCTTFieldNewID,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("renaming field: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Renamed field [%s] to [%s].\n", args[0], s.flagDCTTFieldNewID)
	return emitFormattedTo(s.inv.stdout, got, s.flagDCTTFormat)
}

func (s *dataCatalogCmds) runDCTTFieldRenameEnum(cmd *cobra.Command, args []string) error {
	template, err := s.requireTemplate()
	if err != nil {
		return err
	}
	field, err := s.dcTTFieldName(template, args[0])
	if err != nil {
		return err
	}
	enumName := dcChild("enumValues", args[1], field)
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TagTemplates.Fields.EnumValues.Rename(enumName, &datacatalog.GoogleCloudDatacatalogV1RenameTagTemplateFieldEnumValueRequest{
		NewEnumValueDisplayName: s.flagDCTTFieldEnumRename,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("renaming enum value: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Renamed enum [%s] to [%s].\n", args[1], s.flagDCTTFieldEnumRename)
	return emitFormattedTo(s.inv.stdout, got, s.flagDCTTFormat)
}

func (s *dataCatalogCmds) requireTemplate() (string, error) {
	if s.flagDCTTFieldTemplate == "" {
		return "", fmt.Errorf("--template is required")
	}
	return s.flagDCTTFieldTemplate, nil
}
