package cmd

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
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing managed instances: %w", err)
			}
			allInstances = append(allInstances, resp.ManagedInstances...)
			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		}

		return s.formatManagedInstances(allInstances, s.flagManagedListFormat, true)
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
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing managed instances: %w", err)
		}
		allInstances = append(allInstances, resp.ManagedInstances...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return s.formatManagedInstances(allInstances, s.flagManagedListFormat, false)
}

// formatManagedInstances renders managed instances per --format. It supports
// json, get(FIELD) and csv(FIELDS); otherwise it prints the default table.
// showZone includes the ZONE column in the default table (regional listings
// span zones; zonal listings do not).
func (s *computeCmds) formatManagedInstances(instances []*compute.ManagedInstance, format string, showZone bool) error {
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
	}

	if showZone {
		printTableHeadingTo(s.inv.stdout, len(instances), "%-40s %-10s %-15s %-20s %-15s\n", "NAME", "ZONE", "STATUS", "ACTION", "HEALTH")
		for _, mi := range instances {
			fmt.Fprintf(s.inv.stdout, "%-40s %-10s %-15s %-20s %-15s\n",
				managedInstanceField(mi, "NAME"), managedInstanceField(mi, "ZONE"),
				managedInstanceField(mi, "STATUS"), managedInstanceField(mi, "ACTION"),
				managedInstanceField(mi, "HEALTH"))
		}
		return nil
	}
	printTableHeadingTo(s.inv.stdout, len(instances), "%-40s %-15s %-20s %-15s\n", "NAME", "STATUS", "ACTION", "HEALTH")
	for _, mi := range instances {
		fmt.Fprintf(s.inv.stdout, "%-40s %-15s %-20s %-15s\n",
			managedInstanceField(mi, "NAME"), managedInstanceField(mi, "STATUS"),
			managedInstanceField(mi, "ACTION"), managedInstanceField(mi, "HEALTH"))
	}
	return nil
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
