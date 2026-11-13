package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	networksecurity "google.golang.org/api/networksecurity/v1"
)

// --- address-groups (project scope, issue #819) ---
// --- org-address-groups (organization scope, issue #834) ---

func (s *networkSecurityCmds) registerNSAddressGroups(root *cobra.Command) {
	// project-scoped
	projectCmds := []*cobra.Command{
		s.nsAGCreateCmd, s.nsAGDeleteCmd, s.nsAGDescribeCmd, s.nsAGListCmd, s.nsAGUpdateCmd,
		s.nsAGAddItemsCmd, s.nsAGRemoveItemsCmd, s.nsAGCloneItemsCmd, s.nsAGListRefsCmd,
	}
	s.addNSLocationFlag(projectCmds...)
	s.addNSFormatFlag(s.nsAGDescribeCmd, s.nsAGListCmd, s.nsAGListRefsCmd)
	s.addNSFilterFlag(s.nsAGListCmd, s.nsAGListRefsCmd)
	s.addNSCreateConfigFlag(s.nsAGCreateCmd)
	s.addNSUpdateConfigFlag(s.nsAGUpdateCmd)
	s.addNSAsyncFlag(s.nsAGCreateCmd, s.nsAGDeleteCmd, s.nsAGUpdateCmd, s.nsAGAddItemsCmd, s.nsAGRemoveItemsCmd, s.nsAGCloneItemsCmd)
	s.addNSRequestIDFlag(s.nsAGCreateCmd, s.nsAGDeleteCmd, s.nsAGUpdateCmd)
	for _, c := range []*cobra.Command{s.nsAGAddItemsCmd, s.nsAGRemoveItemsCmd} {
		c.Flags().StringSliceVar(&s.flagNSItems, "items", nil, "Comma-separated list of items to add/remove (required)")
		_ = c.MarkFlagRequired("items")
	}
	s.nsAGCloneItemsCmd.Flags().StringVar(&s.flagNSSourceAddressGroup, "source-address-group", "", "Fully qualified source address group to clone items from (required)")
	_ = s.nsAGCloneItemsCmd.MarkFlagRequired("source-address-group")
	s.nsAGListRefsCmd.Flags().Int64Var(&s.flagNSPageSize, "page-size", 0, "Maximum results per page (optional)")

	s.networkSecurityAddressGroupsCmd.AddCommand(projectCmds...)
	root.AddCommand(s.networkSecurityAddressGroupsCmd)

	// org-scoped
	orgCmds := []*cobra.Command{
		s.nsOAGCreateCmd, s.nsOAGDeleteCmd, s.nsOAGDescribeCmd, s.nsOAGListCmd, s.nsOAGUpdateCmd,
		s.nsOAGAddItemsCmd, s.nsOAGRemoveItemsCmd, s.nsOAGCloneItemsCmd, s.nsOAGListRefsCmd,
	}
	s.addNSOrgFlags(orgCmds...)
	s.addNSFormatFlag(s.nsOAGDescribeCmd, s.nsOAGListCmd, s.nsOAGListRefsCmd)
	s.addNSFilterFlag(s.nsOAGListCmd, s.nsOAGListRefsCmd)
	s.addNSCreateConfigFlag(s.nsOAGCreateCmd)
	s.addNSUpdateConfigFlag(s.nsOAGUpdateCmd)
	s.addNSAsyncFlag(s.nsOAGCreateCmd, s.nsOAGDeleteCmd, s.nsOAGUpdateCmd, s.nsOAGAddItemsCmd, s.nsOAGRemoveItemsCmd, s.nsOAGCloneItemsCmd)
	s.addNSRequestIDFlag(s.nsOAGCreateCmd, s.nsOAGDeleteCmd, s.nsOAGUpdateCmd)
	for _, c := range []*cobra.Command{s.nsOAGAddItemsCmd, s.nsOAGRemoveItemsCmd} {
		c.Flags().StringSliceVar(&s.flagNSItems, "items", nil, "Comma-separated list of items to add/remove (required)")
		_ = c.MarkFlagRequired("items")
	}
	s.nsOAGCloneItemsCmd.Flags().StringVar(&s.flagNSSourceAddressGroup, "source-address-group", "", "Fully qualified source address group to clone items from (required)")
	_ = s.nsOAGCloneItemsCmd.MarkFlagRequired("source-address-group")
	s.nsOAGListRefsCmd.Flags().Int64Var(&s.flagNSPageSize, "page-size", 0, "Maximum results per page (optional)")

	s.networkSecurityOrgAddressGroupsCmd.AddCommand(orgCmds...)
	root.AddCommand(s.networkSecurityOrgAddressGroupsCmd)
}

