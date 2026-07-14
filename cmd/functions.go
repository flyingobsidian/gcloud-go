package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudfunctionsv1 "google.golang.org/api/cloudfunctions/v1"
	cloudfunctions "google.golang.org/api/cloudfunctions/v2"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud functions (#342, #860-#864) ---

var functionsCmd = &cobra.Command{Use: "functions", Short: "Manage Cloud Functions"}

// Shared flags for the IAM-policy, event-types, logs, regions, runtimes subcommands.
var (
	flagFuncRegion        string
	flagFuncFormat        string
	flagFuncMember        string
	flagFuncRole          string
	flagFuncPolicyFile    string
	flagFuncLimit         int64
	flagFuncMinLogLevel   string
	flagFuncExecutionID   string
	flagFuncStartTime     string
	flagFuncEndTime       string
	flagFuncGen2          bool
)

// --- top-level IAM commands ---

var (
	funcGetIamCmd = &cobra.Command{
		Use: "get-iam-policy FUNCTION", Short: "Get the IAM policy for a Cloud Function",
		Args: cobra.ExactArgs(1), RunE: runFuncGetIam,
	}
	funcSetIamCmd = &cobra.Command{
		Use: "set-iam-policy FUNCTION POLICY_FILE", Short: "Set the IAM policy for a Cloud Function",
		Args: cobra.ExactArgs(2), RunE: runFuncSetIam,
	}
	funcAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding FUNCTION", Short: "Add an IAM policy binding to a Cloud Function",
		Args: cobra.ExactArgs(1), RunE: runFuncAddIam,
	}
	funcRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding FUNCTION", Short: "Remove an IAM policy binding from a Cloud Function",
		Args: cobra.ExactArgs(1), RunE: runFuncRemoveIam,
	}
)

// --- subgroups ---

var (
	funcEventTypesCmd = &cobra.Command{Use: "event-types", Short: "List trigger event types for Cloud Functions"}
	funcEventTypesListCmd = &cobra.Command{
		Use: "list", Short: "List event types available to Cloud Functions",
		Args: cobra.NoArgs, RunE: runFuncEventTypesList,
	}

	funcLogsCmd = &cobra.Command{Use: "logs", Short: "Manage Cloud Functions logs"}
	funcLogsReadCmd = &cobra.Command{
		Use: "read [NAME]", Short: "Read log entries produced by Cloud Functions",
		Args: cobra.MaximumNArgs(1), RunE: runFuncLogsRead,
	}

	funcRegionsCmd = &cobra.Command{Use: "regions", Short: "List regions available to Cloud Functions"}
	funcRegionsListCmd = &cobra.Command{
		Use: "list", Short: "List regions available to Cloud Functions",
		Args: cobra.NoArgs, RunE: runFuncRegionsList,
	}

	funcRuntimesCmd = &cobra.Command{Use: "runtimes", Short: "List runtimes available to Cloud Functions"}
	funcRuntimesListCmd = &cobra.Command{
		Use: "list", Short: "List runtimes available to Cloud Functions",
		Args: cobra.NoArgs, RunE: runFuncRuntimesList,
	}
)

