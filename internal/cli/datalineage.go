package cli

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

	"github.com/spf13/cobra"
	datalineage "google.golang.org/api/datalineage/v1"
)

// --- gcloud datalineage (#323) ---

func dlLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func dlChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

// --- config ---

// --- runs ---

// --- lineage-events ---

// --- processes ---

func (s *datalineageCmds) initDatalineage() {
	// config
	for _, c := range []*cobra.Command{s.dlConfigDescribeCmd, s.dlConfigUpdateCmd} {
		c.Flags().StringVar(&s.flagDLLocation, "location", "", "Location containing the config (required)")
		_ = c.MarkFlagRequired("location")
	}
	s.dlConfigDescribeCmd.Flags().StringVar(&s.flagDLFormat, "format", "", "Output format")
	s.dlConfigUpdateCmd.Flags().StringVar(&s.flagDLConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the ProjectLineageConfig body (required)")
	_ = s.dlConfigUpdateCmd.MarkFlagRequired("config-file")
	s.dlConfigUpdateCmd.Flags().StringVar(&s.flagDLUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.datalineageConfigCmd.AddCommand(s.dlConfigDescribeCmd, s.dlConfigUpdateCmd)
	s.datalineageCmd.AddCommand(s.datalineageConfigCmd)

	// processes
	procAll := []*cobra.Command{s.dlProcCreateCmd, s.dlProcDeleteCmd, s.dlProcDescribeCmd, s.dlProcListCmd, s.dlProcUpdateCmd}
	for _, c := range procAll {
		c.Flags().StringVar(&s.flagDLLocation, "location", "", "Location containing the process (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{s.dlProcCreateCmd, s.dlProcUpdateCmd} {
		c.Flags().StringVar(&s.flagDLConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Process body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.dlProcUpdateCmd.Flags().StringVar(&s.flagDLUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.dlProcDescribeCmd.Flags().StringVar(&s.flagDLFormat, "format", "", "Output format")
	s.dlProcListCmd.Flags().StringVar(&s.flagDLFormat, "format", "", "Output format")
	s.datalineageProcessesCmd.AddCommand(procAll...)
	s.datalineageCmd.AddCommand(s.datalineageProcessesCmd)

	// runs
	runAll := []*cobra.Command{s.dlRunCreateCmd, s.dlRunDeleteCmd, s.dlRunDescribeCmd, s.dlRunListCmd, s.dlRunUpdateCmd}
	for _, c := range runAll {
		c.Flags().StringVar(&s.flagDLLocation, "location", "", "Location containing the run (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagDLProcess, "process", "", "Parent process ID (required)")
		_ = c.MarkFlagRequired("process")
	}
	for _, c := range []*cobra.Command{s.dlRunCreateCmd, s.dlRunUpdateCmd} {
		c.Flags().StringVar(&s.flagDLConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Run body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.dlRunUpdateCmd.Flags().StringVar(&s.flagDLUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.dlRunDescribeCmd.Flags().StringVar(&s.flagDLFormat, "format", "", "Output format")
	s.dlRunListCmd.Flags().StringVar(&s.flagDLFormat, "format", "", "Output format")
	s.datalineageRunsCmd.AddCommand(runAll...)
	s.datalineageCmd.AddCommand(s.datalineageRunsCmd)

	// lineage-events
	eventAll := []*cobra.Command{s.dlEventCreateCmd, s.dlEventDeleteCmd, s.dlEventDescribeCmd, s.dlEventListCmd}
	for _, c := range eventAll {
		c.Flags().StringVar(&s.flagDLLocation, "location", "", "Location containing the lineage event (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagDLProcess, "process", "", "Parent process ID (required)")
		_ = c.MarkFlagRequired("process")
		c.Flags().StringVar(&s.flagDLRun, "run", "", "Parent run ID (required)")
		_ = c.MarkFlagRequired("run")
	}
	s.dlEventCreateCmd.Flags().StringVar(&s.flagDLConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the LineageEvent body (required)")
	_ = s.dlEventCreateCmd.MarkFlagRequired("config-file")
	s.dlEventDescribeCmd.Flags().StringVar(&s.flagDLFormat, "format", "", "Output format")
	s.dlEventListCmd.Flags().StringVar(&s.flagDLFormat, "format", "", "Output format")
	s.datalineageLineageEventsCmd.AddCommand(eventAll...)
	s.datalineageCmd.AddCommand(s.datalineageLineageEventsCmd)

}

// --- config impl (raw HTTP: the Go client doesn't expose GetProjectLineageConfig
// or UpdateProjectLineageConfig, but the datalineage.googleapis.com REST API
// does) ---

func dlConfigName(project, location string) string {
	return fmt.Sprintf("%s/projectLineageConfig", dlLocationParent(project, location))
}

func (s *datalineageCmds) runDLConfigDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	return s.dlRawJSON(ctx, http.MethodGet, dlConfigName(project, s.flagDLLocation), "", nil, s.flagDLFormat)
}

func (s *datalineageCmds) runDLConfigUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	// Load user config as raw map so we can pass fields through untouched.
	var body map[string]any
	if err := loadYAMLOrJSONInto(s.flagDLConfigFile, &body); err != nil {
		return err
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	mask := s.flagDLUpdateMask
	if mask == "" {
		keys := make([]string, 0, len(body))
		for k := range body {
			keys = append(keys, k)
		}
		mask = strings.Join(keys, ",")
	}
	ctx := cmd.Context()
	return s.dlRawJSON(ctx, http.MethodPatch, dlConfigName(project, s.flagDLLocation),
		"?updateMask="+mask, payload, "")
}

func (s *datalineageCmds) dlRawJSON(ctx context.Context, method, name, query string, payload []byte, format string) error {
	ts, err := s.inv.clients().PlatformTokenSource(ctx)
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
	return emitFormattedTo(s.inv.stdout, out, format)
}

// --- processes impl ---

func dlProcName(id, project, location string) string {
	return dlChild("processes", id, dlLocationParent(project, location))
}

func (s *datalineageCmds) runDLProcCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	p := &datalineage.GoogleCloudDatacatalogLineageV1Process{}
	if err := loadYAMLOrJSONInto(s.flagDLConfigFile, p); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataLineage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Processes.Create(dlLocationParent(project, s.flagDLLocation), p).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating process: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *datalineageCmds) runDLProcDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataLineage(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Processes.Delete(dlProcName(args[0], project, s.flagDLLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting process: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted process [%s].\n", args[0])
	return nil
}

func (s *datalineageCmds) runDLProcDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataLineage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Processes.Get(dlProcName(args[0], project, s.flagDLLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing process: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDLFormat)
}

func (s *datalineageCmds) runDLProcList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataLineage(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Processes.List(dlLocationParent(project, s.flagDLLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing processes: %w", err)
	}
	if s.flagDLFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Processes, s.flagDLFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Processes), "%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, p := range resp.Processes {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(p.Name), p.DisplayName)
	}
	return nil
}

func (s *datalineageCmds) runDLProcUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	p := &datalineage.GoogleCloudDatacatalogLineageV1Process{}
	if err := loadYAMLOrJSONInto(s.flagDLConfigFile, p); err != nil {
		return err
	}
	mask := s.flagDLUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(p))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataLineage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Processes.Patch(dlProcName(args[0], project, s.flagDLLocation), p).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating process: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

// --- runs impl ---

func dlRunParent(project, location, process string) string {
	return fmt.Sprintf("%s/processes/%s", dlLocationParent(project, location), process)
}

func dlRunName(id, project, location, process string) string {
	return dlChild("runs", id, dlRunParent(project, location, process))
}

func (s *datalineageCmds) runDLRunCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	r := &datalineage.GoogleCloudDatacatalogLineageV1Run{}
	if err := loadYAMLOrJSONInto(s.flagDLConfigFile, r); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataLineage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Processes.Runs.Create(dlRunParent(project, s.flagDLLocation, s.flagDLProcess), r).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating run: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *datalineageCmds) runDLRunDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataLineage(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Processes.Runs.Delete(dlRunName(args[0], project, s.flagDLLocation, s.flagDLProcess)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting run: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted run [%s].\n", args[0])
	return nil
}

func (s *datalineageCmds) runDLRunDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataLineage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Processes.Runs.Get(dlRunName(args[0], project, s.flagDLLocation, s.flagDLProcess)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing run: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDLFormat)
}

func (s *datalineageCmds) runDLRunList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataLineage(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Processes.Runs.List(dlRunParent(project, s.flagDLLocation, s.flagDLProcess)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing runs: %w", err)
	}
	if s.flagDLFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Runs, s.flagDLFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Runs), "%-40s %s\n", "NAME", "STATE")
	for _, r := range resp.Runs {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(r.Name), r.State)
	}
	return nil
}

func (s *datalineageCmds) runDLRunUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	r := &datalineage.GoogleCloudDatacatalogLineageV1Run{}
	if err := loadYAMLOrJSONInto(s.flagDLConfigFile, r); err != nil {
		return err
	}
	mask := s.flagDLUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(r))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataLineage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Processes.Runs.Patch(dlRunName(args[0], project, s.flagDLLocation, s.flagDLProcess), r).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating run: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

// --- lineage-events impl ---

func dlEventParent(project, location, process, run string) string {
	return fmt.Sprintf("%s/runs/%s", dlRunParent(project, location, process), run)
}

func dlEventName(id, project, location, process, run string) string {
	return dlChild("lineageEvents", id, dlEventParent(project, location, process, run))
}

func (s *datalineageCmds) runDLEventCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	e := &datalineage.GoogleCloudDatacatalogLineageV1LineageEvent{}
	if err := loadYAMLOrJSONInto(s.flagDLConfigFile, e); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataLineage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Processes.Runs.LineageEvents.Create(dlEventParent(project, s.flagDLLocation, s.flagDLProcess, s.flagDLRun), e).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating lineage event: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *datalineageCmds) runDLEventDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataLineage(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Processes.Runs.LineageEvents.Delete(dlEventName(args[0], project, s.flagDLLocation, s.flagDLProcess, s.flagDLRun)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting lineage event: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted lineage event [%s].\n", args[0])
	return nil
}

func (s *datalineageCmds) runDLEventDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataLineage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Processes.Runs.LineageEvents.Get(dlEventName(args[0], project, s.flagDLLocation, s.flagDLProcess, s.flagDLRun)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing lineage event: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDLFormat)
}

func (s *datalineageCmds) runDLEventList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataLineage(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Processes.Runs.LineageEvents.List(dlEventParent(project, s.flagDLLocation, s.flagDLProcess, s.flagDLRun)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing lineage events: %w", err)
	}
	if s.flagDLFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.LineageEvents, s.flagDLFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.LineageEvents), "%-40s %s\n", "NAME", "START_TIME")
	for _, e := range resp.LineageEvents {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(e.Name), e.StartTime)
	}
	return nil
}

// datalineageCmds holds the commands and flag values of one invocation's "datalineage"
// command group, so invocations don't share state (#2070).
type datalineageCmds struct {
	inv *invocation

	datalineageCmd              *cobra.Command
	flagDLLocation              string
	flagDLConfigFile            string
	flagDLUpdateMask            string
	flagDLFormat                string
	flagDLProcess               string
	flagDLRun                   string
	datalineageConfigCmd        *cobra.Command
	dlConfigDescribeCmd         *cobra.Command
	dlConfigUpdateCmd           *cobra.Command
	datalineageRunsCmd          *cobra.Command
	dlRunCreateCmd              *cobra.Command
	dlRunDeleteCmd              *cobra.Command
	dlRunDescribeCmd            *cobra.Command
	dlRunListCmd                *cobra.Command
	dlRunUpdateCmd              *cobra.Command
	datalineageLineageEventsCmd *cobra.Command
	dlEventCreateCmd            *cobra.Command
	dlEventDeleteCmd            *cobra.Command
	dlEventDescribeCmd          *cobra.Command
	dlEventListCmd              *cobra.Command
	datalineageProcessesCmd     *cobra.Command
	dlProcCreateCmd             *cobra.Command
	dlProcDeleteCmd             *cobra.Command
	dlProcDescribeCmd           *cobra.Command
	dlProcListCmd               *cobra.Command
	dlProcUpdateCmd             *cobra.Command
}

// newDatalineageCmd returns the "datalineage" command group for inv.
func newDatalineageCmd(inv *invocation) *cobra.Command {
	return newDatalineageCmds(inv).datalineageCmd
}

// newDatalineageCmds builds the "datalineage" command group for inv and returns its state.
func newDatalineageCmds(inv *invocation) *datalineageCmds {
	s := &datalineageCmds{inv: inv}
	s.datalineageCmd = &cobra.Command{Use: "datalineage", Short: "Manage Data Lineage"}
	s.datalineageConfigCmd = &cobra.Command{Use: "config", Short: "Manage Data Lineage config"}
	s.dlConfigDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe the project lineage config for a location",
		Args: cobra.NoArgs, RunE: s.runDLConfigDescribe,
	}
	s.dlConfigUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update the project lineage config from a --config-file",
		Args: cobra.NoArgs, RunE: s.runDLConfigUpdate,
	}
	s.datalineageRunsCmd = &cobra.Command{Use: "runs", Short: "Manage Data Lineage runs"}
	s.dlRunCreateCmd = &cobra.Command{
		Use: "create RUN", Short: "Create a run under a process from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runDLRunCreate,
	}
	s.dlRunDeleteCmd = &cobra.Command{
		Use: "delete RUN", Short: "Delete a run",
		Args: cobra.ExactArgs(1), RunE: s.runDLRunDelete,
	}
	s.dlRunDescribeCmd = &cobra.Command{
		Use: "describe RUN", Short: "Describe a run",
		Args: cobra.ExactArgs(1), RunE: s.runDLRunDescribe,
	}
	s.dlRunListCmd = &cobra.Command{
		Use: "list", Short: "List runs under a process",
		Args: cobra.NoArgs, RunE: s.runDLRunList,
	}
	s.dlRunUpdateCmd = &cobra.Command{
		Use: "update RUN", Short: "Update a run from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runDLRunUpdate,
	}
	s.datalineageLineageEventsCmd = &cobra.Command{Use: "lineage-events", Short: "Manage Data Lineage lineage events"}
	s.dlEventCreateCmd = &cobra.Command{
		Use: "create EVENT", Short: "Create a lineage event under a run from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runDLEventCreate,
	}
	s.dlEventDeleteCmd = &cobra.Command{
		Use: "delete EVENT", Short: "Delete a lineage event",
		Args: cobra.ExactArgs(1), RunE: s.runDLEventDelete,
	}
	s.dlEventDescribeCmd = &cobra.Command{
		Use: "describe EVENT", Short: "Describe a lineage event",
		Args: cobra.ExactArgs(1), RunE: s.runDLEventDescribe,
	}
	s.dlEventListCmd = &cobra.Command{
		Use: "list", Short: "List lineage events under a run",
		Args: cobra.NoArgs, RunE: s.runDLEventList,
	}
	s.datalineageProcessesCmd = &cobra.Command{Use: "processes", Short: "Manage Data Lineage processes"}
	s.dlProcCreateCmd = &cobra.Command{
		Use: "create PROCESS", Short: "Create a process from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runDLProcCreate,
	}
	s.dlProcDeleteCmd = &cobra.Command{
		Use: "delete PROCESS", Short: "Delete a process",
		Args: cobra.ExactArgs(1), RunE: s.runDLProcDelete,
	}
	s.dlProcDescribeCmd = &cobra.Command{
		Use: "describe PROCESS", Short: "Describe a process",
		Args: cobra.ExactArgs(1), RunE: s.runDLProcDescribe,
	}
	s.dlProcListCmd = &cobra.Command{
		Use: "list", Short: "List processes in a location",
		Args: cobra.NoArgs, RunE: s.runDLProcList,
	}
	s.dlProcUpdateCmd = &cobra.Command{
		Use: "update PROCESS", Short: "Update a process from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runDLProcUpdate,
	}
	s.initDatalineage()
	return s
}
