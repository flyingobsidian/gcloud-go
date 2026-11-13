package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	monitoring "google.golang.org/api/monitoring/v3"
)

// --- policies list ---

// --- snoozes create ---

// --- snoozes describe ---

// --- snoozes list ---

// --- snoozes update ---

// --- snoozes cancel ---

// --- policies describe ---

// --- policies create ---

// --- policies update ---

// --- policies delete ---

// --- channels ---

func (g *monitoringCmds) initMonitoring() {
	g.monitoringPoliciesListCmd.Flags().StringVar(&g.flagMonPoliciesFormat, "format", "", "Output format: yaml (default), json")
	g.monitoringPoliciesListCmd.Flags().StringVar(&g.flagMonPoliciesFilter, "filter", "", "Filter expression")
	g.monitoringPoliciesListCmd.Flags().BoolVar(&g.flagMonPoliciesURI, "uri", false, "Print resource names")

	g.monitoringPoliciesCreateCmd.Flags().StringVar(&g.flagPolCreateFromFile, "policy-from-file", "", "JSON file with policy definition")
	g.monitoringPoliciesCreateCmd.Flags().StringVar(&g.flagPolCreateDisplayName, "display-name", "", "Display name")
	g.monitoringPoliciesCreateCmd.Flags().BoolVar(&g.flagPolCreateEnabled, "enabled", true, "Enable the policy")
	g.monitoringPoliciesCreateCmd.Flags().BoolVar(&g.flagPolCreateNoEnabled, "no-enabled", false, "Disable the policy")
	g.monitoringPoliciesCreateCmd.Flags().StringSliceVar(&g.flagPolCreateChannels, "notification-channels", nil, "Notification channel names")
	g.monitoringPoliciesCreateCmd.Flags().StringVar(&g.flagPolCreateDocumentation, "documentation", "", "Documentation content")

	g.monitoringPoliciesUpdateCmd.Flags().StringVar(&g.flagPolUpdateFromFile, "policy-from-file", "", "JSON file with policy definition")
	g.monitoringPoliciesUpdateCmd.Flags().BoolVar(&g.flagPolUpdateEnabled, "enabled", false, "Enable the policy")
	g.monitoringPoliciesUpdateCmd.Flags().BoolVar(&g.flagPolUpdateNoEnabled, "no-enabled", false, "Disable the policy")
	g.monitoringPoliciesUpdateCmd.Flags().StringVar(&g.flagPolUpdateFields, "fields", "", "Comma-separated fields to update")
	g.monitoringPoliciesUpdateCmd.Flags().StringSliceVar(&g.flagPolAddChannels, "add-notification-channels", nil, "Channels to add")
	g.monitoringPoliciesUpdateCmd.Flags().StringSliceVar(&g.flagPolRemoveChannels, "remove-notification-channels", nil, "Channels to remove")

	g.monitoringPoliciesCmd.AddCommand(g.monitoringPoliciesListCmd)
	g.monitoringPoliciesCmd.AddCommand(g.monitoringPoliciesDescribeCmd)
	g.monitoringPoliciesCmd.AddCommand(g.monitoringPoliciesCreateCmd)
	g.monitoringPoliciesCmd.AddCommand(g.monitoringPoliciesUpdateCmd)
	g.monitoringPoliciesCmd.AddCommand(g.monitoringPoliciesDeleteCmd)

	g.monitoringSnoozesCreateCmd.Flags().StringVar(&g.flagSnoozeDisplayName, "display-name", "", "Display name for the snooze")
	g.monitoringSnoozesCreateCmd.MarkFlagRequired("display-name")
	g.monitoringSnoozesCreateCmd.Flags().StringVar(&g.flagSnoozeStartTime, "start-time", "", "Start time (RFC3339)")
	g.monitoringSnoozesCreateCmd.MarkFlagRequired("start-time")
	g.monitoringSnoozesCreateCmd.Flags().StringVar(&g.flagSnoozeEndTime, "end-time", "", "End time (RFC3339)")
	g.monitoringSnoozesCreateCmd.MarkFlagRequired("end-time")
	g.monitoringSnoozesCreateCmd.Flags().StringSliceVar(&g.flagSnoozePolicies, "criteria-policies", nil, "Alert policy resource names to snooze")
	g.monitoringSnoozesCreateCmd.Flags().StringVar(&g.flagSnoozeFilter, "criteria-filter", "", "Filter for snooze criteria")
	g.monitoringSnoozesCreateCmd.Flags().StringVar(&g.flagSnoozeFromFile, "snooze-from-file", "", "JSON file containing snooze definition")
	g.monitoringSnoozesCreateCmd.Flags().StringVar(&g.flagSnoozeCreateFormat, "format", "", "Output format (e.g. json, yaml, 'value(name)')")
	g.monitoringSnoozesCmd.AddCommand(g.monitoringSnoozesCreateCmd)

	g.monitoringSnoozesListCmd.Flags().StringVar(&g.flagSnoozesListFormat, "format", "", "Output format: yaml (default), json, 'csv(NAME,DISPLAY_NAME)'")
	g.monitoringSnoozesListCmd.Flags().StringVar(&g.flagSnoozesListFilter, "filter", "", "Filter expression")
	g.monitoringSnoozesListCmd.Flags().BoolVar(&g.flagSnoozesListURI, "uri", false, "Print resource names")
	g.monitoringSnoozesCmd.AddCommand(g.monitoringSnoozesListCmd)

	g.monitoringSnoozesDescribeCmd.Flags().StringVar(&g.flagSnoozeDescribeFormat, "format", "", "Output format: yaml (default), json, 'table(name,display_name)'")
	g.monitoringSnoozesCmd.AddCommand(g.monitoringSnoozesDescribeCmd)

	g.monitoringSnoozesUpdateCmd.Flags().StringVar(&g.flagSnoozeUpdateDisplayName, "display-name", "", "New display name")
	g.monitoringSnoozesUpdateCmd.Flags().StringVar(&g.flagSnoozeUpdateStartTime, "start-time", "", "New start time (RFC3339)")
	g.monitoringSnoozesUpdateCmd.Flags().StringVar(&g.flagSnoozeUpdateEndTime, "end-time", "", "New end time (RFC3339)")
	g.monitoringSnoozesUpdateCmd.Flags().StringVar(&g.flagSnoozeUpdateFromFile, "snooze-from-file", "", "JSON file containing snooze definition")
	g.monitoringSnoozesCmd.AddCommand(g.monitoringSnoozesUpdateCmd)

	g.monitoringSnoozesCmd.AddCommand(g.monitoringSnoozesCancelCmd)

	// channels
	g.monitoringChannelsListCmd.Flags().StringVar(&g.flagChanListFormat, "format", "", "Output format (e.g. json)")
	g.monitoringChannelsListCmd.Flags().StringVar(&g.flagChanListFilter, "filter", "", "Filter expression")
	g.monitoringChannelsListCmd.Flags().BoolVar(&g.flagChanListURI, "uri", false, "Print resource names")
	g.monitoringChannelsCreateCmd.Flags().StringVar(&g.flagChanType, "type", "", "Channel type (required)")
	g.monitoringChannelsCreateCmd.MarkFlagRequired("type")
	g.monitoringChannelsCreateCmd.Flags().StringVar(&g.flagChanDisplayName, "display-name", "", "Display name")
	g.monitoringChannelsCreateCmd.Flags().StringToStringVar(&g.flagChanLabels, "channel-labels", nil, "Channel labels (key=value)")
	g.monitoringChannelsUpdateCmd.Flags().StringVar(&g.flagChanUpdateDisplayName, "display-name", "", "New display name")
	g.monitoringChannelsUpdateCmd.Flags().StringToStringVar(&g.flagChanUpdateLabels, "channel-labels", nil, "Channel labels")
	g.monitoringChannelsUpdateCmd.Flags().StringVar(&g.flagChanUpdateFields, "fields", "", "Comma-separated fields to update")
	g.monitoringChannelsCmd.AddCommand(g.monitoringChannelsListCmd)
	g.monitoringChannelsCmd.AddCommand(g.monitoringChannelsDescribeCmd)
	g.monitoringChannelsCmd.AddCommand(g.monitoringChannelsCreateCmd)
	g.monitoringChannelsCmd.AddCommand(g.monitoringChannelsUpdateCmd)
	g.monitoringChannelsCmd.AddCommand(g.monitoringChannelsDeleteCmd)

	g.monitoringCmd.AddCommand(g.monitoringPoliciesCmd)
	g.monitoringCmd.AddCommand(g.monitoringSnoozesCmd)
	g.monitoringCmd.AddCommand(g.monitoringChannelsCmd)
	g.monitoringCmd.AddCommand(g.monitoringChannelsAliasCmd)

}