func init() {
	// Legacy stubs still to be implemented under #342.
	for _, name := range []string{
		"add-invoker-policy-binding", "call", "delete", "deploy", "describe", "detach",
		"list", "remove-invoker-policy-binding",
	} {
		registerStubCommand(functionsCmd, name, "Not yet implemented")
	}

	// IAM commands (#860).
	for _, c := range []*cobra.Command{funcGetIamCmd, funcSetIamCmd, funcAddIamCmd, funcRemoveIamCmd} {
		c.Flags().StringVar(&flagFuncRegion, "region", "", "The Cloud region of the function (required)")
		c.Flags().StringVar(&flagFuncFormat, "format", "", "Output format")
		c.Flags().BoolVar(&flagFuncGen2, "gen2", true, "Operate against the Cloud Functions v2 API (default true; --gen2=false uses v1)")
		_ = c.MarkFlagRequired("region")
	}
	funcAddIamCmd.Flags().StringVar(&flagFuncMember, "member", "", "The IAM member (required)")
	funcAddIamCmd.Flags().StringVar(&flagFuncRole, "role", "", "The IAM role (required)")
	_ = funcAddIamCmd.MarkFlagRequired("member")
	_ = funcAddIamCmd.MarkFlagRequired("role")
	funcRemoveIamCmd.Flags().StringVar(&flagFuncMember, "member", "", "The IAM member (required)")
	funcRemoveIamCmd.Flags().StringVar(&flagFuncRole, "role", "", "The IAM role (required)")
	_ = funcRemoveIamCmd.MarkFlagRequired("member")
	_ = funcRemoveIamCmd.MarkFlagRequired("role")
	functionsCmd.AddCommand(funcGetIamCmd, funcSetIamCmd, funcAddIamCmd, funcRemoveIamCmd)

	// event-types (#861).
	funcEventTypesListCmd.Flags().StringVar(&flagFuncFormat, "format", "", "Output format")
	funcEventTypesListCmd.Flags().BoolVar(&flagFuncGen2, "gen2", false, "List Gen2 (Eventarc) event types rather than the Gen1 static list")
	funcEventTypesCmd.AddCommand(funcEventTypesListCmd)
	functionsCmd.AddCommand(funcEventTypesCmd)

	// logs (#862).
	funcLogsReadCmd.Flags().StringVar(&flagFuncRegion, "region", "", "The Cloud region of the function (required)")
	funcLogsReadCmd.Flags().StringVar(&flagFuncExecutionID, "execution-id", "", "Only return log entries for the given execution ID")
	funcLogsReadCmd.Flags().StringVar(&flagFuncMinLogLevel, "min-log-level", "", "Minimum severity to include (DEBUG, INFO, NOTICE, WARNING, ERROR, CRITICAL, ALERT, EMERGENCY)")
	funcLogsReadCmd.Flags().StringVar(&flagFuncStartTime, "start-time", "", "RFC3339 timestamp; defaults to 7 days ago")
	funcLogsReadCmd.Flags().StringVar(&flagFuncEndTime, "end-time", "", "RFC3339 timestamp; if omitted, current time")
	funcLogsReadCmd.Flags().Int64Var(&flagFuncLimit, "limit", 20, "Maximum number of log entries to return (1-1000)")
	funcLogsReadCmd.Flags().StringVar(&flagFuncFormat, "format", "", "Output format")
	funcLogsReadCmd.Flags().BoolVar(&flagFuncGen2, "gen2", false, "Only include Gen2 (Cloud Run) function logs")
	_ = funcLogsReadCmd.MarkFlagRequired("region")
	funcLogsCmd.AddCommand(funcLogsReadCmd)
	functionsCmd.AddCommand(funcLogsCmd)

	// regions (#863).
	funcRegionsListCmd.Flags().StringVar(&flagFuncFormat, "format", "", "Output format")
	funcRegionsCmd.AddCommand(funcRegionsListCmd)
	functionsCmd.AddCommand(funcRegionsCmd)

	// runtimes (#864).
	funcRuntimesListCmd.Flags().StringVar(&flagFuncRegion, "region", "", "Only show runtimes within the region")
	funcRuntimesListCmd.Flags().StringVar(&flagFuncFormat, "format", "", "Output format")
	funcRuntimesCmd.AddCommand(funcRuntimesListCmd)
	functionsCmd.AddCommand(funcRuntimesCmd)

	rootCmd.AddCommand(functionsCmd)
}

// --- helpers ---

func functionResourceName(project, region, name string) string {
	if strings.HasPrefix(name, "projects/") {
		return name
	}
	return fmt.Sprintf("projects/%s/locations/%s/functions/%s", project, region, name)
}

func functionLocationParent(project, region string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, region)
}

// --- IAM implementations ---

func runFuncGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	resource := functionResourceName(project, flagFuncRegion, args[0])
	if flagFuncGen2 {
		svc, err := gcp.CloudFunctionsV2Service(ctx, flagAccount)
		if err != nil {
			return err
		}
		policy, err := svc.Projects.Locations.Functions.GetIamPolicy(resource).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("getting IAM policy: %w", err)
		}
		return emitFormatted(policy, flagFuncFormat)
	}
	svc, err := gcp.CloudFunctionsV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Functions.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagFuncFormat)
}

func runFuncSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	resource := functionResourceName(project, flagFuncRegion, args[0])
	if flagFuncGen2 {
		policy := &cloudfunctions.Policy{}
		if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
			return err
		}
		svc, err := gcp.CloudFunctionsV2Service(ctx, flagAccount)
		if err != nil {
			return err
		}
		got, err := svc.Projects.Locations.Functions.SetIamPolicy(resource, &cloudfunctions.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("setting IAM policy: %w", err)
		}
		return emitFormatted(got, flagFuncFormat)
	}
	svc, err := gcp.CloudFunctionsV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy := &cloudfunctionsv1.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Functions.SetIamPolicy(resource, &cloudfunctionsv1.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, flagFuncFormat)
}

func runFuncAddIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	resource := functionResourceName(project, flagFuncRegion, args[0])
	if flagFuncGen2 {
		svc, err := gcp.CloudFunctionsV2Service(ctx, flagAccount)
		if err != nil {
			return err
		}
		policy, err := svc.Projects.Locations.Functions.GetIamPolicy(resource).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("getting IAM policy: %w", err)
		}
		addBindingV2(policy, flagFuncRole, flagFuncMember)
		got, err := svc.Projects.Locations.Functions.SetIamPolicy(resource, &cloudfunctions.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("updating IAM policy: %w", err)
		}
		return emitFormatted(got, flagFuncFormat)
	}
	svc, err := gcp.CloudFunctionsV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Functions.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	addBindingV1(policy, flagFuncRole, flagFuncMember)
	got, err := svc.Projects.Locations.Functions.SetIamPolicy(resource, &cloudfunctionsv1.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating IAM policy: %w", err)
	}
	return emitFormatted(got, flagFuncFormat)
}

func runFuncRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	resource := functionResourceName(project, flagFuncRegion, args[0])
	if flagFuncGen2 {
		svc, err := gcp.CloudFunctionsV2Service(ctx, flagAccount)
		if err != nil {
			return err
		}
		policy, err := svc.Projects.Locations.Functions.GetIamPolicy(resource).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("getting IAM policy: %w", err)
		}
		removeBindingV2(policy, flagFuncRole, flagFuncMember)
		got, err := svc.Projects.Locations.Functions.SetIamPolicy(resource, &cloudfunctions.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("updating IAM policy: %w", err)
		}
		return emitFormatted(got, flagFuncFormat)
	}
	svc, err := gcp.CloudFunctionsV1Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Functions.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	removeBindingV1(policy, flagFuncRole, flagFuncMember)
	got, err := svc.Projects.Locations.Functions.SetIamPolicy(resource, &cloudfunctionsv1.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating IAM policy: %w", err)
	}
	return emitFormatted(got, flagFuncFormat)
}

func addBindingV1(p *cloudfunctionsv1.Policy, role, member string) {
	for _, b := range p.Bindings {
		if b.Role != role {
			continue
		}
		for _, m := range b.Members {
			if m == member {
				return
			}
		}
		b.Members = append(b.Members, member)
		return
	}
	p.Bindings = append(p.Bindings, &cloudfunctionsv1.Binding{Role: role, Members: []string{member}})
}

func removeBindingV1(p *cloudfunctionsv1.Policy, role, member string) {
	for _, b := range p.Bindings {
		if b.Role != role {
			continue
		}
		filtered := b.Members[:0]
		for _, m := range b.Members {
			if m != member {
				filtered = append(filtered, m)
			}
		}
		b.Members = filtered
	}
}

func addBindingV2(p *cloudfunctions.Policy, role, member string) {
	for _, b := range p.Bindings {
		if b.Role != role {
			continue
		}
		for _, m := range b.Members {
			if m == member {
				return
			}
		}
		b.Members = append(b.Members, member)
		return
	}
	p.Bindings = append(p.Bindings, &cloudfunctions.Binding{Role: role, Members: []string{member}})
}

