package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networksecurity "google.golang.org/api/networksecurity/v1"
)

// --- address-groups (project scope, issue #819) ---
// --- org-address-groups (organization scope, issue #834) ---

var networkSecurityAddressGroupsCmd = &cobra.Command{Use: "address-groups", Short: "Manage address groups"}
var networkSecurityOrgAddressGroupsCmd = &cobra.Command{Use: "org-address-groups", Short: "Manage organization address groups"}

// Project-scoped commands.
var (
	nsAGCreateCmd = &cobra.Command{
		Use: "create ADDRESS_GROUP", Short: "Create an address group from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runNSAGCreate,
	}
	nsAGDeleteCmd = &cobra.Command{
		Use: "delete ADDRESS_GROUP", Short: "Delete an address group",
		Args: cobra.ExactArgs(1), RunE: runNSAGDelete,
	}
	nsAGDescribeCmd = &cobra.Command{
		Use: "describe ADDRESS_GROUP", Short: "Describe an address group",
		Args: cobra.ExactArgs(1), RunE: runNSAGDescribe,
	}
	nsAGListCmd = &cobra.Command{
		Use: "list", Short: "List address groups in a location",
		Args: cobra.NoArgs, RunE: runNSAGList,
	}
	nsAGUpdateCmd = &cobra.Command{
		Use: "update ADDRESS_GROUP", Short: "Update an address group from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runNSAGUpdate,
	}
	nsAGAddItemsCmd = &cobra.Command{
		Use: "add-items ADDRESS_GROUP", Short: "Add items to an address group",
		Args: cobra.ExactArgs(1), RunE: runNSAGAddItems,
	}
	nsAGRemoveItemsCmd = &cobra.Command{
		Use: "remove-items ADDRESS_GROUP", Short: "Remove items from an address group",
		Args: cobra.ExactArgs(1), RunE: runNSAGRemoveItems,
	}
	nsAGCloneItemsCmd = &cobra.Command{
		Use: "clone-items ADDRESS_GROUP", Short: "Clone items into an address group",
		Args: cobra.ExactArgs(1), RunE: runNSAGCloneItems,
	}
	nsAGListRefsCmd = &cobra.Command{
		Use: "list-references ADDRESS_GROUP", Short: "List references to an address group",
		Args: cobra.ExactArgs(1), RunE: runNSAGListRefs,
	}
)

// Org-scoped commands.
var (
	nsOAGCreateCmd = &cobra.Command{
		Use: "create ADDRESS_GROUP", Short: "Create an organization address group from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runNSOAGCreate,
	}
	nsOAGDeleteCmd = &cobra.Command{
		Use: "delete ADDRESS_GROUP", Short: "Delete an organization address group",
		Args: cobra.ExactArgs(1), RunE: runNSOAGDelete,
	}
	nsOAGDescribeCmd = &cobra.Command{
		Use: "describe ADDRESS_GROUP", Short: "Describe an organization address group",
		Args: cobra.ExactArgs(1), RunE: runNSOAGDescribe,
	}
	nsOAGListCmd = &cobra.Command{
		Use: "list", Short: "List organization address groups",
		Args: cobra.NoArgs, RunE: runNSOAGList,
	}
	nsOAGUpdateCmd = &cobra.Command{
		Use: "update ADDRESS_GROUP", Short: "Update an organization address group",
		Args: cobra.ExactArgs(1), RunE: runNSOAGUpdate,
	}
	nsOAGAddItemsCmd = &cobra.Command{
		Use: "add-items ADDRESS_GROUP", Short: "Add items to an organization address group",
		Args: cobra.ExactArgs(1), RunE: runNSOAGAddItems,
	}
	nsOAGRemoveItemsCmd = &cobra.Command{
		Use: "remove-items ADDRESS_GROUP", Short: "Remove items from an organization address group",
		Args: cobra.ExactArgs(1), RunE: runNSOAGRemoveItems,
	}
	nsOAGCloneItemsCmd = &cobra.Command{
		Use: "clone-items ADDRESS_GROUP", Short: "Clone items into an organization address group",
		Args: cobra.ExactArgs(1), RunE: runNSOAGCloneItems,
	}
	nsOAGListRefsCmd = &cobra.Command{
		Use: "list-references ADDRESS_GROUP", Short: "List references to an organization address group",
		Args: cobra.ExactArgs(1), RunE: runNSOAGListRefs,
	}
)

