package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/flyingobsidian/gcloud-go/internal/gfilter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// --- gcloud app-topology (#1908, GA in gcloud-python 585.0.0) ---
//
// The App Topology API (apptopology.googleapis.com/v1) is not exposed by
// google.golang.org/api, so the group talks to it through the shared
// restClient. Locations default to "global" for domains and resources-graph,
// matching gcloud-python's DefaultToGlobal fallthrough.

var appTopologyCmd = &cobra.Command{
	Use:   "app-topology",
	Short: "Manage App Topology resources and query topology graphs",
}

var (
	appTopologyDomainsCmd = &cobra.Command{Use: "domains", Short: "View App Topology domains"}
	appTopologySchemaCmd  = &cobra.Command{Use: "schema", Short: "View App Topology domain schemas"}
	appTopologyLocCmd     = &cobra.Command{Use: "locations", Short: "View App Topology locations"}
	appTopologyGraphCmd   = &cobra.Command{Use: "resources-graph", Short: "Query App Topology resource graphs"}
)

var appTopologyRest = newRESTClient("https://apptopology.googleapis.com/v1")

var (
	flagATLocation    string
	flagATFormat      string
	flagATLimit       int64
	flagATPageSize    int64
	flagATSortBy      []string
	flagATURI         bool
	flagATDomains     []string
	flagATPattern     string
	flagATPatternFile string
)

var (
	appTopologyDomainsListCmd = &cobra.Command{
		Use: "list", Short: "List App Topology domains",
		Args: cobra.NoArgs, RunE: runAppTopologyDomainsList,
	}
	appTopologyDomainsDescribeCmd = &cobra.Command{
		Use: "describe DOMAIN", Short: "Describe an App Topology domain",
		Args: cobra.ExactArgs(1), RunE: runAppTopologyDomainsDescribe,
	}
	appTopologySchemaDescribeCmd = &cobra.Command{
		Use: "describe DOMAIN", Short: "Describe the schema for an App Topology domain",
		Args: cobra.ExactArgs(1), RunE: runAppTopologySchemaDescribe,
	}
	appTopologyLocListCmd = &cobra.Command{
		Use: "list", Short: "List App Topology locations",
		Args: cobra.NoArgs, RunE: runAppTopologyLocList,
	}
	appTopologyLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe an App Topology location",
		Args: cobra.ExactArgs(1), RunE: runAppTopologyLocDescribe,
	}
	appTopologyGraphGenerateCmd = &cobra.Command{
		Use: "generate", Short: "Generate a resource topology graph",
		Args: cobra.NoArgs, RunE: runAppTopologyGraphGenerate,
	}
)

