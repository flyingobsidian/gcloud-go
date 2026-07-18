package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	datacatalog "google.golang.org/api/datacatalog/v1"
)

// --- gcloud data-catalog taxonomies (#1509) ---

var dcTaxCmd = &cobra.Command{Use: "taxonomies", Short: "Manage Data Catalog policy taxonomies"}

var (
	flagDCTaxLocation      string
	flagDCTaxFormat        string
	flagDCTaxConfigFile    string
	flagDCTaxUpdateMask    string
	flagDCTaxFilter        string
	flagDCTaxPageSize      int64
	flagDCTaxTaxIds        []string
	flagDCTaxSerialized    bool
	flagDCTaxImportFile    string
	flagDCTaxImportSource  string
	flagDCTaxCrossRegional string
	flagDCTaxIamMember     string
	flagDCTaxIamRole       string
	flagDCTaxIamCondExpr   string
	flagDCTaxIamCondT      string
	flagDCTaxIamCondD      string
	flagDCTaxIamAllCond    bool
)

var (
	dcTaxDescribeCmd = &cobra.Command{
		Use: "describe TAXONOMY", Short: "Describe a taxonomy",
		Args: cobra.ExactArgs(1), RunE: runDCTaxDescribe,
	}
	dcTaxListCmd = &cobra.Command{
		Use: "list", Short: "List taxonomies",
		Args: cobra.NoArgs, RunE: runDCTaxList,
	}
	dcTaxExportCmd = &cobra.Command{
		Use: "export", Short: "Export taxonomies (returns serialized JSON)",
		Args: cobra.NoArgs, RunE: runDCTaxExport,
	}
	dcTaxImportCmd = &cobra.Command{
		Use: "import", Short: "Import taxonomies (from a serialized JSON file or a cross-regional source)",
		Args: cobra.NoArgs, RunE: runDCTaxImport,
	}
	dcTaxGetIamCmd = &cobra.Command{
		Use: "get-iam-policy TAXONOMY", Short: "Get the IAM policy for a taxonomy",
		Args: cobra.ExactArgs(1), RunE: runDCTaxGetIam,
	}
	dcTaxSetIamCmd = &cobra.Command{
		Use: "set-iam-policy TAXONOMY POLICY_FILE", Short: "Set the IAM policy for a taxonomy",
		Args: cobra.ExactArgs(2), RunE: runDCTaxSetIam,
	}
	dcTaxAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding TAXONOMY", Short: "Add an IAM policy binding to a taxonomy",
		Args: cobra.ExactArgs(1), RunE: runDCTaxAddIam,
	}
	dcTaxRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding TAXONOMY", Short: "Remove an IAM policy binding from a taxonomy",
		Args: cobra.ExactArgs(1), RunE: runDCTaxRemoveIam,
	}
)

func init() {
	all := []*cobra.Command{
		dcTaxDescribeCmd, dcTaxListCmd, dcTaxExportCmd, dcTaxImportCmd,
		dcTaxGetIamCmd, dcTaxSetIamCmd, dcTaxAddIamCmd, dcTaxRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDCTaxLocation, "location", "", "Location for the taxonomy (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagDCTaxFormat, "format", "", "Output format")
	}
	dcTaxListCmd.Flags().StringVar(&flagDCTaxFilter, "filter", "", "Server-side filter expression")
	dcTaxListCmd.Flags().Int64Var(&flagDCTaxPageSize, "page-size", 0, "Maximum results per page")

	dcTaxExportCmd.Flags().StringSliceVar(&flagDCTaxTaxIds, "taxonomies", nil,
		"Taxonomy IDs to export (required, comma-separated or repeated)")
	_ = dcTaxExportCmd.MarkFlagRequired("taxonomies")
	dcTaxExportCmd.Flags().BoolVar(&flagDCTaxSerialized, "serialized", true,
		"Return serialized taxonomies (default true)")

	dcTaxImportCmd.Flags().StringVar(&flagDCTaxImportFile, "source", "",
		"Path to a JSON file with a SerializedTaxonomy list (mutually exclusive with --cross-regional-source)")
	dcTaxImportCmd.Flags().StringVar(&flagDCTaxCrossRegional, "cross-regional-source", "",
		"Fully-qualified source taxonomy resource name to import cross-regionally")

	for _, c := range []*cobra.Command{dcTaxAddIamCmd, dcTaxRemoveIamCmd} {
		dcIamMemberFlags(c, &flagDCTaxIamMember, &flagDCTaxIamRole, &flagDCTaxIamCondExpr, &flagDCTaxIamCondT, &flagDCTaxIamCondD)
	}
	dcTaxRemoveIamCmd.Flags().BoolVar(&flagDCTaxIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	dcTaxCmd.AddCommand(all...)
	dcTaxCmd.AddCommand(dcPTCmd)
	dataCatalogCmd.AddCommand(dcTaxCmd)
}

func dcTaxParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return dcLocationParent(project, flagDCTaxLocation), nil
}