// --- project-scoped implementations ---

func (s *networkSecurityCmds) runNSAGCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.nsProjectParent()
	if err != nil {
		return err
	}
	body := &networksecurity.AddressGroup{}
	if err := loadYAMLOrJSONInto(s.flagNSConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.AddressGroups.Create(parent, body).AddressGroupId(args[0]).Context(ctx)
	if s.flagNSRequestID != "" {
		call = call.RequestId(s.flagNSRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating address group: %w", err)
	}
	return s.nsFinishOp(ctx, svc, op, "Create address group", args[0])
}

func (s *networkSecurityCmds) runNSAGDelete(cmd *cobra.Command, args []string) error {
	parent, err := s.nsProjectParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	call := svc.Projects.Locations.AddressGroups.Delete(name).Context(ctx)
	if s.flagNSRequestID != "" {
		call = call.RequestId(s.flagNSRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting address group: %w", err)
	}
	return s.nsFinishOp(ctx, svc, op, "Delete address group", args[0])
}

func (s *networkSecurityCmds) runNSAGDescribe(cmd *cobra.Command, args []string) error {
	parent, err := s.nsProjectParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	got, err := svc.Projects.Locations.AddressGroups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing address group: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNSFormat)
}

func (s *networkSecurityCmds) runNSAGList(cmd *cobra.Command, args []string) error {
	parent, err := s.nsProjectParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
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
	if s.flagNSFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagNSFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "NAME", "TYPE")
	for _, g := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", nsBasename(g.Name), g.Type)
	}
	return nil
}

