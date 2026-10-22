package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	auditmanager "google.golang.org/api/auditmanager/v1"
)

// --- gcloud audit-manager audit-schedules (#1911, gcloud-python 586.0.0) ---
// --- gcloud audit-manager operations describe (#1911) ---

var auditManagerSchedulesCmd = &cobra.Command{
	Use:   "audit-schedules",
	Short: "Command group for Audit Manager Audit Schedules",
}

var (
	flagAMFolder       string
	flagAMOrganization string
	flagAMLocation     string
	flagAMScheduleID   string
	flagAMFramework    string
	flagAMReportFormat string
	flagAMGcsURI       string
	flagAMStartTime    string
	flagAMEndTime      string
	flagAMFrequency    string
	flagAMTimeZone     string
	flagAMDisplayName  string
	flagAMState        string
	flagAMFormat       string
	flagAMLimit        int64
	flagAMPageSize     int64
)

var (
	auditManagerSchedulesCreateCmd = &cobra.Command{
		Use: "create", Short: "Create Audit Schedule",
		Args: cobra.NoArgs, RunE: runAuditManagerSchedulesCreate,
	}
	auditManagerSchedulesListCmd = &cobra.Command{
		Use: "list", Short: "List Audit Schedules",
		Args: cobra.NoArgs, RunE: runAuditManagerSchedulesList,
	}
	auditManagerSchedulesUpdateCmd = &cobra.Command{
		Use: "update AUDIT_SCHEDULE", Short: "Update Audit Schedule",
		Args: cobra.ExactArgs(1), RunE: runAuditManagerSchedulesUpdate,
	}
	auditManagerOperationsDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe Audit operation",
		Args: cobra.ExactArgs(1), RunE: runAuditManagerOperationsDescribe,
	}
)

// errAuditManagerStopPaging ends Pages iteration once --limit is reached.
var errAuditManagerStopPaging = errors.New("stop paging")

var auditManagerFrequencies = map[string]string{
	"daily": "DAILY", "weekly": "WEEKLY", "monthly": "MONTHLY", "quarterly": "QUARTERLY", "annually": "ANNUALLY",
}

var auditManagerScheduleStates = map[string]string{
	"active": "SCHEDULE_STATE_ACTIVE", "paused": "SCHEDULE_STATE_PAUSED", "deleted": "SCHEDULE_STATE_DELETED",
}

func init() {
	scope := func(c *cobra.Command, help string) {
		c.Flags().StringVar(&flagAMFolder, "folder", "", "Folder Id "+help)
		c.Flags().StringVar(&flagAMOrganization, "organization", "", "Organization Id "+help)
		c.Flags().StringVar(&flagAMFormat, "format", "", "Output format")
	}
	scope(auditManagerSchedulesCreateCmd, "for which to create audit schedule")
	scope(auditManagerSchedulesListCmd, "for which to list audit schedules")
	auditManagerSchedulesCreateCmd.Flags().StringVar(&flagAMLocation, "location", "",
		"The location where the audit schedule should be created.")
	auditManagerSchedulesListCmd.Flags().StringVar(&flagAMLocation, "location", "",
		"The location where audit schedules should be listed.")
	for _, c := range []*cobra.Command{auditManagerSchedulesCreateCmd, auditManagerSchedulesListCmd} {
		_ = c.MarkFlagRequired("location")
	}

	f := auditManagerSchedulesCreateCmd.Flags()
	f.StringVar(&flagAMScheduleID, "audit-schedule-id", "", "ID to use for the audit schedule.")
	f.StringVar(&flagAMFramework, "compliance-framework", "",
		"Compliance Framework against which the Report must be generated. Eg: FEDRAMP_MODERATE")
	f.StringVar(&flagAMReportFormat, "report-format", "", "The format in which the audit report should be created: odf.")
	f.StringVar(&flagAMGcsURI, "gcs-uri", "",
		"Destination Cloud storage bucket where report and evidence must be uploaded.")
	f.StringVar(&flagAMStartTime, "start-time", "",
		"Date and time when the first audit run is triggered in RFC 3339 format.")
	f.StringVar(&flagAMEndTime, "end-time", "", "Date and time that the schedule stops in RFC 3339 format.")
	f.StringVar(&flagAMFrequency, "frequency", "",
		"Frequency of audit runs: daily, weekly, monthly, quarterly, annually.")
	f.StringVar(&flagAMTimeZone, "time-zone", "",
		"Time zone for the audit schedule in IANA format (e.g. America/New_York).")
	f.StringVar(&flagAMDisplayName, "display-name", "", "Display name for the audit schedule.")
	for _, name := range []string{"audit-schedule-id", "compliance-framework", "report-format", "gcs-uri", "start-time", "frequency"} {
		_ = auditManagerSchedulesCreateCmd.MarkFlagRequired(name)
	}

	auditManagerSchedulesListCmd.Flags().Int64Var(&flagAMLimit, "limit", 0, "Maximum number of resources to list")
	auditManagerSchedulesListCmd.Flags().Int64Var(&flagAMPageSize, "page-size", 0, "Number of resources to fetch per page")

	auditManagerSchedulesUpdateCmd.Flags().StringVar(&flagAMState, "state", "",
		"State to update the audit schedule to: active, paused, deleted.")
	_ = auditManagerSchedulesUpdateCmd.MarkFlagRequired("state")
	auditManagerSchedulesUpdateCmd.Flags().StringVar(&flagAMFormat, "format", "", "Output format")

	o := auditManagerOperationsDescribeCmd.Flags()
	o.StringVar(&flagAMFolder, "folder", "", "The folder ID of the operation")
	o.StringVar(&flagAMOrganization, "organization", "", "The organization ID of the operation")
	o.StringVar(&flagAMLocation, "location", "", "The location of the operation")
	o.StringVar(&flagAMFormat, "format", "", "Output format")

	auditManagerSchedulesCmd.AddCommand(auditManagerSchedulesCreateCmd, auditManagerSchedulesListCmd, auditManagerSchedulesUpdateCmd)
	auditManagerCmd.AddCommand(auditManagerSchedulesCmd)
	auditManagerOperationsCmd.AddCommand(auditManagerOperationsDescribeCmd)
}

