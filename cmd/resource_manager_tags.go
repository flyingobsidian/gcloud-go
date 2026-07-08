package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	crm "google.golang.org/api/cloudresourcemanager/v3"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Create and manipulate tag keys, values, and bindings",
}

// --- tag keys ---

var tagKeysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Manage tag keys",
}

var tagKeyCreateCmd = &cobra.Command{
	Use:   "create SHORT_NAME",
	Short: "Create a tag key",
	Args:  cobra.ExactArgs(1),
	RunE:  runTagKeyCreate,
}

var tagKeyDeleteCmd = &cobra.Command{
	Use:   "delete TAG_KEY",
	Short: "Delete a tag key",
	Args:  cobra.ExactArgs(1),
	RunE:  runTagKeyDelete,
}

var tagKeyDescribeCmd = &cobra.Command{
	Use:   "describe TAG_KEY",
	Short: "Describe a tag key",
	Args:  cobra.ExactArgs(1),
	RunE:  runTagKeyDescribe,
}

var tagKeyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tag keys under a parent",
	Args:  cobra.NoArgs,
	RunE:  runTagKeyList,
}

var tagKeyUpdateCmd = &cobra.Command{
	Use:   "update TAG_KEY",
	Short: "Update a tag key's description",
	Args:  cobra.ExactArgs(1),
	RunE:  runTagKeyUpdate,
}

// --- tag values ---

var tagValuesCmd = &cobra.Command{
	Use:   "values",
	Short: "Manage tag values",
}

var tagValueCreateCmd = &cobra.Command{
	Use:   "create SHORT_NAME",
	Short: "Create a tag value",
	Args:  cobra.ExactArgs(1),
	RunE:  runTagValueCreate,
}

var tagValueDeleteCmd = &cobra.Command{
	Use:   "delete TAG_VALUE",
	Short: "Delete a tag value",
	Args:  cobra.ExactArgs(1),
	RunE:  runTagValueDelete,
}

var tagValueDescribeCmd = &cobra.Command{
	Use:   "describe TAG_VALUE",
	Short: "Describe a tag value",
	Args:  cobra.ExactArgs(1),
	RunE:  runTagValueDescribe,
}

var tagValueListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tag values under a parent tag key",
	Args:  cobra.NoArgs,
	RunE:  runTagValueList,
}

var tagValueUpdateCmd = &cobra.Command{
	Use:   "update TAG_VALUE",
	Short: "Update a tag value's description",
	Args:  cobra.ExactArgs(1),
	RunE:  runTagValueUpdate,
}

// --- tag bindings ---

var tagBindingsCmd = &cobra.Command{
	Use:   "bindings",
	Short: "Manage tag bindings",
}

var tagBindingCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a tag binding",
	Args:  cobra.NoArgs,
	RunE:  runTagBindingCreate,
}

var tagBindingDeleteCmd = &cobra.Command{
	Use:   "delete BINDING_NAME",
	Short: "Delete a tag binding",
	Args:  cobra.ExactArgs(1),
	RunE:  runTagBindingDelete,
}

var tagBindingListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tag bindings on a resource",
	Args:  cobra.NoArgs,
	RunE:  runTagBindingList,
}

var (
	flagTagKeyParent       string
	flagTagKeyDescription  string
	flagTagKeyListParent   string
	flagTagKeyListPageSize int64
	flagTagKeyListLimit    int64
	flagTagKeyListFormat   string

	flagTagValueParent       string
	flagTagValueDescription  string
	flagTagValueListParent   string
	flagTagValueListPageSize int64
	flagTagValueListLimit    int64
	flagTagValueListFormat   string

	flagTagBindingParent          string
	flagTagBindingTagValue        string
	flagTagBindingTagValueNS      string
	flagTagBindingLocation        string
	flagTagBindingListParent      string
	flagTagBindingListEffective   bool
	flagTagBindingListLocation    string
	flagTagBindingListPageSize    int64
	flagTagBindingListLimit       int64
	flagTagBindingListFormat      string
)

