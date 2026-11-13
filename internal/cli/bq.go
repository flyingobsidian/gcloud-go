package cli

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

	"github.com/spf13/cobra"
)

// --- gcloud bq (#939) ---
//
// The BigQuery Migration API (bigquerymigration.googleapis.com/v2) is not
// exposed by the google-golang-api client at v0.279.0, so the commands below
// wrap the REST endpoint directly using the same platform token source used
// for eventarcpublishing.

func (s *bqCmds) initBq() {
	for _, c := range []*cobra.Command{s.bqMWCreateCmd, s.bqMWDeleteCmd, s.bqMWDescribeCmd, s.bqMWListCmd, s.bqMWStartCmd} {
		c.Flags().StringVar(&s.flagBQMWLocation, "location", "us", "Location of the workflow")
	}
	s.bqMWCreateCmd.Flags().StringVar(&s.flagBQMWFile, "config-from-file", "", "JSON file with the MigrationWorkflow body (required)")
	s.bqMWCreateCmd.MarkFlagRequired("config-from-file")

	s.bqMWListCmd.Flags().StringVar(&s.flagBQMWFilter, "filter", "", "Server-side filter expression")
	s.bqMWListCmd.Flags().IntVar(&s.flagBQMWPageSize, "page-size", 0, "Number of results per page")
	s.bqMWListCmd.Flags().StringVar(&s.flagBQMWReadMask, "read-mask", "", "Comma-separated list of fields to return")

	s.bqMWDescribeCmd.Flags().StringVar(&s.flagBQMWReadMask, "read-mask", "", "Comma-separated list of fields to return")

	s.bqMigrationWorkflowsCmd.AddCommand(s.bqMWCreateCmd, s.bqMWDeleteCmd, s.bqMWDescribeCmd, s.bqMWListCmd, s.bqMWStartCmd)
	s.bqCmd.AddCommand(s.bqMigrationWorkflowsCmd)

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

func (s *bqCmds) bqDoJSON(ctx context.Context, method, apiURL string, body any) ([]byte, error) {
	ts, err := s.inv.clients().PlatformTokenSource(ctx)
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

func (s *bqCmds) bqPrintJSON(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, b, "", "  "); err != nil {
		s.inv.stdout.Write(b)
		fmt.Fprintln(s.inv.stdout)
		return nil
	}
	fmt.Fprintln(s.inv.stdout, pretty.String())
	return nil
}

// --- impl ---

func (s *bqCmds) runBQMWCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	var body map[string]any
	f, err := os.ReadFile(s.flagBQMWFile)
	if err != nil {
		return fmt.Errorf("reading %s: %w", s.flagBQMWFile, err)
	}
	if err := json.Unmarshal(f, &body); err != nil {
		return fmt.Errorf("parsing %s: %w", s.flagBQMWFile, err)
	}
	// If a workflow id was passed, set display_name if the body omits it.
	if len(args) == 1 && args[0] != "" {
		if _, ok := body["displayName"]; !ok {
			body["displayName"] = args[0]
		}
	}
	apiURL := fmt.Sprintf("%s/v2/%s/workflows", bqMigrationHost, bqMWParent(project, s.flagBQMWLocation))
	ctx := cmd.Context()
	respBody, err := s.bqDoJSON(ctx, http.MethodPost, apiURL, body)
	if err != nil {
		return err
	}
	return s.bqPrintJSON(respBody)
}

func (s *bqCmds) runBQMWDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	apiURL := fmt.Sprintf("%s/v2/%s", bqMigrationHost, bqMWName(args[0], project, s.flagBQMWLocation))
	ctx := cmd.Context()
	if _, err := s.bqDoJSON(ctx, http.MethodDelete, apiURL, nil); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Deleted migration workflow [%s].\n", args[0])
	return nil
}

func (s *bqCmds) runBQMWDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	q := url.Values{}
	if s.flagBQMWReadMask != "" {
		q.Set("readMask", s.flagBQMWReadMask)
	}
	apiURL := fmt.Sprintf("%s/v2/%s", bqMigrationHost, bqMWName(args[0], project, s.flagBQMWLocation))
	if enc := q.Encode(); enc != "" {
		apiURL += "?" + enc
	}
	ctx := cmd.Context()
	respBody, err := s.bqDoJSON(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return err
	}
	return s.bqPrintJSON(respBody)
}

