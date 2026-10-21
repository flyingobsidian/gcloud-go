package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	compute "google.golang.org/api/compute/v1"
)

// --- managed list (#1990) ---

var (
	flagManagedListZones   []string
	flagManagedListRegions []string
)

// managedListFormat is gcloud's default table for `instance-groups managed
// list`, over the rows built by managedListTableRows.
const managedListFormat = "table(name,location,scope,baseInstanceName,size,targetSize,instanceTemplate,autoscaled)"

func init() {
	managedListCmd.Flags().StringSliceVar(&flagManagedListZones, "zones", nil,
		"Only show zonal MIGs; if zones are given, only MIGs in those zones")
	managedListCmd.Flags().StringSliceVar(&flagManagedListRegions, "regions", nil,
		"Only show regional MIGs; if regions are given, only MIGs in those regions")
	managedListCmd.MarkFlagsMutuallyExclusive("zones", "regions")
}

// managedListScope describes which MIGs `managed list` shows, from the
// --zones and --regions flags. With neither set, every zonal and regional MIG
// in the project is listed.
type managedListScope struct {
	zonalOnly    bool
	regionalOnly bool
	zones        []string
	regions      []string
}

// managedListScopeFromFlags builds the scope from --zones/--regions. An
// explicit --zone (a gcloud-go global flag) is treated as --zones=ZONE.
func managedListScopeFromFlags(cmd *cobra.Command) managedListScope {
	switch {
	case cmd.Flags().Changed("zones"):
		return managedListScope{zonalOnly: true, zones: flagManagedListZones}
	case cmd.Flags().Changed("regions"):
		return managedListScope{regionalOnly: true, regions: flagManagedListRegions}
	case flagZone != "":
		return managedListScope{zonalOnly: true, zones: []string{flagZone}}
	}
	return managedListScope{}
}

// keep reports whether mig belongs in an aggregated listing for s.
func (s managedListScope) keep(mig *compute.InstanceGroupManager) bool {
	switch {
	case s.zonalOnly:
		return mig.Zone != ""
	case s.regionalOnly:
		return mig.Region != ""
	}
	return true
}

func runManagedList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	migs, err := fetchManagedListMIGs(ctx, svc, project, managedListScopeFromFlags(cmd))
	if err != nil {
		return err
	}
	if flagManagedListURI {
		for _, m := range migs {
			fmt.Println(m.SelfLink)
		}
		return nil
	}

	var groups []*compute.InstanceGroup
	var autoscalers []*compute.Autoscaler
	if len(migs) > 0 {
		if groups, err = fetchAllInstanceGroups(ctx, svc, project); err != nil {
			return err
		}
		if autoscalers, err = fetchAllAutoscalers(ctx, svc, project); err != nil {
			return err
		}
	}
	records, hadErrors, err := managedListRecords(migs, groups, autoscalers)
	if err != nil {
		return err
	}

	switch {
	case flagFormat != "":
		err = emitFormatted(records, flagFormat)
	case len(records) == 0:
		fmt.Fprintln(os.Stderr, "Listed 0 items.")
	default:
		err = emitFormatted(managedListTableRows(records), managedListFormat)
	}
	if err != nil {
		return err
	}
	if hadErrors {
		fmt.Fprintln(os.Stderr, "(*) - there are errors in your autoscaling setup, please describe the resource to see details")
	}
	return nil
}