func (g *monitoringCmds) runMonitoringPoliciesList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}

	var allPolicies []*monitoring.AlertPolicy
	pageToken := ""
	for {
		call := svc.Projects.AlertPolicies.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if g.flagMonPoliciesFilter != "" {
			call = call.Filter(g.flagMonPoliciesFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing alert policies: %w", err)
		}
		allPolicies = append(allPolicies, resp.AlertPolicies...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if g.flagMonPoliciesURI {
		for _, policy := range allPolicies {
			fmt.Fprintln(g.inv.stdout, policy.Name)
		}
		return nil
	}

	return g.formatPoliciesList(allPolicies, g.flagMonPoliciesFormat)
}

// formatPoliciesList renders alert policies per --format. The default (empty)
// format is yaml, matching gcloud. It also supports json; any other value
// prints the NAME/ENABLED/DISPLAY_NAME table.
func (g *monitoringCmds) formatPoliciesList(policies []*monitoring.AlertPolicy, format string) error {
	switch {
	case format == "" || format == "yaml":
		for _, p := range policies {
			fmt.Fprintln(g.inv.stdout, "---")
			if err := g.inv.yaml(p); err != nil {
				return err
			}
		}
		return nil
	case format == "json":
		// AlertPolicy's generated MarshalJSON HTML-escapes operators (so a
		// filter like "severity>=ERROR" comes out with the ">" unicode-escaped).
		// Round-trip through a generic value, which decodes those escapes, then
		// re-encode with HTML escaping disabled so operators appear literally.
		data, err := json.Marshal(policies)
		if err != nil {
			return err
		}
		var generic any
		if err := json.Unmarshal(data, &generic); err != nil {
			return err
		}
		if generic == nil {
			generic = []any{}
		}
		enc := json.NewEncoder(g.inv.stdout)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		return enc.Encode(generic)
	}

	printTableHeadingTo(g.inv.stdout, len(policies), "%-60s %-10s %s\n", "NAME", "ENABLED", "DISPLAY_NAME")
	for _, p := range policies {
		fmt.Fprintf(g.inv.stdout, "%-60s %-10t %s\n", p.Name, p.Enabled, p.DisplayName)
	}
	return nil
}

func (g *monitoringCmds) runMonitoringSnoozesCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}

	var snooze *monitoring.Snooze
	if g.flagSnoozeFromFile != "" {
		data, err := os.ReadFile(g.flagSnoozeFromFile)
		if err != nil {
			return fmt.Errorf("reading snooze file: %w", err)
		}
		snooze = &monitoring.Snooze{}
		if err := json.Unmarshal(data, snooze); err != nil {
			return fmt.Errorf("parsing snooze file: %w", err)
		}
	} else {
		snooze = &monitoring.Snooze{}
	}
	// Apply flag overrides (flags take precedence over file values).
	if g.flagSnoozeDisplayName != "" {
		snooze.DisplayName = g.flagSnoozeDisplayName
	}
	if g.flagSnoozeStartTime != "" || g.flagSnoozeEndTime != "" {
		if snooze.Interval == nil {
			snooze.Interval = &monitoring.TimeInterval{}
		}
		if g.flagSnoozeStartTime != "" {
			snooze.Interval.StartTime = g.flagSnoozeStartTime
		}
		if g.flagSnoozeEndTime != "" {
			snooze.Interval.EndTime = g.flagSnoozeEndTime
		}
	}
	if len(g.flagSnoozePolicies) > 0 || g.flagSnoozeFilter != "" {
		if snooze.Criteria == nil {
			snooze.Criteria = &monitoring.Criteria{}
		}
		if len(g.flagSnoozePolicies) > 0 {
			snooze.Criteria.Policies = g.flagSnoozePolicies
		}
		if g.flagSnoozeFilter != "" {
			snooze.Criteria.Filter = g.flagSnoozeFilter
		}
	}

	result, err := svc.Projects.Snoozes.Create(fmt.Sprintf("projects/%s", project), snooze).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating snooze: %w", err)
	}

	if g.flagSnoozeCreateFormat != "" {
		return g.inv.formatOutput(result, g.flagSnoozeCreateFormat)
	}

	fmt.Fprintf(g.inv.stdout, "Created snooze [%s].\n", result.Name)
	return nil
}