func registerNSAddressGroups(root *cobra.Command) {
	// project-scoped
	projectCmds := []*cobra.Command{
		nsAGCreateCmd, nsAGDeleteCmd, nsAGDescribeCmd, nsAGListCmd, nsAGUpdateCmd,
		nsAGAddItemsCmd, nsAGRemoveItemsCmd, nsAGCloneItemsCmd, nsAGListRefsCmd,
	}
	addNSLocationFlag(projectCmds...)
	addNSFormatFlag(nsAGDescribeCmd, nsAGListCmd, nsAGListRefsCmd)
	addNSFilterFlag(nsAGListCmd, nsAGListRefsCmd)
	addNSCreateConfigFlag(nsAGCreateCmd)
	addNSUpdateConfigFlag(nsAGUpdateCmd)
	addNSAsyncFlag(nsAGCreateCmd, nsAGDeleteCmd, nsAGUpdateCmd, nsAGAddItemsCmd, nsAGRemoveItemsCmd, nsAGCloneItemsCmd)
	addNSRequestIDFlag(nsAGCreateCmd, nsAGDeleteCmd, nsAGUpdateCmd)
	for _, c := range []*cobra.Command{nsAGAddItemsCmd, nsAGRemoveItemsCmd} {
		c.Flags().StringSliceVar(&flagNSItems, "items", nil, "Comma-separated list of items to add/remove (required)")
		_ = c.MarkFlagRequired("items")
	}
	nsAGCloneItemsCmd.Flags().StringVar(&flagNSSourceAddressGroup, "source-address-group", "", "Fully qualified source address group to clone items from (required)")
	_ = nsAGCloneItemsCmd.MarkFlagRequired("source-address-group")
	nsAGListRefsCmd.Flags().Int64Var(&flagNSPageSize, "page-size", 0, "Maximum results per page (optional)")

	networkSecurityAddressGroupsCmd.AddCommand(projectCmds...)
	root.AddCommand(networkSecurityAddressGroupsCmd)

	// org-scoped
	orgCmds := []*cobra.Command{
		nsOAGCreateCmd, nsOAGDeleteCmd, nsOAGDescribeCmd, nsOAGListCmd, nsOAGUpdateCmd,
		nsOAGAddItemsCmd, nsOAGRemoveItemsCmd, nsOAGCloneItemsCmd, nsOAGListRefsCmd,
	}
	addNSOrgFlags(orgCmds...)
	addNSFormatFlag(nsOAGDescribeCmd, nsOAGListCmd, nsOAGListRefsCmd)
	addNSFilterFlag(nsOAGListCmd, nsOAGListRefsCmd)
	addNSCreateConfigFlag(nsOAGCreateCmd)
	addNSUpdateConfigFlag(nsOAGUpdateCmd)
	addNSAsyncFlag(nsOAGCreateCmd, nsOAGDeleteCmd, nsOAGUpdateCmd, nsOAGAddItemsCmd, nsOAGRemoveItemsCmd, nsOAGCloneItemsCmd)
	addNSRequestIDFlag(nsOAGCreateCmd, nsOAGDeleteCmd, nsOAGUpdateCmd)
	for _, c := range []*cobra.Command{nsOAGAddItemsCmd, nsOAGRemoveItemsCmd} {
		c.Flags().StringSliceVar(&flagNSItems, "items", nil, "Comma-separated list of items to add/remove (required)")
		_ = c.MarkFlagRequired("items")
	}
	nsOAGCloneItemsCmd.Flags().StringVar(&flagNSSourceAddressGroup, "source-address-group", "", "Fully qualified source address group to clone items from (required)")
	_ = nsOAGCloneItemsCmd.MarkFlagRequired("source-address-group")
	nsOAGListRefsCmd.Flags().Int64Var(&flagNSPageSize, "page-size", 0, "Maximum results per page (optional)")

	networkSecurityOrgAddressGroupsCmd.AddCommand(orgCmds...)
	root.AddCommand(networkSecurityOrgAddressGroupsCmd)
}

