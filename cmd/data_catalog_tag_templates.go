package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	datacatalog "google.golang.org/api/datacatalog/v1"
)

// --- gcloud data-catalog tag-templates (#1508) ---

var dcTTCmd = &cobra.Command{Use: "tag-templates", Short: "Manage Data Catalog tag templates"}

var (
	flagDCTTLocation   string
	flagDCTTFormat     string
	flagDCTTConfigFile string
	flagDCTTUpdateMask string
	flagDCTTForce      bool

	flagDCTTIamMember   string
	flagDCTTIamRole     string
	flagDCTTIamCondExpr string
	flagDCTTIamCondT    string
	flagDCTTIamCondD    string
	flagDCTTIamAllCond  bool
)

var (
	dcTTCreateCmd = &cobra.Command{
		Use: "create TAG_TEMPLATE", Short: "Create a tag template",
		Args: cobra.ExactArgs(1), RunE: runDCTTCreate,
	}
	dcTTDeleteCmd = &cobra.Command{
		Use: "delete TAG_TEMPLATE", Short: "Delete a tag template",
		Args: cobra.ExactArgs(1), RunE: runDCTTDelete,
	}
	dcTTDescribeCmd = &cobra.Command{
		Use: "describe TAG_TEMPLATE", Short: "Describe a tag template",
		Args: cobra.ExactArgs(1), RunE: runDCTTDescribe,
	}
	dcTTUpdateCmd = &cobra.Command{
		Use: "update TAG_TEMPLATE", Short: "Update a tag template",
		Args: cobra.ExactArgs(1), RunE: runDCTTUpdate,
	}
	dcTTGetIamCmd = &cobra.Command{
		Use: "get-iam-policy TAG_TEMPLATE", Short: "Get the IAM policy for a tag template",
		Args: cobra.ExactArgs(1), RunE: runDCTTGetIam,
	}
	dcTTSetIamCmd = &cobra.Command{
		Use: "set-iam-policy TAG_TEMPLATE POLICY_FILE", Short: "Set the IAM policy for a tag template",
		Args: cobra.ExactArgs(2), RunE: runDCTTSetIam,
	}
	dcTTAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding TAG_TEMPLATE", Short: "Add an IAM policy binding to a tag template",
		Args: cobra.ExactArgs(1), RunE: runDCTTAddIam,
	}
	dcTTRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding TAG_TEMPLATE", Short: "Remove an IAM policy binding from a tag template",
		Args: cobra.ExactArgs(1), RunE: runDCTTRemoveIam,
	}
)