func init() {
	tagKeyCreateCmd.Flags().StringVar(&flagTagKeyParent, "parent", "", "Parent (e.g. organizations/123 or projects/my-project) (required)")
	tagKeyCreateCmd.Flags().StringVar(&flagTagKeyDescription, "description", "", "Description for the tag key")
	tagKeyCreateCmd.MarkFlagRequired("parent")

	tagKeyListCmd.Flags().StringVar(&flagTagKeyListParent, "parent", "", "Parent to list tag keys under (required)")
	tagKeyListCmd.Flags().Int64Var(&flagTagKeyListPageSize, "page-size", 0, "Page size for API pagination")
	tagKeyListCmd.Flags().Int64Var(&flagTagKeyListLimit, "limit", 0, "Maximum number of tag keys to list (0 = no limit)")
	tagKeyListCmd.Flags().StringVar(&flagTagKeyListFormat, "format", "", "Output format (json, yaml, or table)")
	tagKeyListCmd.MarkFlagRequired("parent")

	tagKeyUpdateCmd.Flags().StringVar(&flagTagKeyDescription, "description", "", "New description for the tag key (required)")
	tagKeyUpdateCmd.MarkFlagRequired("description")

	tagKeysCmd.AddCommand(tagKeyCreateCmd, tagKeyDeleteCmd, tagKeyDescribeCmd, tagKeyListCmd, tagKeyUpdateCmd)

	tagValueCreateCmd.Flags().StringVar(&flagTagValueParent, "parent", "", "Parent tag key (e.g. tagKeys/123) (required)")
	tagValueCreateCmd.Flags().StringVar(&flagTagValueDescription, "description", "", "Description for the tag value")
	tagValueCreateCmd.MarkFlagRequired("parent")

	tagValueListCmd.Flags().StringVar(&flagTagValueListParent, "parent", "", "Parent tag key to list values under (required)")
	tagValueListCmd.Flags().Int64Var(&flagTagValueListPageSize, "page-size", 0, "Page size for API pagination")
	tagValueListCmd.Flags().Int64Var(&flagTagValueListLimit, "limit", 0, "Maximum number of tag values to list (0 = no limit)")
	tagValueListCmd.Flags().StringVar(&flagTagValueListFormat, "format", "", "Output format (json, yaml, or table)")
	tagValueListCmd.MarkFlagRequired("parent")

	tagValueUpdateCmd.Flags().StringVar(&flagTagValueDescription, "description", "", "New description for the tag value (required)")
	tagValueUpdateCmd.MarkFlagRequired("description")

	tagValuesCmd.AddCommand(tagValueCreateCmd, tagValueDeleteCmd, tagValueDescribeCmd, tagValueListCmd, tagValueUpdateCmd)

	tagBindingCreateCmd.Flags().StringVar(&flagTagBindingParent, "parent", "", "Full resource name that the tag will be bound to, e.g. //cloudresourcemanager.googleapis.com/projects/123 (required)")
	tagBindingCreateCmd.Flags().StringVar(&flagTagBindingTagValue, "tag-value", "", "Tag value resource name (e.g. tagValues/456)")
	tagBindingCreateCmd.Flags().StringVar(&flagTagBindingTagValueNS, "tag-value-namespaced-name", "", "Tag value namespaced name (e.g. 12345/env/prod)")
	tagBindingCreateCmd.Flags().StringVar(&flagTagBindingLocation, "location", "", "Location for the binding (leave empty for global)")
	tagBindingCreateCmd.MarkFlagRequired("parent")

	tagBindingListCmd.Flags().StringVar(&flagTagBindingListParent, "parent", "", "Full resource name to list bindings for (required)")
	tagBindingListCmd.Flags().BoolVar(&flagTagBindingListEffective, "effective", false, "List effective tags (including inherited)")
	tagBindingListCmd.Flags().StringVar(&flagTagBindingListLocation, "location", "", "Location scope for the listing")
	tagBindingListCmd.Flags().Int64Var(&flagTagBindingListPageSize, "page-size", 0, "Page size for API pagination")
	tagBindingListCmd.Flags().Int64Var(&flagTagBindingListLimit, "limit", 0, "Maximum number of bindings to list (0 = no limit)")
	tagBindingListCmd.Flags().StringVar(&flagTagBindingListFormat, "format", "", "Output format (json, yaml, or table)")
	tagBindingListCmd.MarkFlagRequired("parent")

	tagBindingsCmd.AddCommand(tagBindingCreateCmd, tagBindingDeleteCmd, tagBindingListCmd)

	tagsCmd.AddCommand(tagKeysCmd, tagValuesCmd, tagBindingsCmd)
	resourceManagerCmd.AddCommand(tagsCmd)
}

