package cmd

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	auditmanager "google.golang.org/api/auditmanager/v1"
)

// --- gcloud audit-manager audit-schedules (#1911, gcloud-python 586.0.0) ---
// --- gcloud audit-manager operations describe (#1911) ---

// errAuditManagerStopPaging ends Pages iteration once --limit is reached.
var errAuditManagerStopPaging = errors.New("stop paging")

func auditManagerFrequencies() map[string]string {
	return map[string]string{
		"daily": "DAILY", "weekly": "WEEKLY", "monthly": "MONTHLY", "quarterly": "QUARTERLY", "annually": "ANNUALLY",
	}
}

func auditManagerScheduleStates() map[string]string {
	return map[string]string{
		"active": "SCHEDULE_STATE_ACTIVE", "paused": "SCHEDULE_STATE_PAUSED", "deleted": "SCHEDULE_STATE_DELETED",
	}
}

func (g *auditManagerCmds) initAuditManagerAuditSchedules() {
	scope := func(c *cobra.Command, help string) {
		c.Flags().StringVar(&g.flagAMFolder, "folder", "", "Folder Id "+help)
		c.Flags().StringVar(&g.flagAMOrganization, "organization", "", "Organization Id "+help)
		c.Flags().StringVar(&g.flagAMFormat, "format", "", "Output format")
	}
	scope(g.auditManagerSchedulesCreateCmd, "for which to create audit schedule")
	scope(g.auditManagerSchedulesListCmd, "for which to list audit schedules")
	g.auditManagerSchedulesCreateCmd.Flags().StringVar(&g.flagAMLocation, "location", "",
		"The location where the audit schedule should be created.")
	g.auditManagerSchedulesListCmd.Flags().StringVar(&g.flagAMLocation, "location", "",
		"The location where audit schedules should be listed.")
	for _, c := range []*cobra.Command{g.auditManagerSchedulesCreateCmd, g.auditManagerSchedulesListCmd} {
		_ = c.MarkFlagRequired("location")
	}

	f := g.auditManagerSchedulesCreateCmd.Flags()
	f.StringVar(&g.flagAMScheduleID, "audit-schedule-id", "", "ID to use for the audit schedule.")
	f.StringVar(&g.flagAMFramework, "compliance-framework", "",
		"Compliance Framework against which the Report must be generated. Eg: FEDRAMP_MODERATE")
	f.StringVar(&g.flagAMReportFormat, "report-format", "", "The format in which the audit report should be created: odf.")
	f.StringVar(&g.flagAMGcsURI, "gcs-uri", "",
		"Destination Cloud storage bucket where report and evidence must be uploaded.")
	f.StringVar(&g.flagAMStartTime, "start-time", "",
		"Date and time when the first audit run is triggered in RFC 3339 format.")
	f.StringVar(&g.flagAMEndTime, "end-time", "", "Date and time that the schedule stops in RFC 3339 format.")
	f.StringVar(&g.flagAMFrequency, "frequency", "",
		"Frequency of audit runs: daily, weekly, monthly, quarterly, annually.")
	f.StringVar(&g.flagAMTimeZone, "time-zone", "",
		"Time zone for the audit schedule in IANA format (e.g. America/New_York).")
	f.StringVar(&g.flagAMDisplayName, "display-name", "", "Display name for the audit schedule.")
	for _, name := range []string{"audit-schedule-id", "compliance-framework", "report-format", "gcs-uri", "start-time", "frequency"} {
		_ = g.auditManagerSchedulesCreateCmd.MarkFlagRequired(name)
	}

	g.auditManagerSchedulesListCmd.Flags().Int64Var(&g.flagAMLimit, "limit", 0, "Maximum number of resources to list")
	g.auditManagerSchedulesListCmd.Flags().Int64Var(&g.flagAMPageSize, "page-size", 0, "Number of resources to fetch per page")

	g.auditManagerSchedulesUpdateCmd.Flags().StringVar(&g.flagAMState, "state", "",
		"State to update the audit schedule to: active, paused, deleted.")
	_ = g.auditManagerSchedulesUpdateCmd.MarkFlagRequired("state")
	g.auditManagerSchedulesUpdateCmd.Flags().StringVar(&g.flagAMFormat, "format", "", "Output format")

	o := g.auditManagerOperationsDescribeCmd.Flags()
	o.StringVar(&g.flagAMFolder, "folder", "", "The folder ID of the operation")
	o.StringVar(&g.flagAMOrganization, "organization", "", "The organization ID of the operation")
	o.StringVar(&g.flagAMLocation, "location", "", "The location of the operation")
	o.StringVar(&g.flagAMFormat, "format", "", "Output format")

	g.auditManagerSchedulesCmd.AddCommand(g.auditManagerSchedulesCreateCmd, g.auditManagerSchedulesListCmd, g.auditManagerSchedulesUpdateCmd)
	g.auditManagerCmd.AddCommand(g.auditManagerSchedulesCmd)
	g.auditManagerOperationsCmd.AddCommand(g.auditManagerOperationsDescribeCmd)
}