func snoozeName(project, snoozeID string) string {
	// Accept either a bare snooze ID or a full resource name.
	if strings.HasPrefix(snoozeID, "projects/") {
		return snoozeID
	}
	return fmt.Sprintf("projects/%s/snoozes/%s", project, snoozeID)
}

func (g *monitoringCmds) runMonitoringSnoozesDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}

	snooze, err := svc.Projects.Snoozes.Get(snoozeName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing snooze: %w", err)
	}

	return g.formatSnooze(snooze, g.flagSnoozeDescribeFormat)
}

// formatSnooze renders a single snooze per --format. The default (empty) format
// is yaml, matching gcloud. It also supports json and csv(COLS)/table(COLS).
func (g *monitoringCmds) formatSnooze(s *monitoring.Snooze, format string) error {
	switch {
	case format == "" || format == "yaml":
		return g.inv.yaml(s)
	case format == "json":
		return g.inv.printJSON(s)
	case isCsvFormat(format):
		return g.printSnoozeColumns([]*monitoring.Snooze{s}, extractCsvFields(format), ",")
	case isTableFormat(format):
		return g.printSnoozeTable([]*monitoring.Snooze{s}, extractTableFields(format))
	}
	return g.inv.yaml(s)
}

// printSnoozeTable prints an aligned table with uppercased column headings,
// auto-sizing each column to the widest value (matching gcloud table output).
func (g *monitoringCmds) printSnoozeTable(snoozes []*monitoring.Snooze, fields []string) error {
	headers := make([]string, len(fields))
	widths := make([]int, len(fields))
	for i, f := range fields {
		headers[i] = strings.ToUpper(f)
		widths[i] = len(headers[i])
	}
	rows := make([][]string, 0, len(snoozes))
	for _, s := range snoozes {
		row := make([]string, len(fields))
		for i, f := range fields {
			row[i] = snoozeField(s, f)
			if len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
		rows = append(rows, row)
	}

	printRow := func(cols []string) {
		var b strings.Builder
		for i, c := range cols {
			if i > 0 {
				b.WriteString("  ")
			}
			if i < len(cols)-1 {
				fmt.Fprintf(&b, "%-*s", widths[i], c)
			} else {
				b.WriteString(c)
			}
		}
		fmt.Fprintln(g.inv.stdout, b.String())
	}

	printRow(headers)
	for _, row := range rows {
		printRow(row)
	}
	return nil
}

func (g *monitoringCmds) runMonitoringSnoozesList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}

	var allSnoozes []*monitoring.Snooze
	pageToken := ""
	for {
		call := svc.Projects.Snoozes.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if g.flagSnoozesListFilter != "" {
			call = call.Filter(g.flagSnoozesListFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing snoozes: %w", err)
		}
		allSnoozes = append(allSnoozes, resp.Snoozes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if g.flagSnoozesListURI {
		for _, snooze := range allSnoozes {
			fmt.Fprintln(g.inv.stdout, snooze.Name)
		}
		return nil
	}

	return g.formatSnoozesList(allSnoozes, g.flagSnoozesListFormat)
}

// formatSnoozesList renders snoozes per --format. The default (empty) format is
// yaml, matching gcloud. It also supports json and csv(COLS)/table(COLS); any
// other value falls back to the default NAME/DISPLAY_NAME table.
func (g *monitoringCmds) formatSnoozesList(snoozes []*monitoring.Snooze, format string) error {
	switch {
	case format == "" || format == "yaml":
		for _, s := range snoozes {
			fmt.Fprintln(g.inv.stdout, "---")
			if err := g.inv.yaml(s); err != nil {
				return err
			}
		}
		return nil
	case format == "json":
		return g.inv.printJSON(snoozes)
	case isCsvFormat(format):
		return g.printSnoozeColumns(snoozes, extractCsvFields(format), ",")
	case isTableFormat(format):
		return g.printSnoozeColumns(snoozes, extractTableFields(format), "\t")
	}

	// Default columnar table.
	printTableHeadingTo(g.inv.stdout, len(snoozes), "%-60s %s\n", "NAME", "DISPLAY_NAME")
	for _, s := range snoozes {
		fmt.Fprintf(g.inv.stdout, "%-60s %s\n", s.Name, s.DisplayName)
	}
	return nil
}

// printSnoozeColumns prints a heading row of the requested column names (as
// given) followed by one row per snooze, joined by sep.
func (g *monitoringCmds) printSnoozeColumns(snoozes []*monitoring.Snooze, fields []string, sep string) error {
	fmt.Fprintln(g.inv.stdout, strings.Join(fields, sep))
	for _, s := range snoozes {
		vals := make([]string, len(fields))
		for i, f := range fields {
			vals[i] = snoozeField(s, f)
		}
		fmt.Fprintln(g.inv.stdout, strings.Join(vals, sep))
	}
	return nil
}

// snoozeField extracts a display column from a snooze. Column names are
// case-insensitive.
func snoozeField(s *monitoring.Snooze, field string) string {
	switch strings.ToUpper(field) {
	case "NAME":
		return s.Name
	case "DISPLAY_NAME", "DISPLAYNAME":
		return s.DisplayName
	}
	return ""
}

func (g *monitoringCmds) runMonitoringSnoozesUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}

	name := snoozeName(project, args[0])

	if g.flagSnoozeUpdateFromFile != "" {
		data, err := os.ReadFile(g.flagSnoozeUpdateFromFile)
		if err != nil {
			return fmt.Errorf("reading snooze file: %w", err)
		}
		snooze := &monitoring.Snooze{}
		if err := json.Unmarshal(data, snooze); err != nil {
			return fmt.Errorf("parsing snooze file: %w", err)
		}
		result, err := svc.Projects.Snoozes.Patch(name, snooze).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("updating snooze: %w", err)
		}
		fmt.Fprintf(g.inv.stdout, "Updated snooze [%s].\n", result.Name)
		return nil
	}

	// Read-modify-write: get current snooze then apply updates.
	snooze, err := svc.Projects.Snoozes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting snooze: %w", err)
	}

	var updateFields []string
	if g.flagSnoozeUpdateDisplayName != "" {
		snooze.DisplayName = g.flagSnoozeUpdateDisplayName
		updateFields = append(updateFields, "display_name")
	}
	if g.flagSnoozeUpdateStartTime != "" {
		if snooze.Interval == nil {
			snooze.Interval = &monitoring.TimeInterval{}
		}
		snooze.Interval.StartTime = g.flagSnoozeUpdateStartTime
		updateFields = append(updateFields, "interval.start_time")
	}
	if g.flagSnoozeUpdateEndTime != "" {
		if snooze.Interval == nil {
			snooze.Interval = &monitoring.TimeInterval{}
		}
		snooze.Interval.EndTime = g.flagSnoozeUpdateEndTime
		updateFields = append(updateFields, "interval.end_time")
	}

	if len(updateFields) == 0 {
		return fmt.Errorf("at least one of --display-name, --start-time, or --end-time is required")
	}

	result, err := svc.Projects.Snoozes.Patch(name, snooze).
		UpdateMask(joinFields(updateFields)).
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating snooze: %w", err)
	}

	fmt.Fprintf(g.inv.stdout, "Updated snooze [%s].\n", result.Name)
	return nil
}

