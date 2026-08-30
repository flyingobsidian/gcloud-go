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
	"path/filepath"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
)

// --- gcloud apigee apis (#1375, import #1753) ---

var apigeeApisCmd = &cobra.Command{Use: "apis", Short: "Manage Apigee API proxies"}

var (
	flagApigeeApiOrganization string
	flagApigeeApiFormat       string
	flagApigeeApiFromBundle   string
	flagApigeeApiFromTemplate string
	flagApigeeApiSpace        string
)

var (
	apigeeApisDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an API proxy",
		Args: cobra.ExactArgs(1), RunE: runApigeeApisDelete,
	}
	apigeeApisDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an API proxy",
		Args: cobra.ExactArgs(1), RunE: runApigeeApisDescribe,
	}
	apigeeApisListCmd = &cobra.Command{
		Use: "list", Short: "List API proxies in an organization",
		Args: cobra.NoArgs, RunE: runApigeeApisList,
	}
	apigeeApisImportCmd = &cobra.Command{
		Use: "import API", Short: "Import an API proxy from a local bundle or feature template",
		Args: cobra.ExactArgs(1), RunE: runApigeeApisImport,
	}
)

func init() {
	all := []*cobra.Command{apigeeApisDeleteCmd, apigeeApisDescribeCmd, apigeeApisListCmd, apigeeApisImportCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagApigeeApiOrganization, "organization", "", "Apigee organization (required)")
		_ = c.MarkFlagRequired("organization")
		c.Flags().StringVar(&flagApigeeApiFormat, "format", "", "Output format")
	}
	apigeeApisImportCmd.Flags().StringVar(&flagApigeeApiFromBundle, "from-bundle", "",
		"Path to a local API proxy bundle ZIP to import "+
			"(mutually exclusive with --from-template)")
	apigeeApisImportCmd.Flags().StringVar(&flagApigeeApiFromTemplate, "from-template", "",
		"Path to an Apigee Feature Template YAML (not yet supported in gcloud-go; "+
			"use --from-bundle for now)")
	apigeeApisImportCmd.Flags().StringVar(&flagApigeeApiSpace, "space", "",
		"Optional Apigee space to import the API proxy into")

	apigeeApisCmd.AddCommand(all...)
	apigeeCmd.AddCommand(apigeeApisCmd)
}

func apigeeApiName(id string) (string, error) {
	return apigeeResource(flagApigeeApiOrganization, "apis", id)
}

func runApigeeApisDelete(cmd *cobra.Command, args []string) error {
	name, err := apigeeApiName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Organizations.Apis.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting api: %w", err)
	}
	fmt.Printf("Deleted api [%s].\n", args[0])
	return nil
}

func runApigeeApisDescribe(cmd *cobra.Command, args []string) error {
	name, err := apigeeApiName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Apis.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing api: %w", err)
	}
	return emitFormatted(got, flagApigeeApiFormat)
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
func runApigeeApisImport(cmd *cobra.Command, args []string) error {
	if flagApigeeApiFromBundle == "" && flagApigeeApiFromTemplate == "" {
		return fmt.Errorf("either --from-bundle or --from-template must be specified")
	}
	if flagApigeeApiFromBundle != "" && flagApigeeApiFromTemplate != "" {
		return fmt.Errorf("--from-bundle and --from-template are mutually exclusive")
	}
	if flagApigeeApiFromTemplate != "" {
		return fmt.Errorf("--from-template (Apigee Feature Template compilation) is not yet supported in gcloud-go; " +
			"compile the template locally and use --from-bundle=BUNDLE.zip instead")
	}

	// The API proxy positional may already be a fully qualified name.
	proxyName := args[0]
	org := flagApigeeApiOrganization
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

	bundlePath := flagApigeeApiFromBundle
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

	ctx := context.Background()
	ts, err := gcp.PlatformTokenSource(ctx, flagAccount)
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
	if flagApigeeApiSpace != "" {
		q.Set("space", flagApigeeApiSpace)
	}
	target := fmt.Sprintf("https://apigee.googleapis.com/v1/organizations/%s/apis?%s",
		url.PathEscape(org), q.Encode())

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
	fmt.Fprintf(os.Stderr, "Imported API proxy [%s] revision %s from %s.\n",
		proxyName, rev.Revision, bundlePath)
	return emitFormatted(&rev, flagApigeeApiFormat)
}

func runApigeeApisList(cmd *cobra.Command, args []string) error {
	parent, err := apigeeOrgName(flagApigeeApiOrganization)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApigeeService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Organizations.Apis.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing apis: %w", err)
	}
	var all []*apigee.GoogleCloudApigeeV1ApiProxy
	all = append(all, resp.Proxies...)
	return emitFormatted(all, flagApigeeApiFormat)
}