func removeBindingV2(p *cloudfunctions.Policy, role, member string) {
	for _, b := range p.Bindings {
		if b.Role != role {
			continue
		}
		filtered := b.Members[:0]
		for _, m := range b.Members {
			if m != member {
				filtered = append(filtered, m)
			}
		}
		b.Members = filtered
	}
}

// --- event-types (#861) ---

type funcTriggerEventRow struct {
	Provider        string `json:"provider"`
	EventType       string `json:"event_type"`
	Default         string `json:"event_type_default,omitempty"`
	ResourceType    string `json:"resource_type,omitempty"`
	ResourceOption  string `json:"resource_optional,omitempty"`
}

var gen1TriggerEvents = []funcTriggerEventRow{
	{Provider: "cloud.pubsub", EventType: "google.pubsub.topic.publish", Default: "Yes", ResourceType: "topic"},
	{Provider: "cloud.pubsub", EventType: "providers/cloud.pubsub/eventTypes/topic.publish", ResourceType: "topic"},
	{Provider: "cloud.storage", EventType: "google.storage.object.finalize", Default: "Yes", ResourceType: "bucket"},
	{Provider: "cloud.storage", EventType: "providers/cloud.storage/eventTypes/object.change", ResourceType: "bucket"},
	{Provider: "cloud.storage", EventType: "google.storage.object.archive", ResourceType: "bucket"},
	{Provider: "cloud.storage", EventType: "google.storage.object.delete", ResourceType: "bucket"},
	{Provider: "cloud.storage", EventType: "google.storage.object.metadataUpdate", ResourceType: "bucket"},
	{Provider: "google.firebase.database.ref", EventType: "providers/google.firebase.database/eventTypes/ref.create", Default: "Yes", ResourceType: "firebase database"},
	{Provider: "google.firebase.database.ref", EventType: "providers/google.firebase.database/eventTypes/ref.update", ResourceType: "firebase database"},
	{Provider: "google.firebase.database.ref", EventType: "providers/google.firebase.database/eventTypes/ref.delete", ResourceType: "firebase database"},
	{Provider: "google.firebase.database.ref", EventType: "providers/google.firebase.database/eventTypes/ref.write", ResourceType: "firebase database"},
	{Provider: "google.firestore.document", EventType: "providers/cloud.firestore/eventTypes/document.create", Default: "Yes", ResourceType: "firestore document"},
	{Provider: "google.firestore.document", EventType: "providers/cloud.firestore/eventTypes/document.update", ResourceType: "firestore document"},
	{Provider: "google.firestore.document", EventType: "providers/cloud.firestore/eventTypes/document.delete", ResourceType: "firestore document"},
	{Provider: "google.firestore.document", EventType: "providers/cloud.firestore/eventTypes/document.write", ResourceType: "firestore document"},
	{Provider: "google.firebase.analytics.event", EventType: "providers/google.firebase.analytics/eventTypes/event.log", Default: "Yes", ResourceType: "firebase analytics"},
	{Provider: "google.firebase.remoteConfig", EventType: "google.firebase.remoteconfig.update", Default: "Yes", ResourceType: "project", ResourceOption: "Yes"},
	{Provider: "firebase.auth", EventType: "providers/firebase.auth/eventTypes/user.create", Default: "Yes", ResourceType: "project", ResourceOption: "Yes"},
	{Provider: "firebase.auth", EventType: "providers/firebase.auth/eventTypes/user.delete", ResourceType: "project", ResourceOption: "Yes"},
}

