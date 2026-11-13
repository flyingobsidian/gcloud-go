package cli

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
	cloudfunctionsv1 "google.golang.org/api/cloudfunctions/v1"
	cloudfunctions "google.golang.org/api/cloudfunctions/v2"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud functions (#342, #860-#864) ---

// Shared flags for the IAM-policy, event-types, logs, regions, runtimes subcommands.

// --- top-level IAM commands ---

// --- subgroups ---

func (s *functionsCmds) initFunctions() {
	// Legacy stubs still to be implemented under #342.
	for _, name := range []string{
		"add-invoker-policy-binding", "call", "delete", "deploy", "describe", "detach",
		"remove-invoker-policy-binding",
	} {
		registerStubCommand(s.functionsCmd, name, "Not yet implemented")
	}

	// IAM commands (#860).
	for _, c := range []*cobra.Command{s.funcGetIamCmd, s.funcSetIamCmd, s.funcAddIamCmd, s.funcRemoveIamCmd} {
		c.Flags().StringVar(&s.flagFuncRegion, "region", "", "The Cloud region of the function (required)")
		c.Flags().StringVar(&s.flagFuncFormat, "format", "", "Output format")
		c.Flags().BoolVar(&s.flagFuncGen2, "gen2", true, "Operate against the Cloud Functions v2 API (default true; --gen2=false uses v1)")
		_ = c.MarkFlagRequired("region")
	}
	s.funcAddIamCmd.Flags().StringVar(&s.flagFuncMember, "member", "", "The IAM member (required)")
	s.funcAddIamCmd.Flags().StringVar(&s.flagFuncRole, "role", "", "The IAM role (required)")
	_ = s.funcAddIamCmd.MarkFlagRequired("member")
	_ = s.funcAddIamCmd.MarkFlagRequired("role")
	s.funcRemoveIamCmd.Flags().StringVar(&s.flagFuncMember, "member", "", "The IAM member (required)")
	s.funcRemoveIamCmd.Flags().StringVar(&s.flagFuncRole, "role", "", "The IAM role (required)")
	_ = s.funcRemoveIamCmd.MarkFlagRequired("member")
	_ = s.funcRemoveIamCmd.MarkFlagRequired("role")
	s.functionsCmd.AddCommand(s.funcGetIamCmd, s.funcSetIamCmd, s.funcAddIamCmd, s.funcRemoveIamCmd)

	// event-types (#861).
	s.funcEventTypesListCmd.Flags().StringVar(&s.flagFuncFormat, "format", "", "Output format")
	s.funcEventTypesListCmd.Flags().BoolVar(&s.flagFuncGen2, "gen2", false, "List Gen2 (Eventarc) event types rather than the Gen1 static list")
	s.funcEventTypesCmd.AddCommand(s.funcEventTypesListCmd)
	s.functionsCmd.AddCommand(s.funcEventTypesCmd)

	// logs (#862).
	s.funcLogsReadCmd.Flags().StringVar(&s.flagFuncRegion, "region", "", "The Cloud region of the function (required)")
	s.funcLogsReadCmd.Flags().StringVar(&s.flagFuncExecutionID, "execution-id", "", "Only return log entries for the given execution ID")
	s.funcLogsReadCmd.Flags().StringVar(&s.flagFuncMinLogLevel, "min-log-level", "", "Minimum severity to include (DEBUG, INFO, NOTICE, WARNING, ERROR, CRITICAL, ALERT, EMERGENCY)")
	s.funcLogsReadCmd.Flags().StringVar(&s.flagFuncStartTime, "start-time", "", "RFC3339 timestamp; defaults to 7 days ago")
	s.funcLogsReadCmd.Flags().StringVar(&s.flagFuncEndTime, "end-time", "", "RFC3339 timestamp; if omitted, current time")
	s.funcLogsReadCmd.Flags().Int64Var(&s.flagFuncLimit, "limit", 20, "Maximum number of log entries to return (1-1000)")
	s.funcLogsReadCmd.Flags().StringVar(&s.flagFuncFormat, "format", "", "Output format")
	s.funcLogsReadCmd.Flags().BoolVar(&s.flagFuncGen2, "gen2", false, "Only include Gen2 (Cloud Run) function logs")
	_ = s.funcLogsReadCmd.MarkFlagRequired("region")
	s.funcLogsCmd.AddCommand(s.funcLogsReadCmd)
	s.functionsCmd.AddCommand(s.funcLogsCmd)

	// regions (#863).
	s.funcRegionsListCmd.Flags().StringVar(&s.flagFuncFormat, "format", "", "Output format")
	s.funcRegionsCmd.AddCommand(s.funcRegionsListCmd)
	s.functionsCmd.AddCommand(s.funcRegionsCmd)

	// runtimes (#864).
	s.funcRuntimesListCmd.Flags().StringVar(&s.flagFuncRegion, "region", "", "Only show runtimes within the region")
	s.funcRuntimesListCmd.Flags().StringVar(&s.flagFuncFormat, "format", "", "Output format")
	s.funcRuntimesCmd.AddCommand(s.funcRuntimesListCmd)
	s.functionsCmd.AddCommand(s.funcRuntimesCmd)

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

func (s *functionsCmds) runFuncGetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	resource := functionResourceName(project, s.flagFuncRegion, args[0])
	if s.flagFuncGen2 {
		svc, err := s.inv.clients().CloudFunctionsV2(ctx)
		if err != nil {
			return err
		}
		policy, err := svc.Projects.Locations.Functions.GetIamPolicy(resource).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("getting IAM policy: %w", err)
		}
		return emitFormattedTo(s.inv.stdout, policy, s.flagFuncFormat)
	}
	svc, err := s.inv.clients().CloudFunctionsV1(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Functions.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagFuncFormat)
}

