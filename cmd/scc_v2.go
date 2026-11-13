package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
)

// sccV2ParentPath builds the V2-style parent used for findings list:
//
//	organizations/{org}/sources/{source}/locations/{location}
//	folders/{folder}/sources/{source}/locations/{location}
//	projects/{project}/sources/{source}/locations/{location}
//
// `parent` here is the scope (organizations/N | folders/N | projects/P) and
// `source` is "-" (all sources) or a numeric source id / full path.
func sccV2ParentPath(parent, source, location string) string {
	// If the caller passed a fully-qualified sources/... string, honour it.
	if strings.Contains(source, "/") && strings.HasPrefix(source, "sources/") {
		return parent + "/" + source + "/locations/" + location
	}
	if source == "" {
		source = "-"
	}
	return fmt.Sprintf("%s/sources/%s/locations/%s", parent, source, location)
}

// sccV2ScopePath builds the V2-style parent used by collection-level methods
// that are not scoped to a source:
//
//	organizations/{org}/locations/{location}
//	folders/{folder}/locations/{location}
//	projects/{project}/locations/{location}
//
// This mirrors ValidateLocationAndGetRegionalizedParent in gcloud-python.
func sccV2ScopePath(parent, location string) string {
	if strings.HasPrefix(location, "locations/") {
		return parent + "/" + location
	}
	return parent + "/locations/" + location
}

// sccMuteStateEnum maps the --mute-state choice onto the enum both API
// versions expect. gcloud-python accepts "muted" and "undefined" only (see
// ConvertMuteStateInput), case-insensitively via base.ChoiceArgument.
func sccMuteStateEnum(muteState string) (string, error) {
	switch strings.ToLower(muteState) {
	case "", "muted":
		return "MUTED", nil
	case "undefined":
		return "UNDEFINED", nil
	default:
		return "", fmt.Errorf("--mute-state must be one of [muted, undefined], got %q", muteState)
	}
}

// sccV2BulkMuteFindings performs
// `POST /v2/{parent}/findings:bulkMute`, which mutes every finding matching
// the filter server-side and returns a long-running operation. It is the only
// bulk write in the findings surface: there is no bulk state change.
func (g *sccCmds) sccV2BulkMuteFindings(parent string) error {
	if g.inv.quotaProject() == "" {
		return errNoBillingProject("the SCC V2 API")
	}
	muteState, err := sccMuteStateEnum(g.flagSccFindingBulkMuteState)
	if err != nil {
		return err
	}
	scopePath := sccV2ScopePath(parent, g.flagSccFindingLocation)
	body := map[string]any{
		"filter":    g.flagSccFindingBulkMuteFilter,
		"muteState": muteState,
	}

	ctx := g.inv.context()
	var op map[string]any
	if err := g.sccV2Rest.do(ctx, http.MethodPost, "/"+scopePath+"/findings:bulkMute", nil, body, &op); err != nil {
		return fmt.Errorf("bulk muting findings (V2): %w", err)
	}
	return emitFormattedTo(g.inv.stdout, op, g.flagSccFormat)
}

// sccDatasetPattern is the shape gcloud-python validates --dataset against
// (ValidateDataset in command_lib/scc/findings/util.py).
var sccDatasetPattern = regexp.MustCompile(`^projects/[a-zA-Z0-9-]+/datasets/[a-zA-Z0-9_]+$`)

// sccV2ExportFindings performs `POST /v2/{parent}/findings:export`, exporting
// the findings under a source to a BigQuery dataset. Unlike the other findings
// commands this one has no V1 equivalent: gcloud-python pins it to V2.
func (g *sccCmds) sccV2ExportFindings(parent string) error {
	if g.inv.quotaProject() == "" {
		return errNoBillingProject("the SCC V2 API")
	}
	if !sccDatasetPattern.MatchString(g.flagSccFindingDataset) {
		return fmt.Errorf(
			"--dataset must match projects/[a-zA-Z0-9-]+/datasets/[a-zA-Z0-9_]+, got %q",
			g.flagSccFindingDataset)
	}
	parentPath := sccV2ParentPath(parent, g.flagSccFindingSource, g.flagSccFindingLocation)
	body := map[string]any{
		"bigQueryDestination": map[string]any{"dataset": g.flagSccFindingDataset},
	}

	ctx := g.inv.context()
	var resp map[string]any
	if err := g.sccV2Rest.do(ctx, http.MethodPost, "/"+parentPath+"/findings:export", nil, body, &resp); err != nil {
		return fmt.Errorf("exporting findings to BigQuery: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, resp, g.flagSccFormat)
}

// sccV2ListFindings performs `GET /v2/{parent}/findings` and prints the
// results using the same table/format branches as the V1 path.
func (g *sccCmds) sccV2ListFindings(parent string) error {
	// SCC V2 rejects unauthenticated ADC user creds with PERMISSION_DENIED
	// unless a quota project is set (#1740). Resolve it up-front so the
	// user gets a clear message rather than an opaque server error; the
	// value is transparently forwarded as x-goog-user-project by the
	// shared restClient. Goes through restUserProject so tests can inject
	// a synthetic project without touching global config or ADC.
	if g.inv.quotaProject() == "" {
		return errNoBillingProject("the SCC V2 API")
	}
	parentPath := sccV2ParentPath(parent, g.flagSccFindingSource, g.flagSccFindingLocation)

	q := url.Values{}
	if g.flagSccFilter != "" {
		q.Set("filter", g.flagSccFilter)
	}
	if g.flagSccOrderBy != "" {
		q.Set("orderBy", g.flagSccOrderBy)
	}
	if g.flagSccFindingFieldMask != "" {
		q.Set("fieldMask", g.flagSccFindingFieldMask)
	}
	// The V2 findings.list endpoint does not accept compareDuration /
	// readTime (those were V1-only). If the caller provided them we return
	// a clear error rather than silently dropping them.
	if g.flagSccFindingCompareDur != "" {
		return fmt.Errorf("--compare-duration is not supported by the SCC V2 API (--location=%q)", g.flagSccFindingLocation)
	}
	if g.flagSccFindingReadTime != "" {
		return fmt.Errorf("--read-time is not supported by the SCC V2 API (--location=%q)", g.flagSccFindingLocation)
	}

	ctx := g.inv.context()
	results, err := g.sccV2Rest.paginate(ctx, "/"+parentPath+"/findings", q, "listFindingsResults", g.flagSccPageSize)
	if err != nil {
		return fmt.Errorf("listing findings (V2): %w", err)
	}

	if g.flagSccFormat != "" {
		return emitFormattedTo(g.inv.stdout, results, g.flagSccFormat)
	}
	// Default table matching the V1 codepath.
	printTableHeadingTo(g.inv.stdout, len(results), "%-40s %-10s %s\n", "NAME", "STATE", "CATEGORY")
	for _, r := range results {
		finding, _ := r["finding"].(map[string]any)
		if finding == nil {
			continue
		}
		name, _ := finding["name"].(string)
		state, _ := finding["state"].(string)
		category, _ := finding["category"].(string)
		fmt.Fprintf(g.inv.stdout, "%-40s %-10s %s\n", path.Base(name), state, category)
	}
	return nil
}

// sccV2RestClientForTest is exposed for tests that want to override the
// endpoint (e.g. point at an httptest.Server).
func (g *sccCmds) sccV2RestClientForTest(endpoint string) {
	g.sccV2Rest = g.inv.restClient(endpoint)
}