func joinFields(fields []string) string {
	result := ""
	for i, f := range fields {
		if i > 0 {
			result += ","
		}
		result += f
	}
	return result
}

func (g *monitoringCmds) runMonitoringSnoozesCancel(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}

	name := snoozeName(project, args[0])

	// Get current snooze to check timing.
	snooze, err := svc.Projects.Snoozes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting snooze: %w", err)
	}

	// Cancel = set end time to now (+ 1 second).
	// If start time is in the future, also set start to now.
	now := time.Now().UTC().Add(time.Second).Format(time.RFC3339)
	updateMask := "interval.end_time"

	if snooze.Interval != nil && snooze.Interval.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, snooze.Interval.StartTime)
		if err == nil && startTime.After(time.Now()) {
			snooze.Interval.StartTime = now
			updateMask = "interval.start_time,interval.end_time"
		}
	}

	if snooze.Interval == nil {
		snooze.Interval = &monitoring.TimeInterval{}
	}
	snooze.Interval.EndTime = now

	result, err := svc.Projects.Snoozes.Patch(name, snooze).
		UpdateMask(updateMask).
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cancelling snooze: %w", err)
	}

	fmt.Fprintf(g.inv.stdout, "Cancelled snooze [%s].\n", result.Name)
	return nil
}

