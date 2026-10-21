package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	compute "google.golang.org/api/compute/v1"
)

// --- Compute Engine multi-scope lister (#1998) ---
//
// A port of gcloud's compute lister (api_lib/compute/lister.py): list
// commands aggregate across every zone and region by default, and the
// --zones, --regions and --global flags narrow the scope.

// computeLister describes one Compute Engine resource type's list command.
// Fetchers left nil mark scopes the resource doesn't have.
type computeLister[T any] struct {
	noun         string // resource name in error messages, e.g. "disks"
	aggregated   func(ctx context.Context, svc *compute.Service, project string) ([]T, error)
	zonalList    func(ctx context.Context, svc *compute.Service, project, zone string) ([]T, error)
	regionalList func(ctx context.Context, svc *compute.Service, project, region string) ([]T, error)
	globalList   func(ctx context.Context, svc *compute.Service, project string) ([]T, error)
	// location returns the zone and region URLs of an item; both are empty
	// for a global resource.
	location    func(T) (zone, region string)
	selfLink    func(T) string
	tableFormat string // gcloud's default table, over the rows from tableRow
	tableRow    func(T) map[string]any
	notice      string // printed to stderr after the default table
}

// computeListRequest holds the scope flags of one list invocation.
type computeListRequest struct {
	zones, regions       []string
	zonesSet, regionsSet bool
	global               bool
}

// list fetches the items selected by req, following gcloud's lister: named
// zones or regions are listed one by one, --global lists global resources,
// and otherwise an aggregated list is filtered to the requested scope type.
// It never returns a nil slice, so an empty result renders as [] in JSON.
func (l computeLister[T]) list(ctx context.Context, svc *compute.Service, project string, req computeListRequest) ([]T, error) {
	all := []T{}
	switch {
	case len(req.zones) > 0:
		for _, zone := range req.zones {
			items, err := l.zonalList(ctx, svc, project, zone)
			if err != nil {
				return nil, fmt.Errorf("listing %s in zone %s: %w", l.noun, zone, err)
			}
			all = append(all, items...)
		}
	case len(req.regions) > 0:
		for _, region := range req.regions {
			items, err := l.regionalList(ctx, svc, project, region)
			if err != nil {
				return nil, fmt.Errorf("listing %s in region %s: %w", l.noun, region, err)
			}
			all = append(all, items...)
		}
	case req.global || l.aggregated == nil:
		items, err := l.globalList(ctx, svc, project)
		if err != nil {
			return nil, fmt.Errorf("listing %s: %w", l.noun, err)
		}
		all = append(all, items...)
	default:
		items, err := l.aggregated(ctx, svc, project)
		if err != nil {
			return nil, fmt.Errorf("listing %s: %w", l.noun, err)
		}
		for _, item := range items {
			zone, region := l.location(item)
			if (req.zonesSet && zone == "") || (req.regionsSet && region == "") {
				continue
			}
			all = append(all, item)
		}
	}
	return all, nil
}

// emit prints items with --format, their self links with --uri, or gcloud's
// default table, which (like gcloud) is replaced by "Listed 0 items." on errW
// when there is nothing to show.
func (l computeLister[T]) emit(w, errW io.Writer, items []T, format string, uri bool) error {
	if uri {
		for _, item := range items {
			if _, err := fmt.Fprintln(w, l.selfLink(item)); err != nil {
				return err
			}
		}
		return nil
	}
	if format != "" {
		return emitFormattedTo(w, items, format)
	}
	if len(items) == 0 {
		_, err := fmt.Fprintln(errW, "Listed 0 items.")
		return err
	}
	rows := make([]map[string]any, 0, len(items))
	for _, item := range items {
		rows = append(rows, l.tableRow(item))
	}
	if err := emitFormattedTo(w, rows, l.tableFormat); err != nil {
		return err
	}
	if l.notice != "" {
		_, err := fmt.Fprint(errW, l.notice)
		return err
	}
	return nil
}

// newComputeListCmd builds the `list` command for l, with the scope flags of
// the scopes l supports.
func newComputeListCmd[T any](short string, l computeLister[T]) *cobra.Command {
	var req computeListRequest
	var uri bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: short,
		Args:  cobra.NoArgs,
	}
	var scopeFlags []string
	if l.zonalList != nil {
		cmd.Flags().StringSliceVar(&req.zones, "zones", nil,
			"Only show zonal resources; if zones are given, only resources in those zones")
		scopeFlags = append(scopeFlags, "zones")
	}
	if l.regionalList != nil {
		cmd.Flags().StringSliceVar(&req.regions, "regions", nil,
			"Only show regional resources; if regions are given, only resources in those regions")
		scopeFlags = append(scopeFlags, "regions")
	}
	if l.globalList != nil && l.aggregated != nil {
		cmd.Flags().BoolVar(&req.global, "global", false, "Only show global resources")
		scopeFlags = append(scopeFlags, "global")
	}
	if len(scopeFlags) > 1 {
		cmd.MarkFlagsMutuallyExclusive(scopeFlags...)
	}
	cmd.Flags().BoolVar(&uri, "uri", false, "Print a list of resource URIs instead of the default output")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		req.zonesSet = cmd.Flags().Changed("zones")
		req.regionsSet = cmd.Flags().Changed("regions")
		project, err := resolveProject()
		if err != nil {
			return err
		}
		ctx := context.Background()
		svc, err := icompute.NewService(ctx, flagAccount)
		if err != nil {
			return err
		}
		items, err := l.list(ctx, svc, project, req)
		if err != nil {
			return err
		}
		return l.emit(os.Stdout, os.Stderr, items, flagFormat, uri)
	}
	return cmd
}

// sortedScopes returns the keys of an aggregated list's items ("global",
// "regions/...", "zones/...") in a stable order.
func sortedScopes[V any](items map[string]V) []string {
	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