// tagKeyResourceName normalizes a tag key identifier ("tagKeys/123" or "123").
func tagKeyResourceName(id string) string {
	id = strings.TrimPrefix(id, "tagKeys/")
	return "tagKeys/" + id
}

// tagValueResourceName normalizes a tag value identifier.
func tagValueResourceName(id string) string {
	id = strings.TrimPrefix(id, "tagValues/")
	return "tagValues/" + id
}

// regionalCRMEndpoint returns a service endpoint for the given location, or
// the empty string for the global endpoint. Tag bindings are location-scoped.
func regionalCRMEndpoint(location string) string {
	if location == "" {
		return ""
	}
	return fmt.Sprintf("https://%s-cloudresourcemanager.googleapis.com/", location)
}

func runTagKeyCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.TagKeys.Create(&crm.TagKey{
		Parent:      flagTagKeyParent,
		ShortName:   args[0],
		Description: flagTagKeyDescription,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating tag key: %w", err)
	}
	fmt.Printf("Create tag key in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runTagKeyDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.TagKeys.Delete(tagKeyResourceName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting tag key: %w", err)
	}
	fmt.Printf("Delete tag key in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runTagKeyDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	// Numeric IDs resolve via Get; namespaced names resolve via GetNamespaced.
	name := args[0]
	if strings.HasPrefix(name, "tagKeys/") || isAllDigits(strings.TrimPrefix(name, "tagKeys/")) {
		key, err := svc.TagKeys.Get(tagKeyResourceName(name)).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("describing tag key: %w", err)
		}
		return yamlEncode(key)
	}
	key, err := svc.TagKeys.GetNamespaced().Name(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing tag key: %w", err)
	}
	return yamlEncode(key)
}

// isAllDigits reports whether s is non-empty and contains only ASCII digits.
func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func runTagKeyList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var all []*crm.TagKey
	pageToken := ""
	for {
		call := svc.TagKeys.List().Parent(flagTagKeyListParent).Context(ctx)
		if flagTagKeyListPageSize > 0 {
			call = call.PageSize(flagTagKeyListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing tag keys: %w", err)
		}
		all = append(all, resp.TagKeys...)
		if flagTagKeyListLimit > 0 && int64(len(all)) >= flagTagKeyListLimit {
			all = all[:flagTagKeyListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return printListResults(all, flagTagKeyListFormat, func() {
		fmt.Printf("%-40s %-30s %s\n", "NAME", "SHORT_NAME", "NAMESPACED_NAME")
		for _, k := range all {
			fmt.Printf("%-40s %-30s %s\n", k.Name, k.ShortName, k.NamespacedName)
		}
	})
}

func runTagKeyUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.TagKeys.Patch(tagKeyResourceName(args[0]), &crm.TagKey{
		Description: flagTagKeyDescription,
	}).UpdateMask("description").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating tag key: %w", err)
	}
	fmt.Printf("Update tag key in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runTagValueCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.TagValues.Create(&crm.TagValue{
		Parent:      tagKeyResourceName(flagTagValueParent),
		ShortName:   args[0],
		Description: flagTagValueDescription,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating tag value: %w", err)
	}
	fmt.Printf("Create tag value in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runTagValueDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.TagValues.Delete(tagValueResourceName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting tag value: %w", err)
	}
	fmt.Printf("Delete tag value in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runTagValueDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := args[0]
	if strings.HasPrefix(name, "tagValues/") || isAllDigits(strings.TrimPrefix(name, "tagValues/")) {
		value, err := svc.TagValues.Get(tagValueResourceName(name)).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("describing tag value: %w", err)
		}
		return yamlEncode(value)
	}
	value, err := svc.TagValues.GetNamespaced().Name(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing tag value: %w", err)
	}
	return yamlEncode(value)
}