type funcGen2EventRow struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// gen2TriggerEvents mirrors the Eventarc-backed events surfaced by
// `gcloud functions event-types list --gen2`.
var gen2TriggerEvents = []funcGen2EventRow{
	{Name: "google.cloud.pubsub.topic.v1.messagePublished", Description: "A Pub/Sub message is published to a topic."},
	{Name: "google.cloud.storage.object.v1.finalized", Description: "An object is created or overwritten in a Cloud Storage bucket."},
	{Name: "google.cloud.storage.object.v1.deleted", Description: "An object is deleted from a Cloud Storage bucket."},
	{Name: "google.cloud.storage.object.v1.archived", Description: "An object is archived in a Cloud Storage bucket."},
	{Name: "google.cloud.storage.object.v1.metadataUpdated", Description: "An object's metadata is updated in a Cloud Storage bucket."},
	{Name: "google.cloud.firestore.document.v1.created", Description: "A Firestore document is created."},
	{Name: "google.cloud.firestore.document.v1.updated", Description: "A Firestore document is updated."},
	{Name: "google.cloud.firestore.document.v1.deleted", Description: "A Firestore document is deleted."},
	{Name: "google.cloud.firestore.document.v1.written", Description: "A Firestore document is written (create, update, or delete)."},
	{Name: "google.cloud.audit.log.v1.written", Description: "An audit log entry is written."},
}

func runFuncEventTypesList(cmd *cobra.Command, args []string) error {
	if flagFuncGen2 {
		if flagFuncFormat != "" {
			return emitFormatted(gen2TriggerEvents, flagFuncFormat)
		}
		fmt.Printf("%-60s %s\n", "NAME", "DESCRIPTION")
		for _, e := range gen2TriggerEvents {
			fmt.Printf("%-60s %s\n", e.Name, e.Description)
		}
		return nil
	}
	if flagFuncFormat != "" {
		return emitFormatted(gen1TriggerEvents, flagFuncFormat)
	}
	fmt.Printf("%-32s %-60s %-18s %-22s %s\n",
		"EVENT_PROVIDER", "EVENT_TYPE", "EVENT_TYPE_DEFAULT", "RESOURCE_TYPE", "RESOURCE_OPTIONAL")
	for _, e := range gen1TriggerEvents {
		fmt.Printf("%-32s %-60s %-18s %-22s %s\n",
			e.Provider, e.EventType, e.Default, e.ResourceType, e.ResourceOption)
	}
	return nil
}

// --- logs (#862) ---

type funcLogRow struct {
	Level       string `json:"level,omitempty"`
	Name        string `json:"name,omitempty"`
	ExecutionID string `json:"execution_id,omitempty"`
	TimeUTC     string `json:"time_utc,omitempty"`
	Log         string `json:"log,omitempty"`
}

func funcParseTime(v string) (time.Time, error) {
	if v == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, v)
}

func funcBuildLogFilter(project, region, name string) (string, error) {
	gen1 := []string{
		`resource.type="cloud_function"`,
		fmt.Sprintf(`resource.labels.region="%s"`, region),
		`logName:"cloud-functions"`,
	}
	gen2 := []string{
		`resource.type="cloud_run_revision"`,
		fmt.Sprintf(`resource.labels.location="%s"`, region),
		`logName:"run.googleapis.com"`,
		`labels."goog-managed-by"="cloudfunctions"`,
	}
	if name != "" {
		gen1 = append(gen1, fmt.Sprintf(`resource.labels.function_name="%s"`, name))
		service := strings.ReplaceAll(strings.ToLower(name), "_", "-")
		gen2 = append(gen2, fmt.Sprintf(`resource.labels.service_name="%s"`, service))
	}
	var parts []string
	if flagFuncGen2 {
		parts = append(parts, strings.Join(gen2, " "))
	} else {
		// Match python behaviour: without --gen2 we return both by default.
		parts = append(parts, fmt.Sprintf("(%s) OR (%s)", strings.Join(gen1, " "), strings.Join(gen2, " ")))
	}
	if flagFuncExecutionID != "" {
		parts = append(parts, fmt.Sprintf(`labels.execution_id="%s"`, flagFuncExecutionID))
	}
	if flagFuncMinLogLevel != "" {
		parts = append(parts, fmt.Sprintf(`severity>=%s`, strings.ToUpper(flagFuncMinLogLevel)))
	}
	end, err := funcParseTime(flagFuncEndTime)
	if err != nil {
		return "", fmt.Errorf("invalid --end-time: %w", err)
	}
	if !end.IsZero() {
		parts = append(parts, fmt.Sprintf(`timestamp<="%s"`, end.UTC().Format(time.RFC3339)))
	}
	start, err := funcParseTime(flagFuncStartTime)
	if err != nil {
		return "", fmt.Errorf("invalid --start-time: %w", err)
	}
	if start.IsZero() {
		start = time.Now().UTC().Add(-7 * 24 * time.Hour)
	}
	parts = append(parts, fmt.Sprintf(`timestamp>="%s"`, start.UTC().Format(time.RFC3339)))
	_ = project
	return strings.Join(parts, " "), nil
}