// fetchManagedListMIGs lists MIGs per requested zone or region, or with an
// aggregated list across all zones and regions when none are given.
func fetchManagedListMIGs(ctx context.Context, svc *compute.Service, project string, scope managedListScope) ([]*compute.InstanceGroupManager, error) {
	migs := []*compute.InstanceGroupManager{}
	for _, zone := range scope.zones {
		err := svc.InstanceGroupManagers.List(project, zone).Pages(ctx, func(resp *compute.InstanceGroupManagerList) error {
			migs = append(migs, resp.Items...)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("listing managed instance groups in zone %s: %w", zone, err)
		}
	}
	for _, region := range scope.regions {
		err := svc.RegionInstanceGroupManagers.List(project, region).Pages(ctx, func(resp *compute.RegionInstanceGroupManagerList) error {
			migs = append(migs, resp.Items...)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("listing managed instance groups in region %s: %w", region, err)
		}
	}
	if len(scope.zones) > 0 || len(scope.regions) > 0 {
		return migs, nil
	}
	err := svc.InstanceGroupManagers.AggregatedList(project).Pages(ctx, func(resp *compute.InstanceGroupManagerAggregatedList) error {
		for _, k := range managedListSortedScopes(resp.Items) {
			for _, m := range resp.Items[k].InstanceGroupManagers {
				if scope.keep(m) {
					migs = append(migs, m)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("listing managed instance groups: %w", err)
	}
	return migs, nil
}

// fetchAllInstanceGroups lists every zonal and regional instance group, which
// carry the current size of their managing MIG.
func fetchAllInstanceGroups(ctx context.Context, svc *compute.Service, project string) ([]*compute.InstanceGroup, error) {
	var groups []*compute.InstanceGroup
	err := svc.InstanceGroups.AggregatedList(project).Pages(ctx, func(resp *compute.InstanceGroupAggregatedList) error {
		for _, k := range managedListSortedScopes(resp.Items) {
			groups = append(groups, resp.Items[k].InstanceGroups...)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("listing instance groups: %w", err)
	}
	return groups, nil
}

// fetchAllAutoscalers lists every zonal and regional autoscaler.
func fetchAllAutoscalers(ctx context.Context, svc *compute.Service, project string) ([]*compute.Autoscaler, error) {
	var autoscalers []*compute.Autoscaler
	err := svc.Autoscalers.AggregatedList(project).Pages(ctx, func(resp *compute.AutoscalerAggregatedList) error {
		for _, k := range managedListSortedScopes(resp.Items) {
			autoscalers = append(autoscalers, resp.Items[k].Autoscalers...)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("listing autoscalers: %w", err)
	}
	return autoscalers, nil
}

// managedListSortedScopes returns the keys of an aggregated list's items
// ("regions/...", "zones/...") in a stable order, so output order doesn't
// depend on map iteration.
func managedListSortedScopes[V any](items map[string]V) []string {
	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// computeResourcePath strips the API host and version from a Compute URL,
// leaving "projects/...", so self links and autoscaler targets compare equal
// regardless of API version.
func computeResourcePath(url string) string {
	if i := strings.Index(url, "projects/"); i >= 0 {
		return url[i:]
	}
	return url
}

// managedListRecords converts migs into gcloud's list records: each MIG's own
// fields plus the "size" of its instance group, the "autoscaler" targeting it
// (if any) and "autoscaled" ("yes", "yes (*)" for an autoscaler in error, or
// "no"). hadErrors reports whether any autoscaler is in error. The result is
// never nil, so an empty list renders as [] in JSON.
func managedListRecords(migs []*compute.InstanceGroupManager, groups []*compute.InstanceGroup,
	autoscalers []*compute.Autoscaler) (records []map[string]any, hadErrors bool, err error) {
	sizes := make(map[string]int64, len(groups))
	for _, g := range groups {
		sizes[computeResourcePath(g.SelfLink)] = g.Size
	}
	byTarget := make(map[string]*compute.Autoscaler, len(autoscalers))
	for _, a := range autoscalers {
		byTarget[computeResourcePath(a.Target)] = a
	}

	records = []map[string]any{}
	for _, mig := range migs {
		rows, err := jsonRows(mig)
		if err != nil {
			return nil, false, fmt.Errorf("converting managed instance group %s: %w", mig.Name, err)
		}
		rec := rows[0]
		migPath := computeResourcePath(mig.SelfLink)
		rec["size"] = ""
		if size, ok := sizes[strings.Replace(migPath, "/instanceGroupManagers/", "/instanceGroups/", 1)]; ok {
			rec["size"] = strconv.FormatInt(size, 10)
		}
		rec["autoscaled"] = "no"
		if a, ok := byTarget[migPath]; ok {
			asRows, err := jsonRows(a)
			if err != nil {
				return nil, false, fmt.Errorf("converting autoscaler %s: %w", a.Name, err)
			}
			rec["autoscaler"] = asRows[0]
			rec["autoscaled"] = "yes"
			if a.Status == "ERROR" {
				rec["autoscaled"] = "yes (*)"
				hadErrors = true
			}
		}
		records = append(records, rec)
	}
	return records, hadErrors, nil
}

// managedListTableRows projects records onto the columns of gcloud's default
// table: the MIG's location and scope (zone or region) and the basename of its
// instance template.
func managedListTableRows(records []map[string]any) []map[string]any {
	rows := make([]map[string]any, 0, len(records))
	for _, rec := range records {
		location, scope := "", ""
		if z, _ := rec["zone"].(string); z != "" {
			location, scope = path.Base(z), "zone"
		} else if r, _ := rec["region"].(string); r != "" {
			location, scope = path.Base(r), "region"
		}
		targetSize, ok := rec["targetSize"]
		if !ok {
			// encoding/json omits a zero targetSize; gcloud shows 0.
			targetSize = 0
		}
		template, _ := rec["instanceTemplate"].(string)
		if template != "" {
			template = path.Base(template)
		}
		rows = append(rows, map[string]any{
			"name":             rec["name"],
			"location":         location,
			"scope":            scope,
			"baseInstanceName": rec["baseInstanceName"],
			"size":             rec["size"],
			"targetSize":       targetSize,
			"instanceTemplate": template,
			"autoscaled":       rec["autoscaled"],
		})
	}
	return rows
}