func (s *networkSecurityCmds) runNSAGUpdate(cmd *cobra.Command, args []string) error {
	parent, err := s.nsProjectParent()
	if err != nil {
		return err
	}
	body := &networksecurity.AddressGroup{}
	if err := loadYAMLOrJSONInto(s.flagNSConfigFile, body); err != nil {
		return err
	}
	mask := s.nsResolveMask(body)
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	call := svc.Projects.Locations.AddressGroups.Patch(name, body).UpdateMask(mask).Context(ctx)
	if s.flagNSRequestID != "" {
		call = call.RequestId(s.flagNSRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating address group: %w", err)
	}
	return s.nsFinishOp(ctx, svc, op, "Update address group", args[0])
}

func (s *networkSecurityCmds) runNSAGAddItems(cmd *cobra.Command, args []string) error {
	parent, err := s.nsProjectParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	req := &networksecurity.AddAddressGroupItemsRequest{Items: s.flagNSItems, RequestId: s.flagNSRequestID}
	op, err := svc.Projects.Locations.AddressGroups.AddItems(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding items: %w", err)
	}
	return s.nsFinishOp(ctx, svc, op, "Add items", args[0])
}

func (s *networkSecurityCmds) runNSAGRemoveItems(cmd *cobra.Command, args []string) error {
	parent, err := s.nsProjectParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	req := &networksecurity.RemoveAddressGroupItemsRequest{Items: s.flagNSItems, RequestId: s.flagNSRequestID}
	op, err := svc.Projects.Locations.AddressGroups.RemoveItems(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing items: %w", err)
	}
	return s.nsFinishOp(ctx, svc, op, "Remove items", args[0])
}

func (s *networkSecurityCmds) runNSAGCloneItems(cmd *cobra.Command, args []string) error {
	parent, err := s.nsProjectParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	req := &networksecurity.CloneAddressGroupItemsRequest{
		SourceAddressGroup: s.flagNSSourceAddressGroup,
		RequestId:          s.flagNSRequestID,
	}
	op, err := svc.Projects.Locations.AddressGroups.CloneItems(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cloning items: %w", err)
	}
	return s.nsFinishOp(ctx, svc, op, "Clone items", args[0])
}

func (s *networkSecurityCmds) runNSAGListRefs(cmd *cobra.Command, args []string) error {
	parent, err := s.nsProjectParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	var all []*networksecurity.ListAddressGroupReferencesResponseAddressGroupReference
	pageToken := ""
	for {
		call := svc.Projects.Locations.AddressGroups.ListReferences(name).Context(ctx)
		if s.flagNSPageSize > 0 {
			call = call.PageSize(s.flagNSPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagNSFormat)
}

// --- org-scoped implementations ---

func (s *networkSecurityCmds) runNSOAGCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.nsOrgParent()
	if err != nil {
		return err
	}
	body := &networksecurity.AddressGroup{}
	if err := loadYAMLOrJSONInto(s.flagNSConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.AddressGroups.Create(parent, body).AddressGroupId(args[0]).Context(ctx)
	if s.flagNSRequestID != "" {
		call = call.RequestId(s.flagNSRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating org address group: %w", err)
	}
	return s.nsFinishOp(ctx, svc, op, "Create org address group", args[0])
}

func (s *networkSecurityCmds) runNSOAGDelete(cmd *cobra.Command, args []string) error {
	parent, err := s.nsOrgParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	call := svc.Organizations.Locations.AddressGroups.Delete(name).Context(ctx)
	if s.flagNSRequestID != "" {
		call = call.RequestId(s.flagNSRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting org address group: %w", err)
	}
	return s.nsFinishOp(ctx, svc, op, "Delete org address group", args[0])
}

func (s *networkSecurityCmds) runNSOAGDescribe(cmd *cobra.Command, args []string) error {
	parent, err := s.nsOrgParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	got, err := svc.Organizations.Locations.AddressGroups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing org address group: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNSFormat)
}

func (s *networkSecurityCmds) runNSOAGList(cmd *cobra.Command, args []string) error {
	parent, err := s.nsOrgParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
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
	if s.flagNSFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagNSFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "NAME", "TYPE")
	for _, g := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", nsBasename(g.Name), g.Type)
	}
	return nil
}

func (s *networkSecurityCmds) runNSOAGUpdate(cmd *cobra.Command, args []string) error {
	parent, err := s.nsOrgParent()
	if err != nil {
		return err
	}
	body := &networksecurity.AddressGroup{}
	if err := loadYAMLOrJSONInto(s.flagNSConfigFile, body); err != nil {
		return err
	}
	mask := s.nsResolveMask(body)
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	call := svc.Organizations.Locations.AddressGroups.Patch(name, body).UpdateMask(mask).Context(ctx)
	if s.flagNSRequestID != "" {
		call = call.RequestId(s.flagNSRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating org address group: %w", err)
	}
	return s.nsFinishOp(ctx, svc, op, "Update org address group", args[0])
}

func (s *networkSecurityCmds) runNSOAGAddItems(cmd *cobra.Command, args []string) error {
	parent, err := s.nsOrgParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	req := &networksecurity.AddAddressGroupItemsRequest{Items: s.flagNSItems, RequestId: s.flagNSRequestID}
	op, err := svc.Organizations.Locations.AddressGroups.AddItems(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding items: %w", err)
	}
	return s.nsFinishOp(ctx, svc, op, "Add items", args[0])
}

func (s *networkSecurityCmds) runNSOAGRemoveItems(cmd *cobra.Command, args []string) error {
	parent, err := s.nsOrgParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	req := &networksecurity.RemoveAddressGroupItemsRequest{Items: s.flagNSItems, RequestId: s.flagNSRequestID}
	op, err := svc.Organizations.Locations.AddressGroups.RemoveItems(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing items: %w", err)
	}
	return s.nsFinishOp(ctx, svc, op, "Remove items", args[0])
}

func (s *networkSecurityCmds) runNSOAGCloneItems(cmd *cobra.Command, args []string) error {
	parent, err := s.nsOrgParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	req := &networksecurity.CloneAddressGroupItemsRequest{
		SourceAddressGroup: s.flagNSSourceAddressGroup,
		RequestId:          s.flagNSRequestID,
	}
	op, err := svc.Organizations.Locations.AddressGroups.CloneItems(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cloning items: %w", err)
	}
	return s.nsFinishOp(ctx, svc, op, "Clone items", args[0])
}

func (s *networkSecurityCmds) runNSOAGListRefs(cmd *cobra.Command, args []string) error {
	parent, err := s.nsOrgParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	name := nsChild(parent, "addressGroups", args[0])
	var all []*networksecurity.ListAddressGroupReferencesResponseAddressGroupReference
	pageToken := ""
	for {
		call := svc.Organizations.Locations.AddressGroups.ListReferences(name).Context(ctx)
		if s.flagNSPageSize > 0 {
			call = call.PageSize(s.flagNSPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagNSFormat)
}
