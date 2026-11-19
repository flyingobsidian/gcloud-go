package cli

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	compute "google.golang.org/api/compute/v1"
)

// --- managed list-instances ---

// --- managed describe ---

// --- managed resize ---

func (s *computeCmds) initInstanceGroupsManaged() {
	// list-instances
	s.managedListInstancesCmd.Flags().StringVar(&s.flagManagedRegion, "region", "", "Region of the managed instance group")
	s.managedListInstancesCmd.Flags().StringVar(&s.flagManagedListFormat, "format", "", "Output format (e.g. json, 'csv(NAME,ZONE,STATUS)', 'get(NAME)')")
	s.managedListInstancesCmd.Flags().IntVar(&s.flagManagedListInstLimit, "limit", 0, "Maximum number of resources to list")
	s.managedListInstancesCmd.Flags().Int64Var(&s.flagManagedListInstPageSize, "page-size", 0, "Number of resources per page")

	// describe
	s.managedDescribeCmd.Flags().StringVar(&s.flagManagedDescribeRegion, "region", "", "Region")

	// resize
	s.managedResizeCmd.Flags().StringVar(&s.flagManagedResizeRegion, "region", "", "Region")
	s.managedResizeCmd.Flags().Int64Var(&s.flagManagedResizeSize, "size", 0, "New target size")
	s.managedResizeCmd.MarkFlagRequired("size")

	s.managedCmd.AddCommand(s.managedListInstancesCmd)
	s.managedCmd.AddCommand(s.managedDescribeCmd)
	s.managedCmd.AddCommand(s.managedResizeCmd)
	s.instanceGroupsCmd.AddCommand(s.managedCmd)
}

