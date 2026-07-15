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
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
)

// --- gcloud bq (#939) ---
//
// The BigQuery Migration API (bigquerymigration.googleapis.com/v2) is not
// exposed by the google-golang-api client at v0.279.0, so the commands below
// wrap the REST endpoint directly using the same platform token source used
// for eventarcpublishing.

var bqCmd = &cobra.Command{
	Use:   "bq",
	Short: "Manage BigQuery Migration resources",
}

var bqMigrationWorkflowsCmd = &cobra.Command{
	Use:   "migration-workflows",
	Short: "Manage BigQuery migration workflows",
}

var (
	flagBQMWLocation string
	flagBQMWFile     string
	flagBQMWFilter   string
	flagBQMWPageSize int
	flagBQMWReadMask string
)

var bqMWCreateCmd = &cobra.Command{
	Use:   "create WORKFLOW_ID",
	Short: "Create a BigQuery migration workflow",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runBQMWCreate,
}

var bqMWDeleteCmd = &cobra.Command{
	Use:   "delete WORKFLOW",
	Short: "Delete a BigQuery migration workflow",
	Args:  cobra.ExactArgs(1),
	RunE:  runBQMWDelete,
}

var bqMWDescribeCmd = &cobra.Command{
	Use:   "describe WORKFLOW",
	Short: "Describe a BigQuery migration workflow",
	Args:  cobra.ExactArgs(1),
	RunE:  runBQMWDescribe,
}

var bqMWListCmd = &cobra.Command{
	Use:   "list",
	Short: "List BigQuery migration workflows in a location",
	Args:  cobra.NoArgs,
	RunE:  runBQMWList,
}

var bqMWStartCmd = &cobra.Command{
	Use:   "start WORKFLOW",
	Short: "Start a BigQuery migration workflow",
	Args:  cobra.ExactArgs(1),
	RunE:  runBQMWStart,
}

func init() {
	for _, c := range []*cobra.Command{bqMWCreateCmd, bqMWDeleteCmd, bqMWDescribeCmd, bqMWListCmd, bqMWStartCmd} {
		c.Flags().StringVar(&flagBQMWLocation, "location", "us", "Location of the workflow")
	}
	bqMWCreateCmd.Flags().StringVar(&flagBQMWFile, "config-from-file", "", "JSON file with the MigrationWorkflow body (required)")
	bqMWCreateCmd.MarkFlagRequired("config-from-file")

	bqMWListCmd.Flags().StringVar(&flagBQMWFilter, "filter", "", "Server-side filter expression")
	bqMWListCmd.Flags().IntVar(&flagBQMWPageSize, "page-size", 0, "Number of results per page")
	bqMWListCmd.Flags().StringVar(&flagBQMWReadMask, "read-mask", "", "Comma-separated list of fields to return")

	bqMWDescribeCmd.Flags().StringVar(&flagBQMWReadMask, "read-mask", "", "Comma-separated list of fields to return")

	bqMigrationWorkflowsCmd.AddCommand(bqMWCreateCmd, bqMWDeleteCmd, bqMWDescribeCmd, bqMWListCmd, bqMWStartCmd)
	bqCmd.AddCommand(bqMigrationWorkflowsCmd)
	rootCmd.AddCommand(bqCmd)
}

// --- helpers ---

const bqMigrationHost = "https://bigquerymigration.googleapis.com"

func bqMWParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func bqMWName(id, project, location string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/workflows/%s", bqMWParent(project, location), id)
}

func bqDoJSON(ctx context.Context, method, apiURL string, body any) ([]byte, error) {
	ts, err := gcp.PlatformTokenSource(ctx, flagAccount)
	if err != nil {
		return nil, err
	}
	tok, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("obtaining access token: %w", err)
	}
	var reqBody io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, apiURL, reqBody)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	tok.SetAuthHeader(req)
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return respBody, fmt.Errorf("bigquerymigration: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return respBody, nil
}

