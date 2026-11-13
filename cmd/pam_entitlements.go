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

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// --- gcloud pam entitlements (#962) ---
//
// The Privileged Access Manager REST API is not covered by the generated
// google.golang.org/api Go client, so this command surface talks directly to
// https://privilegedaccessmanager.googleapis.com/v1.

func (s *pamCmds) initPamEntitlements() {
	all := []*cobra.Command{
		s.pamEntCreateCmd, s.pamEntDeleteCmd, s.pamEntDescribeCmd, s.pamEntExportCmd,
		s.pamEntListCmd, s.pamEntSearchCmd, s.pamEntUpdateCmd,
	}
	for _, c := range all {
		s.addPAMScopeFlags(c)
		c.Flags().StringVar(&s.flagPAMLocation, "location", "global",
			"Location that owns the entitlements (e.g. \"global\")")
	}
	for _, c := range []*cobra.Command{s.pamEntCreateCmd, s.pamEntUpdateCmd} {
		c.Flags().StringVar(&s.flagPAMEntitlementFile, "entitlement-file", "",
			"Path to a YAML/JSON file containing the entitlement configuration (required)")
		_ = c.MarkFlagRequired("entitlement-file")
	}
	s.pamEntExportCmd.Flags().StringVar(&s.flagPAMDestination, "destination", "",
		"Path to the local YAML file to write (defaults to stdout)")
	s.pamEntSearchCmd.Flags().StringVar(&s.flagPAMCallerAccessType, "caller-access-type", "",
		"Search access type: grant-requester or grant-approver (required)")
	_ = s.pamEntSearchCmd.MarkFlagRequired("caller-access-type")
	s.pamEntListCmd.Flags().StringVar(&s.flagPAMFilter, "filter", "", "Server-side filter expression")
	s.pamEntListCmd.Flags().StringVar(&s.flagPAMOrderBy, "order-by", "", "Server-side ordering hint")
	s.pamEntListCmd.Flags().Int64Var(&s.flagPAMPageSize, "page-size", 0, "Maximum number of results per page")
	s.pamEntSearchCmd.Flags().Int64Var(&s.flagPAMPageSize, "page-size", 0, "Maximum number of results per page")
	for _, c := range []*cobra.Command{s.pamEntCreateCmd, s.pamEntDeleteCmd, s.pamEntDescribeCmd, s.pamEntListCmd, s.pamEntSearchCmd, s.pamEntUpdateCmd} {
		c.Flags().StringVar(&s.flagPAMFormat, "format", "", "Output format")
	}

	s.pamEntitlementsCmd.AddCommand(all...)
	s.pamCmd.AddCommand(s.pamEntitlementsCmd)
}

func (s *pamCmds) addPAMScopeFlags(c *cobra.Command) {
	c.Flags().StringVar(&s.flagPAMFolder, "folder", "",
		"Folder scope (alternative to --project/--organization)")
	c.Flags().StringVar(&s.flagPAMOrganization, "organization", "",
		"Organization scope (alternative to --project/--folder)")
}

// pamScopeParent returns "projects/PROJECT", "folders/FOLDER" or "organizations/ORG".
func (s *pamCmds) pamScopeParent() (string, error) {
	set := 0
	if s.inv.flags.project != "" {
		set++
	}
	if s.flagPAMFolder != "" {
		set++
	}
	if s.flagPAMOrganization != "" {
		set++
	}
	if set > 1 {
		return "", fmt.Errorf("only one of --project, --folder, --organization may be set")
	}
	if s.flagPAMFolder != "" {
		return "folders/" + s.flagPAMFolder, nil
	}
	if s.flagPAMOrganization != "" {
		return "organizations/" + s.flagPAMOrganization, nil
	}
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return "projects/" + project, nil
}

func (s *pamCmds) pamEntParent() (string, error) {
	scope, err := s.pamScopeParent()
	if err != nil {
		return "", err
	}
	loc := s.flagPAMLocation
	if loc == "" {
		loc = "global"
	}
	return fmt.Sprintf("%s/locations/%s", scope, loc), nil
}