func dcTaxName(id string) (string, error) {
	parent, err := dcTaxParent()
	if err != nil {
		return "", err
	}
	return dcChild("taxonomies", id, parent), nil
}

func runDCTaxDescribe(cmd *cobra.Command, args []string) error {
	name, err := dcTaxName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Taxonomies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing taxonomy: %w", err)
	}
	return emitFormatted(got, flagDCTaxFormat)
}

func runDCTaxList(cmd *cobra.Command, args []string) error {
	parent, err := dcTaxParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*datacatalog.GoogleCloudDatacatalogV1Taxonomy
	pageToken := ""
	for {
		call := svc.Projects.Locations.Taxonomies.List(parent).Context(ctx)
		if flagDCTaxFilter != "" {
			call = call.Filter(flagDCTaxFilter)
		}
		if flagDCTaxPageSize > 0 {
			call = call.PageSize(flagDCTaxPageSize)
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
	return emitFormatted(all, flagDCTaxFormat)
}

func runDCTaxExport(cmd *cobra.Command, args []string) error {
	parent, err := dcTaxParent()
	if err != nil {
		return err
	}
	// Build fully-qualified taxonomy names.
	names := make([]string, 0, len(flagDCTaxTaxIds))
	for _, id := range flagDCTaxTaxIds {
		if strings.HasPrefix(id, "projects/") {
			names = append(names, id)
		} else {
			names = append(names, dcChild("taxonomies", id, parent))
		}
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Taxonomies.Export(parent).
		Taxonomies(names...).SerializedTaxonomies(flagDCTaxSerialized).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting taxonomies: %w", err)
	}
	return emitFormatted(resp, flagDCTaxFormat)
}

func runDCTaxImport(cmd *cobra.Command, args []string) error {
	parent, err := dcTaxParent()
	if err != nil {
		return err
	}
	if (flagDCTaxImportFile == "" && flagDCTaxCrossRegional == "") ||
		(flagDCTaxImportFile != "" && flagDCTaxCrossRegional != "") {
		return fmt.Errorf("exactly one of --source or --cross-regional-source is required")
	}
	req := &datacatalog.GoogleCloudDatacatalogV1ImportTaxonomiesRequest{}
	if flagDCTaxImportFile != "" {
		var payload struct {
			Taxonomies []*datacatalog.GoogleCloudDatacatalogV1SerializedTaxonomy `json:"taxonomies"`
		}
		if err := loadYAMLOrJSONInto(flagDCTaxImportFile, &payload); err != nil {
			return err
		}
		if len(payload.Taxonomies) == 0 {
			return fmt.Errorf("no taxonomies in source file")
		}
		req.InlineSource = &datacatalog.GoogleCloudDatacatalogV1InlineSource{Taxonomies: payload.Taxonomies}
	} else {
		req.CrossRegionalSource = &datacatalog.GoogleCloudDatacatalogV1CrossRegionalSource{
			Taxonomy: flagDCTaxCrossRegional,
		}
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Taxonomies.Import(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("importing taxonomies: %w", err)
	}
	return emitFormatted(resp, flagDCTaxFormat)
}

func runDCTaxGetIam(cmd *cobra.Command, args []string) error {
	name, err := dcTaxName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Taxonomies.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagDCTaxFormat)
}

func runDCTaxSetIam(cmd *cobra.Command, args []string) error {
	name, err := dcTaxName(args[0])
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
	updated, err := svc.Projects.Locations.Taxonomies.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	dcUpdatedIam(fmt.Sprintf("taxonomy [%s]", args[0]))
	return emitFormatted(updated, flagDCTaxFormat)
}

func runDCTaxAddIam(cmd *cobra.Command, args []string) error {
	name, err := dcTaxName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Taxonomies.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	dcAddBinding(policy, flagDCTaxIamRole, flagDCTaxIamMember,
		dcBuildCondition(flagDCTaxIamCondExpr, flagDCTaxIamCondT, flagDCTaxIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.Taxonomies.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	dcUpdatedIam(fmt.Sprintf("taxonomy [%s]", args[0]))
	return emitFormatted(updated, flagDCTaxFormat)
}

func runDCTaxRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := dcTaxName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Taxonomies.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !dcRemoveBinding(policy, flagDCTaxIamRole, flagDCTaxIamMember,
		dcBuildCondition(flagDCTaxIamCondExpr, flagDCTaxIamCondT, flagDCTaxIamCondD), flagDCTaxIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagDCTaxIamRole, flagDCTaxIamMember)
	}
	updated, err := svc.Projects.Locations.Taxonomies.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	dcUpdatedIam(fmt.Sprintf("taxonomy [%s]", args[0]))
	return emitFormatted(updated, flagDCTaxFormat)
}

// --- policy-tags subgroup ---

var dcPTCmd = &cobra.Command{Use: "policy-tags", Short: "Manage taxonomy policy tags"}

var (
	flagDCPTTaxonomy   string
	flagDCPTConfigFile string
	flagDCPTUpdateMask string
	flagDCPTPageSize   int64

	flagDCPTIamMember   string
	flagDCPTIamRole     string
	flagDCPTIamCondExpr string
	flagDCPTIamCondT    string
	flagDCPTIamCondD    string
	flagDCPTIamAllCond  bool
)

var (
	dcPTCreateCmd = &cobra.Command{
		Use: "create POLICY_TAG", Short: "Create a policy tag",
		Args: cobra.ExactArgs(1), RunE: runDCPTCreate,
	}
	dcPTDeleteCmd = &cobra.Command{
		Use: "delete POLICY_TAG", Short: "Delete a policy tag",
		Args: cobra.ExactArgs(1), RunE: runDCPTDelete,
	}
	dcPTDescribeCmd = &cobra.Command{
		Use: "describe POLICY_TAG", Short: "Describe a policy tag",
		Args: cobra.ExactArgs(1), RunE: runDCPTDescribe,
	}
	dcPTListCmd = &cobra.Command{
		Use: "list", Short: "List policy tags",
		Args: cobra.NoArgs, RunE: runDCPTList,
	}
	dcPTUpdateCmd = &cobra.Command{
		Use: "update POLICY_TAG", Short: "Update a policy tag",
		Args: cobra.ExactArgs(1), RunE: runDCPTUpdate,
	}
	dcPTGetIamCmd = &cobra.Command{
		Use: "get-iam-policy POLICY_TAG", Short: "Get the IAM policy for a policy tag",
		Args: cobra.ExactArgs(1), RunE: runDCPTGetIam,
	}
	dcPTSetIamCmd = &cobra.Command{
		Use: "set-iam-policy POLICY_TAG POLICY_FILE", Short: "Set the IAM policy for a policy tag",
		Args: cobra.ExactArgs(2), RunE: runDCPTSetIam,
	}
	dcPTAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding POLICY_TAG", Short: "Add an IAM policy binding to a policy tag",
		Args: cobra.ExactArgs(1), RunE: runDCPTAddIam,
	}
	dcPTRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding POLICY_TAG", Short: "Remove an IAM policy binding from a policy tag",
		Args: cobra.ExactArgs(1), RunE: runDCPTRemoveIam,
	}
)

