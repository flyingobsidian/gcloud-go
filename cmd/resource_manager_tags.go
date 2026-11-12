package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	crm "google.golang.org/api/cloudresourcemanager/v3"
)

// tagsFlags holds the flag values of the commands built by newTagsCmd.
type tagsFlags struct {
	tagKeyParent            string
	tagKeyDescription       string
	tagKeyListParent        string
	tagKeyListPageSize      int64
	tagKeyListLimit         int64
	tagKeyListFormat        string
	tagValueParent          string
	tagValueDescription     string
	tagValueListParent      string
	tagValueListPageSize    int64
	tagValueListLimit       int64
	tagValueListFormat      string
	tagBindingParent        string
	tagBindingTagValue      string
	tagBindingTagValueNS    string
	tagBindingLocation      string
	tagBindingListParent    string
	tagBindingListEffective bool
	tagBindingListLocation  string
	tagBindingListPageSize  int64
	tagBindingListLimit     int64
	tagBindingListFormat    string
}

func newTagsCmd(inv *invocation) *cobra.Command {
	f := &tagsFlags{}

	tagsCmd := &cobra.Command{
		Use:   "tags",
		Short: "Create and manipulate tag keys, values, and bindings",
	}

	tagKeysCmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage tag keys",
	}

	tagKeyCreateCmd := &cobra.Command{
		Use:   "create SHORT_NAME",
		Short: "Create a tag key",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runTagKeyCreate(c.Context(), inv, f, c, args)
		},
	}

	tagKeyDeleteCmd := &cobra.Command{
		Use:   "delete TAG_KEY",
		Short: "Delete a tag key",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runTagKeyDelete(c.Context(), inv, f, c, args)
		},
	}

	tagKeyDescribeCmd := &cobra.Command{
		Use:   "describe TAG_KEY",
		Short: "Describe a tag key",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runTagKeyDescribe(c.Context(), inv, f, c, args)
		},
	}

	tagKeyListCmd := &cobra.Command{
		Use:   "list",
		Short: "List tag keys under a parent",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return runTagKeyList(c.Context(), inv, f, c, args)
		},
	}

	tagKeyUpdateCmd := &cobra.Command{
		Use:   "update TAG_KEY",
		Short: "Update a tag key's description",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runTagKeyUpdate(c.Context(), inv, f, c, args)
		},
	}

	tagValuesCmd := &cobra.Command{
		Use:   "values",
		Short: "Manage tag values",
	}

	tagValueCreateCmd := &cobra.Command{
		Use:   "create SHORT_NAME",
		Short: "Create a tag value",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runTagValueCreate(c.Context(), inv, f, c, args)
		},
	}

	tagValueDeleteCmd := &cobra.Command{
		Use:   "delete TAG_VALUE",
		Short: "Delete a tag value",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runTagValueDelete(c.Context(), inv, f, c, args)
		},
	}

	tagValueDescribeCmd := &cobra.Command{
		Use:   "describe TAG_VALUE",
		Short: "Describe a tag value",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runTagValueDescribe(c.Context(), inv, f, c, args)
		},
	}

	tagValueListCmd := &cobra.Command{
		Use:   "list",
		Short: "List tag values under a parent tag key",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return runTagValueList(c.Context(), inv, f, c, args)
		},
	}

	tagValueUpdateCmd := &cobra.Command{
		Use:   "update TAG_VALUE",
		Short: "Update a tag value's description",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runTagValueUpdate(c.Context(), inv, f, c, args)
		},
	}

	tagBindingsCmd := &cobra.Command{
		Use:   "bindings",
		Short: "Manage tag bindings",
	}

	tagBindingCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a tag binding",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return runTagBindingCreate(c.Context(), inv, f, c, args)
		},
	}

	tagBindingDeleteCmd := &cobra.Command{
		Use:   "delete BINDING_NAME",
		Short: "Delete a tag binding",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runTagBindingDelete(c.Context(), inv, f, c, args)
		},
	}

	tagBindingListCmd := &cobra.Command{
		Use:   "list",
		Short: "List tag bindings on a resource",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return runTagBindingList(c.Context(), inv, f, c, args)
		},
	}
	tagKeyCreateCmd.Flags().StringVar(&f.tagKeyParent, "parent", "", "Parent (e.g. organizations/123 or projects/my-project) (required)")
	tagKeyCreateCmd.Flags().StringVar(&f.tagKeyDescription, "description", "", "Description for the tag key")
	tagKeyCreateCmd.MarkFlagRequired("parent")

	tagKeyListCmd.Flags().StringVar(&f.tagKeyListParent, "parent", "", "Parent to list tag keys under (required)")
	tagKeyListCmd.Flags().Int64Var(&f.tagKeyListPageSize, "page-size", 0, "Page size for API pagination")
	tagKeyListCmd.Flags().Int64Var(&f.tagKeyListLimit, "limit", 0, "Maximum number of tag keys to list (0 = no limit)")
	tagKeyListCmd.Flags().StringVar(&f.tagKeyListFormat, "format", "", "Output format (json, yaml, or table)")
	tagKeyListCmd.MarkFlagRequired("parent")

	tagKeyUpdateCmd.Flags().StringVar(&f.tagKeyDescription, "description", "", "New description for the tag key (required)")
	tagKeyUpdateCmd.MarkFlagRequired("description")

	tagKeysCmd.AddCommand(tagKeyCreateCmd, tagKeyDeleteCmd, tagKeyDescribeCmd, tagKeyListCmd, tagKeyUpdateCmd)

	tagValueCreateCmd.Flags().StringVar(&f.tagValueParent, "parent", "", "Parent tag key (e.g. tagKeys/123) (required)")
	tagValueCreateCmd.Flags().StringVar(&f.tagValueDescription, "description", "", "Description for the tag value")
	tagValueCreateCmd.MarkFlagRequired("parent")

	tagValueListCmd.Flags().StringVar(&f.tagValueListParent, "parent", "", "Parent tag key to list values under (required)")
	tagValueListCmd.Flags().Int64Var(&f.tagValueListPageSize, "page-size", 0, "Page size for API pagination")
	tagValueListCmd.Flags().Int64Var(&f.tagValueListLimit, "limit", 0, "Maximum number of tag values to list (0 = no limit)")
	tagValueListCmd.Flags().StringVar(&f.tagValueListFormat, "format", "", "Output format (json, yaml, or table)")
	tagValueListCmd.MarkFlagRequired("parent")

	tagValueUpdateCmd.Flags().StringVar(&f.tagValueDescription, "description", "", "New description for the tag value (required)")
	tagValueUpdateCmd.MarkFlagRequired("description")

	tagValuesCmd.AddCommand(tagValueCreateCmd, tagValueDeleteCmd, tagValueDescribeCmd, tagValueListCmd, tagValueUpdateCmd)

	tagBindingCreateCmd.Flags().StringVar(&f.tagBindingParent, "parent", "", "Full resource name that the tag will be bound to, e.g. //cloudresourcemanager.googleapis.com/projects/123 (required)")
	tagBindingCreateCmd.Flags().StringVar(&f.tagBindingTagValue, "tag-value", "", "Tag value resource name (e.g. tagValues/456)")
	tagBindingCreateCmd.Flags().StringVar(&f.tagBindingTagValueNS, "tag-value-namespaced-name", "", "Tag value namespaced name (e.g. 12345/env/prod)")
	tagBindingCreateCmd.Flags().StringVar(&f.tagBindingLocation, "location", "", "Location for the binding (leave empty for global)")
	tagBindingCreateCmd.MarkFlagRequired("parent")

	tagBindingListCmd.Flags().StringVar(&f.tagBindingListParent, "parent", "", "Full resource name to list bindings for (required)")
	tagBindingListCmd.Flags().BoolVar(&f.tagBindingListEffective, "effective", false, "List effective tags (including inherited)")
	tagBindingListCmd.Flags().StringVar(&f.tagBindingListLocation, "location", "", "Location scope for the listing")
	tagBindingListCmd.Flags().Int64Var(&f.tagBindingListPageSize, "page-size", 0, "Page size for API pagination")
	tagBindingListCmd.Flags().Int64Var(&f.tagBindingListLimit, "limit", 0, "Maximum number of bindings to list (0 = no limit)")
	tagBindingListCmd.Flags().StringVar(&f.tagBindingListFormat, "format", "", "Output format (json, yaml, or table)")
	tagBindingListCmd.MarkFlagRequired("parent")

	tagBindingsCmd.AddCommand(tagBindingCreateCmd, tagBindingDeleteCmd, tagBindingListCmd)

	tagsCmd.AddCommand(tagKeysCmd, tagValuesCmd, tagBindingsCmd)
	return tagsCmd
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

