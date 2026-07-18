package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	datacatalog "google.golang.org/api/datacatalog/v1"
)

// --- gcloud data-catalog entry-groups (#1505) ---

var dcEntryGroupsCmd = &cobra.Command{Use: "entry-groups", Short: "Manage Data Catalog entry groups"}

var (
	flagDCEGLocation    string
	flagDCEGFormat      string
	flagDCEGConfigFile  string
	flagDCEGUpdateMask  string
	flagDCEGPageSize    int64
	flagDCEGIamMember   string
	flagDCEGIamRole     string
	flagDCEGIamCondExpr string
	flagDCEGIamCondT    string
	flagDCEGIamCondD    string
	flagDCEGIamAllCond  bool
)

var (
	dcEGCreateCmd = &cobra.Command{
		Use: "create ENTRY_GROUP", Short: "Create an entry group",
		Args: cobra.ExactArgs(1), RunE: runDCEGCreate,
	}
	dcEGDeleteCmd = &cobra.Command{
		Use: "delete ENTRY_GROUP", Short: "Delete an entry group",
		Args: cobra.ExactArgs(1), RunE: runDCEGDelete,
	}
	dcEGDescribeCmd = &cobra.Command{
		Use: "describe ENTRY_GROUP", Short: "Describe an entry group",
		Args: cobra.ExactArgs(1), RunE: runDCEGDescribe,
	}
	dcEGListCmd = &cobra.Command{
		Use: "list", Short: "List entry groups",
		Args: cobra.NoArgs, RunE: runDCEGList,
	}
	dcEGUpdateCmd = &cobra.Command{
		Use: "update ENTRY_GROUP", Short: "Update an entry group",
		Args: cobra.ExactArgs(1), RunE: runDCEGUpdate,
	}
	dcEGGetIamCmd = &cobra.Command{
		Use: "get-iam-policy ENTRY_GROUP", Short: "Get the IAM policy for an entry group",
		Args: cobra.ExactArgs(1), RunE: runDCEGGetIam,
	}
	dcEGSetIamCmd = &cobra.Command{
		Use: "set-iam-policy ENTRY_GROUP POLICY_FILE", Short: "Set the IAM policy for an entry group",
		Args: cobra.ExactArgs(2), RunE: runDCEGSetIam,
	}
	dcEGAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding ENTRY_GROUP", Short: "Add an IAM policy binding to an entry group",
		Args: cobra.ExactArgs(1), RunE: runDCEGAddIam,
	}
	dcEGRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding ENTRY_GROUP", Short: "Remove an IAM policy binding from an entry group",
		Args: cobra.ExactArgs(1), RunE: runDCEGRemoveIam,
	}
)