func init() {
	all := []*cobra.Command{
		dcPTCreateCmd, dcPTDeleteCmd, dcPTDescribeCmd, dcPTListCmd, dcPTUpdateCmd,
		dcPTGetIamCmd, dcPTSetIamCmd, dcPTAddIamCmd, dcPTRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDCTaxLocation, "location", "", "Location for the taxonomy (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagDCPTTaxonomy, "taxonomy", "", "Taxonomy ID that owns the policy tag (required)")
		_ = c.MarkFlagRequired("taxonomy")
		c.Flags().StringVar(&flagDCTaxFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{dcPTCreateCmd, dcPTUpdateCmd} {
		c.Flags().StringVar(&flagDCPTConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the PolicyTag body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	dcPTUpdateCmd.Flags().StringVar(&flagDCPTUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	dcPTListCmd.Flags().Int64Var(&flagDCPTPageSize, "page-size", 0, "Maximum results per page")

	for _, c := range []*cobra.Command{dcPTAddIamCmd, dcPTRemoveIamCmd} {
		dcIamMemberFlags(c, &flagDCPTIamMember, &flagDCPTIamRole, &flagDCPTIamCondExpr, &flagDCPTIamCondT, &flagDCPTIamCondD)
	}
	dcPTRemoveIamCmd.Flags().BoolVar(&flagDCPTIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	dcPTCmd.AddCommand(all...)
}

func dcPTParent() (string, error) {
	return dcTaxName(flagDCPTTaxonomy)
}

func dcPTName(id string) (string, error) {
	parent, err := dcPTParent()
	if err != nil {
		return "", err
	}
	return dcChild("policyTags", id, parent), nil
}

func runDCPTCreate(cmd *cobra.Command, args []string) error {
	parent, err := dcPTParent()
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1PolicyTag{}
	if err := loadYAMLOrJSONInto(flagDCPTConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Taxonomies.PolicyTags.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating policy tag: %w", err)
	}
	fmt.Printf("Created policy tag [%s].\n", got.Name)
	return emitFormatted(got, flagDCTaxFormat)
}

func runDCPTDelete(cmd *cobra.Command, args []string) error {
	name, err := dcPTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Taxonomies.PolicyTags.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting policy tag: %w", err)
	}
	fmt.Printf("Deleted policy tag [%s].\n", args[0])
	return nil
}

func runDCPTDescribe(cmd *cobra.Command, args []string) error {
	name, err := dcPTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Taxonomies.PolicyTags.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing policy tag: %w", err)
	}
	return emitFormatted(got, flagDCTaxFormat)
}

func runDCPTList(cmd *cobra.Command, args []string) error {
	parent, err := dcPTParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*datacatalog.GoogleCloudDatacatalogV1PolicyTag
	pageToken := ""
	for {
		call := svc.Projects.Locations.Taxonomies.PolicyTags.List(parent).Context(ctx)
		if flagDCPTPageSize > 0 {
			call = call.PageSize(flagDCPTPageSize)
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
	return emitFormatted(all, flagDCTaxFormat)
}

func runDCPTUpdate(cmd *cobra.Command, args []string) error {
	name, err := dcPTName(args[0])
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1PolicyTag{}
	if err := loadYAMLOrJSONInto(flagDCPTConfigFile, body); err != nil {
		return err
	}
	mask := flagDCPTUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
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
	fmt.Printf("Updated policy tag [%s].\n", args[0])
	return emitFormatted(got, flagDCTaxFormat)
}

func runDCPTGetIam(cmd *cobra.Command, args []string) error {
	name, err := dcPTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Taxonomies.PolicyTags.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagDCTaxFormat)
}

func runDCPTSetIam(cmd *cobra.Command, args []string) error {
	name, err := dcPTName(args[0])
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
	updated, err := svc.Projects.Locations.Taxonomies.PolicyTags.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	dcUpdatedIam(fmt.Sprintf("policy tag [%s]", args[0]))
	return emitFormatted(updated, flagDCTaxFormat)
}

func runDCPTAddIam(cmd *cobra.Command, args []string) error {
	name, err := dcPTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Taxonomies.PolicyTags.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	dcAddBinding(policy, flagDCPTIamRole, flagDCPTIamMember,
		dcBuildCondition(flagDCPTIamCondExpr, flagDCPTIamCondT, flagDCPTIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.Taxonomies.PolicyTags.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	dcUpdatedIam(fmt.Sprintf("policy tag [%s]", args[0]))
	return emitFormatted(updated, flagDCTaxFormat)
}

func runDCPTRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := dcPTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Taxonomies.PolicyTags.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !dcRemoveBinding(policy, flagDCPTIamRole, flagDCPTIamMember,
		dcBuildCondition(flagDCPTIamCondExpr, flagDCPTIamCondT, flagDCPTIamCondD), flagDCPTIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagDCPTIamRole, flagDCPTIamMember)
	}
	updated, err := svc.Projects.Locations.Taxonomies.PolicyTags.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	dcUpdatedIam(fmt.Sprintf("policy tag [%s]", args[0]))
	return emitFormatted(updated, flagDCTaxFormat)
}
