package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	networksecurity "google.golang.org/api/networksecurity/v1"
)

// nsCRUD is a generic Create/Delete/Describe/List/Update binding for network
// security resources. Each resource plugs the API service calls in via closures
// so registerNS* only has to build the cobra subcommands.
type nsCRUD[T any] struct {
	s *networkSecurityCmds
	// group is the cobra subcommand name (e.g. "url-lists"). It is not used
	// for API calls; it exists so error messages can mention the group.
	group string
	// singular is a human label such as "URL list" used in progress logging.
	singular string
	// collection is the URL segment used to synthesise fully-qualified names
	// (e.g. "urlLists"). Must not be empty.
	collection string

	// parentFn resolves the parent resource (projects/PROJECT/locations/LOC
	// or organizations/ORG/locations/LOC).
	parentFn func() (string, error)

	// API closures. Any nil closure disables the corresponding subcommand.
	createFn func(ctx context.Context, svc *networksecurity.Service, parent, id string, body *T, requestID string) (*networksecurity.Operation, error)
	deleteFn func(ctx context.Context, svc *networksecurity.Service, name, requestID string) (*networksecurity.Operation, error)
	getFn    func(ctx context.Context, svc *networksecurity.Service, name string) (*T, error)
	listFn   func(ctx context.Context, svc *networksecurity.Service, parent, pageToken, filter string) ([]*T, string, error)
	patchFn  func(ctx context.Context, svc *networksecurity.Service, name string, body *T, mask, requestID string) (*networksecurity.Operation, error)

	// nameCol/secondaryCol return the columns shown in the default plain-text
	// list output. If nameCol is nil the resource's name basename is used and
	// secondaryCol is dropped.
	nameCol      func(*T) string
	secondaryCol func(*T) string
	// secondaryHeader is the header printed above secondaryCol.
	secondaryHeader string
}

func (b *nsCRUD[T]) runCreate(c *cobra.Command, args []string) error {
	parent, err := b.parentFn()
	if err != nil {
		return err
	}
	body := new(T)
	if err := loadYAMLOrJSONInto(b.s.flagNSConfigFile, body); err != nil {
		return err
	}
	ctx := c.Context()
	svc, err := b.s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	op, err := b.createFn(ctx, svc, parent, args[0], body, b.s.flagNSRequestID)
	if err != nil {
		return fmt.Errorf("creating %s: %w", b.singular, err)
	}
	return b.s.nsFinishOp(ctx, svc, op, "Create "+b.singular, args[0])
}

func (b *nsCRUD[T]) runDelete(c *cobra.Command, args []string) error {
	parent, err := b.parentFn()
	if err != nil {
		return err
	}
	ctx := c.Context()
	svc, err := b.s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	name := nsChild(parent, b.collection, args[0])
	op, err := b.deleteFn(ctx, svc, name, b.s.flagNSRequestID)
	if err != nil {
		return fmt.Errorf("deleting %s: %w", b.singular, err)
	}
	return b.s.nsFinishOp(ctx, svc, op, "Delete "+b.singular, args[0])
}

func (b *nsCRUD[T]) runDescribe(c *cobra.Command, args []string) error {
	parent, err := b.parentFn()
	if err != nil {
		return err
	}
	ctx := c.Context()
	svc, err := b.s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	name := nsChild(parent, b.collection, args[0])
	got, err := b.getFn(ctx, svc, name)
	if err != nil {
		return fmt.Errorf("describing %s: %w", b.singular, err)
	}
	return emitFormattedTo(b.s.inv.stdout, got, b.s.flagNSFormat)
}

func (b *nsCRUD[T]) runList(c *cobra.Command, _ []string) error {
	parent, err := b.parentFn()
	if err != nil {
		return err
	}
	ctx := c.Context()
	svc, err := b.s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	var all []*T
	pageToken := ""
	for {
		items, next, err := b.listFn(ctx, svc, parent, pageToken, b.s.flagNSFilter)
		if err != nil {
			return fmt.Errorf("listing %s: %w", b.group, err)
		}
		all = append(all, items...)
		if next == "" {
			break
		}
		pageToken = next
	}
	if b.s.flagNSFormat != "" {
		return emitFormattedTo(b.s.inv.stdout, all, b.s.flagNSFormat)
	}
	if b.nameCol == nil {
		return emitFormattedTo(b.s.inv.stdout, all, "")
	}
	if b.secondaryCol == nil {
		printTableHeadingTo(b.s.inv.stdout, len(all), "%-50s\n", "NAME")
		for _, it := range all {
			fmt.Fprintf(b.s.inv.stdout, "%-50s\n", b.nameCol(it))
		}
		return nil
	}
	printTableHeadingTo(b.s.inv.stdout, len(all), "%-50s %s\n", "NAME", b.secondaryHeader)
	for _, it := range all {
		fmt.Fprintf(b.s.inv.stdout, "%-50s %s\n", b.nameCol(it), b.secondaryCol(it))
	}
	return nil
}

func (b *nsCRUD[T]) runUpdate(c *cobra.Command, args []string) error {
	parent, err := b.parentFn()
	if err != nil {
		return err
	}
	body := new(T)
	if err := loadYAMLOrJSONInto(b.s.flagNSConfigFile, body); err != nil {
		return err
	}
	mask := b.s.nsResolveMask(body)
	ctx := c.Context()
	svc, err := b.s.inv.clients().NetworkSecurity(ctx)
	if err != nil {
		return err
	}
	name := nsChild(parent, b.collection, args[0])
	op, err := b.patchFn(ctx, svc, name, body, mask, b.s.flagNSRequestID)
	if err != nil {
		return fmt.Errorf("updating %s: %w", b.singular, err)
	}
	return b.s.nsFinishOp(ctx, svc, op, "Update "+b.singular, args[0])
}

// build wires the CRUD subcommands under group, honouring the parent-scope
// flag setter. It returns the leaf commands so registerNS* can add extra flags
// if needed. scopeFlags is a function that adds --location (and --organization
// for org resources) to each supplied command.
func (b *nsCRUD[T]) build(parent *cobra.Command, group, groupShort string, scopeFlags func(...*cobra.Command)) *cobra.Command {
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
	upd := &cobra.Command{
		Use: "update " + toUpperArg(singular), Short: "Update a " + singular + " from a --config-file",
		Args: cobra.ExactArgs(1), RunE: b.runUpdate,
	}

	scopeFlags(create, del, desc, list, upd)
	b.s.addNSFormatFlag(desc, list)
	b.s.addNSFilterFlag(list)
	b.s.addNSCreateConfigFlag(create)
	b.s.addNSUpdateConfigFlag(upd)
	b.s.addNSAsyncFlag(create, del, upd)
	b.s.addNSRequestIDFlag(create, del, upd)

	g.AddCommand(create, del, desc, list, upd)
	parent.AddCommand(g)
	return g
}

// toUpperArg converts a human-readable singular ("URL list") into an ARG token
// ("URL_LIST") used in the cobra Use line.
func toUpperArg(singular string) string {
	out := make([]byte, 0, len(singular))
	for i := 0; i < len(singular); i++ {
		c := singular[i]
		switch {
		case c == ' ':
			out = append(out, '_')
		case c >= 'a' && c <= 'z':
			out = append(out, c-32)
		default:
			out = append(out, c)
		}
	}
	return string(out)
}