func init() {
	all := []*cobra.Command{
		dcEGCreateCmd, dcEGDeleteCmd, dcEGDescribeCmd, dcEGListCmd, dcEGUpdateCmd,
		dcEGGetIamCmd, dcEGSetIamCmd, dcEGAddIamCmd, dcEGRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDCEGLocation, "location", "", "Location for the entry group (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagDCEGFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{dcEGCreateCmd, dcEGUpdateCmd} {
		c.Flags().StringVar(&flagDCEGConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the EntryGroup body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	dcEGUpdateCmd.Flags().StringVar(&flagDCEGUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	dcEGListCmd.Flags().Int64Var(&flagDCEGPageSize, "page-size", 0, "Maximum results per page")

	for _, c := range []*cobra.Command{dcEGAddIamCmd, dcEGRemoveIamCmd} {
		dcIamMemberFlags(c, &flagDCEGIamMember, &flagDCEGIamRole, &flagDCEGIamCondExpr, &flagDCEGIamCondT, &flagDCEGIamCondD)
	}
	dcEGRemoveIamCmd.Flags().BoolVar(&flagDCEGIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	dcEntryGroupsCmd.AddCommand(all...)
	dataCatalogCmd.AddCommand(dcEntryGroupsCmd)
}

func dcEGParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return dcLocationParent(project, flagDCEGLocation), nil
}

func dcEGName(id string) (string, error) {
	parent, err := dcEGParent()
	if err != nil {
		return "", err
	}
	return dcChild("entryGroups", id, parent), nil
}

func runDCEGCreate(cmd *cobra.Command, args []string) error {
	parent, err := dcEGParent()
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1EntryGroup{}
	if err := loadYAMLOrJSONInto(flagDCEGConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.EntryGroups.Create(parent, body).EntryGroupId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating entry group: %w", err)
	}
	fmt.Printf("Created entry group [%s].\n", args[0])
	return emitFormatted(got, flagDCEGFormat)
}

func runDCEGDelete(cmd *cobra.Command, args []string) error {
	name, err := dcEGName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.EntryGroups.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting entry group: %w", err)
	}
	fmt.Printf("Deleted entry group [%s].\n", args[0])
	return nil
}

func runDCEGDescribe(cmd *cobra.Command, args []string) error {
	name, err := dcEGName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.EntryGroups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing entry group: %w", err)
	}
	return emitFormatted(got, flagDCEGFormat)
}

func runDCEGList(cmd *cobra.Command, args []string) error {
	parent, err := dcEGParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*datacatalog.GoogleCloudDatacatalogV1EntryGroup
	pageToken := ""
	for {
		call := svc.Projects.Locations.EntryGroups.List(parent).Context(ctx)
		if flagDCEGPageSize > 0 {
			call = call.PageSize(flagDCEGPageSize)
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
	return emitFormatted(all, flagDCEGFormat)
}

func runDCEGUpdate(cmd *cobra.Command, args []string) error {
	name, err := dcEGName(args[0])
	if err != nil {
		return err
	}
	body := &datacatalog.GoogleCloudDatacatalogV1EntryGroup{}
	if err := loadYAMLOrJSONInto(flagDCEGConfigFile, body); err != nil {
		return err
	}
	mask := flagDCEGUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
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
	fmt.Printf("Updated entry group [%s].\n", args[0])
	return emitFormatted(got, flagDCEGFormat)
}

func runDCEGGetIam(cmd *cobra.Command, args []string) error {
	name, err := dcEGName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.EntryGroups.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagDCEGFormat)
}

func runDCEGSetIam(cmd *cobra.Command, args []string) error {
	name, err := dcEGName(args[0])
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
	updated, err := svc.Projects.Locations.EntryGroups.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	dcUpdatedIam(fmt.Sprintf("entry group [%s]", args[0]))
	return emitFormatted(updated, flagDCEGFormat)
}

func runDCEGAddIam(cmd *cobra.Command, args []string) error {
	name, err := dcEGName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.EntryGroups.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	dcAddBinding(policy, flagDCEGIamRole, flagDCEGIamMember,
		dcBuildCondition(flagDCEGIamCondExpr, flagDCEGIamCondT, flagDCEGIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.EntryGroups.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	dcUpdatedIam(fmt.Sprintf("entry group [%s]", args[0]))
	return emitFormatted(updated, flagDCEGFormat)
}

func runDCEGRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := dcEGName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.EntryGroups.GetIamPolicy(name, &datacatalog.GetIamPolicyRequest{
		Options: &datacatalog.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !dcRemoveBinding(policy, flagDCEGIamRole, flagDCEGIamMember,
		dcBuildCondition(flagDCEGIamCondExpr, flagDCEGIamCondT, flagDCEGIamCondD), flagDCEGIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagDCEGIamRole, flagDCEGIamMember)
	}
	updated, err := svc.Projects.Locations.EntryGroups.SetIamPolicy(name, &datacatalog.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	dcUpdatedIam(fmt.Sprintf("entry group [%s]", args[0]))
	return emitFormatted(updated, flagDCEGFormat)
}