func (s *bqCmds) runBQMWList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	pageToken := ""
	first := true
	fmt.Fprintln(s.inv.stdout, "[")
	for {
		q := url.Values{}
		if s.flagBQMWFilter != "" {
			q.Set("filter", s.flagBQMWFilter)
		}
		if s.flagBQMWPageSize > 0 {
			q.Set("pageSize", fmt.Sprintf("%d", s.flagBQMWPageSize))
		}
		if s.flagBQMWReadMask != "" {
			q.Set("readMask", s.flagBQMWReadMask)
		}
		if pageToken != "" {
			q.Set("pageToken", pageToken)
		}
		apiURL := fmt.Sprintf("%s/v2/%s/workflows", bqMigrationHost, bqMWParent(project, s.flagBQMWLocation))
		if enc := q.Encode(); enc != "" {
			apiURL += "?" + enc
		}
		respBody, err := s.bqDoJSON(ctx, http.MethodGet, apiURL, nil)
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
				fmt.Fprintln(s.inv.stdout, ",")
			}
			first = false
			b, _ := json.MarshalIndent(wf, "  ", "  ")
			fmt.Fprintf(s.inv.stdout, "  %s", string(b))
		}
		if page.NextPageToken == "" {
			break
		}
		pageToken = page.NextPageToken
	}
	fmt.Fprintln(s.inv.stdout)
	fmt.Fprintln(s.inv.stdout, "]")
	return nil
}

func (s *bqCmds) runBQMWStart(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	apiURL := fmt.Sprintf("%s/v2/%s:start", bqMigrationHost, bqMWName(args[0], project, s.flagBQMWLocation))
	ctx := cmd.Context()
	respBody, err := s.bqDoJSON(ctx, http.MethodPost, apiURL, map[string]any{})
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Started migration workflow [%s].\n", args[0])
	return s.bqPrintJSON(respBody)
}

// bqCmds holds the commands and flag values of one invocation's "bq"
// command group, so invocations don't share state (#2070).
type bqCmds struct {
	inv *invocation

	bqCmd                   *cobra.Command
	bqMigrationWorkflowsCmd *cobra.Command
	flagBQMWLocation        string
	flagBQMWFile            string
	flagBQMWFilter          string
	flagBQMWPageSize        int
	flagBQMWReadMask        string
	bqMWCreateCmd           *cobra.Command
	bqMWDeleteCmd           *cobra.Command
	bqMWDescribeCmd         *cobra.Command
	bqMWListCmd             *cobra.Command
	bqMWStartCmd            *cobra.Command
}

// newBqCmd returns the "bq" command group for inv.
func newBqCmd(inv *invocation) *cobra.Command {
	return newBqCmds(inv).bqCmd
}

// newBqCmds builds the "bq" command group for inv and returns its state.
func newBqCmds(inv *invocation) *bqCmds {
	s := &bqCmds{inv: inv}
	s.bqCmd = &cobra.Command{
		Use:   "bq",
		Short: "Manage BigQuery Migration resources",
	}
	s.bqMigrationWorkflowsCmd = &cobra.Command{
		Use:   "migration-workflows",
		Short: "Manage BigQuery migration workflows",
	}
	s.bqMWCreateCmd = &cobra.Command{
		Use:   "create WORKFLOW_ID",
		Short: "Create a BigQuery migration workflow",
		Args:  cobra.MaximumNArgs(1),
		RunE:  s.runBQMWCreate,
	}
	s.bqMWDeleteCmd = &cobra.Command{
		Use:   "delete WORKFLOW",
		Short: "Delete a BigQuery migration workflow",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runBQMWDelete,
	}
	s.bqMWDescribeCmd = &cobra.Command{
		Use:   "describe WORKFLOW",
		Short: "Describe a BigQuery migration workflow",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runBQMWDescribe,
	}
	s.bqMWListCmd = &cobra.Command{
		Use:   "list",
		Short: "List BigQuery migration workflows in a location",
		Args:  cobra.NoArgs,
		RunE:  s.runBQMWList,
	}
	s.bqMWStartCmd = &cobra.Command{
		Use:   "start WORKFLOW",
		Short: "Start a BigQuery migration workflow",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runBQMWStart,
	}
	s.initBq()
	return s
}