// --- project-scoped implementations ---

func runNSAGCreate(cmd *cobra.Command, args []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	body := &networksecurity.AddressGroup{}
	if err := loadYAMLOrJSONInto(flagNSConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.AddressGroups.Create(parent, body).AddressGroupId(args[0]).Context(ctx)
	if flagNSRequestID != "" {
		call = call.RequestId(flagNSRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating address group: %w", err)
	}
	return nsFinishOp(ctx, svc, op, "Create address group", args[0])
}

func runNSAGDelete(cmd *cobra.Command, args []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	call := svc.Projects.Locations.AddressGroups.Delete(name).Context(ctx)
	if flagNSRequestID != "" {
		call = call.RequestId(flagNSRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting address group: %w", err)
	}
	return nsFinishOp(ctx, svc, op, "Delete address group", args[0])
}

func runNSAGDescribe(cmd *cobra.Command, args []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	got, err := svc.Projects.Locations.AddressGroups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing address group: %w", err)
	}
	return emitFormatted(got, flagNSFormat)
}

func runNSAGList(cmd *cobra.Command, args []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networksecurity.AddressGroup
	pageToken := ""
	for {
		call := svc.Projects.Locations.AddressGroups.List(parent).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing address groups: %w", err)
		}
		all = append(all, resp.AddressGroups...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagNSFormat != "" {
		return emitFormatted(all, flagNSFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "TYPE")
	for _, g := range all {
		fmt.Printf("%-40s %s\n", nsBasename(g.Name), g.Type)
	}
	return nil
}

func runNSAGUpdate(cmd *cobra.Command, args []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	body := &networksecurity.AddressGroup{}
	if err := loadYAMLOrJSONInto(flagNSConfigFile, body); err != nil {
		return err
	}
	mask := nsResolveMask(body)
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	call := svc.Projects.Locations.AddressGroups.Patch(name, body).UpdateMask(mask).Context(ctx)
	if flagNSRequestID != "" {
		call = call.RequestId(flagNSRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating address group: %w", err)
	}
	return nsFinishOp(ctx, svc, op, "Update address group", args[0])
}

func runNSAGAddItems(cmd *cobra.Command, args []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	req := &networksecurity.AddAddressGroupItemsRequest{Items: flagNSItems, RequestId: flagNSRequestID}
	op, err := svc.Projects.Locations.AddressGroups.AddItems(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding items: %w", err)
	}
	return nsFinishOp(ctx, svc, op, "Add items", args[0])
}

func runNSAGRemoveItems(cmd *cobra.Command, args []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	req := &networksecurity.RemoveAddressGroupItemsRequest{Items: flagNSItems, RequestId: flagNSRequestID}
	op, err := svc.Projects.Locations.AddressGroups.RemoveItems(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing items: %w", err)
	}
	return nsFinishOp(ctx, svc, op, "Remove items", args[0])
}

func runNSAGCloneItems(cmd *cobra.Command, args []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	req := &networksecurity.CloneAddressGroupItemsRequest{
		SourceAddressGroup: flagNSSourceAddressGroup,
		RequestId:          flagNSRequestID,
	}
	op, err := svc.Projects.Locations.AddressGroups.CloneItems(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cloning items: %w", err)
	}
	return nsFinishOp(ctx, svc, op, "Clone items", args[0])
}

func runNSAGListRefs(cmd *cobra.Command, args []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	var all []*networksecurity.ListAddressGroupReferencesResponseAddressGroupReference
	pageToken := ""
	for {
		call := svc.Projects.Locations.AddressGroups.ListReferences(name).Context(ctx)
		if flagNSPageSize > 0 {
			call = call.PageSize(flagNSPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing references: %w", err)
		}
		all = append(all, resp.AddressGroupReferences...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNSFormat)
}

// --- org-scoped implementations ---

func runNSOAGCreate(cmd *cobra.Command, args []string) error {
	parent, err := nsOrgParent()
	if err != nil {
		return err
	}
	body := &networksecurity.AddressGroup{}
	if err := loadYAMLOrJSONInto(flagNSConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.AddressGroups.Create(parent, body).AddressGroupId(args[0]).Context(ctx)
	if flagNSRequestID != "" {
		call = call.RequestId(flagNSRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating org address group: %w", err)
	}
	return nsFinishOp(ctx, svc, op, "Create org address group", args[0])
}

func runNSOAGDelete(cmd *cobra.Command, args []string) error {
	parent, err := nsOrgParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	call := svc.Organizations.Locations.AddressGroups.Delete(name).Context(ctx)
	if flagNSRequestID != "" {
		call = call.RequestId(flagNSRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting org address group: %w", err)
	}
	return nsFinishOp(ctx, svc, op, "Delete org address group", args[0])
}

func runNSOAGDescribe(cmd *cobra.Command, args []string) error {
	parent, err := nsOrgParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	got, err := svc.Organizations.Locations.AddressGroups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing org address group: %w", err)
	}
	return emitFormatted(got, flagNSFormat)
}

func runNSOAGList(cmd *cobra.Command, args []string) error {
	parent, err := nsOrgParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*networksecurity.AddressGroup
	pageToken := ""
	for {
		call := svc.Organizations.Locations.AddressGroups.List(parent).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing org address groups: %w", err)
		}
		all = append(all, resp.AddressGroups...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagNSFormat != "" {
		return emitFormatted(all, flagNSFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "TYPE")
	for _, g := range all {
		fmt.Printf("%-40s %s\n", nsBasename(g.Name), g.Type)
	}
	return nil
}

func runNSOAGUpdate(cmd *cobra.Command, args []string) error {
	parent, err := nsOrgParent()
	if err != nil {
		return err
	}
	body := &networksecurity.AddressGroup{}
	if err := loadYAMLOrJSONInto(flagNSConfigFile, body); err != nil {
		return err
	}
	mask := nsResolveMask(body)
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	call := svc.Organizations.Locations.AddressGroups.Patch(name, body).UpdateMask(mask).Context(ctx)
	if flagNSRequestID != "" {
		call = call.RequestId(flagNSRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating org address group: %w", err)
	}
	return nsFinishOp(ctx, svc, op, "Update org address group", args[0])
}

func runNSOAGAddItems(cmd *cobra.Command, args []string) error {
	parent, err := nsOrgParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	req := &networksecurity.AddAddressGroupItemsRequest{Items: flagNSItems, RequestId: flagNSRequestID}
	op, err := svc.Organizations.Locations.AddressGroups.AddItems(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding items: %w", err)
	}
	return nsFinishOp(ctx, svc, op, "Add items", args[0])
}

func runNSOAGRemoveItems(cmd *cobra.Command, args []string) error {
	parent, err := nsOrgParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	req := &networksecurity.RemoveAddressGroupItemsRequest{Items: flagNSItems, RequestId: flagNSRequestID}
	op, err := svc.Organizations.Locations.AddressGroups.RemoveItems(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing items: %w", err)
	}
	return nsFinishOp(ctx, svc, op, "Remove items", args[0])
}

func runNSOAGCloneItems(cmd *cobra.Command, args []string) error {
	parent, err := nsOrgParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	req := &networksecurity.CloneAddressGroupItemsRequest{
		SourceAddressGroup: flagNSSourceAddressGroup,
		RequestId:          flagNSRequestID,
	}
	op, err := svc.Organizations.Locations.AddressGroups.CloneItems(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cloning items: %w", err)
	}
	return nsFinishOp(ctx, svc, op, "Clone items", args[0])
}

func runNSOAGListRefs(cmd *cobra.Command, args []string) error {
	parent, err := nsOrgParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	var all []*networksecurity.ListAddressGroupReferencesResponseAddressGroupReference
	pageToken := ""
	for {
		call := svc.Organizations.Locations.AddressGroups.ListReferences(name).Context(ctx)
		if flagNSPageSize > 0 {
			call = call.PageSize(flagNSPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing references: %w", err)
		}
		all = append(all, resp.AddressGroupReferences...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNSFormat)
}