// auditManagerScope resolves the parent scope from exactly one of --project,
// --folder or --organization, plus --location.
func auditManagerScope(cmd *cobra.Command) (string, error) {
	var set []string
	for _, f := range []string{"folder", "organization", "project"} {
		if cmd.Flags().Changed(f) {
			set = append(set, f)
		}
	}
	if len(set) != 1 {
		return "", fmt.Errorf("Exactly one of (--folder | --organization | --project) must be specified.")
	}
	var scope string
	switch set[0] {
	case "folder":
		scope = "folders/" + flagAMFolder
	case "organization":
		scope = "organizations/" + flagAMOrganization
	default:
		scope = "projects/" + flagProject
	}
	return scope + "/locations/" + flagAMLocation, nil
}

// auditManagerChoice lower-cases v and checks it against choices.
func auditManagerChoice(flag, v string, choices map[string]string) (string, error) {
	if got, ok := choices[strings.ToLower(v)]; ok {
		return got, nil
	}
	keys := make([]string, 0, len(choices))
	for k := range choices {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return "", fmt.Errorf("argument --%s: Invalid choice: '%s'.\nValid choices are [%s].", flag, v, strings.Join(keys, ", "))
}

// auditManagerBuildSchedule builds the AuditSchedule body for create.
func auditManagerBuildSchedule() (*auditmanager.AuditSchedule, error) {
	format, err := auditManagerChoice("report-format", flagAMReportFormat, map[string]string{"odf": "AUDIT_REPORT_FORMAT_ODF"})
	if err != nil {
		return nil, err
	}
	freq, err := auditManagerChoice("frequency", flagAMFrequency, auditManagerFrequencies)
	if err != nil {
		return nil, err
	}
	return &auditmanager.AuditSchedule{
		GcsUri:              flagAMGcsURI,
		ComplianceFramework: flagAMFramework,
		ReportFormat:        format,
		DisplayName:         flagAMDisplayName,
		ScheduleConfig: &auditmanager.ScheduleConfig{
			StartTime: flagAMStartTime,
			EndTime:   flagAMEndTime,
			Frequency: freq,
			TimeZone:  flagAMTimeZone,
		},
	}, nil
}

func runAuditManagerSchedulesCreate(cmd *cobra.Command, args []string) error {
	parent, err := auditManagerScope(cmd)
	if err != nil {
		return err
	}
	body, err := auditManagerBuildSchedule()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AuditManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	// The create/list/patch paths are v1/{+parent}/auditSchedules and
	// v1/{+name} for every parent type, so the projects service serves
	// folder and organization scopes too.
	got, err := svc.Projects.Locations.AuditSchedules.Create(parent, body).
		AuditScheduleId(flagAMScheduleID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating audit schedule: %w", err)
	}
	return emitFormatted(got, flagAMFormat)
}

func runAuditManagerSchedulesList(cmd *cobra.Command, args []string) error {
	parent, err := auditManagerScope(cmd)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AuditManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*auditmanager.AuditSchedule
	call := svc.Projects.Locations.AuditSchedules.List(parent).Context(ctx)
	if flagAMPageSize > 0 {
		call = call.PageSize(flagAMPageSize)
	}
	err = call.Pages(ctx, func(resp *auditmanager.ListAuditSchedulesResponse) error {
		all = append(all, resp.AuditSchedules...)
		if flagAMLimit > 0 && int64(len(all)) >= flagAMLimit {
			return errAuditManagerStopPaging
		}
		return nil
	})
	if err != nil && err != errAuditManagerStopPaging {
		return fmt.Errorf("listing audit schedules: %w", err)
	}
	if flagAMLimit > 0 && int64(len(all)) > flagAMLimit {
		all = all[:flagAMLimit]
	}
	return emitAuditSchedules(os.Stdout, os.Stderr, all, flagAMFormat)
}

// emitAuditSchedules prints schedules with gcloud-python's list columns
// unless --format is given.
func emitAuditSchedules(w, errW io.Writer, all []*auditmanager.AuditSchedule, format string) error {
	if format != "" {
		if all == nil {
			all = []*auditmanager.AuditSchedule{}
		}
		return emitFormattedTo(w, all, format)
	}
	if len(all) == 0 {
		_, err := fmt.Fprintln(errW, "Listed 0 items.")
		return err
	}
	rows, err := jsonRows(all)
	if err != nil {
		return err
	}
	for _, r := range rows {
		name, _ := r["name"].(string)
		r["id"] = name[strings.LastIndex(name, "/")+1:]
	}
	return emitFormattedTo(w, rows, "table(id,complianceFramework,displayName,state,gcsUri,reportFormat,"+
		"scheduleConfig,createTime,updateTime,nextRunTime,lastTriggerTime,errorMessage)")
}

func runAuditManagerSchedulesUpdate(cmd *cobra.Command, args []string) error {
	state, err := auditManagerChoice("state", flagAMState, auditManagerScheduleStates)
	if err != nil {
		return err
	}
	name := args[0]
	if !strings.HasPrefix(name, "projects/") && !strings.HasPrefix(name, "folders/") && !strings.HasPrefix(name, "organizations/") {
		return fmt.Errorf("Invalid value for [name]: Audit schedule name must be of the form " +
			"projects/{project}/locations/{location}/auditSchedules/{audit_schedule}, " +
			"folders/{folder}/locations/{location}/auditSchedules/{audit_schedule}, or " +
			"organizations/{organization}/locations/{location}/auditSchedules/{audit_schedule}.")
	}
	ctx := context.Background()
	svc, err := gcp.AuditManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AuditSchedules.Patch(name, &auditmanager.AuditSchedule{Name: name, State: state}).
		UpdateMask("state").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating audit schedule: %w", err)
	}
	return emitFormatted(got, flagAMFormat)
}