func (s *functionsCmds) runFuncSetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	resource := functionResourceName(project, s.flagFuncRegion, args[0])
	if s.flagFuncGen2 {
		policy := &cloudfunctions.Policy{}
		if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
			return err
		}
		svc, err := s.inv.clients().CloudFunctionsV2(ctx)
		if err != nil {
			return err
		}
		got, err := svc.Projects.Locations.Functions.SetIamPolicy(resource, &cloudfunctions.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("setting IAM policy: %w", err)
		}
		return emitFormattedTo(s.inv.stdout, got, s.flagFuncFormat)
	}
	svc, err := s.inv.clients().CloudFunctionsV1(ctx)
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
	return emitFormattedTo(s.inv.stdout, got, s.flagFuncFormat)
}

func (s *functionsCmds) runFuncAddIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	resource := functionResourceName(project, s.flagFuncRegion, args[0])
	if s.flagFuncGen2 {
		svc, err := s.inv.clients().CloudFunctionsV2(ctx)
		if err != nil {
			return err
		}
		policy, err := svc.Projects.Locations.Functions.GetIamPolicy(resource).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("getting IAM policy: %w", err)
		}
		addBindingV2(policy, s.flagFuncRole, s.flagFuncMember)
		got, err := svc.Projects.Locations.Functions.SetIamPolicy(resource, &cloudfunctions.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("updating IAM policy: %w", err)
		}
		return emitFormattedTo(s.inv.stdout, got, s.flagFuncFormat)
	}
	svc, err := s.inv.clients().CloudFunctionsV1(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Functions.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	addBindingV1(policy, s.flagFuncRole, s.flagFuncMember)
	got, err := svc.Projects.Locations.Functions.SetIamPolicy(resource, &cloudfunctionsv1.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagFuncFormat)
}

func (s *functionsCmds) runFuncRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	resource := functionResourceName(project, s.flagFuncRegion, args[0])
	if s.flagFuncGen2 {
		svc, err := s.inv.clients().CloudFunctionsV2(ctx)
		if err != nil {
			return err
		}
		policy, err := svc.Projects.Locations.Functions.GetIamPolicy(resource).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("getting IAM policy: %w", err)
		}
		removeBindingV2(policy, s.flagFuncRole, s.flagFuncMember)
		got, err := svc.Projects.Locations.Functions.SetIamPolicy(resource, &cloudfunctions.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("updating IAM policy: %w", err)
		}
		return emitFormattedTo(s.inv.stdout, got, s.flagFuncFormat)
	}
	svc, err := s.inv.clients().CloudFunctionsV1(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Functions.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	removeBindingV1(policy, s.flagFuncRole, s.flagFuncMember)
	got, err := svc.Projects.Locations.Functions.SetIamPolicy(resource, &cloudfunctionsv1.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagFuncFormat)
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
	Provider       string `json:"provider"`
	EventType      string `json:"event_type"`
	Default        string `json:"event_type_default,omitempty"`
	ResourceType   string `json:"resource_type,omitempty"`
	ResourceOption string `json:"resource_optional,omitempty"`
}

