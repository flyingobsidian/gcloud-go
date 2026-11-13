package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
)

// --- gcloud apigee apis (#1375, import #1753) ---

func (s *apigeeCmds) initApigeeApis() {
	all := []*cobra.Command{s.apigeeApisDeleteCmd, s.apigeeApisDescribeCmd, s.apigeeApisListCmd, s.apigeeApisImportCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagApigeeApiOrganization, "organization", "", "Apigee organization (required)")
		_ = c.MarkFlagRequired("organization")
		c.Flags().StringVar(&s.flagApigeeApiFormat, "format", "", "Output format")
	}
	s.apigeeApisImportCmd.Flags().StringVar(&s.flagApigeeApiFromBundle, "from-bundle", "",
		"Path to a local API proxy bundle ZIP to import "+
			"(mutually exclusive with --from-template)")
	s.apigeeApisImportCmd.Flags().StringVar(&s.flagApigeeApiFromTemplate, "from-template", "",
		"Path to an Apigee Feature Template YAML (not yet supported in gcloud-go; "+
			"use --from-bundle for now)")
	s.apigeeApisImportCmd.Flags().StringVar(&s.flagApigeeApiSpace, "space", "",
		"Optional Apigee space to import the API proxy into")

	s.apigeeApisCmd.AddCommand(all...)
	s.apigeeCmd.AddCommand(s.apigeeApisCmd)
}

func (s *apigeeCmds) apigeeApiName(id string) (string, error) {
	return apigeeResource(s.flagApigeeApiOrganization, "apis", id)
}

// runApigeeApisDelete deletes an API proxy and all of its revisions after a
// confirmation prompt, then prints the deleted proxy, matching gcloud-python's
// `apigee apis delete` (584.0.0, GA in 589.0.0).
func (s *apigeeCmds) runApigeeApisDelete(cmd *cobra.Command, args []string) error {
	name, err := s.apigeeApiName(args[0])
	if err != nil {
		return err
	}
	apiID := path.Base(name)
	if err := s.inv.promptContinue(
		fmt.Sprintf("API proxy [%s] and all of its revisions will be deleted.", apiID),
		"Deletion aborted by user.",
	); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, name)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Apis.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting API proxy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Deleted API proxy [%s].\n", apiID)
	return emitFormattedTo(s.inv.stdout, got, s.flagApigeeApiFormat)
}

func (s *apigeeCmds) runApigeeApisDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.apigeeApiName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, name)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Apis.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing api: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagApigeeApiFormat)
}

// runApigeeApisImport uploads a local API proxy bundle ZIP to Apigee using the
// `apis?action=import` REST endpoint. The Go SDK's OrganizationsApisService
// only exposes a JSON body Create, so this reaches for a direct POST with the
// bundle as `application/octet-stream`, matching gcloud-python's
// `APIsClient.Create`.
//
// --from-template compilation (Apigee Feature Templates) is not yet supported
// in gcloud-go; users should compile the template into a bundle upstream and
// then invoke `apis import --from-bundle=...`.
func (s *apigeeCmds) runApigeeApisImport(cmd *cobra.Command, args []string) error {
	if s.flagApigeeApiFromBundle == "" && s.flagApigeeApiFromTemplate == "" {
		return fmt.Errorf("either --from-bundle or --from-template must be specified")
	}
	if s.flagApigeeApiFromBundle != "" && s.flagApigeeApiFromTemplate != "" {
		return fmt.Errorf("--from-bundle and --from-template are mutually exclusive")
	}
	if s.flagApigeeApiFromTemplate != "" {
		return fmt.Errorf("--from-template (Apigee Feature Template compilation) is not yet supported in gcloud-go; " +
			"compile the template locally and use --from-bundle=BUNDLE.zip instead")
	}

	// The API proxy positional may already be a fully qualified name.
	proxyName := args[0]
	org := s.flagApigeeApiOrganization
	if strings.HasPrefix(proxyName, "organizations/") {
		parts := strings.Split(proxyName, "/")
		if len(parts) != 4 || parts[2] != "apis" {
			return fmt.Errorf("unrecognised fully-qualified API proxy name %q", proxyName)
		}
		org, proxyName = parts[1], parts[3]
	}
	if org == "" {
		return fmt.Errorf("--organization is required (or supply organizations/ORG/apis/API as the positional arg)")
	}

	bundlePath := s.flagApigeeApiFromBundle
	if !strings.HasSuffix(strings.ToLower(bundlePath), ".zip") {
		return fmt.Errorf("--from-bundle must point at a .zip file (got %q)", filepath.Base(bundlePath))
	}
	bundle, err := os.ReadFile(bundlePath)
	if err != nil {
		return fmt.Errorf("reading bundle: %w", err)
	}
	if len(bundle) == 0 {
		return fmt.Errorf("bundle %q is empty", bundlePath)
	}

	ctx := cmd.Context()
	ts, err := s.inv.clients().PlatformTokenSource(ctx)
	if err != nil {
		return err
	}
	tok, err := ts.Token()
	if err != nil {
		return fmt.Errorf("obtaining access token: %w", err)
	}

	q := url.Values{}
	q.Set("action", "import")
	q.Set("name", proxyName)
	if s.flagApigeeApiSpace != "" {
		q.Set("space", s.flagApigeeApiSpace)
	}
	host, err := s.inv.clients().ApigeeOrgHost(ctx, org)
	if err != nil {
		return err
	}
	target := fmt.Sprintf("https://%s/v1/organizations/%s/apis?%s",
		host, url.PathEscape(org), q.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(bundle))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = int64(len(bundle))
	tok.SetAuthHeader(req)

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("uploading bundle: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("apigee apis import HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var rev apigee.GoogleCloudApigeeV1ApiProxyRevision
	if err := json.Unmarshal(body, &rev); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Imported API proxy [%s] revision %s from %s.\n",
		proxyName, rev.Revision, bundlePath)
	return emitFormattedTo(s.inv.stdout, &rev, s.flagApigeeApiFormat)
}

func (s *apigeeCmds) runApigeeApisList(cmd *cobra.Command, args []string) error {
	parent, err := apigeeOrgName(s.flagApigeeApiOrganization)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, parent)
	if err != nil {
		return err
	}
	resp, err := svc.Organizations.Apis.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing apis: %w", err)
	}
	var all []*apigee.GoogleCloudApigeeV1ApiProxy
	all = append(all, resp.Proxies...)
	return emitFormattedTo(s.inv.stdout, all, s.flagApigeeApiFormat)
}