// --- policies describe (#179) ---

func (g *monitoringCmds) runMonitoringPoliciesDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/alertPolicies/%s", project, args[0])
	policy, err := svc.Projects.AlertPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing alert policy: %w", err)
	}

	return g.inv.formatOutput(policy, "")
}

// --- policies create (#180) ---

func (g *monitoringCmds) runMonitoringPoliciesCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}

	var policy *monitoring.AlertPolicy
	if g.flagPolCreateFromFile != "" {
		data, err := os.ReadFile(g.flagPolCreateFromFile)
		if err != nil {
			return fmt.Errorf("reading policy file: %w", err)
		}
		policy = &monitoring.AlertPolicy{}
		if err := json.Unmarshal(data, policy); err != nil {
			return fmt.Errorf("parsing policy file: %w", err)
		}
	} else {
		policy = &monitoring.AlertPolicy{}
	}

	if g.flagPolCreateDisplayName != "" {
		policy.DisplayName = g.flagPolCreateDisplayName
	}
	if g.flagPolCreateNoEnabled {
		policy.Enabled = false
	} else {
		policy.Enabled = g.flagPolCreateEnabled
	}
	if len(g.flagPolCreateChannels) > 0 {
		policy.NotificationChannels = g.flagPolCreateChannels
	}
	if g.flagPolCreateDocumentation != "" {
		policy.Documentation = &monitoring.Documentation{
			Content: g.flagPolCreateDocumentation,
		}
	}

	result, err := svc.Projects.AlertPolicies.Create(fmt.Sprintf("projects/%s", project), policy).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating alert policy: %w", err)
	}

	fmt.Fprintf(g.inv.stdout, "Created alert policy [%s].\n", result.Name)
	return nil
}

// --- policies update (#181) ---

func (g *monitoringCmds) runMonitoringPoliciesUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/alertPolicies/%s", project, args[0])

	if g.flagPolUpdateFromFile != "" {
		data, err := os.ReadFile(g.flagPolUpdateFromFile)
		if err != nil {
			return fmt.Errorf("reading policy file: %w", err)
		}
		policy := &monitoring.AlertPolicy{}
		if err := json.Unmarshal(data, policy); err != nil {
			return fmt.Errorf("parsing policy file: %w", err)
		}
		call := svc.Projects.AlertPolicies.Patch(name, policy).Context(ctx)
		if g.flagPolUpdateFields != "" {
			call = call.UpdateMask(g.flagPolUpdateFields)
		}
		result, err := call.Do()
		if err != nil {
			return fmt.Errorf("updating alert policy: %w", err)
		}
		fmt.Fprintf(g.inv.stdout, "Updated alert policy [%s].\n", result.Name)
		return nil
	}

	// Read-modify-write.
	policy, err := svc.Projects.AlertPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting alert policy: %w", err)
	}

	var updateFields []string
	if g.flagPolUpdateEnabled {
		policy.Enabled = true
		updateFields = append(updateFields, "enabled")
	}
	if g.flagPolUpdateNoEnabled {
		policy.Enabled = false
		updateFields = append(updateFields, "enabled")
	}
	if len(g.flagPolAddChannels) > 0 {
		policy.NotificationChannels = append(policy.NotificationChannels, g.flagPolAddChannels...)
		updateFields = append(updateFields, "notification_channels")
	}
	if len(g.flagPolRemoveChannels) > 0 {
		removeSet := make(map[string]bool)
		for _, ch := range g.flagPolRemoveChannels {
			removeSet[ch] = true
		}
		var kept []string
		for _, ch := range policy.NotificationChannels {
			if !removeSet[ch] {
				kept = append(kept, ch)
			}
		}
		policy.NotificationChannels = kept
		updateFields = append(updateFields, "notification_channels")
	}

	if len(updateFields) == 0 && g.flagPolUpdateFields == "" {
		return fmt.Errorf("no update flags specified")
	}

	mask := g.flagPolUpdateFields
	if mask == "" {
		mask = strings.Join(updateFields, ",")
	}

	result, err := svc.Projects.AlertPolicies.Patch(name, policy).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating alert policy: %w", err)
	}

	fmt.Fprintf(g.inv.stdout, "Updated alert policy [%s].\n", result.Name)
	return nil
}

// --- policies delete (#182) ---

func (g *monitoringCmds) runMonitoringPoliciesDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	if !g.inv.flags.quiet {
		fmt.Fprintf(g.inv.stdout, "You are about to delete alert policy [%s].\n", args[0])
		fmt.Fprint(g.inv.stdout, "Do you want to continue (Y/n)? ")
		var answer string
		fmt.Scanln(&answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "" && answer != "y" && answer != "yes" {
			fmt.Fprintln(g.inv.stdout, "Aborted.")
			return nil
		}
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/alertPolicies/%s", project, args[0])
	if _, err := svc.Projects.AlertPolicies.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting alert policy: %w", err)
	}

	fmt.Fprintf(g.inv.stdout, "Deleted alert policy [%s].\n", args[0])
	return nil
}

