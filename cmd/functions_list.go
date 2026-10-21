package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudfunctionsv1 "google.golang.org/api/cloudfunctions/v1"
	cloudfunctions "google.golang.org/api/cloudfunctions/v2"
)

// --- gcloud functions list (#2004) ---
//
// Mirrors surface/functions/list.py (GA): 2nd gen functions from the v2 API,
// then 1st gen functions from the v1 API, unless --v2 asks for the v2 API
// alone.

var (
	flagFuncListRegions []string
	flagFuncListV2      bool
	flagFuncListLimit   int64
)

func init() {
	list := &cobra.Command{
		Use:   "list",
		Short: "List Google Cloud Functions",
		Args:  cobra.NoArgs,
		RunE:  runFuncList,
	}
	list.Flags().StringSliceVar(&flagFuncListRegions, "regions", []string{"-"},
		"Regions containing functions to list; default all regions")
	list.Flags().BoolVar(&flagFuncListV2, "v2", false, "List functions using the Cloud Functions v2 API only")
	list.Flags().Int64Var(&flagFuncListLimit, "limit", 0, "Maximum number of functions to list (0 = no limit)")
	functionsCmd.AddCommand(list)
}

// funcUnreachableWarning is gcloud's warning for regions a list could not
// reach.
func funcUnreachableWarning(w io.Writer, unreachable []string) {
	if len(unreachable) == 0 {
		return
	}
	fmt.Fprintf(w, "WARNING: The following regions were fully or partially unreachable for query: %s\n"+
		"This could be due to permission setup. Additional informationcan be found in: "+
		"https://cloud.google.com/functions/docs/troubleshooting\n", strings.Join(unreachable, ", "))
}

// funcV1Regions keeps the regions the v1 API can list: "-" (all regions) and
// regions that v1 supports, as gcloud does before listing 1st gen functions.
func funcV1Regions(regions, v1Supported []string) []string {
	supported := make(map[string]bool, len(v1Supported))
	for _, r := range v1Supported {
		supported[r] = true
	}
	kept := []string{}
	for _, r := range regions {
		if r == "-" || supported[r] {
			kept = append(kept, r)
		}
	}
	return kept
}

