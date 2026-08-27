package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// sccV2Rest is the REST client for the SCC V2 API. google.golang.org/api
// (through v0.293.0 at the time of writing) does not ship a generated
// securitycenter/v2 client, so V2 endpoints are called directly. Requests
// automatically pick up the caller's ADC / --account credentials via the
// shared restClient.
var sccV2Rest = newRESTClient("https://securitycenter.googleapis.com/v2")

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

// sccV2ListFindings performs `GET /v2/{parent}/findings` and prints the
// results using the same table/format branches as the V1 path.
func sccV2ListFindings(parent string) error {
	// SCC V2 rejects unauthenticated ADC user creds with PERMISSION_DENIED
	// unless a quota project is set (#1740). Resolve it up-front so the
	// user gets a clear message rather than an opaque server error; the
	// value is transparently forwarded as x-goog-user-project by the
	// shared restClient. Goes through restUserProject so tests can inject
	// a synthetic project without touching global config or ADC.
	if restUserProject() == "" {
		return errNoBillingProject("the SCC V2 API")
	}
	parentPath := sccV2ParentPath(parent, flagSccFindingSource, flagSccFindingLocation)

	q := url.Values{}
	if flagSccFilter != "" {
		q.Set("filter", flagSccFilter)
	}
	if flagSccOrderBy != "" {
		q.Set("orderBy", flagSccOrderBy)
	}
	if flagSccFindingFieldMask != "" {
		q.Set("fieldMask", flagSccFindingFieldMask)
	}
	// The V2 findings.list endpoint does not accept compareDuration /
	// readTime (those were V1-only). If the caller provided them we return
	// a clear error rather than silently dropping them.
	if flagSccFindingCompareDur != "" {
		return fmt.Errorf("--compare-duration is not supported by the SCC V2 API (--location=%q)", flagSccFindingLocation)
	}
	if flagSccFindingReadTime != "" {
		return fmt.Errorf("--read-time is not supported by the SCC V2 API (--location=%q)", flagSccFindingLocation)
	}

	ctx := context.Background()
	results, err := sccV2Rest.paginate(ctx, "/"+parentPath+"/findings", q, "listFindingsResults", flagSccPageSize)
	if err != nil {
		return fmt.Errorf("listing findings (V2): %w", err)
	}

	if flagSccFormat != "" {
		return emitFormatted(results, flagSccFormat)
	}
	// Default table matching the V1 codepath.
	fmt.Printf("%-40s %-10s %s\n", "NAME", "STATE", "CATEGORY")
	for _, r := range results {
		finding, _ := r["finding"].(map[string]any)
		if finding == nil {
			continue
		}
		name, _ := finding["name"].(string)
		state, _ := finding["state"].(string)
		category, _ := finding["category"].(string)
		fmt.Printf("%-40s %-10s %s\n", path.Base(name), state, category)
	}
	return nil
}

// sccV2RestClientForTest is exposed for tests that want to override the
// endpoint (e.g. point at an httptest.Server).
func sccV2RestClientForTest(endpoint string) {
	sccV2Rest = newRESTClient(endpoint)
}

// dummy import guard: net/http is imported so tests in the same file that
// intercept requests can reference it without pulling it in themselves.
var _ = http.MethodGet