func runFuncLogsRead(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	if flagFuncLimit < 1 || flagFuncLimit > 1000 {
		return fmt.Errorf("--limit must be between 1 and 1000")
	}
	name := ""
	if len(args) == 1 {
		name = args[0]
	}
	filter, err := funcBuildLogFilter(project, flagFuncRegion, name)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.LoggingService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &logging.ListLogEntriesRequest{
		Filter:        filter,
		OrderBy:       "timestamp desc",
		PageSize:      flagFuncLimit,
		ResourceNames: []string{"projects/" + project},
	}
	resp, err := svc.Entries.List(req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing log entries: %w", err)
	}
	rows := make([]funcLogRow, 0, len(resp.Entries))
	for _, e := range resp.Entries {
		rows = append(rows, funcLogEntryToRow(e))
	}
	if flagFuncFormat != "" {
		return emitFormatted(rows, flagFuncFormat)
	}
	fmt.Printf("%-6s %-32s %-32s %-24s %s\n", "LEVEL", "NAME", "EXECUTION_ID", "TIME_UTC", "LOG")
	for _, r := range rows {
		fmt.Printf("%-6s %-32s %-32s %-24s %s\n", r.Level, r.Name, r.ExecutionID, r.TimeUTC, r.Log)
	}
	return nil
}

func funcLogEntryToRow(e *logging.LogEntry) funcLogRow {
	row := funcLogRow{}
	if e.TextPayload != "" {
		row.Log = e.TextPayload
	} else if len(e.JsonPayload) > 0 {
		row.Log = string(e.JsonPayload)
	} else if len(e.ProtoPayload) > 0 {
		row.Log = string(e.ProtoPayload)
	}
	switch strings.ToUpper(e.Severity) {
	case "DEBUG", "INFO", "NOTICE", "WARNING", "ERROR", "CRITICAL", "ALERT", "EMERGENCY":
		row.Level = string(e.Severity[0])
	default:
		row.Level = e.Severity
	}
	if e.Resource != nil {
		if v, ok := e.Resource.Labels["function_name"]; ok {
			row.Name = v
		} else if v, ok := e.Resource.Labels["service_name"]; ok {
			row.Name = v
		}
	}
	if v, ok := e.Labels["execution_id"]; ok {
		row.ExecutionID = v
	}
	row.TimeUTC = e.Timestamp
	return row
}

// --- regions (#863) ---

func runFuncRegionsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudFunctionsV2Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*cloudfunctions.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List("projects/" + project).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing regions: %w", err)
		}
		all = append(all, resp.Locations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagFuncFormat != "" {
		return emitFormatted(all, flagFuncFormat)
	}
	fmt.Printf("%s\n", "NAME")
	for _, l := range all {
		id := l.LocationId
		if id == "" {
			id = path.Base(l.Name)
		}
		fmt.Printf("%s\n", id)
	}
	return nil
}

// --- runtimes (#864) ---

func runFuncRuntimesList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	region := flagFuncRegion
	if region == "" {
		region = "-"
	}
	ctx := context.Background()
	svc, err := gcp.CloudFunctionsV2Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Runtimes.List(functionLocationParent(project, region)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing runtimes: %w", err)
	}
	if flagFuncFormat != "" {
		return emitFormatted(resp.Runtimes, flagFuncFormat)
	}
	fmt.Printf("%-24s %-16s %s\n", "NAME", "STAGE", "ENVIRONMENT")
	for _, r := range resp.Runtimes {
		fmt.Printf("%-24s %-16s %s\n", r.Name, r.Stage, r.Environment)
	}
	return nil
}