func init() {
	all := []*cobra.Command{
		dcTTCreateCmd, dcTTDeleteCmd, dcTTDescribeCmd, dcTTUpdateCmd,
		dcTTGetIamCmd, dcTTSetIamCmd, dcTTAddIamCmd, dcTTRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDCTTLocation, "location", "", "Location for the tag template (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagDCTTFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{dcTTCreateCmd, dcTTUpdateCmd} {
		c.Flags().StringVar(&flagDCTTConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the TagTemplate body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	dcTTUpdateCmd.Flags().StringVar(&flagDCTTUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	dcTTDeleteCmd.Flags().BoolVar(&flagDCTTForce, "force", false, "Force delete even if the template has tags")

	for _, c := range []*cobra.Command{dcTTAddIamCmd, dcTTRemoveIamCmd} {
		dcIamMemberFlags(c, &flagDCTTIamMember, &flagDCTTIamRole, &flagDCTTIamCondExpr, &flagDCTTIamCondT, &flagDCTTIamCondD)
	}
	dcTTRemoveIamCmd.Flags().BoolVar(&flagDCTTIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	dcTTCmd.AddCommand(all...)
	dcTTCmd.AddCommand(dcTTFieldsCmd)
	dataCatalogCmd.AddCommand(dcTTCmd)
}

func dcTTParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return dcLocationParent(project, flagDCTTLocation), nil
}

func dcTTName(id string) (string, error) {
	parent, err := dcTTParent()
	if err != nil {
		return "", err
	}
	return dcChild("tagTemplates", id, parent), nil
}

func runDCTTCreate(cmd *cobra.Command, args []string) error {
	parent, err := dcTTParent()
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1TagTemplate{}
	if err := loadYAMLOrJSONInto(flagDCTTConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TagTemplates.Create(parent, body).TagTemplateId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating tag template: %w", err)
	}
	fmt.Printf("Created tag template [%s].\n", args[0])
	return emitFormatted(got, flagDCTTFormat)
}

func runDCTTDelete(cmd *cobra.Command, args []string) error {
	name, err := dcTTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.TagTemplates.Delete(name).Context(ctx)
	if flagDCTTForce {
		call = call.Force(true)
	}
	if _, err := call.Do(); err != nil {
		return fmt.Errorf("deleting tag template: %w", err)
	}
	fmt.Printf("Deleted tag template [%s].\n", args[0])
	return nil
}

func runDCTTDescribe(cmd *cobra.Command, args []string) error {
	name, err := dcTTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TagTemplates.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing tag template: %w", err)
	}
	return emitFormatted(got, flagDCTTFormat)
}

func runDCTTUpdate(cmd *cobra.Command, args []string) error {
	name, err := dcTTName(args[0])
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1TagTemplate{}
	if err := loadYAMLOrJSONInto(flagDCTTConfigFile, body); err != nil {
		return err
	}
	mask := flagDCTTUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
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
	fmt.Printf("Updated tag template [%s].\n", args[0])
	return emitFormatted(got, flagDCTTFormat)
}

func runDCTTGetIam(cmd *cobra.Command, args []string) error {
	name, err := dcTTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.TagTemplates.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagDCTTFormat)
}

func runDCTTSetIam(cmd *cobra.Command, args []string) error {
	name, err := dcTTName(args[0])
	if err != nil {
		return err
	}
	policy := &datacatalog.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.TagTemplates.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	dcUpdatedIam(fmt.Sprintf("tag template [%s]", args[0]))
	return emitFormatted(updated, flagDCTTFormat)
}

func runDCTTAddIam(cmd *cobra.Command, args []string) error {
	name, err := dcTTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.TagTemplates.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	dcAddBinding(policy, flagDCTTIamRole, flagDCTTIamMember,
		dcBuildCondition(flagDCTTIamCondExpr, flagDCTTIamCondT, flagDCTTIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.TagTemplates.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	dcUpdatedIam(fmt.Sprintf("tag template [%s]", args[0]))
	return emitFormatted(updated, flagDCTTFormat)
}

func runDCTTRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := dcTTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.TagTemplates.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !dcRemoveBinding(policy, flagDCTTIamRole, flagDCTTIamMember,
		dcBuildCondition(flagDCTTIamCondExpr, flagDCTTIamCondT, flagDCTTIamCondD), flagDCTTIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagDCTTIamRole, flagDCTTIamMember)
	}
	updated, err := svc.Projects.Locations.TagTemplates.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	dcUpdatedIam(fmt.Sprintf("tag template [%s]", args[0]))
	return emitFormatted(updated, flagDCTTFormat)
}

// --- tag-templates fields subgroup ---

var dcTTFieldsCmd = &cobra.Command{Use: "fields", Short: "Manage tag template fields"}

var (
	flagDCTTFieldConfigFile string
	flagDCTTFieldUpdateMask string
	flagDCTTFieldNewID      string
	flagDCTTFieldEnumRename string
)

var (
	dcTTFieldCreateCmd = &cobra.Command{
		Use: "create FIELD", Short: "Create a tag template field",
		Args: cobra.ExactArgs(1), RunE: runDCTTFieldCreate,
	}
	dcTTFieldDeleteCmd = &cobra.Command{
		Use: "delete FIELD", Short: "Delete a tag template field",
		Args: cobra.ExactArgs(1), RunE: runDCTTFieldDelete,
	}
	dcTTFieldUpdateCmd = &cobra.Command{
		Use: "update FIELD", Short: "Update a tag template field",
		Args: cobra.ExactArgs(1), RunE: runDCTTFieldUpdate,
	}
	dcTTFieldRenameCmd = &cobra.Command{
		Use: "rename FIELD", Short: "Rename a tag template field ID",
		Args: cobra.ExactArgs(1), RunE: runDCTTFieldRename,
	}
	dcTTFieldRenameEnumCmd = &cobra.Command{
		Use: "rename-enum-value FIELD ENUM_VALUE", Short: "Rename an enum value on a tag template field",
		Args: cobra.ExactArgs(2), RunE: runDCTTFieldRenameEnum,
	}
)

func init() {
	all := []*cobra.Command{dcTTFieldCreateCmd, dcTTFieldDeleteCmd, dcTTFieldUpdateCmd, dcTTFieldRenameCmd, dcTTFieldRenameEnumCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagDCTTLocation, "location", "", "Location for the tag template (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagDCTTFieldTemplate, "template", "", "Tag template ID that owns the field (required)")
		_ = c.MarkFlagRequired("template")
		c.Flags().StringVar(&flagDCTTFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{dcTTFieldCreateCmd, dcTTFieldUpdateCmd} {
		c.Flags().StringVar(&flagDCTTFieldConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the TagTemplateField body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	dcTTFieldUpdateCmd.Flags().StringVar(&flagDCTTFieldUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	dcTTFieldRenameCmd.Flags().StringVar(&flagDCTTFieldNewID, "new-id", "", "New tag template field ID (required)")
	_ = dcTTFieldRenameCmd.MarkFlagRequired("new-id")
	dcTTFieldRenameEnumCmd.Flags().StringVar(&flagDCTTFieldEnumRename, "new-display-name", "",
		"New display name for the enum value (required)")
	_ = dcTTFieldRenameEnumCmd.MarkFlagRequired("new-display-name")

	dcTTFieldsCmd.AddCommand(all...)
}

func dcTTFieldName(template, field string) (string, error) {
	parent, err := dcTTName(template)
	if err != nil {
		return "", err
	}
	return dcChild("fields", field, parent), nil
}

func runDCTTFieldCreate(cmd *cobra.Command, args []string) error {
	// args[0] = FIELD id; require --template as flag
	template, err := requireTemplate()
	if err != nil {
		return err
	}
	parent, err := dcTTName(template)
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1TagTemplateField{}
	if err := loadYAMLOrJSONInto(flagDCTTFieldConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TagTemplates.Fields.Create(parent, body).TagTemplateFieldId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating tag template field: %w", err)
	}
	fmt.Printf("Created field [%s].\n", args[0])
	return emitFormatted(got, flagDCTTFormat)
}

func runDCTTFieldDelete(cmd *cobra.Command, args []string) error {
	template, err := requireTemplate()
	if err != nil {
		return err
	}
	name, err := dcTTFieldName(template, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.TagTemplates.Fields.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting field: %w", err)
	}
	fmt.Printf("Deleted field [%s].\n", args[0])
	return nil
}

func runDCTTFieldUpdate(cmd *cobra.Command, args []string) error {
	template, err := requireTemplate()
	if err != nil {
		return err
	}
	name, err := dcTTFieldName(template, args[0])
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1TagTemplateField{}
	if err := loadYAMLOrJSONInto(flagDCTTFieldConfigFile, body); err != nil {
		return err
	}
	mask := flagDCTTFieldUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
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
	fmt.Printf("Updated field [%s].\n", args[0])
	return emitFormatted(got, flagDCTTFormat)
}

func runDCTTFieldRename(cmd *cobra.Command, args []string) error {
	template, err := requireTemplate()
	if err != nil {
		return err
	}
	name, err := dcTTFieldName(template, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TagTemplates.Fields.Rename(name, &datacatalog.GoogleCloudDatacatalogV1RenameTagTemplateFieldRequest{
		NewTagTemplateFieldId: flagDCTTFieldNewID,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("renaming field: %w", err)
	}
	fmt.Printf("Renamed field [%s] to [%s].\n", args[0], flagDCTTFieldNewID)
	return emitFormatted(got, flagDCTTFormat)
}

func runDCTTFieldRenameEnum(cmd *cobra.Command, args []string) error {
	template, err := requireTemplate()
	if err != nil {
		return err
	}
	field, err := dcTTFieldName(template, args[0])
	if err != nil {
		return err
	}
	enumName := dcChild("enumValues", args[1], field)
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.TagTemplates.Fields.EnumValues.Rename(enumName, &datacatalog.GoogleCloudDatacatalogV1RenameTagTemplateFieldEnumValueRequest{
		NewEnumValueDisplayName: flagDCTTFieldEnumRename,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("renaming enum value: %w", err)
	}
	fmt.Printf("Renamed enum [%s] to [%s].\n", args[1], flagDCTTFieldEnumRename)
	return emitFormatted(got, flagDCTTFormat)
}

var flagDCTTFieldTemplate string

func requireTemplate() (string, error) {
	if flagDCTTFieldTemplate == "" {
		return "", fmt.Errorf("--template is required")
	}
	return flagDCTTFieldTemplate, nil
}