// --- channels list (#183) ---

func (g *monitoringCmds) runMonitoringChannelsList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}

	var allChannels []*monitoring.NotificationChannel
	pageToken := ""
	for {
		call := svc.Projects.NotificationChannels.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if g.flagChanListFilter != "" {
			call = call.Filter(g.flagChanListFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing notification channels: %w", err)
		}
		allChannels = append(allChannels, resp.NotificationChannels...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if g.flagChanListURI {
		for _, ch := range allChannels {
			fmt.Fprintln(g.inv.stdout, ch.Name)
		}
		return nil
	}

	if g.flagChanListFormat == "json" {
		return g.inv.printJSON(allChannels)
	}

	printTableHeadingTo(g.inv.stdout, len(allChannels), "%-60s %-15s %-10s %s\n", "NAME", "TYPE", "ENABLED", "DISPLAY_NAME")
	for _, ch := range allChannels {
		fmt.Fprintf(g.inv.stdout, "%-60s %-15s %-10t %s\n", ch.Name, ch.Type, ch.Enabled, ch.DisplayName)
	}
	return nil
}

// --- channels describe (#184) ---

func (g *monitoringCmds) runMonitoringChannelsDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/notificationChannels/%s", project, args[0])
	ch, err := svc.Projects.NotificationChannels.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing notification channel: %w", err)
	}

	return g.inv.formatOutput(ch, "")
}

// --- channels create (#184) ---

func (g *monitoringCmds) runMonitoringChannelsCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}

	ch := &monitoring.NotificationChannel{
		Type: g.flagChanType,
	}
	if g.flagChanDisplayName != "" {
		ch.DisplayName = g.flagChanDisplayName
	}
	if len(g.flagChanLabels) > 0 {
		ch.Labels = g.flagChanLabels
	}

	result, err := svc.Projects.NotificationChannels.Create(fmt.Sprintf("projects/%s", project), ch).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating notification channel: %w", err)
	}

	fmt.Fprintf(g.inv.stdout, "Created notification channel [%s].\n", result.Name)
	return nil
}

// --- channels update (#184) ---

func (g *monitoringCmds) runMonitoringChannelsUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/notificationChannels/%s", project, args[0])

	ch, err := svc.Projects.NotificationChannels.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting notification channel: %w", err)
	}

	var updateFields []string
	if g.flagChanUpdateDisplayName != "" {
		ch.DisplayName = g.flagChanUpdateDisplayName
		updateFields = append(updateFields, "display_name")
	}
	if len(g.flagChanUpdateLabels) > 0 {
		if ch.Labels == nil {
			ch.Labels = make(map[string]string)
		}
		for k, v := range g.flagChanUpdateLabels {
			ch.Labels[k] = v
		}
		updateFields = append(updateFields, "labels")
	}

	mask := g.flagChanUpdateFields
	if mask == "" && len(updateFields) > 0 {
		mask = strings.Join(updateFields, ",")
	}

	result, err := svc.Projects.NotificationChannels.Patch(name, ch).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating notification channel: %w", err)
	}

	fmt.Fprintf(g.inv.stdout, "Updated notification channel [%s].\n", result.Name)
	return nil
}

// --- channels delete (#184) ---

func (g *monitoringCmds) runMonitoringChannelsDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	if !g.inv.flags.quiet {
		fmt.Fprintf(g.inv.stdout, "You are about to delete notification channel [%s].\n", args[0])
		fmt.Fprint(g.inv.stdout, "Do you want to continue (Y/n)? ")
		var answer string
		fmt.Scanln(&answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "" && answer != "y" && answer != "yes" {
			fmt.Fprintln(g.inv.stdout, "Aborted.")
			return nil
		}
	}

	ctx := cmd.Context()
	svc, err := g.inv.clients().Monitoring(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/notificationChannels/%s", project, args[0])
	if _, err := svc.Projects.NotificationChannels.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting notification channel: %w", err)
	}

	fmt.Fprintf(g.inv.stdout, "Deleted notification channel [%s].\n", args[0])
	return nil
}