// auditManagerScope resolves the parent scope from exactly one of --project,
// --folder or --organization, plus --location.
func (g *auditManagerCmds) auditManagerScope(cmd *cobra.Command) (string, error) {
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
		scope = "folders/" + g.flagAMFolder
	case "organization":
		scope = "organizations/" + g.flagAMOrganization
	default:
		scope = "projects/" + g.inv.flags.project
	}
	return scope + "/locations/" + g.flagAMLocation, nil
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
func (g *auditManagerCmds) auditManagerBuildSchedule() (*auditmanager.AuditSchedule, error) {
	format, err := auditManagerChoice("report-format", g.flagAMReportFormat, map[string]string{"odf": "AUDIT_REPORT_FORMAT_ODF"})
	if err != nil {
		return nil, err
	}
	freq, err := auditManagerChoice("frequency", g.flagAMFrequency, auditManagerFrequencies())
	if err != nil {
		return nil, err
	}
	return &auditmanager.AuditSchedule{
		GcsUri:              g.flagAMGcsURI,
		ComplianceFramework: g.flagAMFramework,
		ReportFormat:        format,
		DisplayName:         g.flagAMDisplayName,
		ScheduleConfig: &auditmanager.ScheduleConfig{
			StartTime: g.flagAMStartTime,
			EndTime:   g.flagAMEndTime,
			Frequency: freq,
			TimeZone:  g.flagAMTimeZone,
		},
	}, nil
}

func (g *auditManagerCmds) runAuditManagerSchedulesCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.auditManagerScope(cmd)
	if err != nil {
		return err
	}
	body, err := g.auditManagerBuildSchedule()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AuditManager(ctx)
	if err != nil {
		return err
	}
	// The create/list/patch paths are v1/{+parent}/auditSchedules and
	// v1/{+name} for every parent type, so the projects service serves
	// folder and organization scopes too.
	got, err := svc.Projects.Locations.AuditSchedules.Create(parent, body).
		AuditScheduleId(g.flagAMScheduleID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating audit schedule: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAMFormat)
}

func (g *auditManagerCmds) runAuditManagerSchedulesList(cmd *cobra.Command, args []string) error {
	parent, err := g.auditManagerScope(cmd)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AuditManager(ctx)
	if err != nil {
		return err
	}
	var all []*auditmanager.AuditSchedule
	call := svc.Projects.Locations.AuditSchedules.List(parent).Context(ctx)
	if g.flagAMPageSize > 0 {
		call = call.PageSize(g.flagAMPageSize)
	}
	err = call.Pages(ctx, func(resp *auditmanager.ListAuditSchedulesResponse) error {
		all = append(all, resp.AuditSchedules...)
		if g.flagAMLimit > 0 && int64(len(all)) >= g.flagAMLimit {
			return errAuditManagerStopPaging
		}
		return nil
	})
	if err != nil && err != errAuditManagerStopPaging {
		return fmt.Errorf("listing audit schedules: %w", err)
	}
	if g.flagAMLimit > 0 && int64(len(all)) > g.flagAMLimit {
		all = all[:g.flagAMLimit]
	}
	return emitAuditSchedules(g.inv.stdout, g.inv.stderr, all, g.flagAMFormat)
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

func (g *auditManagerCmds) runAuditManagerSchedulesUpdate(cmd *cobra.Command, args []string) error {
	state, err := auditManagerChoice("state", g.flagAMState, auditManagerScheduleStates())
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
	ctx := cmd.Context()
	svc, err := g.inv.clients().AuditManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.AuditSchedules.Patch(name, &auditmanager.AuditSchedule{Name: name, State: state}).
		UpdateMask("state").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating audit schedule: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAMFormat)
}

// auditManagerOperationName resolves OPERATION to
// {projects|folders|organizations}/X/locations/L/operationDetails/OP.
func (g *auditManagerCmds) auditManagerOperationName(cmd *cobra.Command, op string) (string, error) {
	for _, p := range []string{"projects/", "folders/", "organizations/"} {
		if strings.HasPrefix(op, p) {
			return op, nil
		}
	}
	if g.flagAMLocation == "" {
		return "", fmt.Errorf("Error parsing [operation]. The [operation] resource is not properly specified. Failed to find attribute [location].")
	}
	if cmd.Flags().Changed("folder") && cmd.Flags().Changed("organization") {
		return "", fmt.Errorf("At most one of --folder | --organization can be specified.")
	}
	var scope string
	switch {
	case g.flagAMFolder != "":
		scope = "folders/" + g.flagAMFolder
	case g.flagAMOrganization != "":
		scope = "organizations/" + g.flagAMOrganization
	default:
		project, err := g.inv.project()
		if err != nil {
			return "", err
		}
		scope = "projects/" + project
	}
	return fmt.Sprintf("%s/locations/%s/operationDetails/%s", scope, g.flagAMLocation, op), nil
}

func (g *auditManagerCmds) runAuditManagerOperationsDescribe(cmd *cobra.Command, args []string) error {
	name, err := g.auditManagerOperationName(cmd, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().AuditManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.OperationDetails.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagAMFormat)
}