func runTagValueList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var all []*crm.TagValue
	pageToken := ""
	for {
		call := svc.TagValues.List().Parent(tagKeyResourceName(flagTagValueListParent)).Context(ctx)
		if flagTagValueListPageSize > 0 {
			call = call.PageSize(flagTagValueListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing tag values: %w", err)
		}
		all = append(all, resp.TagValues...)
		if flagTagValueListLimit > 0 && int64(len(all)) >= flagTagValueListLimit {
			all = all[:flagTagValueListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return printListResults(all, flagTagValueListFormat, func() {
		fmt.Printf("%-40s %-30s %s\n", "NAME", "SHORT_NAME", "NAMESPACED_NAME")
		for _, v := range all {
			fmt.Printf("%-40s %-30s %s\n", v.Name, v.ShortName, v.NamespacedName)
		}
	})
}

func runTagValueUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.TagValues.Patch(tagValueResourceName(args[0]), &crm.TagValue{
		Description: flagTagValueDescription,
	}).UpdateMask("description").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating tag value: %w", err)
	}
	fmt.Printf("Update tag value in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

// bindingSvc returns a CRM service configured for the given location, either
// the global endpoint or the regional one.
func bindingSvc(ctx context.Context, location string) (*crm.Service, error) {
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return nil, err
	}
	if ep := regionalCRMEndpoint(location); ep != "" {
		svc.BasePath = ep
	}
	return svc, nil
}

func runTagBindingCreate(cmd *cobra.Command, args []string) error {
	if flagTagBindingTagValue == "" && flagTagBindingTagValueNS == "" {
		return fmt.Errorf("one of --tag-value or --tag-value-namespaced-name is required")
	}
	if flagTagBindingTagValue != "" && flagTagBindingTagValueNS != "" {
		return fmt.Errorf("specify only one of --tag-value or --tag-value-namespaced-name")
	}
	ctx := context.Background()
	svc, err := bindingSvc(ctx, flagTagBindingLocation)
	if err != nil {
		return err
	}
	binding := &crm.TagBinding{Parent: flagTagBindingParent}
	if flagTagBindingTagValue != "" {
		binding.TagValue = tagValueResourceName(flagTagBindingTagValue)
	} else {
		binding.TagValueNamespacedName = flagTagBindingTagValueNS
	}
	op, err := svc.TagBindings.Create(binding).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating tag binding: %w", err)
	}
	fmt.Printf("Create tag binding in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runTagBindingDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := bindingSvc(ctx, flagTagBindingLocation)
	if err != nil {
		return err
	}
	op, err := svc.TagBindings.Delete(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting tag binding: %w", err)
	}
	fmt.Printf("Delete tag binding in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runTagBindingList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := bindingSvc(ctx, flagTagBindingListLocation)
	if err != nil {
		return err
	}

	parent := flagTagBindingListParent
	if flagTagBindingListEffective {
		return listEffectiveTags(ctx, svc, parent)
	}

	var all []*crm.TagBinding
	pageToken := ""
	for {
		call := svc.TagBindings.List().Parent(parent).Context(ctx)
		if flagTagBindingListPageSize > 0 {
			call = call.PageSize(flagTagBindingListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing tag bindings: %w", err)
		}
		all = append(all, resp.TagBindings...)
		if flagTagBindingListLimit > 0 && int64(len(all)) >= flagTagBindingListLimit {
			all = all[:flagTagBindingListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return printListResults(all, flagTagBindingListFormat, func() {
		fmt.Printf("%-60s %-25s %s\n", "NAME", "TAG_VALUE", "NAMESPACED_NAME")
		for _, b := range all {
			fmt.Printf("%-60s %-25s %s\n", b.Name, b.TagValue, b.TagValueNamespacedName)
		}
	})
}

func listEffectiveTags(ctx context.Context, svc *crm.Service, parent string) error {
	var all []*crm.EffectiveTag
	pageToken := ""
	for {
		call := svc.EffectiveTags.List().Parent(parent).Context(ctx)
		if flagTagBindingListPageSize > 0 {
			call = call.PageSize(flagTagBindingListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing effective tags: %w", err)
		}
		all = append(all, resp.EffectiveTags...)
		if flagTagBindingListLimit > 0 && int64(len(all)) >= flagTagBindingListLimit {
			all = all[:flagTagBindingListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return printListResults(all, flagTagBindingListFormat, func() {
		fmt.Printf("%-40s %-25s %s\n", "NAMESPACED_TAG_KEY", "TAG_VALUE", "INHERITED")
		for _, t := range all {
			fmt.Printf("%-40s %-25s %v\n", t.NamespacedTagKey, t.TagValue, t.Inherited)
		}
	})
}

// printListResults writes results in the requested format, falling back to
// tableFn for the default table view.
func printListResults(v any, format string, tableFn func()) error {
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	case "yaml":
		return yamlEncode(v)
	}
	tableFn()
	return nil
}