func gen1TriggerEvents() []funcTriggerEventRow {
	return []funcTriggerEventRow{
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
}

type funcGen2EventRow struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// gen2TriggerEvents mirrors the Eventarc-backed events surfaced by
// `gcloud functions event-types list --gen2`.
func gen2TriggerEvents() []funcGen2EventRow {
	return []funcGen2EventRow{
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
}

func (s *functionsCmds) runFuncEventTypesList(cmd *cobra.Command, args []string) error {
	if s.flagFuncGen2 {
		if s.flagFuncFormat != "" {
			return emitFormattedTo(s.inv.stdout, gen2TriggerEvents(), s.flagFuncFormat)
		}
		printTableHeadingTo(s.inv.stdout, len(gen2TriggerEvents()), "%-60s %s\n", "NAME", "DESCRIPTION")
		for _, e := range gen2TriggerEvents() {
			fmt.Fprintf(s.inv.stdout, "%-60s %s\n", e.Name, e.Description)
		}
		return nil
	}
	if s.flagFuncFormat != "" {
		return emitFormattedTo(s.inv.stdout, gen1TriggerEvents(), s.flagFuncFormat)
	}
	fmt.Fprintf(s.inv.stdout, "%-32s %-60s %-18s %-22s %s\n",
		"EVENT_PROVIDER", "EVENT_TYPE", "EVENT_TYPE_DEFAULT", "RESOURCE_TYPE", "RESOURCE_OPTIONAL")
	for _, e := range gen1TriggerEvents() {
		fmt.Fprintf(s.inv.stdout, "%-32s %-60s %-18s %-22s %s\n",
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

func (s *functionsCmds) funcBuildLogFilter(project, region, name string) (string, error) {
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
	if s.flagFuncGen2 {
		parts = append(parts, strings.Join(gen2, " "))
	} else {
		// Match python behaviour: without --gen2 we return both by default.
		parts = append(parts, fmt.Sprintf("(%s) OR (%s)", strings.Join(gen1, " "), strings.Join(gen2, " ")))
	}
	if s.flagFuncExecutionID != "" {
		parts = append(parts, fmt.Sprintf(`labels.execution_id="%s"`, s.flagFuncExecutionID))
	}
	if s.flagFuncMinLogLevel != "" {
		parts = append(parts, fmt.Sprintf(`severity>=%s`, strings.ToUpper(s.flagFuncMinLogLevel)))
	}
	end, err := funcParseTime(s.flagFuncEndTime)
	if err != nil {
		return "", fmt.Errorf("invalid --end-time: %w", err)
	}
	if !end.IsZero() {
		parts = append(parts, fmt.Sprintf(`timestamp<="%s"`, end.UTC().Format(time.RFC3339)))
	}
	start, err := funcParseTime(s.flagFuncStartTime)
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

func (s *functionsCmds) runFuncLogsRead(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	if s.flagFuncLimit < 1 || s.flagFuncLimit > 1000 {
		return fmt.Errorf("--limit must be between 1 and 1000")
	}
	name := ""
	if len(args) == 1 {
		name = args[0]
	}
	filter, err := s.funcBuildLogFilter(project, s.flagFuncRegion, name)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Logging(ctx)
	if err != nil {
		return err
	}
	req := &logging.ListLogEntriesRequest{
		Filter:        filter,
		OrderBy:       "timestamp desc",
		PageSize:      s.flagFuncLimit,
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
	if s.flagFuncFormat != "" {
		return emitFormattedTo(s.inv.stdout, rows, s.flagFuncFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(rows), "%-6s %-32s %-32s %-24s %s\n", "LEVEL", "NAME", "EXECUTION_ID", "TIME_UTC", "LOG")
	for _, r := range rows {
		fmt.Fprintf(s.inv.stdout, "%-6s %-32s %-32s %-24s %s\n", r.Level, r.Name, r.ExecutionID, r.TimeUTC, r.Log)
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

func (s *functionsCmds) runFuncRegionsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudFunctionsV2(ctx)
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
	if s.flagFuncFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagFuncFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%s\n", "NAME")
	for _, l := range all {
		id := l.LocationId
		if id == "" {
			id = path.Base(l.Name)
		}
		fmt.Fprintf(s.inv.stdout, "%s\n", id)
	}
	return nil
}

// --- runtimes (#864) ---

func (s *functionsCmds) runFuncRuntimesList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	region := s.flagFuncRegion
	if region == "" {
		region = "-"
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudFunctionsV2(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Runtimes.List(functionLocationParent(project, region)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing runtimes: %w", err)
	}
	if s.flagFuncFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Runtimes, s.flagFuncFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Runtimes), "%-24s %-16s %s\n", "NAME", "STAGE", "ENVIRONMENT")
	for _, r := range resp.Runtimes {
		fmt.Fprintf(s.inv.stdout, "%-24s %-16s %s\n", r.Name, r.Stage, r.Environment)
	}
	return nil
}

// functionsCmds holds the commands and flag values of one invocation's "functions"
// command group, so invocations don't share state (#2070).
type functionsCmds struct {
	inv *invocation

	functionsCmd          *cobra.Command
	flagFuncRegion        string
	flagFuncFormat        string
	flagFuncMember        string
	flagFuncRole          string
	flagFuncLimit         int64
	flagFuncMinLogLevel   string
	flagFuncExecutionID   string
	flagFuncStartTime     string
	flagFuncEndTime       string
	flagFuncGen2          bool
	funcGetIamCmd         *cobra.Command
	funcSetIamCmd         *cobra.Command
	funcAddIamCmd         *cobra.Command
	funcRemoveIamCmd      *cobra.Command
	funcEventTypesCmd     *cobra.Command
	funcEventTypesListCmd *cobra.Command
	funcLogsCmd           *cobra.Command
	funcLogsReadCmd       *cobra.Command
	funcRegionsCmd        *cobra.Command
	funcRegionsListCmd    *cobra.Command
	funcRuntimesCmd       *cobra.Command
	funcRuntimesListCmd   *cobra.Command
	flagFuncListRegions   []string
	flagFuncListV2        bool
	flagFuncListLimit     int64
}

// newFunctionsCmd returns the "functions" command group for inv.
func newFunctionsCmd(inv *invocation) *cobra.Command {
	return newFunctionsCmds(inv).functionsCmd
}

// newFunctionsCmds builds the "functions" command group for inv and returns its state.
func newFunctionsCmds(inv *invocation) *functionsCmds {
	s := &functionsCmds{inv: inv}
	s.functionsCmd = &cobra.Command{Use: "functions", Short: "Manage Cloud Functions"}
	s.funcGetIamCmd = &cobra.Command{
		Use: "get-iam-policy FUNCTION", Short: "Get the IAM policy for a Cloud Function",
		Args: cobra.ExactArgs(1), RunE: s.runFuncGetIam,
	}
	s.funcSetIamCmd = &cobra.Command{
		Use: "set-iam-policy FUNCTION POLICY_FILE", Short: "Set the IAM policy for a Cloud Function",
		Args: cobra.ExactArgs(2), RunE: s.runFuncSetIam,
	}
	s.funcAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding FUNCTION", Short: "Add an IAM policy binding to a Cloud Function",
		Args: cobra.ExactArgs(1), RunE: s.runFuncAddIam,
	}
	s.funcRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding FUNCTION", Short: "Remove an IAM policy binding from a Cloud Function",
		Args: cobra.ExactArgs(1), RunE: s.runFuncRemoveIam,
	}
	s.funcEventTypesCmd = &cobra.Command{Use: "event-types", Short: "List trigger event types for Cloud Functions"}
	s.funcEventTypesListCmd = &cobra.Command{
		Use: "list", Short: "List event types available to Cloud Functions",
		Args: cobra.NoArgs, RunE: s.runFuncEventTypesList,
	}
	s.funcLogsCmd = &cobra.Command{Use: "logs", Short: "Manage Cloud Functions logs"}
	s.funcLogsReadCmd = &cobra.Command{
		Use: "read [NAME]", Short: "Read log entries produced by Cloud Functions",
		Args: cobra.MaximumNArgs(1), RunE: s.runFuncLogsRead,
	}
	s.funcRegionsCmd = &cobra.Command{Use: "regions", Short: "List regions available to Cloud Functions"}
	s.funcRegionsListCmd = &cobra.Command{
		Use: "list", Short: "List regions available to Cloud Functions",
		Args: cobra.NoArgs, RunE: s.runFuncRegionsList,
	}
	s.funcRuntimesCmd = &cobra.Command{Use: "runtimes", Short: "List runtimes available to Cloud Functions"}
	s.funcRuntimesListCmd = &cobra.Command{
		Use: "list", Short: "List runtimes available to Cloud Functions",
		Args: cobra.NoArgs, RunE: s.runFuncRuntimesList,
	}
	s.initFunctions()
	s.initFunctionsList()
	return s
}
