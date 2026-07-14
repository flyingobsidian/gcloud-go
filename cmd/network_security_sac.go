package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networksecuritybeta "google.golang.org/api/networksecurity/v1beta1"
)

// secure-access-connect only exists on the v1beta1 surface (issue #835). The
// v1 Go client does not expose SacAttachments / SacRealms, so this file uses
// the beta client directly.

func registerNSSecureAccessConnect(root *cobra.Command) {
	group := &cobra.Command{Use: "secure-access-connect", Short: "Manage Secure Access Connect resources"}
	group.AddCommand(newNSSacAttachments(), newNSSacRealms())
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

func nsSacFinishOp(ctx context.Context, svc *networksecuritybeta.Service, op *networksecuritybeta.Operation, verb, name string) error {
	if flagNSAsync {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := nsSacWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
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

func newNSSacAttachments() *cobra.Command {
	group := &cobra.Command{Use: "attachments", Short: "Manage Secure Access Connect attachments"}
	create := &cobra.Command{
		Use: "create ATTACHMENT", Short: "Create a Secure Access Connect attachment from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runNSSacAttachmentCreate,
	}
	del := &cobra.Command{
		Use: "delete ATTACHMENT", Short: "Delete a Secure Access Connect attachment",
		Args: cobra.ExactArgs(1), RunE: runNSSacAttachmentDelete,
	}
	desc := &cobra.Command{
		Use: "describe ATTACHMENT", Short: "Describe a Secure Access Connect attachment",
		Args: cobra.ExactArgs(1), RunE: runNSSacAttachmentDescribe,
	}
	list := &cobra.Command{
		Use: "list", Short: "List Secure Access Connect attachments in a location",
		Args: cobra.NoArgs, RunE: runNSSacAttachmentList,
	}
	addNSLocationFlag(create, del, desc, list)
	addNSFormatFlag(desc, list)
	addNSCreateConfigFlag(create)
	addNSAsyncFlag(create, del)
	addNSRequestIDFlag(create, del)
	group.AddCommand(create, del, desc, list)
	return group
}

func runNSSacAttachmentCreate(_ *cobra.Command, args []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	body := &networksecuritybeta.SACAttachment{}
	if err := loadYAMLOrJSONInto(flagNSConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityBetaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	c := svc.Projects.Locations.SacAttachments.Create(parent, body).SacAttachmentId(args[0]).Context(ctx)
	if flagNSRequestID != "" {
		c = c.RequestId(flagNSRequestID)
	}
	op, err := c.Do()
	if err != nil {
		return fmt.Errorf("creating attachment: %w", err)
	}
	return nsSacFinishOp(ctx, svc, op, "Create attachment", args[0])
}

func runNSSacAttachmentDelete(_ *cobra.Command, args []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityBetaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := nsSacChild(parent, "sacAttachments", args[0])
	c := svc.Projects.Locations.SacAttachments.Delete(name).Context(ctx)
	if flagNSRequestID != "" {
		c = c.RequestId(flagNSRequestID)
	}
	op, err := c.Do()
	if err != nil {
		return fmt.Errorf("deleting attachment: %w", err)
	}
	return nsSacFinishOp(ctx, svc, op, "Delete attachment", args[0])
}

func runNSSacAttachmentDescribe(_ *cobra.Command, args []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityBetaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := nsSacChild(parent, "sacAttachments", args[0])
	got, err := svc.Projects.Locations.SacAttachments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing attachment: %w", err)
	}
	return emitFormatted(got, flagNSFormat)
}

func runNSSacAttachmentList(_ *cobra.Command, _ []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityBetaService(ctx, flagAccount)
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
	if flagNSFormat != "" {
		return emitFormatted(all, flagNSFormat)
	}
	fmt.Printf("%-50s %s\n", "NAME", "STATE")
	for _, a := range all {
		fmt.Printf("%-50s %s\n", nsBasename(a.Name), a.State)
	}
	return nil
}

// --- realms ---

func newNSSacRealms() *cobra.Command {
	group := &cobra.Command{Use: "realms", Short: "Manage Secure Access Connect realms"}
	create := &cobra.Command{
		Use: "create REALM", Short: "Create a Secure Access Connect realm from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runNSSacRealmCreate,
	}
	del := &cobra.Command{
		Use: "delete REALM", Short: "Delete a Secure Access Connect realm",
		Args: cobra.ExactArgs(1), RunE: runNSSacRealmDelete,
	}
	desc := &cobra.Command{
		Use: "describe REALM", Short: "Describe a Secure Access Connect realm",
		Args: cobra.ExactArgs(1), RunE: runNSSacRealmDescribe,
	}
	list := &cobra.Command{
		Use: "list", Short: "List Secure Access Connect realms in a location",
		Args: cobra.NoArgs, RunE: runNSSacRealmList,
	}
	addNSLocationFlag(create, del, desc, list)
	addNSFormatFlag(desc, list)
	addNSCreateConfigFlag(create)
	addNSAsyncFlag(create, del)
	addNSRequestIDFlag(create, del)
	group.AddCommand(create, del, desc, list)
	return group
}

func runNSSacRealmCreate(_ *cobra.Command, args []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	body := &networksecuritybeta.SACRealm{}
	if err := loadYAMLOrJSONInto(flagNSConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityBetaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	c := svc.Projects.Locations.SacRealms.Create(parent, body).SacRealmId(args[0]).Context(ctx)
	if flagNSRequestID != "" {
		c = c.RequestId(flagNSRequestID)
	}
	op, err := c.Do()
	if err != nil {
		return fmt.Errorf("creating realm: %w", err)
	}
	return nsSacFinishOp(ctx, svc, op, "Create realm", args[0])
}

func runNSSacRealmDelete(_ *cobra.Command, args []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityBetaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := nsSacChild(parent, "sacRealms", args[0])
	c := svc.Projects.Locations.SacRealms.Delete(name).Context(ctx)
	if flagNSRequestID != "" {
		c = c.RequestId(flagNSRequestID)
	}
	op, err := c.Do()
	if err != nil {
		return fmt.Errorf("deleting realm: %w", err)
	}
	return nsSacFinishOp(ctx, svc, op, "Delete realm", args[0])
}

func runNSSacRealmDescribe(_ *cobra.Command, args []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityBetaService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := nsSacChild(parent, "sacRealms", args[0])
	got, err := svc.Projects.Locations.SacRealms.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing realm: %w", err)
	}
	return emitFormatted(got, flagNSFormat)
}

func runNSSacRealmList(_ *cobra.Command, _ []string) error {
	parent, err := nsProjectParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkSecurityBetaService(ctx, flagAccount)
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
	if flagNSFormat != "" {
		return emitFormatted(all, flagNSFormat)
	}
	fmt.Printf("%-50s\n", "NAME")
	for _, r := range all {
		fmt.Printf("%-50s\n", nsBasename(r.Name))
	}
	return nil
}
