package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// --- gcloud pam entitlements (#962) ---
//
// The Privileged Access Manager REST API is not covered by the generated
// google.golang.org/api Go client, so this command surface talks directly to
// https://privilegedaccessmanager.googleapis.com/v1.

var pamEntitlementsCmd = &cobra.Command{
	Use:   "entitlements",
	Short: "Manage Privileged Access Manager entitlements",
}

var (
	flagPAMFolder            string
	flagPAMOrganization      string
	flagPAMLocation          string
	flagPAMEntitlementFile   string
	flagPAMDestination       string
	flagPAMCallerAccessType  string
	flagPAMFormat            string
	flagPAMFilter            string
	flagPAMOrderBy           string
	flagPAMPageSize          int64
)

var (
	pamEntCreateCmd = &cobra.Command{
		Use: "create ENTITLEMENT", Short: "Create a Privileged Access Manager entitlement",
		Args: cobra.ExactArgs(1), RunE: runPAMEntCreate,
	}
	pamEntDeleteCmd = &cobra.Command{
		Use: "delete ENTITLEMENT", Short: "Delete a Privileged Access Manager entitlement",
		Args: cobra.ExactArgs(1), RunE: runPAMEntDelete,
	}
	pamEntDescribeCmd = &cobra.Command{
		Use: "describe ENTITLEMENT", Short: "Show details of a Privileged Access Manager entitlement",
		Args: cobra.ExactArgs(1), RunE: runPAMEntDescribe,
	}
	pamEntExportCmd = &cobra.Command{
		Use: "export ENTITLEMENT", Short: "Export a Privileged Access Manager entitlement to a local YAML file",
		Args: cobra.ExactArgs(1), RunE: runPAMEntExport,
	}
	pamEntListCmd = &cobra.Command{
		Use: "list", Short: "List Privileged Access Manager entitlements under a parent",
		Args: cobra.NoArgs, RunE: runPAMEntList,
	}
	pamEntSearchCmd = &cobra.Command{
		Use: "search", Short: "Search Privileged Access Manager entitlements you can request/approve",
		Args: cobra.NoArgs, RunE: runPAMEntSearch,
	}
	pamEntUpdateCmd = &cobra.Command{
		Use: "update ENTITLEMENT", Short: "Update an existing Privileged Access Manager entitlement",
		Args: cobra.ExactArgs(1), RunE: runPAMEntUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		pamEntCreateCmd, pamEntDeleteCmd, pamEntDescribeCmd, pamEntExportCmd,
		pamEntListCmd, pamEntSearchCmd, pamEntUpdateCmd,
	}
	for _, c := range all {
		addPAMScopeFlags(c)
		c.Flags().StringVar(&flagPAMLocation, "location", "global",
			"Location that owns the entitlements (e.g. \"global\")")
	}
	for _, c := range []*cobra.Command{pamEntCreateCmd, pamEntUpdateCmd} {
		c.Flags().StringVar(&flagPAMEntitlementFile, "entitlement-file", "",
			"Path to a YAML/JSON file containing the entitlement configuration (required)")
		_ = c.MarkFlagRequired("entitlement-file")
	}
	pamEntExportCmd.Flags().StringVar(&flagPAMDestination, "destination", "",
		"Path to the local YAML file to write (defaults to stdout)")
	pamEntSearchCmd.Flags().StringVar(&flagPAMCallerAccessType, "caller-access-type", "",
		"Search access type: grant-requester or grant-approver (required)")
	_ = pamEntSearchCmd.MarkFlagRequired("caller-access-type")
	pamEntListCmd.Flags().StringVar(&flagPAMFilter, "filter", "", "Server-side filter expression")
	pamEntListCmd.Flags().StringVar(&flagPAMOrderBy, "order-by", "", "Server-side ordering hint")
	pamEntListCmd.Flags().Int64Var(&flagPAMPageSize, "page-size", 0, "Maximum number of results per page")
	pamEntSearchCmd.Flags().Int64Var(&flagPAMPageSize, "page-size", 0, "Maximum number of results per page")
	for _, c := range []*cobra.Command{pamEntCreateCmd, pamEntDeleteCmd, pamEntDescribeCmd, pamEntListCmd, pamEntSearchCmd, pamEntUpdateCmd} {
		c.Flags().StringVar(&flagPAMFormat, "format", "", "Output format")
	}

	pamEntitlementsCmd.AddCommand(all...)
	pamCmd.AddCommand(pamEntitlementsCmd)
}