// monitoringCmds holds the commands and flag values of one invocation's "monitoring"
// command group, so invocations don't share state (#2070).
type monitoringCmds struct {
	inv *invocation

	monitoringCmd                 *cobra.Command
	monitoringPoliciesCmd         *cobra.Command
	monitoringPoliciesListCmd     *cobra.Command
	flagMonPoliciesFormat         string
	flagMonPoliciesFilter         string
	flagMonPoliciesURI            bool
	monitoringSnoozesCmd          *cobra.Command
	monitoringSnoozesCreateCmd    *cobra.Command
	flagSnoozeDisplayName         string
	flagSnoozeStartTime           string
	flagSnoozeEndTime             string
	flagSnoozePolicies            []string
	flagSnoozeFilter              string
	flagSnoozeFromFile            string
	flagSnoozeCreateFormat        string
	monitoringSnoozesDescribeCmd  *cobra.Command
	flagSnoozeDescribeFormat      string
	monitoringSnoozesListCmd      *cobra.Command
	flagSnoozesListFormat         string
	flagSnoozesListURI            bool
	monitoringSnoozesUpdateCmd    *cobra.Command
	flagSnoozeUpdateDisplayName   string
	flagSnoozeUpdateStartTime     string
	flagSnoozeUpdateEndTime       string
	flagSnoozeUpdateFromFile      string
	monitoringSnoozesCancelCmd    *cobra.Command
	monitoringPoliciesDescribeCmd *cobra.Command
	monitoringPoliciesCreateCmd   *cobra.Command
	flagPolCreateFromFile         string
	flagPolCreateDisplayName      string
	flagPolCreateEnabled          bool
	flagPolCreateNoEnabled        bool
	flagPolCreateChannels         []string
	flagPolCreateDocumentation    string
	monitoringPoliciesUpdateCmd   *cobra.Command
	flagPolUpdateFromFile         string
	flagPolUpdateEnabled          bool
	flagPolUpdateNoEnabled        bool
	flagPolUpdateFields           string
	flagPolAddChannels            []string
	flagPolRemoveChannels         []string
	monitoringPoliciesDeleteCmd   *cobra.Command
	// monitoringChannelsCmd is registered under its gcloud-python name
	// "notification-channels", with "channels" kept as a deprecated alias for
	// backwards compatibility (#543).
	monitoringChannelsCmd         *cobra.Command
	monitoringChannelsAliasCmd    *cobra.Command
	monitoringChannelsListCmd     *cobra.Command
	flagChanListFormat            string
	flagChanListFilter            string
	flagChanListURI               bool
	monitoringChannelsDescribeCmd *cobra.Command
	monitoringChannelsCreateCmd   *cobra.Command
	flagChanType                  string
	flagChanDisplayName           string
	flagChanLabels                map[string]string
	monitoringChannelsUpdateCmd   *cobra.Command
	flagChanUpdateDisplayName     string
	flagChanUpdateLabels          map[string]string
	flagChanUpdateFields          string
	monitoringChannelsDeleteCmd   *cobra.Command
	// --- snoozes list filter ---
	flagSnoozesListFilter           string
	monitoringAlertsCmd             *cobra.Command
	monitoringAlertsListCmd         *cobra.Command
	monitoringAlertsDescribeCmd     *cobra.Command
	flagMonAlertsFormat             string
	flagMonAlertsFilter             string
	flagMonAlertsSortBy             []string
	flagMonAlertsPageSize           int64
	flagMonAlertsLimit              int
	flagMonAlertsURI                bool
	flagMonDashFile                 string
	flagMonDashFilter               string
	flagMonDashPage                 int64
	flagMonDashLimit                int64
	flagMonDashDisplay              string
	flagMonUCFile                   string
	flagMonUCDisplayName            string
	flagMonUCResourceType           string
	flagMonUCHost                   string
	flagMonUCPath                   string
	flagMonUCPort                   int64
	flagMonUCProtocol               string
	flagMonUCPeriod                 string
	flagMonUCTimeout                string
	flagMonUCFilter                 string
	flagMonUCPage                   int64
	flagMonUCLimit                  int64
	monitoringDashboardsCmd         *cobra.Command
	monitoringDashboardsCreateCmd   *cobra.Command
	monitoringDashboardsDeleteCmd   *cobra.Command
	monitoringDashboardsDescribeCmd *cobra.Command
	monitoringDashboardsListCmd     *cobra.Command
	monitoringDashboardsUpdateCmd   *cobra.Command
	monitoringUptimeCmd             *cobra.Command
	monitoringUptimeCreateCmd       *cobra.Command
	monitoringUptimeDeleteCmd       *cobra.Command
	monitoringUptimeDescribeCmd     *cobra.Command
	monitoringUptimeListCmd         *cobra.Command
	monitoringUptimeUpdateCmd       *cobra.Command
}

// newMonitoringCmd returns the "monitoring" command group for inv.
func newMonitoringCmd(inv *invocation) *cobra.Command {
	return newMonitoringCmds(inv).monitoringCmd
}