func init() {
	for _, c := range []*cobra.Command{appTopologyDomainsListCmd, appTopologyDomainsDescribeCmd,
		appTopologySchemaDescribeCmd, appTopologyGraphGenerateCmd} {
		c.Flags().StringVar(&flagATLocation, "location", "", `Location of the resource. Defaults to "global".`)
	}
	for _, c := range []*cobra.Command{appTopologyDomainsListCmd, appTopologyDomainsDescribeCmd,
		appTopologySchemaDescribeCmd, appTopologyLocListCmd, appTopologyLocDescribeCmd, appTopologyGraphGenerateCmd} {
		c.Flags().StringVar(&flagATFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{appTopologyDomainsListCmd, appTopologyLocListCmd} {
		c.Flags().Int64Var(&flagATLimit, "limit", 0, "Maximum number of resources to list")
		c.Flags().Int64Var(&flagATPageSize, "page-size", 0, "Number of resources to fetch per page")
		c.Flags().StringSliceVar(&flagATSortBy, "sort-by", nil,
			"Comma-separated list of resource field keys to sort by (prefix with ~ to reverse)")
		c.Flags().BoolVar(&flagATURI, "uri", false, "Print a list of resource URIs instead of the default output")
	}
	f := appTopologyGraphGenerateCmd.Flags()
	f.StringSliceVar(&flagATDomains, "domains", nil,
		"List of App Topology domains to include in the topology query (e.g., SRE, DEVOPS, SECURITY)")
	_ = appTopologyGraphGenerateCmd.MarkFlagRequired("domains")
	f.StringVar(&flagATPattern, "pattern", "",
		"A GraphPattern query filter definition in proto3 YAML/JSON string format")
	f.StringVar(&flagATPatternFile, "pattern-file", "",
		`Path to a YAML or JSON file (or "-" for standard input) containing the GraphPattern query filter definition`)
	appTopologyGraphGenerateCmd.MarkFlagsMutuallyExclusive("pattern", "pattern-file")

	appTopologySchemaCmd.AddCommand(appTopologySchemaDescribeCmd)
	appTopologyDomainsCmd.AddCommand(appTopologyDomainsListCmd, appTopologyDomainsDescribeCmd, appTopologySchemaCmd)
	appTopologyLocCmd.AddCommand(appTopologyLocListCmd, appTopologyLocDescribeCmd)
	appTopologyGraphCmd.AddCommand(appTopologyGraphGenerateCmd)
	appTopologyCmd.AddCommand(appTopologyDomainsCmd, appTopologyLocCmd, appTopologyGraphCmd)
	rootCmd.AddCommand(appTopologyCmd)
}

// appTopologyLocationName resolves projects/P/locations/L from a location
// value, defaulting to "global" when empty.
func appTopologyLocationName(location string) (string, error) {
	if strings.HasPrefix(location, "projects/") {
		return location, nil
	}
	if location == "" {
		location = "global"
	}
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, location), nil
}

// appTopologyDomainName resolves projects/P/locations/L/domains/D.
func appTopologyDomainName(domain string) (string, error) {
	if strings.HasPrefix(domain, "projects/") {
		return domain, nil
	}
	loc, err := appTopologyLocationName(flagATLocation)
	if err != nil {
		return "", err
	}
	return loc + "/domains/" + domain, nil
}

func runAppTopologyDomainsDescribe(cmd *cobra.Command, args []string) error {
	name, err := appTopologyDomainName(args[0])
	if err != nil {
		return err
	}
	var got map[string]any
	if err := appTopologyRest.do(context.Background(), http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing domain: %w", err)
	}
	return emitFormatted(got, flagATFormat)
}

// appTopologySchemaName appends "/schema" to a domain resource name unless
// it already names the schema subresource (7 path segments), so a domain
// literally called "schema" still resolves correctly.
func appTopologySchemaName(domain string) string {
	base := strings.TrimRight(domain, "/")
	if len(strings.Split(base, "/")) == 6 {
		return base + "/schema"
	}
	return base
}

func runAppTopologySchemaDescribe(cmd *cobra.Command, args []string) error {
	name, err := appTopologyDomainName(args[0])
	if err != nil {
		return err
	}
	var got map[string]any
	if err := appTopologyRest.do(context.Background(), http.MethodGet, "/"+appTopologySchemaName(name), nil, nil, &got); err != nil {
		return fmt.Errorf("describing domain schema: %w", err)
	}
	return emitFormatted(got, flagATFormat)
}

func runAppTopologyLocDescribe(cmd *cobra.Command, args []string) error {
	name, err := appTopologyLocationName(args[0])
	if err != nil {
		return err
	}
	var got map[string]any
	if err := appTopologyRest.do(context.Background(), http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormatted(got, flagATFormat)
}

func runAppTopologyDomainsList(cmd *cobra.Command, args []string) error {
	parent, err := appTopologyLocationName(flagATLocation)
	if err != nil {
		return err
	}
	all, err := appTopologyRest.paginate(context.Background(), "/"+parent+"/domains", nil, "domains", flagATPageSize)
	if err != nil {
		return fmt.Errorf("listing domains: %w", err)
	}
	return emitAppTopologyList(os.Stdout, os.Stderr, all, "table(name)", func(r map[string]any) map[string]any {
		// name.scope("domains"): the part of the name after ".../domains/".
		name, _ := r["name"].(string)
		if i := strings.LastIndex(name, "/domains/"); i >= 0 {
			name = name[i+len("/domains/"):]
		}
		return map[string]any{"name": name}
	})
}

func runAppTopologyLocList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	parent := project
	if !strings.HasPrefix(parent, "projects/") {
		parent = "projects/" + parent
	}
	all, err := appTopologyRest.paginate(context.Background(), "/"+parent+"/locations", nil, "locations", flagATPageSize)
	if err != nil {
		return fmt.Errorf("listing locations: %w", err)
	}
	return emitAppTopologyList(os.Stdout, os.Stderr, all, "table(locationId,name)", nil)
}

// emitAppTopologyList applies the gcloud list pipeline (--filter, --sort-by,
// --limit, then --uri or --format) to records. defaultTable is used when no
// --format is given, after mapping each record through tableRow if non-nil.
func emitAppTopologyList(w, errW io.Writer, records []map[string]any, defaultTable string, tableRow func(map[string]any) map[string]any) error {
	if flagFilter != "" {
		filter, err := gfilter.Compile(flagFilter)
		if err != nil {
			return fmt.Errorf("invalid --filter expression: %w", err)
		}
		kept := records[:0]
		for _, r := range records {
			if filter.Match(r) {
				kept = append(kept, r)
			}
		}
		records = kept
	}
	sortRecordsBy(records, flagATSortBy)
	if flagATLimit > 0 && int64(len(records)) > flagATLimit {
		records = records[:flagATLimit]
	}
	if flagATURI {
		for _, r := range records {
			name, _ := r["name"].(string)
			fmt.Fprintln(w, "https://apptopology.googleapis.com/v1/"+name)
		}
		return nil
	}
	format := flagATFormat
	if format == "" {
		format = flagFormat
	}
	if format == "" {
		if len(records) == 0 {
			_, err := fmt.Fprintln(errW, "Listed 0 items.")
			return err
		}
		rows := records
		if tableRow != nil {
			rows = make([]map[string]any, len(records))
			for i, r := range records {
				rows[i] = tableRow(r)
			}
		}
		return emitFormattedTo(w, rows, defaultTable)
	}
	if records == nil {
		records = []map[string]any{}
	}
	return emitFormattedTo(w, records, format)
}

// sortRecordsBy stably sorts records by the given --sort-by keys; a leading
// "~" reverses that key. Values compare as their string rendering.
func sortRecordsBy(records []map[string]any, keys []string) {
	for i := len(keys) - 1; i >= 0; i-- {
		key, desc := strings.CutPrefix(keys[i], "~")
		sort.SliceStable(records, func(a, b int) bool {
			va, vb := fieldValue(records[a], key), fieldValue(records[b], key)
			if desc {
				return va > vb
			}
			return va < vb
		})
	}
}

var appTopologyProjectsDomainRE = regexp.MustCompile(`projects/[^/]+/locations/[^/]+/domains/[^/]+`)
var appTopologyURLPrefixRE = regexp.MustCompile(`^(https?://[^/]+/(?:v[0-9a-zA-Z_]+/)?|//[^/]+/)`)

// appTopologyNormalizeDomains converts --domains values (bare names,
// "domains/D", "locations/L/domains/D", full names or URLs) into fully
// qualified domain names under location, deduplicating while preserving
// order, as gcloud-python's resources-graph generate does.
func appTopologyNormalizeDomains(raw []string, location string) []string {
	project := strings.SplitN(strings.TrimPrefix(location, "projects/"), "/", 2)[0]
	var out []string
	seen := map[string]bool{}
	for _, d := range raw {
		d = strings.Trim(strings.TrimSpace(d), "/")
		if d == "" {
			continue
		}
		if strings.HasPrefix(d, "http://") || strings.HasPrefix(d, "https://") || strings.HasPrefix(d, "//") {
			if m := appTopologyProjectsDomainRE.FindString(d); m != "" {
				d = m
			} else if i := strings.Index(d, "domains/"); strings.Contains(d, "/domains/") && i >= 0 {
				d = d[i:]
			} else {
				d = appTopologyURLPrefixRE.ReplaceAllString(d, "")
			}
		}
		d = strings.Trim(d, "/")
		d = strings.TrimPrefix(d, "domains/")
		switch {
		case strings.HasPrefix(d, "locations/"):
			d = "projects/" + project + "/" + d
		case !strings.HasPrefix(d, "projects/"):
			d = location + "/domains/" + d
		}
		if !seen[d] {
			seen[d] = true
			out = append(out, d)
		}
	}
	return out
}

// appTopologySnakeToCamel normalizes snake_case, kebab-case,
// SCREAMING_SNAKE_CASE and UpperCamelCase keys to lowerCamelCase.
func appTopologySnakeToCamel(key string) string {
	if key == "" {
		return key
	}
	if strings.ContainsAny(key, "_-") {
		var parts []string
		for _, p := range strings.Split(strings.ReplaceAll(key, "-", "_"), "_") {
			if p != "" {
				parts = append(parts, p)
			}
		}
		if len(parts) == 0 {
			return key
		}
		var b strings.Builder
		b.WriteString(strings.ToLower(parts[0]))
		for _, p := range parts[1:] {
			r := []rune(strings.ToLower(p))
			r[0] = unicode.ToUpper(r[0])
			b.WriteString(string(r))
		}
		return b.String()
	}
	if strings.ToUpper(key) == key && strings.ToLower(key) != key {
		return strings.ToLower(key)
	}
	r := []rune(key)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

// appTopologyNormalizePattern recursively camel-cases keys and upper-cases
// "direction" enum values in a GraphPattern parsed from YAML/JSON.
func appTopologyNormalizePattern(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			if s, ok := val.(string); ok && strings.ToLower(k) == "direction" {
				out[appTopologySnakeToCamel(k)] = strings.ReplaceAll(strings.ToUpper(strings.TrimSpace(s)), "-", "_")
				continue
			}
			out[appTopologySnakeToCamel(k)] = appTopologyNormalizePattern(val)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			out[i] = appTopologyNormalizePattern(e)
		}
		return out
	}
	return v
}