func addPAMScopeFlags(c *cobra.Command) {
	c.Flags().StringVar(&flagPAMFolder, "folder", "",
		"Folder scope (alternative to --project/--organization)")
	c.Flags().StringVar(&flagPAMOrganization, "organization", "",
		"Organization scope (alternative to --project/--folder)")
}

// pamScopeParent returns "projects/PROJECT", "folders/FOLDER" or "organizations/ORG".
func pamScopeParent() (string, error) {
	set := 0
	if flagProject != "" {
		set++
	}
	if flagPAMFolder != "" {
		set++
	}
	if flagPAMOrganization != "" {
		set++
	}
	if set > 1 {
		return "", fmt.Errorf("only one of --project, --folder, --organization may be set")
	}
	if flagPAMFolder != "" {
		return "folders/" + flagPAMFolder, nil
	}
	if flagPAMOrganization != "" {
		return "organizations/" + flagPAMOrganization, nil
	}
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return "projects/" + project, nil
}

func pamEntParent() (string, error) {
	scope, err := pamScopeParent()
	if err != nil {
		return "", err
	}
	loc := flagPAMLocation
	if loc == "" {
		loc = "global"
	}
	return fmt.Sprintf("%s/locations/%s", scope, loc), nil
}

func pamEntName(id string) (string, error) {
	if strings.Contains(id, "/entitlements/") {
		return id, nil
	}
	parent, err := pamEntParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/entitlements/%s", parent, id), nil
}

// pamCallerAccessValue maps the CLI-facing --caller-access-type flag value to
// the enum expected by the Privileged Access Manager REST API.
func pamCallerAccessValue(v string) (string, error) {
	switch strings.ToLower(v) {
	case "grant-requester":
		return "GRANT_REQUESTER", nil
	case "grant-approver":
		return "GRANT_APPROVER", nil
	default:
		return "", fmt.Errorf("--caller-access-type must be one of grant-requester, grant-approver (got %q)", v)
	}
}

// pamLoadEntitlement reads and normalises a YAML/JSON entitlement config.
func pamLoadEntitlement(p string) (map[string]any, error) {
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("reading entitlement file: %w", err)
	}
	var raw any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing entitlement file: %w", err)
	}
	normalised, ok := convertYAMLKeys(raw).(map[string]any)
	if !ok {
		return nil, fmt.Errorf("entitlement file must decode to a JSON object")
	}
	return normalised, nil
}

func runPAMEntCreate(cmd *cobra.Command, args []string) error {
	parent, err := pamEntParent()
	if err != nil {
		return err
	}
	body, err := pamLoadEntitlement(flagPAMEntitlementFile)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	ctx := context.Background()
	return pamHTTP(ctx, http.MethodPost, parent+"/entitlements",
		"?entitlementId="+url.QueryEscape(args[0]), payload, flagPAMFormat)
}

func runPAMEntDelete(cmd *cobra.Command, args []string) error {
	name, err := pamEntName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	// gcloud python routes delete through SetForceFieldInDeleteEntitlementRequest,
	// which sets force=true on the request so grants under the entitlement are
	// cleaned up alongside it.
	return pamHTTP(ctx, http.MethodDelete, name, "?force=true", nil, flagPAMFormat)
}

func runPAMEntDescribe(cmd *cobra.Command, args []string) error {
	name, err := pamEntName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	return pamHTTP(ctx, http.MethodGet, name, "", nil, flagPAMFormat)
}

func runPAMEntExport(cmd *cobra.Command, args []string) error {
	name, err := pamEntName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	body, err := pamHTTPRaw(ctx, http.MethodGet, name, "", nil)
	if err != nil {
		return err
	}
	var decoded any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return fmt.Errorf("decoding entitlement: %w", err)
	}
	yamlBytes, err := yaml.Marshal(decoded)
	if err != nil {
		return fmt.Errorf("encoding entitlement as YAML: %w", err)
	}
	if flagPAMDestination == "" {
		_, err := os.Stdout.Write(yamlBytes)
		return err
	}
	if err := os.WriteFile(flagPAMDestination, yamlBytes, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", flagPAMDestination, err)
	}
	fmt.Printf("Exported entitlement [%s] to %s.\n", path.Base(name), flagPAMDestination)
	return nil
}