// newMonitoringCmds builds the "monitoring" command group for inv and returns its state.
func newMonitoringCmds(inv *invocation) *monitoringCmds {
	g := &monitoringCmds{inv: inv}
	g.monitoringCmd = &cobra.Command{
		Use:   "monitoring",
		Short: "Manage Cloud Monitoring",
	}
	g.monitoringPoliciesCmd = &cobra.Command{
		Use:   "policies",
		Short: "Manage alerting policies",
	}
	g.monitoringPoliciesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List alerting policies",
		Args:  cobra.NoArgs,
		RunE:  g.runMonitoringPoliciesList,
	}
	g.monitoringSnoozesCmd = &cobra.Command{
		Use:   "snoozes",
		Short: "Manage alert snoozes",
	}
	g.monitoringSnoozesCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create an alert snooze",
		Args:  cobra.NoArgs,
		RunE:  g.runMonitoringSnoozesCreate,
	}
	g.monitoringSnoozesDescribeCmd = &cobra.Command{
		Use:   "describe SNOOZE_ID",
		Short: "Describe an alert snooze",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runMonitoringSnoozesDescribe,
	}
	g.monitoringSnoozesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List alert snoozes",
		Args:  cobra.NoArgs,
		RunE:  g.runMonitoringSnoozesList,
	}
	g.monitoringSnoozesUpdateCmd = &cobra.Command{
		Use:   "update SNOOZE_ID",
		Short: "Update an alert snooze",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runMonitoringSnoozesUpdate,
	}
	g.monitoringSnoozesCancelCmd = &cobra.Command{
		Use:   "cancel SNOOZE_ID",
		Short: "Cancel an alert snooze",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runMonitoringSnoozesCancel,
	}
	g.monitoringPoliciesDescribeCmd = &cobra.Command{
		Use:   "describe POLICY_ID",
		Short: "Describe an alerting policy",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runMonitoringPoliciesDescribe,
	}
	g.monitoringPoliciesCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create an alerting policy",
		Args:  cobra.NoArgs,
		RunE:  g.runMonitoringPoliciesCreate,
	}
	g.monitoringPoliciesUpdateCmd = &cobra.Command{
		Use:   "update POLICY_ID",
		Short: "Update an alerting policy",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runMonitoringPoliciesUpdate,
	}
	g.monitoringPoliciesDeleteCmd = &cobra.Command{
		Use:   "delete POLICY_ID",
		Short: "Delete an alerting policy",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runMonitoringPoliciesDelete,
	}
	g.monitoringChannelsCmd = &cobra.Command{
		Use:   "notification-channels",
		Short: "Manage notification channels",
	}
	g.monitoringChannelsAliasCmd = &cobra.Command{
		Use:        "channels",
		Short:      "(DEPRECATED) alias for gcloud monitoring notification-channels",
		Deprecated: "use `gcloud monitoring notification-channels` instead.",
	}
	g.monitoringChannelsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List notification channels",
		Args:  cobra.NoArgs,
		RunE:  g.runMonitoringChannelsList,
	}
	g.monitoringChannelsDescribeCmd = &cobra.Command{
		Use:   "describe CHANNEL_ID",
		Short: "Describe a notification channel",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runMonitoringChannelsDescribe,
	}
	g.monitoringChannelsCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a notification channel",
		Args:  cobra.NoArgs,
		RunE:  g.runMonitoringChannelsCreate,
	}
	g.monitoringChannelsUpdateCmd = &cobra.Command{
		Use:   "update CHANNEL_ID",
		Short: "Update a notification channel",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runMonitoringChannelsUpdate,
	}
	g.monitoringChannelsDeleteCmd = &cobra.Command{
		Use:   "delete CHANNEL_ID",
		Short: "Delete a notification channel",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runMonitoringChannelsDelete,
	}
	g.monitoringAlertsCmd = &cobra.Command{Use: "alerts", Short: "Manage Cloud Monitoring alerts"}
	g.monitoringAlertsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List alerts for a project",
		Args:  cobra.NoArgs,
		RunE:  g.runMonitoringAlertsList,
	}
	g.monitoringAlertsDescribeCmd = &cobra.Command{
		Use:   "describe ALERT",
		Short: "Describe an alert",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runMonitoringAlertsDescribe,
	}
	g.monitoringDashboardsCmd = &cobra.Command{
		Use:   "dashboards",
		Short: "Manage Cloud Monitoring custom dashboards",
	}
	g.monitoringDashboardsCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a dashboard from a JSON/YAML definition",
		Args:  cobra.NoArgs,
		RunE:  g.runMonDashCreate,
	}
	g.monitoringDashboardsDeleteCmd = &cobra.Command{
		Use:   "delete DASHBOARD",
		Short: "Delete a dashboard",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runMonDashDelete,
	}
	g.monitoringDashboardsDescribeCmd = &cobra.Command{
		Use:   "describe DASHBOARD",
		Short: "Describe a dashboard",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runMonDashDescribe,
	}
	g.monitoringDashboardsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List dashboards",
		Args:  cobra.NoArgs,
		RunE:  g.runMonDashList,
	}
	g.monitoringDashboardsUpdateCmd = &cobra.Command{
		Use:   "update DASHBOARD",
		Short: "Update a dashboard from a JSON/YAML definition",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runMonDashUpdate,
	}
	g.monitoringUptimeCmd = &cobra.Command{
		Use:   "uptime",
		Short: "Manage Cloud Monitoring uptime check configurations",
	}
	g.monitoringUptimeCreateCmd = &cobra.Command{
		Use:   "create DISPLAY_NAME",
		Short: "Create an uptime check",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runMonUCCreate,
	}
	g.monitoringUptimeDeleteCmd = &cobra.Command{
		Use:   "delete UPTIME_CHECK",
		Short: "Delete an uptime check",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runMonUCDelete,
	}
	g.monitoringUptimeDescribeCmd = &cobra.Command{
		Use:   "describe UPTIME_CHECK",
		Short: "Describe an uptime check",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runMonUCDescribe,
	}
	g.monitoringUptimeListCmd = &cobra.Command{
		Use:   "list",
		Short: "List uptime checks",
		Args:  cobra.NoArgs,
		RunE:  g.runMonUCList,
	}
	g.monitoringUptimeUpdateCmd = &cobra.Command{
		Use:   "update UPTIME_CHECK",
		Short: "Update an uptime check",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runMonUCUpdate,
	}
	g.initMonitoring()
	g.initMonitoringAlerts()
	g.initMonitoringDashboardsUptime()
	return g
}