func runFuncList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	v2, err := gcp.CloudFunctionsV2Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	listV2 := func(filter string) ([]any, error) {
		var out []any
		for _, region := range flagFuncListRegions {
			call := v2.Projects.Locations.Functions.List(fmt.Sprintf("projects/%s/locations/%s", project, region))
			if filter != "" {
				call = call.Filter(filter)
			}
			err := call.Pages(ctx, func(r *cloudfunctions.ListFunctionsResponse) error {
				funcUnreachableWarning(os.Stderr, r.Unreachable)
				for _, f := range r.Functions {
					out = append(out, f)
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("listing functions: %w", err)
			}
		}
		return out, nil
	}

	all := []any{}
	if flagFuncListV2 {
		items, err := listV2("")
		if err != nil {
			return err
		}
		all = append(all, items...)
	} else {
		items, err := listV2(`environment="GEN_2"`)
		if err != nil {
			return err
		}
		all = append(all, items...)

		v1, err := gcp.CloudFunctionsV1Service(ctx, flagAccount)
		if err != nil {
			return err
		}
		var v1Supported []string
		err = v1.Projects.Locations.List("projects/"+project).Pages(ctx, func(r *cloudfunctionsv1.ListLocationsResponse) error {
			for _, l := range r.Locations {
				v1Supported = append(v1Supported, l.LocationId)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("listing Cloud Functions v1 regions: %w", err)
		}
		for _, region := range funcV1Regions(flagFuncListRegions, v1Supported) {
			err := v1.Projects.Locations.Functions.List(fmt.Sprintf("projects/%s/locations/%s", project, region)).Pages(ctx,
				func(r *cloudfunctionsv1.ListFunctionsResponse) error {
					funcUnreachableWarning(os.Stderr, r.Unreachable)
					for _, f := range r.Functions {
						all = append(all, f)
					}
					return nil
				})
			if err != nil {
				return fmt.Errorf("listing 1st gen functions: %w", err)
			}
		}
	}
	if flagFuncListLimit > 0 && int64(len(all)) > flagFuncListLimit {
		all = all[:flagFuncListLimit]
	}
	return emitFuncList(os.Stdout, os.Stderr, all, flagFormat)
}

// funcListFormat is gcloud's default `functions list` table, over the rows
// built by funcListTableRow.
const funcListFormat = "table(name,state,trigger,region,environment)"

// emitFuncList prints functions with format, or gcloud's default table sorted
// by name ("Listed 0 items." on errW when empty).
func emitFuncList(w, errW io.Writer, functions []any, format string) error {
	if format != "" {
		return emitFormattedTo(w, functions, format)
	}
	if len(functions) == 0 {
		_, err := fmt.Fprintln(errW, "Listed 0 items.")
		return err
	}
	records, err := jsonRows(functions)
	if err != nil {
		return err
	}
	rows := make([]map[string]any, 0, len(records))
	for _, rec := range records {
		rows = append(rows, funcListTableRow(rec))
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i]["name"].(string) < rows[j]["name"].(string) })
	return emitFormattedTo(w, rows, funcListFormat)
}

// funcListTableRow applies gcloud's functions transforms (state(),
// trigger(), generation()) to a v1 CloudFunction or v2 Function record.
func funcListTableRow(rec map[string]any) map[string]any {
	name, _ := rec["name"].(string)
	region := ""
	if _, rest, ok := strings.Cut(name, "/locations/"); ok {
		region, _, _ = strings.Cut(rest, "/")
	}
	return map[string]any{
		"name":        path.Base(name),
		"state":       funcState(rec),
		"trigger":     funcTrigger(rec),
		"region":      region,
		"environment": funcGeneration(rec),
	}
}

// funcState is gcloud's state(): v1 status, else v2 state.
func funcState(rec map[string]any) string {
	if s, ok := rec["status"].(string); ok {
		return s
	}
	s, _ := rec["state"].(string)
	return s
}

// funcMessageFormat is gcloud's _InferFunctionMessageFormat: "CloudFunction"
// for v1 records, "Function" for v2, "" when it can't tell.
func funcMessageFormat(rec map[string]any) string {
	for _, k := range []string{"entryPoint", "buildId", "runtime"} {
		if v, _ := rec[k].(string); v != "" {
			return "CloudFunction"
		}
	}
	for _, k := range []string{"buildConfig", "serviceConfig"} {
		if v, ok := rec[k]; ok && v != nil {
			return "Function"
		}
	}
	return ""
}

// funcGeneration is gcloud's generation(): "1st gen" or "2nd gen".
func funcGeneration(rec map[string]any) string {
	switch env, _ := rec["environment"].(string); env {
	case "GEN_1":
		return "1st gen"
	case "GEN_2":
		return "2nd gen"
	}
	switch funcMessageFormat(rec) {
	case "CloudFunction":
		return "1st gen"
	case "Function":
		return "2nd gen"
	}
	return "-"
}

// funcStorageEventTypes are the Cloud Storage event types gcloud's trigger()
// shows as a bucket.
var funcStorageEventTypes = map[string]bool{
	"google.cloud.storage.object.v1.archived":        true,
	"google.cloud.storage.object.v1.deleted":         true,
	"google.cloud.storage.object.v1.finalized":       true,
	"google.cloud.storage.object.v1.metadataUpdated": true,
}

// funcTrigger is gcloud's trigger(): a short description of how the function
// is invoked.
func funcTrigger(rec map[string]any) string {
	switch funcMessageFormat(rec) {
	case "CloudFunction":
		switch {
		case rec["httpsTrigger"] != nil:
			return "HTTP Trigger"
		case rec["gcsTrigger"] != nil:
			bucket, _ := rec["gcsTrigger"].(string)
			return "bucket: " + bucket
		case rec["pubsubTrigger"] != nil:
			topic, _ := rec["pubsubTrigger"].(string)
			return "topic: " + path.Base(topic)
		case rec["eventTrigger"] != nil:
			return "Event Trigger"
		}
		return ""
	case "Function":
		trigger, ok := rec["eventTrigger"].(map[string]any)
		if !ok {
			return "HTTP Trigger"
		}
		eventType, _ := trigger["eventType"].(string)
		if eventType == "google.cloud.audit.log.v1.written" {
			return "Cloud Audit Log"
		}
		if funcStorageEventTypes[eventType] {
			filters, _ := trigger["eventFilters"].([]any)
			for _, f := range filters {
				if m, _ := f.(map[string]any); m["attribute"] == "bucket" {
					if bucket, _ := m["value"].(string); bucket != "" {
						return "bucket: " + bucket
					}
				}
			}
		}
		if topic, _ := trigger["pubsubTopic"].(string); topic != "" {
			return "topic: " + path.Base(topic)
		}
		return "Event Trigger"
	}
	return ""
}