func runTagKeyCreate(ctx context.Context, inv *invocation, f *tagsFlags, cmd *cobra.Command, args []string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.TagKeys.Create(&crm.TagKey{
		Parent:      f.tagKeyParent,
		ShortName:   args[0],
		Description: f.tagKeyDescription,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating tag key: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Create tag key in progress (operation: %s).\n", op.Name)
	return inv.yaml(op)
}

func runTagKeyDelete(ctx context.Context, inv *invocation, f *tagsFlags, cmd *cobra.Command, args []string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.TagKeys.Delete(tagKeyResourceName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting tag key: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Delete tag key in progress (operation: %s).\n", op.Name)
	return inv.yaml(op)
}

func runTagKeyDescribe(ctx context.Context, inv *invocation, f *tagsFlags, cmd *cobra.Command, args []string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
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
		return inv.yaml(key)
	}
	key, err := svc.TagKeys.GetNamespaced().Name(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing tag key: %w", err)
	}
	return inv.yaml(key)
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

func runTagKeyList(ctx context.Context, inv *invocation, f *tagsFlags, cmd *cobra.Command, args []string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}

	var all []*crm.TagKey
	pageToken := ""
	for {
		call := svc.TagKeys.List().Parent(f.tagKeyListParent).Context(ctx)
		if f.tagKeyListPageSize > 0 {
			call = call.PageSize(f.tagKeyListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing tag keys: %w", err)
		}
		all = append(all, resp.TagKeys...)
		if f.tagKeyListLimit > 0 && int64(len(all)) >= f.tagKeyListLimit {
			all = all[:f.tagKeyListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return inv.printListResults(all, f.tagKeyListFormat, func() { printTagKeyTable(inv.stdout, all) })
}

// printTagKeyTable writes the default `tags keys list` table to w.
func printTagKeyTable(w io.Writer, all []*crm.TagKey) {
	printTableHeadingTo(w, len(all), "%-40s %-30s %s\n", "NAME", "SHORT_NAME", "NAMESPACED_NAME")
	for _, k := range all {
		fmt.Fprintf(w, "%-40s %-30s %s\n", k.Name, k.ShortName, k.NamespacedName)
	}
}

func runTagKeyUpdate(ctx context.Context, inv *invocation, f *tagsFlags, cmd *cobra.Command, args []string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.TagKeys.Patch(tagKeyResourceName(args[0]), &crm.TagKey{
		Description: f.tagKeyDescription,
	}).UpdateMask("description").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating tag key: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Update tag key in progress (operation: %s).\n", op.Name)
	return inv.yaml(op)
}

func runTagValueCreate(ctx context.Context, inv *invocation, f *tagsFlags, cmd *cobra.Command, args []string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.TagValues.Create(&crm.TagValue{
		Parent:      tagKeyResourceName(f.tagValueParent),
		ShortName:   args[0],
		Description: f.tagValueDescription,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating tag value: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Create tag value in progress (operation: %s).\n", op.Name)
	return inv.yaml(op)
}

func runTagValueDelete(ctx context.Context, inv *invocation, f *tagsFlags, cmd *cobra.Command, args []string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.TagValues.Delete(tagValueResourceName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting tag value: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Delete tag value in progress (operation: %s).\n", op.Name)
	return inv.yaml(op)
}

func runTagValueDescribe(ctx context.Context, inv *invocation, f *tagsFlags, cmd *cobra.Command, args []string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}
	name := args[0]
	if strings.HasPrefix(name, "tagValues/") || isAllDigits(strings.TrimPrefix(name, "tagValues/")) {
		value, err := svc.TagValues.Get(tagValueResourceName(name)).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("describing tag value: %w", err)
		}
		return inv.yaml(value)
	}
	value, err := svc.TagValues.GetNamespaced().Name(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing tag value: %w", err)
	}
	return inv.yaml(value)
}

func runTagValueList(ctx context.Context, inv *invocation, f *tagsFlags, cmd *cobra.Command, args []string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}

	var all []*crm.TagValue
	pageToken := ""
	for {
		call := svc.TagValues.List().Parent(tagKeyResourceName(f.tagValueListParent)).Context(ctx)
		if f.tagValueListPageSize > 0 {
			call = call.PageSize(f.tagValueListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing tag values: %w", err)
		}
		all = append(all, resp.TagValues...)
		if f.tagValueListLimit > 0 && int64(len(all)) >= f.tagValueListLimit {
			all = all[:f.tagValueListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return inv.printListResults(all, f.tagValueListFormat, func() {
		printTableHeadingTo(inv.stdout, len(all), "%-40s %-30s %s\n", "NAME", "SHORT_NAME", "NAMESPACED_NAME")
		for _, v := range all {
			fmt.Fprintf(inv.stdout, "%-40s %-30s %s\n", v.Name, v.ShortName, v.NamespacedName)
		}
	})
}

func runTagValueUpdate(ctx context.Context, inv *invocation, f *tagsFlags, cmd *cobra.Command, args []string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.TagValues.Patch(tagValueResourceName(args[0]), &crm.TagValue{
		Description: f.tagValueDescription,
	}).UpdateMask("description").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating tag value: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Update tag value in progress (operation: %s).\n", op.Name)
	return inv.yaml(op)
}

// bindingSvc returns a CRM service configured for the given location, either
// the global endpoint or the regional one.
func bindingSvc(ctx context.Context, inv *invocation, location string) (*crm.Service, error) {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return nil, err
	}
	if ep := regionalCRMEndpoint(location); ep != "" {
		svc.BasePath = ep
	}
	return svc, nil
}

func runTagBindingCreate(ctx context.Context, inv *invocation, f *tagsFlags, cmd *cobra.Command, args []string) error {
	if f.tagBindingTagValue == "" && f.tagBindingTagValueNS == "" {
		return fmt.Errorf("one of --tag-value or --tag-value-namespaced-name is required")
	}
	if f.tagBindingTagValue != "" && f.tagBindingTagValueNS != "" {
		return fmt.Errorf("specify only one of --tag-value or --tag-value-namespaced-name")
	}
	svc, err := bindingSvc(ctx, inv, f.tagBindingLocation)
	if err != nil {
		return err
	}
	binding := &crm.TagBinding{Parent: f.tagBindingParent}
	if f.tagBindingTagValue != "" {
		binding.TagValue = tagValueResourceName(f.tagBindingTagValue)
	} else {
		binding.TagValueNamespacedName = f.tagBindingTagValueNS
	}
	op, err := svc.TagBindings.Create(binding).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating tag binding: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Create tag binding in progress (operation: %s).\n", op.Name)
	return inv.yaml(op)
}

func runTagBindingDelete(ctx context.Context, inv *invocation, f *tagsFlags, cmd *cobra.Command, args []string) error {
	svc, err := bindingSvc(ctx, inv, f.tagBindingLocation)
	if err != nil {
		return err
	}
	op, err := svc.TagBindings.Delete(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting tag binding: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Delete tag binding in progress (operation: %s).\n", op.Name)
	return inv.yaml(op)
}

func runTagBindingList(ctx context.Context, inv *invocation, f *tagsFlags, cmd *cobra.Command, args []string) error {
	svc, err := bindingSvc(ctx, inv, f.tagBindingListLocation)
	if err != nil {
		return err
	}

	parent := f.tagBindingListParent
	if f.tagBindingListEffective {
		return listEffectiveTags(ctx, inv, f, svc, parent)
	}

	var all []*crm.TagBinding
	pageToken := ""
	for {
		call := svc.TagBindings.List().Parent(parent).Context(ctx)
		if f.tagBindingListPageSize > 0 {
			call = call.PageSize(f.tagBindingListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing tag bindings: %w", err)
		}
		all = append(all, resp.TagBindings...)
		if f.tagBindingListLimit > 0 && int64(len(all)) >= f.tagBindingListLimit {
			all = all[:f.tagBindingListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return inv.printListResults(all, f.tagBindingListFormat, func() {
		printTableHeadingTo(inv.stdout, len(all), "%-60s %-25s %s\n", "NAME", "TAG_VALUE", "NAMESPACED_NAME")
		for _, b := range all {
			fmt.Fprintf(inv.stdout, "%-60s %-25s %s\n", b.Name, b.TagValue, b.TagValueNamespacedName)
		}
	})
}

func listEffectiveTags(ctx context.Context, inv *invocation, f *tagsFlags, svc *crm.Service, parent string) error {
	var all []*crm.EffectiveTag
	pageToken := ""
	for {
		call := svc.EffectiveTags.List().Parent(parent).Context(ctx)
		if f.tagBindingListPageSize > 0 {
			call = call.PageSize(f.tagBindingListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing effective tags: %w", err)
		}
		all = append(all, resp.EffectiveTags...)
		if f.tagBindingListLimit > 0 && int64(len(all)) >= f.tagBindingListLimit {
			all = all[:f.tagBindingListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return inv.printListResults(all, f.tagBindingListFormat, func() {
		printTableHeadingTo(inv.stdout, len(all), "%-40s %-25s %s\n", "NAMESPACED_TAG_KEY", "TAG_VALUE", "INHERITED")
		for _, t := range all {
			fmt.Fprintf(inv.stdout, "%-40s %-25s %v\n", t.NamespacedTagKey, t.TagValue, t.Inherited)
		}
	})
}

// printListResults writes results to stdout in the requested format,
// falling back to tableFn for the default table view.
func (inv *invocation) printListResults(v any, format string, tableFn func()) error {
	switch format {
	case "json":
		return inv.printJSON(v)
	case "yaml":
		return inv.yaml(v)
	}
	tableFn()
	return nil
}

// printListResults is the legacy form of inv.printListResults that writes to
// the process stdout, for groups not yet migrated (#2070).
func printListResults(v any, format string, tableFn func()) error {
	switch format {
	case "json":
		return printJSON(v)
	case "yaml":
		return yamlEncode(v)
	}
	tableFn()
	return nil
}