func bqPrintJSON(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, b, "", "  "); err != nil {
		os.Stdout.Write(b)
		fmt.Println()
		return nil
	}
	fmt.Println(pretty.String())
	return nil
}

// --- impl ---

func runBQMWCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	var body map[string]any
	f, err := os.ReadFile(flagBQMWFile)
	if err != nil {
		return fmt.Errorf("reading %s: %w", flagBQMWFile, err)
	}
	if err := json.Unmarshal(f, &body); err != nil {
		return fmt.Errorf("parsing %s: %w", flagBQMWFile, err)
	}
	// If a workflow id was passed, set display_name if the body omits it.
	if len(args) == 1 && args[0] != "" {
		if _, ok := body["displayName"]; !ok {
			body["displayName"] = args[0]
		}
	}
	apiURL := fmt.Sprintf("%s/v2/%s/workflows", bqMigrationHost, bqMWParent(project, flagBQMWLocation))
	ctx := context.Background()
	respBody, err := bqDoJSON(ctx, http.MethodPost, apiURL, body)
	if err != nil {
		return err
	}
	return bqPrintJSON(respBody)
}

func runBQMWDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	apiURL := fmt.Sprintf("%s/v2/%s", bqMigrationHost, bqMWName(args[0], project, flagBQMWLocation))
	ctx := context.Background()
	if _, err := bqDoJSON(ctx, http.MethodDelete, apiURL, nil); err != nil {
		return err
	}
	fmt.Printf("Deleted migration workflow [%s].\n", args[0])
	return nil
}

func runBQMWDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	q := url.Values{}
	if flagBQMWReadMask != "" {
		q.Set("readMask", flagBQMWReadMask)
	}
	apiURL := fmt.Sprintf("%s/v2/%s", bqMigrationHost, bqMWName(args[0], project, flagBQMWLocation))
	if enc := q.Encode(); enc != "" {
		apiURL += "?" + enc
	}
	ctx := context.Background()
	respBody, err := bqDoJSON(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return err
	}
	return bqPrintJSON(respBody)
}

func runBQMWList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	pageToken := ""
	first := true
	fmt.Println("[")
	for {
		q := url.Values{}
		if flagBQMWFilter != "" {
			q.Set("filter", flagBQMWFilter)
		}
		if flagBQMWPageSize > 0 {
			q.Set("pageSize", fmt.Sprintf("%d", flagBQMWPageSize))
		}
		if flagBQMWReadMask != "" {
			q.Set("readMask", flagBQMWReadMask)
		}
		if pageToken != "" {
			q.Set("pageToken", pageToken)
		}
		apiURL := fmt.Sprintf("%s/v2/%s/workflows", bqMigrationHost, bqMWParent(project, flagBQMWLocation))
		if enc := q.Encode(); enc != "" {
			apiURL += "?" + enc
		}
		respBody, err := bqDoJSON(ctx, http.MethodGet, apiURL, nil)
		if err != nil {
			return err
		}
		var page struct {
			MigrationWorkflows []map[string]any `json:"migrationWorkflows"`
			NextPageToken      string           `json:"nextPageToken"`
		}
		if err := json.Unmarshal(respBody, &page); err != nil {
			return fmt.Errorf("parsing list response: %w", err)
		}
		for _, wf := range page.MigrationWorkflows {
			if !first {
				fmt.Println(",")
			}
			first = false
			b, _ := json.MarshalIndent(wf, "  ", "  ")
			fmt.Printf("  %s", string(b))
		}
		if page.NextPageToken == "" {
			break
		}
		pageToken = page.NextPageToken
	}
	fmt.Println()
	fmt.Println("]")
	return nil
}

func runBQMWStart(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	apiURL := fmt.Sprintf("%s/v2/%s:start", bqMigrationHost, bqMWName(args[0], project, flagBQMWLocation))
	ctx := context.Background()
	respBody, err := bqDoJSON(ctx, http.MethodPost, apiURL, map[string]any{})
	if err != nil {
		return err
	}
	fmt.Printf("Started migration workflow [%s].\n", args[0])
	return bqPrintJSON(respBody)
}
