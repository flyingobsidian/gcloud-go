package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	datalineage "google.golang.org/api/datalineage/v1"
)

// --- gcloud datalineage (#323) ---

var datalineageCmd = &cobra.Command{Use: "datalineage", Short: "Manage Data Lineage"}

func dlLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func dlChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

var (
	flagDLLocation   string
	flagDLConfigFile string
	flagDLUpdateMask string
	flagDLFormat     string
)

// --- config ---

var datalineageConfigCmd = &cobra.Command{Use: "config", Short: "Manage Data Lineage config"}

var (
	dlConfigDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe the project lineage config for a location",
		Args: cobra.NoArgs, RunE: runDLConfigDescribe,
	}
	dlConfigUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update the project lineage config from a --config-file",
		Args: cobra.NoArgs, RunE: runDLConfigUpdate,
	}
)

// --- processes ---

var datalineageProcessesCmd = &cobra.Command{Use: "processes", Short: "Manage Data Lineage processes"}

var (
	dlProcCreateCmd = &cobra.Command{
		Use: "create PROCESS", Short: "Create a process from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runDLProcCreate,
	}
	dlProcDeleteCmd = &cobra.Command{
		Use: "delete PROCESS", Short: "Delete a process",
		Args: cobra.ExactArgs(1), RunE: runDLProcDelete,
	}
	dlProcDescribeCmd = &cobra.Command{
		Use: "describe PROCESS", Short: "Describe a process",
		Args: cobra.ExactArgs(1), RunE: runDLProcDescribe,
	}
	dlProcListCmd = &cobra.Command{
		Use: "list", Short: "List processes in a location",
		Args: cobra.NoArgs, RunE: runDLProcList,
	}
	dlProcUpdateCmd = &cobra.Command{
		Use: "update PROCESS", Short: "Update a process from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runDLProcUpdate,
	}
)

func init() {
	// config
	for _, c := range []*cobra.Command{dlConfigDescribeCmd, dlConfigUpdateCmd} {
		c.Flags().StringVar(&flagDLLocation, "location", "", "Location containing the config (required)")
		_ = c.MarkFlagRequired("location")
	}
	dlConfigDescribeCmd.Flags().StringVar(&flagDLFormat, "format", "", "Output format")
	dlConfigUpdateCmd.Flags().StringVar(&flagDLConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the ProjectLineageConfig body (required)")
	_ = dlConfigUpdateCmd.MarkFlagRequired("config-file")
	dlConfigUpdateCmd.Flags().StringVar(&flagDLUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	datalineageConfigCmd.AddCommand(dlConfigDescribeCmd, dlConfigUpdateCmd)
	datalineageCmd.AddCommand(datalineageConfigCmd)

	// processes
	procAll := []*cobra.Command{dlProcCreateCmd, dlProcDeleteCmd, dlProcDescribeCmd, dlProcListCmd, dlProcUpdateCmd}
	for _, c := range procAll {
		c.Flags().StringVar(&flagDLLocation, "location", "", "Location containing the process (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{dlProcCreateCmd, dlProcUpdateCmd} {
		c.Flags().StringVar(&flagDLConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Process body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	dlProcUpdateCmd.Flags().StringVar(&flagDLUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	dlProcDescribeCmd.Flags().StringVar(&flagDLFormat, "format", "", "Output format")
	dlProcListCmd.Flags().StringVar(&flagDLFormat, "format", "", "Output format")
	datalineageProcessesCmd.AddCommand(procAll...)
	datalineageCmd.AddCommand(datalineageProcessesCmd)

	rootCmd.AddCommand(datalineageCmd)
}

// --- config impl (raw HTTP: the Go client doesn't expose GetProjectLineageConfig
// or UpdateProjectLineageConfig, but the datalineage.googleapis.com REST API
// does) ---

func dlConfigName(project, location string) string {
	return fmt.Sprintf("%s/projectLineageConfig", dlLocationParent(project, location))
}

func runDLConfigDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	return dlRawJSON(ctx, http.MethodGet, dlConfigName(project, flagDLLocation), "", nil, flagDLFormat)
}

func runDLConfigUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	// Load user config as raw map so we can pass fields through untouched.
	var body map[string]any
	if err := loadYAMLOrJSONInto(flagDLConfigFile, &body); err != nil {
		return err
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	mask := flagDLUpdateMask
	if mask == "" {
		keys := make([]string, 0, len(body))
		for k := range body {
			keys = append(keys, k)
		}
		mask = strings.Join(keys, ",")
	}
	ctx := context.Background()
	return dlRawJSON(ctx, http.MethodPatch, dlConfigName(project, flagDLLocation),
		"?updateMask="+mask, payload, "")
}

func dlRawJSON(ctx context.Context, method, name, query string, payload []byte, format string) error {
	ts, err := gcp.PlatformTokenSource(ctx, flagAccount)
	if err != nil {
		return err
	}
	tok, err := ts.Token()
	if err != nil {
		return fmt.Errorf("obtaining access token: %w", err)
	}
	url := fmt.Sprintf("https://datalineage.googleapis.com/v1/%s%s", name, query)
	var reqBody io.Reader
	if payload != nil {
		reqBody = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	tok.SetAuthHeader(req)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP %s: %w", method, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	var out any
	if len(body) > 0 {
		_ = json.Unmarshal(body, &out)
	}
	return emitFormatted(out, format)
}

// --- processes impl ---

func dlProcName(id, project, location string) string {
	return dlChild("processes", id, dlLocationParent(project, location))
}

func runDLProcCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	p := &datalineage.GoogleCloudDatacatalogLineageV1Process{}
	if err := loadYAMLOrJSONInto(flagDLConfigFile, p); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataLineageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Processes.Create(dlLocationParent(project, flagDLLocation), p).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating process: %w", err)
	}
	return emitFormatted(got, "")
}

func runDLProcDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataLineageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Processes.Delete(dlProcName(args[0], project, flagDLLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting process: %w", err)
	}
	fmt.Printf("Deleted process [%s].\n", args[0])
	return nil
}

func runDLProcDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataLineageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Processes.Get(dlProcName(args[0], project, flagDLLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing process: %w", err)
	}
	return emitFormatted(got, flagDLFormat)
}

func runDLProcList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataLineageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Processes.List(dlLocationParent(project, flagDLLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing processes: %w", err)
	}
	if flagDLFormat != "" {
		return emitFormatted(resp.Processes, flagDLFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, p := range resp.Processes {
		fmt.Printf("%-40s %s\n", path.Base(p.Name), p.DisplayName)
	}
	return nil
}

func runDLProcUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	p := &datalineage.GoogleCloudDatacatalogLineageV1Process{}
	if err := loadYAMLOrJSONInto(flagDLConfigFile, p); err != nil {
		return err
	}
	mask := flagDLUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(p))
	}
	ctx := context.Background()
	svc, err := gcp.DataLineageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Processes.Patch(dlProcName(args[0], project, flagDLLocation), p).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating process: %w", err)
	}
	return emitFormatted(got, "")
}