func runPAMEntList(cmd *cobra.Command, args []string) error {
	parent, err := pamEntParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	var all []any
	pageToken := ""
	for {
		q := url.Values{}
		if flagPAMFilter != "" {
			q.Set("filter", flagPAMFilter)
		}
		if flagPAMOrderBy != "" {
			q.Set("orderBy", flagPAMOrderBy)
		}
		if flagPAMPageSize > 0 {
			q.Set("pageSize", fmt.Sprintf("%d", flagPAMPageSize))
		}
		if pageToken != "" {
			q.Set("pageToken", pageToken)
		}
		body, err := pamHTTPRaw(ctx, http.MethodGet, parent+"/entitlements", pamQuery(q), nil)
		if err != nil {
			return err
		}
		var page struct {
			Entitlements  []any  `json:"entitlements"`
			NextPageToken string `json:"nextPageToken"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return fmt.Errorf("decoding entitlements list: %w", err)
		}
		all = append(all, page.Entitlements...)
		if page.NextPageToken == "" {
			break
		}
		pageToken = page.NextPageToken
	}
	return emitFormatted(all, flagPAMFormat)
}

func runPAMEntSearch(cmd *cobra.Command, args []string) error {
	parent, err := pamEntParent()
	if err != nil {
		return err
	}
	enum, err := pamCallerAccessValue(flagPAMCallerAccessType)
	if err != nil {
		return err
	}
	ctx := context.Background()
	var all []any
	pageToken := ""
	for {
		q := url.Values{}
		q.Set("callerAccessType", enum)
		if flagPAMPageSize > 0 {
			q.Set("pageSize", fmt.Sprintf("%d", flagPAMPageSize))
		}
		if pageToken != "" {
			q.Set("pageToken", pageToken)
		}
		body, err := pamHTTPRaw(ctx, http.MethodGet, parent+"/entitlements:search", pamQuery(q), nil)
		if err != nil {
			return err
		}
		var page struct {
			Entitlements  []any  `json:"entitlements"`
			NextPageToken string `json:"nextPageToken"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return fmt.Errorf("decoding entitlements search: %w", err)
		}
		all = append(all, page.Entitlements...)
		if page.NextPageToken == "" {
			break
		}
		pageToken = page.NextPageToken
	}
	return emitFormatted(all, flagPAMFormat)
}

func runPAMEntUpdate(cmd *cobra.Command, args []string) error {
	name, err := pamEntName(args[0])
	if err != nil {
		return err
	}
	body, err := pamLoadEntitlement(flagPAMEntitlementFile)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	// gcloud python auto-derives the update mask from the populated fields via
	// SetUpdateMaskInUpdateEntitlementRequest; match that behaviour here.
	keys := make([]string, 0, len(body))
	for k := range body {
		keys = append(keys, k)
	}
	q := "?updateMask=" + url.QueryEscape(strings.Join(keys, ","))
	ctx := context.Background()
	return pamHTTP(ctx, http.MethodPatch, name, q, payload, flagPAMFormat)
}

func pamQuery(q url.Values) string {
	if len(q) == 0 {
		return ""
	}
	return "?" + q.Encode()
}

// pamHTTP issues an authenticated JSON request against the Privileged Access
// Manager REST API and emits the decoded body in the requested format.
func pamHTTP(ctx context.Context, method, name, query string, payload []byte, format string) error {
	body, err := pamHTTPRaw(ctx, method, name, query, payload)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}
	var out any
	if err := json.Unmarshal(body, &out); err != nil {
		// Non-JSON response: emit it as-is.
		_, werr := os.Stdout.Write(body)
		return werr
	}
	return emitFormatted(out, format)
}

func pamHTTPRaw(ctx context.Context, method, name, query string, payload []byte) ([]byte, error) {
	ts, err := gcp.PlatformTokenSource(ctx, flagAccount)
	if err != nil {
		return nil, err
	}
	tok, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("obtaining access token: %w", err)
	}
	target := fmt.Sprintf("https://privilegedaccessmanager.googleapis.com/v1/%s%s", name, query)
	var reqBody io.Reader
	if payload != nil {
		reqBody = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, target, reqBody)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	tok.SetAuthHeader(req)
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP %s: %w", method, err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}