func (s *pamCmds) pamEntName(id string) (string, error) {
	if strings.Contains(id, "/entitlements/") {
		return id, nil
	}
	parent, err := s.pamEntParent()
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

func (s *pamCmds) runPAMEntCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.pamEntParent()
	if err != nil {
		return err
	}
	body, err := pamLoadEntitlement(s.flagPAMEntitlementFile)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	return s.pamHTTP(ctx, http.MethodPost, parent+"/entitlements",
		"?entitlementId="+url.QueryEscape(args[0]), payload, s.flagPAMFormat)
}

func (s *pamCmds) runPAMEntDelete(cmd *cobra.Command, args []string) error {
	name, err := s.pamEntName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	// gcloud python routes delete through SetForceFieldInDeleteEntitlementRequest,
	// which sets force=true on the request so grants under the entitlement are
	// cleaned up alongside it.
	return s.pamHTTP(ctx, http.MethodDelete, name, "?force=true", nil, s.flagPAMFormat)
}

func (s *pamCmds) runPAMEntDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.pamEntName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	return s.pamHTTP(ctx, http.MethodGet, name, "", nil, s.flagPAMFormat)
}

func (s *pamCmds) runPAMEntExport(cmd *cobra.Command, args []string) error {
	name, err := s.pamEntName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	body, err := s.pamHTTPRaw(ctx, http.MethodGet, name, "", nil)
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
	if s.flagPAMDestination == "" {
		_, err := s.inv.stdout.Write(yamlBytes)
		return err
	}
	if err := os.WriteFile(s.flagPAMDestination, yamlBytes, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", s.flagPAMDestination, err)
	}
	fmt.Fprintf(s.inv.stdout, "Exported entitlement [%s] to %s.\n", path.Base(name), s.flagPAMDestination)
	return nil
}

func (s *pamCmds) runPAMEntList(cmd *cobra.Command, args []string) error {
	parent, err := s.pamEntParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var all []any
	pageToken := ""
	for {
		q := url.Values{}
		if s.flagPAMFilter != "" {
			q.Set("filter", s.flagPAMFilter)
		}
		if s.flagPAMOrderBy != "" {
			q.Set("orderBy", s.flagPAMOrderBy)
		}
		if s.flagPAMPageSize > 0 {
			q.Set("pageSize", fmt.Sprintf("%d", s.flagPAMPageSize))
		}
		if pageToken != "" {
			q.Set("pageToken", pageToken)
		}
		body, err := s.pamHTTPRaw(ctx, http.MethodGet, parent+"/entitlements", pamQuery(q), nil)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagPAMFormat)
}

func (s *pamCmds) runPAMEntSearch(cmd *cobra.Command, args []string) error {
	parent, err := s.pamEntParent()
	if err != nil {
		return err
	}
	enum, err := pamCallerAccessValue(s.flagPAMCallerAccessType)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var all []any
	pageToken := ""
	for {
		q := url.Values{}
		q.Set("callerAccessType", enum)
		if s.flagPAMPageSize > 0 {
			q.Set("pageSize", fmt.Sprintf("%d", s.flagPAMPageSize))
		}
		if pageToken != "" {
			q.Set("pageToken", pageToken)
		}
		body, err := s.pamHTTPRaw(ctx, http.MethodGet, parent+"/entitlements:search", pamQuery(q), nil)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagPAMFormat)
}

func (s *pamCmds) runPAMEntUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.pamEntName(args[0])
	if err != nil {
		return err
	}
	body, err := pamLoadEntitlement(s.flagPAMEntitlementFile)
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
	ctx := cmd.Context()
	return s.pamHTTP(ctx, http.MethodPatch, name, q, payload, s.flagPAMFormat)
}

func pamQuery(q url.Values) string {
	if len(q) == 0 {
		return ""
	}
	return "?" + q.Encode()
}

// pamHTTP issues an authenticated JSON request against the Privileged Access
// Manager REST API and emits the decoded body in the requested format.
func (s *pamCmds) pamHTTP(ctx context.Context, method, name, query string, payload []byte, format string) error {
	body, err := s.pamHTTPRaw(ctx, method, name, query, payload)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}
	var out any
	if err := json.Unmarshal(body, &out); err != nil {
		// Non-JSON response: emit it as-is.
		_, werr := s.inv.stdout.Write(body)
		return werr
	}
	return emitFormattedTo(s.inv.stdout, out, format)
}

func (s *pamCmds) pamHTTPRaw(ctx context.Context, method, name, query string, payload []byte) ([]byte, error) {
	ts, err := s.inv.clients().PlatformTokenSource(ctx)
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
