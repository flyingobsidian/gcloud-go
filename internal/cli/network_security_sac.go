package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	networksecuritybeta "google.golang.org/api/networksecurity/v1beta1"
)

// secure-access-connect only exists on the v1beta1 surface (issue #835). The
// v1 Go client does not expose SacAttachments / SacRealms, so this file uses
// the beta client directly.

func (s *networkSecurityCmds) registerNSSecureAccessConnect(root *cobra.Command) {
	group := &cobra.Command{Use: "secure-access-connect", Short: "Manage Secure Access Connect resources"}
	group.AddCommand(s.newNSSacAttachments(), s.newNSSacRealms())
	root.AddCommand(group)
}

func nsSacWaitOp(ctx context.Context, svc *networksecuritybeta.Service, op *networksecuritybeta.Operation) (*networksecuritybeta.Operation, error) {
	for !op.Done {
		got, err := svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		return op, fmt.Errorf("operation %s failed: %s", op.Name, op.Error.Message)
	}
	return op, nil
}

func (s *networkSecurityCmds) nsSacFinishOp(ctx context.Context, svc *networksecuritybeta.Service, op *networksecuritybeta.Operation, verb, name string) error {
	if s.flagNSAsync {
		fmt.Fprintf(s.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	final, err := nsSacWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(s.inv.stdout, final.Response, "")
	}
	return nil
}

func nsSacChild(parent, collection, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

// --- attachments ---

func (s *networkSecurityCmds) newNSSacAttachments() *cobra.Command {
	group := &cobra.Command{Use: "attachments", Short: "Manage Secure Access Connect attachments"}
	create := &cobra.Command{
		Use: "create ATTACHMENT", Short: "Create a Secure Access Connect attachment from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runNSSacAttachmentCreate,
	}
	del := &cobra.Command{
		Use: "delete ATTACHMENT", Short: "Delete a Secure Access Connect attachment",
		Args: cobra.ExactArgs(1), RunE: s.runNSSacAttachmentDelete,
	}
	desc := &cobra.Command{
		Use: "describe ATTACHMENT", Short: "Describe a Secure Access Connect attachment",
		Args: cobra.ExactArgs(1), RunE: s.runNSSacAttachmentDescribe,
	}
	list := &cobra.Command{
		Use: "list", Short: "List Secure Access Connect attachments in a location",
		Args: cobra.NoArgs, RunE: s.runNSSacAttachmentList,
	}
	s.addNSLocationFlag(create, del, desc, list)
	s.addNSFormatFlag(desc, list)
	s.addNSCreateConfigFlag(create)
	s.addNSAsyncFlag(create, del)
	s.addNSRequestIDFlag(create, del)
	group.AddCommand(create, del, desc, list)
	return group
}

func (s *networkSecurityCmds) runNSSacAttachmentCreate(_ *cobra.Command, args []string) error {
	parent, err := s.nsProjectParent()
	if err != nil {
		return err
	}
	body := &networksecuritybeta.SACAttachment{}
	if err := loadYAMLOrJSONInto(s.flagNSConfigFile, body); err != nil {
		return err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().NetworkSecurityBeta(ctx)
	if err != nil {
		return err
	}
	c := svc.Projects.Locations.SacAttachments.Create(parent, body).SacAttachmentId(args[0]).Context(ctx)
	if s.flagNSRequestID != "" {
		c = c.RequestId(s.flagNSRequestID)
	}
	op, err := c.Do()
	if err != nil {
		return fmt.Errorf("creating attachment: %w", err)
	}
	return s.nsSacFinishOp(ctx, svc, op, "Create attachment", args[0])
}

func (s *networkSecurityCmds) runNSSacAttachmentDelete(_ *cobra.Command, args []string) error {
	parent, err := s.nsProjectParent()
	if err != nil {
		return err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().NetworkSecurityBeta(ctx)
	if err != nil {
		return err
	}
	name := nsSacChild(parent, "sacAttachments", args[0])
	c := svc.Projects.Locations.SacAttachments.Delete(name).Context(ctx)
	if s.flagNSRequestID != "" {
		c = c.RequestId(s.flagNSRequestID)
	}
	op, err := c.Do()
	if err != nil {
		return fmt.Errorf("deleting attachment: %w", err)
	}
	return s.nsSacFinishOp(ctx, svc, op, "Delete attachment", args[0])
}

func (s *networkSecurityCmds) runNSSacAttachmentDescribe(_ *cobra.Command, args []string) error {
	parent, err := s.nsProjectParent()
	if err != nil {
		return err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().NetworkSecurityBeta(ctx)
	if err != nil {
		return err
	}
	name := nsSacChild(parent, "sacAttachments", args[0])
	got, err := svc.Projects.Locations.SacAttachments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing attachment: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNSFormat)
}

func (s *networkSecurityCmds) runNSSacAttachmentList(_ *cobra.Command, _ []string) error {
	parent, err := s.nsProjectParent()
	if err != nil {
		return err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().NetworkSecurityBeta(ctx)
	if err != nil {
		return err
	}
	var all []*networksecuritybeta.SACAttachment
	pageToken := ""
	for {
		c := svc.Projects.Locations.SacAttachments.List(parent).Context(ctx)
		if pageToken != "" {
			c = c.PageToken(pageToken)
		}
		resp, err := c.Do()
		if err != nil {
			return fmt.Errorf("listing attachments: %w", err)
		}
		all = append(all, resp.SacAttachments...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagNSFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagNSFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-50s %s\n", "NAME", "STATE")
	for _, a := range all {
		fmt.Fprintf(s.inv.stdout, "%-50s %s\n", nsBasename(a.Name), a.State)
	}
	return nil
}

// --- realms ---

func (s *networkSecurityCmds) newNSSacRealms() *cobra.Command {
	group := &cobra.Command{Use: "realms", Short: "Manage Secure Access Connect realms"}
	create := &cobra.Command{
		Use: "create REALM", Short: "Create a Secure Access Connect realm from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runNSSacRealmCreate,
	}
	del := &cobra.Command{
		Use: "delete REALM", Short: "Delete a Secure Access Connect realm",
		Args: cobra.ExactArgs(1), RunE: s.runNSSacRealmDelete,
	}
	desc := &cobra.Command{
		Use: "describe REALM", Short: "Describe a Secure Access Connect realm",
		Args: cobra.ExactArgs(1), RunE: s.runNSSacRealmDescribe,
	}
	list := &cobra.Command{
		Use: "list", Short: "List Secure Access Connect realms in a location",
		Args: cobra.NoArgs, RunE: s.runNSSacRealmList,
	}
	s.addNSLocationFlag(create, del, desc, list)
	s.addNSFormatFlag(desc, list)
	s.addNSCreateConfigFlag(create)
	s.addNSAsyncFlag(create, del)
	s.addNSRequestIDFlag(create, del)
	group.AddCommand(create, del, desc, list)
	return group
}

func (s *networkSecurityCmds) runNSSacRealmCreate(_ *cobra.Command, args []string) error {
	parent, err := s.nsProjectParent()
	if err != nil {
		return err
	}
	body := &networksecuritybeta.SACRealm{}
	if err := loadYAMLOrJSONInto(s.flagNSConfigFile, body); err != nil {
		return err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().NetworkSecurityBeta(ctx)
	if err != nil {
		return err
	}
	c := svc.Projects.Locations.SacRealms.Create(parent, body).SacRealmId(args[0]).Context(ctx)
	if s.flagNSRequestID != "" {
		c = c.RequestId(s.flagNSRequestID)
	}
	op, err := c.Do()
	if err != nil {
		return fmt.Errorf("creating realm: %w", err)
	}
	return s.nsSacFinishOp(ctx, svc, op, "Create realm", args[0])
}

func (s *networkSecurityCmds) runNSSacRealmDelete(_ *cobra.Command, args []string) error {
	parent, err := s.nsProjectParent()
	if err != nil {
		return err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().NetworkSecurityBeta(ctx)
	if err != nil {
		return err
	}
	name := nsSacChild(parent, "sacRealms", args[0])
	c := svc.Projects.Locations.SacRealms.Delete(name).Context(ctx)
	if s.flagNSRequestID != "" {
		c = c.RequestId(s.flagNSRequestID)
	}
	op, err := c.Do()
	if err != nil {
		return fmt.Errorf("deleting realm: %w", err)
	}
	return s.nsSacFinishOp(ctx, svc, op, "Delete realm", args[0])
}

func (s *networkSecurityCmds) runNSSacRealmDescribe(_ *cobra.Command, args []string) error {
	parent, err := s.nsProjectParent()
	if err != nil {
		return err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().NetworkSecurityBeta(ctx)
	if err != nil {
		return err
	}
	name := nsSacChild(parent, "sacRealms", args[0])
	got, err := svc.Projects.Locations.SacRealms.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing realm: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNSFormat)
}

func (s *networkSecurityCmds) runNSSacRealmList(_ *cobra.Command, _ []string) error {
	parent, err := s.nsProjectParent()
	if err != nil {
		return err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().NetworkSecurityBeta(ctx)
	if err != nil {
		return err
	}
	var all []*networksecuritybeta.SACRealm
	pageToken := ""
	for {
		c := svc.Projects.Locations.SacRealms.List(parent).Context(ctx)
		if pageToken != "" {
			c = c.PageToken(pageToken)
		}
		resp, err := c.Do()
		if err != nil {
			return fmt.Errorf("listing realms: %w", err)
		}
		all = append(all, resp.SacRealms...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagNSFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagNSFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-50s\n", "NAME")
	for _, r := range all {
		fmt.Fprintf(s.inv.stdout, "%-50s\n", nsBasename(r.Name))
	}
	return nil
}