// appTopologyParsePattern reads --pattern / --pattern-file into a normalized
// GraphPattern object. Empty or comment-only input yields nil.
func appTopologyParsePattern(pattern, patternFile string, stdin io.Reader) (map[string]any, error) {
	content := pattern
	if patternFile != "" {
		var data []byte
		var err error
		if patternFile == "-" {
			data, err = io.ReadAll(stdin)
		} else {
			data, err = os.ReadFile(patternFile)
		}
		if err != nil {
			return nil, fmt.Errorf("Failed to read pattern file [%s]: %v", patternFile, err)
		}
		content = string(data)
	}
	if strings.TrimSpace(content) == "" {
		return nil, nil
	}
	var parsed any
	if err := yaml.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, fmt.Errorf("Failed to parse GraphPattern filter: %v", err)
	}
	if parsed == nil {
		return nil, nil
	}
	m, ok := parsed.(map[string]any)
	if !ok {
		return nil, errors.New("Invalid filter format: Expected YAML/JSON object for GraphPattern.")
	}
	return appTopologyNormalizePattern(m).(map[string]any), nil
}

func runAppTopologyGraphGenerate(cmd *cobra.Command, args []string) error {
	location, err := appTopologyLocationName(flagATLocation)
	if err != nil {
		return err
	}
	domains := appTopologyNormalizeDomains(flagATDomains, location)
	if len(domains) == 0 {
		return errors.New("Invalid value for [--domains]: At least one non-empty domain name must be specified.")
	}
	filter, err := appTopologyParsePattern(flagATPattern, flagATPatternFile, os.Stdin)
	if err != nil {
		return err
	}
	body := map[string]any{"topologyDomains": domains}
	if filter != nil {
		body["filter"] = filter
	}
	var resp map[string]any
	path := "/" + location + "/discoveredResourcesTopology:generate"
	if err := appTopologyRest.do(context.Background(), http.MethodPost, path, url.Values{}, body, &resp); err != nil {
		return fmt.Errorf("generating resource topology: %w", err)
	}
	graph := resp["graph"]
	if graph == nil {
		return nil
	}
	return emitFormatted(graph, flagATFormat)
}