// auditManagerOperationName resolves OPERATION to
// {projects|folders|organizations}/X/locations/L/operationDetails/OP.
func auditManagerOperationName(cmd *cobra.Command, op string) (string, error) {
	for _, p := range []string{"projects/", "folders/", "organizations/"} {
		if strings.HasPrefix(op, p) {
			return op, nil
		}
	}
	if flagAMLocation == "" {
		return "", fmt.Errorf("Error parsing [operation]. The [operation] resource is not properly specified. Failed to find attribute [location].")
	}
	if cmd.Flags().Changed("folder") && cmd.Flags().Changed("organization") {
		return "", fmt.Errorf("At most one of --folder | --organization can be specified.")
	}
	var scope string
	switch {
	case flagAMFolder != "":
		scope = "folders/" + flagAMFolder
	case flagAMOrganization != "":
		scope = "organizations/" + flagAMOrganization
	default:
		project, err := resolveProject()
		if err != nil {
			return "", err
		}
		scope = "projects/" + project
	}
	return fmt.Sprintf("%s/locations/%s/operationDetails/%s", scope, flagAMLocation, op), nil
}

func runAuditManagerOperationsDescribe(cmd *cobra.Command, args []string) error {
	name, err := auditManagerOperationName(cmd, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AuditManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.OperationDetails.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagAMFormat)
}
