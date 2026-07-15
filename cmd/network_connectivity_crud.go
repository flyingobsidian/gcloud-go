package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	networkconnectivity "google.golang.org/api/networkconnectivity/v1"
)

// ncCRUD is a generic Create/Delete/Describe/List/Update binding for network
// connectivity resources. Each resource plugs API service calls in via
// closures so registerNC* only has to build the cobra subcommands.
type ncCRUD[T any] struct {
	group      string
	singular   string
	collection string

	parentFn func() (string, error)

	// Any nil closure disables the corresponding subcommand.
	createFn func(ctx context.Context, svc *networkconnectivity.Service, parent, id string, body *T, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error)
	deleteFn func(ctx context.Context, svc *networkconnectivity.Service, name, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error)
	getFn    func(ctx context.Context, svc *networkconnectivity.Service, name string) (*T, error)
	listFn   func(ctx context.Context, svc *networkconnectivity.Service, parent, pageToken, filter string) ([]*T, string, error)
	patchFn  func(ctx context.Context, svc *networkconnectivity.Service, name string, body *T, mask, requestID string) (*networkconnectivity.GoogleLongrunningOperation, error)

	nameCol         func(*T) string
	secondaryCol    func(*T) string
	secondaryHeader string
}

func (b *ncCRUD[T]) runCreate(_ *cobra.Command, args []string) error {
	parent, err := b.parentFn()
	if err != nil {
		return err
	}
	body := new(T)
	if err := loadYAMLOrJSONInto(flagNCConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkConnectivityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := b.createFn(ctx, svc, parent, args[0], body, flagNCRequestID)
	if err != nil {
		return fmt.Errorf("creating %s: %w", b.singular, err)
	}
	return ncFinishOp(ctx, svc, op, "Create "+b.singular, args[0])
}

func (b *ncCRUD[T]) runDelete(_ *cobra.Command, args []string) error {
	parent, err := b.parentFn()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkConnectivityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := ncChild(parent, b.collection, args[0])
	op, err := b.deleteFn(ctx, svc, name, flagNCRequestID)
	if err != nil {
		return fmt.Errorf("deleting %s: %w", b.singular, err)
	}
	return ncFinishOp(ctx, svc, op, "Delete "+b.singular, args[0])
}

func (b *ncCRUD[T]) runDescribe(_ *cobra.Command, args []string) error {
	parent, err := b.parentFn()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkConnectivityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := ncChild(parent, b.collection, args[0])
	got, err := b.getFn(ctx, svc, name)
	if err != nil {
		return fmt.Errorf("describing %s: %w", b.singular, err)
	}
	return emitFormatted(got, flagNCFormat)
}

func (b *ncCRUD[T]) runList(_ *cobra.Command, _ []string) error {
	parent, err := b.parentFn()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetworkConnectivityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*T
	pageToken := ""
	for {
		items, next, err := b.listFn(ctx, svc, parent, pageToken, flagNCFilter)
		if err != nil {
			return fmt.Errorf("listing %s: %w", b.group, err)
		}
		all = append(all, items...)
		if next == "" {
			break
		}
		pageToken = next
	}
	if flagNCFormat != "" {
		return emitFormatted(all, flagNCFormat)
	}
	if b.nameCol == nil {
		return emitFormatted(all, "")
	}
	if b.secondaryCol == nil {
		fmt.Printf("%-50s\n", "NAME")
		for _, it := range all {
			fmt.Printf("%-50s\n", b.nameCol(it))
		}
		return nil
	}
	fmt.Printf("%-50s %s\n", "NAME", b.secondaryHeader)
	for _, it := range all {
		fmt.Printf("%-50s %s\n", b.nameCol(it), b.secondaryCol(it))
	}
	return nil
}

func (b *ncCRUD[T]) runUpdate(_ *cobra.Command, args []string) error {
	parent, err := b.parentFn()
	if err != nil {
		return err
	}
	body := new(T)
	if err := loadYAMLOrJSONInto(flagNCConfigFile, body); err != nil {
		return err
	}
	mask := ncResolveMask(body)
	ctx := context.Background()
	svc, err := gcp.NetworkConnectivityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := ncChild(parent, b.collection, args[0])
	op, err := b.patchFn(ctx, svc, name, body, mask, flagNCRequestID)
	if err != nil {
		return fmt.Errorf("updating %s: %w", b.singular, err)
	}
	return ncFinishOp(ctx, svc, op, "Update "+b.singular, args[0])
}

// build wires the CRUD subcommands under group. hasUpdate controls whether the
// patch subcommand is exposed (some resources are create/delete/describe/list
// only).
func (b *ncCRUD[T]) build(parent *cobra.Command, group, groupShort string, scopeFlags func(...*cobra.Command), hasUpdate bool) *cobra.Command {
	g := &cobra.Command{Use: group, Short: groupShort}
	singular := b.singular

	create := &cobra.Command{
		Use: "create " + toUpperArg(singular), Short: "Create a " + singular + " from a --config-file",
		Args: cobra.ExactArgs(1), RunE: b.runCreate,
	}
	del := &cobra.Command{
		Use: "delete " + toUpperArg(singular), Short: "Delete a " + singular,
		Args: cobra.ExactArgs(1), RunE: b.runDelete,
	}
	desc := &cobra.Command{
		Use: "describe " + toUpperArg(singular), Short: "Describe a " + singular,
		Args: cobra.ExactArgs(1), RunE: b.runDescribe,
	}
	list := &cobra.Command{
		Use: "list", Short: "List " + group + " in a location",
		Args: cobra.NoArgs, RunE: b.runList,
	}
	leaves := []*cobra.Command{create, del, desc, list}
	var upd *cobra.Command
	if hasUpdate {
		upd = &cobra.Command{
			Use: "update " + toUpperArg(singular), Short: "Update a " + singular + " from a --config-file",
			Args: cobra.ExactArgs(1), RunE: b.runUpdate,
		}
		leaves = append(leaves, upd)
	}

	scopeFlags(leaves...)
	addNCFormatFlag(desc, list)
	addNCFilterFlag(list)
	addNCCreateConfigFlag(create)
	if upd != nil {
		addNCUpdateConfigFlag(upd)
		addNCAsyncFlag(create, del, upd)
		addNCRequestIDFlag(create, del, upd)
	} else {
		addNCAsyncFlag(create, del)
		addNCRequestIDFlag(create, del)
	}

	g.AddCommand(leaves...)
	parent.AddCommand(g)
	return g
}