func (s *computeCmds) runManagedListInstances(cmd *cobra.Command, args []string) error {
	group := args[0]
	// Resolve only the project here. The zone is resolved later, on the zonal
	// path, so that supplying --region does not trigger a zone prompt/requirement.
	props, err := s.loadProps()
	if err != nil {
		return err
	}
	project := s.resolveProjectOnly(props)
	if project == "" {
		return fmt.Errorf("project is required; set via --project flag, CLOUDSDK_CORE_PROJECT env, or config")
	}
	region := s.flagManagedRegion
	maxResults, limit := s.managedListInstancesPaging()

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	if region != "" {
		// Regional MIG.
		var allInstances []*compute.ManagedInstance
		pageToken := ""
		for {
			call := svc.RegionInstanceGroupManagers.ListManagedInstances(project, region, group).Context(ctx)
			if maxResults > 0 {
				call = call.MaxResults(maxResults)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing managed instances: %w", err)
			}
			allInstances = append(allInstances, resp.ManagedInstances...)
			if limit > 0 && len(allInstances) >= limit {
				allInstances = allInstances[:limit]
				break
			}
			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		}

		return s.formatManagedInstances(allInstances, s.flagManagedListFormat)
	}

	// Zonal MIG fallback.
	_, zone, err := s.inv.projectZone()
	if err != nil {
		return fmt.Errorf("either --region or --zone is required")
	}

	var allInstances []*compute.ManagedInstance
	pageToken := ""
	for {
		call := svc.InstanceGroupManagers.ListManagedInstances(project, zone, group).Context(ctx)
		if maxResults > 0 {
			call = call.MaxResults(maxResults)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing managed instances: %w", err)
		}
		allInstances = append(allInstances, resp.ManagedInstances...)
		if limit > 0 && len(allInstances) >= limit {
			allInstances = allInstances[:limit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return s.formatManagedInstances(allInstances, s.flagManagedListFormat)
}

// managedListInstancesPaging mirrors gcloud-python 589.0.0's
// list-instances paging: without client-side transforms (--filter,
// --flatten), --page-size and --limit (the smaller of the two,
// capped at 500) are sent as maxResults and the listing stops once --limit
// instances are fetched. It returns maxResults (0 for none) and the limit.
func (s *computeCmds) managedListInstancesPaging() (int64, int) {
	if s.inv.flags.filter != "" || len(s.inv.flags.flatten) > 0 {
		return 0, 0
	}
	maxResults := s.flagManagedListInstPageSize
	if l := int64(s.flagManagedListInstLimit); l > 0 && (maxResults <= 0 || l < maxResults) {
		maxResults = l
	}
	return min(maxResults, 500), s.flagManagedListInstLimit
}

// managedListInstancesColumns returns gcloud-python GA's list-instances
// table columns; GRACEFUL_SHUTDOWN_TIMESTAMP and TERMINATION_TIMESTAMP are
// dynamic fields shown only when some instance has a value (589.0.0).
func managedListInstancesColumns(instances []*compute.ManagedInstance) []string {
	cols := []string{"name", "zone", "status", "healthState", "action", "instanceTemplate", "versionName", "lastError"}
	var graceful, termination bool
	for _, mi := range instances {
		if mi.Scheduling != nil {
			graceful = graceful || mi.Scheduling.GracefulShutdownTimestamp != ""
			termination = termination || mi.Scheduling.TerminationTimestamp != ""
		}
	}
	if graceful {
		cols = append(cols, "gracefulShutdownTimestamp")
	}
	if termination {
		cols = append(cols, "terminationTimestamp")
	}
	return cols
}

// managedListInstancesRows builds the rows of the default table.
func managedListInstancesRows(instances []*compute.ManagedInstance) []map[string]any {
	rows := make([]map[string]any, 0, len(instances))
	for _, mi := range instances {
		row := map[string]any{
			"name":             mi.Name,
			"zone":             extractZoneFromURL(mi.Instance),
			"status":           mi.InstanceStatus,
			"healthState":      "",
			"action":           mi.CurrentAction,
			"instanceTemplate": "",
			"versionName":      "",
			"lastError":        "",
		}
		if row["name"] == "" {
			row["name"] = path.Base(mi.Instance)
		}
		if len(mi.InstanceHealth) > 0 {
			row["healthState"] = mi.InstanceHealth[0].DetailedHealthState
		}
		if mi.Version != nil {
			if mi.Version.InstanceTemplate != "" {
				row["instanceTemplate"] = path.Base(mi.Version.InstanceTemplate)
			}
			row["versionName"] = mi.Version.Name
		}
		if mi.LastAttempt != nil && mi.LastAttempt.Errors != nil {
			var errs []string
			for _, e := range mi.LastAttempt.Errors.Errors {
				errs = append(errs, fmt.Sprintf("Error %s: %s", e.Code, e.Message))
			}
			row["lastError"] = strings.Join(errs, ", ")
		}
		if mi.Scheduling != nil {
			row["gracefulShutdownTimestamp"] = mi.Scheduling.GracefulShutdownTimestamp
			row["terminationTimestamp"] = mi.Scheduling.TerminationTimestamp
		}
		rows = append(rows, row)
	}
	return rows
}

// formatManagedInstances renders managed instances per --format. It supports
// json, get(FIELD) and csv(FIELDS) over the table's column names; otherwise
// it prints gcloud-python's default table.
func (s *computeCmds) formatManagedInstances(instances []*compute.ManagedInstance, format string) error {
	if format == "" {
		format = s.inv.flags.format
	}
	switch {
	case format == "json":
		return s.inv.printJSON(instances)
	case isGetFormat(format):
		field := extractGetField(format)
		for _, mi := range instances {
			fmt.Fprintln(s.inv.stdout, managedInstanceField(mi, field))
		}
		return nil
	case isCsvFormat(format):
		fields := extractCsvFields(format)
		// Heading row, using the canonical lowercase column names (matching gcloud).
		headings := make([]string, len(fields))
		for i, f := range fields {
			headings[i] = strings.ToLower(f)
		}
		fmt.Fprintln(s.inv.stdout, strings.Join(headings, ","))
		for _, mi := range instances {
			vals := make([]string, 0, len(fields))
			for _, f := range fields {
				vals = append(vals, managedInstanceField(mi, f))
			}
			fmt.Fprintln(s.inv.stdout, strings.Join(vals, ","))
		}
		return nil
	case format != "":
		return emitFormattedTo(s.inv.stdout, instances, format)
	case len(instances) == 0:
		fmt.Fprintln(s.inv.stderr, "Listed 0 items.")
		return nil
	}
	table := "table(" + strings.Join(managedListInstancesColumns(instances), ",") + ")"
	return emitFormattedTo(s.inv.stdout, managedListInstancesRows(instances), table)
}

// managedInstanceField extracts a display column from a managed instance.
// Field names match the default table columns and are case-insensitive.
func managedInstanceField(mi *compute.ManagedInstance, field string) string {
	switch strings.ToUpper(field) {
	case "NAME":
		return path.Base(mi.Instance)
	case "ZONE":
		return extractZoneFromURL(mi.Instance)
	case "STATUS":
		return mi.InstanceStatus
	case "ACTION":
		return mi.CurrentAction
	case "HEALTH":
		health := ""
		if mi.InstanceHealth != nil {
			for _, h := range mi.InstanceHealth {
				health = h.DetailedHealthState
			}
		}
		return health
	}
	return ""
}

func (s *computeCmds) runManagedDescribe(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, _, err := s.inv.projectZone()
	if err != nil {
		return err
	}
	region := s.flagManagedDescribeRegion

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	if region != "" {
		mig, err := svc.RegionInstanceGroupManagers.Get(project, region, group).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("describing managed instance group: %w", err)
		}
		return s.inv.formatOutput(mig, "")
	}

	_, zone, err := s.inv.projectZone()
	if err != nil {
		return fmt.Errorf("either --region or --zone is required")
	}

	mig, err := svc.InstanceGroupManagers.Get(project, zone, group).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing managed instance group: %w", err)
	}
	return s.inv.formatOutput(mig, "")
}

func (s *computeCmds) runManagedResize(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, _, err := s.inv.projectZone()
	if err != nil {
		return err
	}
	region := s.flagManagedResizeRegion

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	if region != "" {
		op, err := svc.RegionInstanceGroupManagers.Resize(project, region, group, s.flagManagedResizeSize).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("resizing managed instance group: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Resize operation started: %s\n", op.Name)
		return nil
	}

	_, zone, err := s.inv.projectZone()
	if err != nil {
		return fmt.Errorf("either --region or --zone is required")
	}

	op, err := svc.InstanceGroupManagers.Resize(project, zone, group, s.flagManagedResizeSize).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resizing managed instance group: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Resize operation started: %s\n", op.Name)
	return nil
}

// extractZoneFromURL extracts the zone name from a resource URL like
// "projects/p/zones/us-central1-a/instances/name".
func extractZoneFromURL(url string) string {
	parts := strings.Split(url, "/")
	for i, p := range parts {
		if p == "zones" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
