package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	compute "google.golang.org/api/compute/v1"
)

var managedCmd = &cobra.Command{
	Use:   "managed",
	Short: "Manage managed instance groups",
}

// --- managed list-instances ---

var managedListInstancesCmd = &cobra.Command{
	Use:   "list-instances INSTANCE_GROUP",
	Short: "List instances in a managed instance group",
	Args:  cobra.ExactArgs(1),
	RunE:  runManagedListInstances,
}

var (
	flagManagedRegion     string
	flagManagedListFormat string
)

// --- managed describe ---

var managedDescribeCmd = &cobra.Command{
	Use:   "describe INSTANCE_GROUP",
	Short: "Describe a managed instance group",
	Args:  cobra.ExactArgs(1),
	RunE:  runManagedDescribe,
}

var flagManagedDescribeRegion string

// --- managed resize ---

var managedResizeCmd = &cobra.Command{
	Use:   "resize INSTANCE_GROUP",
	Short: "Resize a managed instance group",
	Args:  cobra.ExactArgs(1),
	RunE:  runManagedResize,
}

var (
	flagManagedResizeRegion string
	flagManagedResizeSize   int64
)

func init() {
	// list-instances
	managedListInstancesCmd.Flags().StringVar(&flagManagedRegion, "region", "", "Region of the managed instance group")
	managedListInstancesCmd.Flags().StringVar(&flagManagedListFormat, "format", "", "Output format (e.g. json, 'csv(NAME,ZONE,STATUS)', 'get(NAME)')")

	// describe
	managedDescribeCmd.Flags().StringVar(&flagManagedDescribeRegion, "region", "", "Region")

	// resize
	managedResizeCmd.Flags().StringVar(&flagManagedResizeRegion, "region", "", "Region")
	managedResizeCmd.Flags().Int64Var(&flagManagedResizeSize, "size", 0, "New target size")
	managedResizeCmd.MarkFlagRequired("size")

	managedCmd.AddCommand(managedListInstancesCmd)
	managedCmd.AddCommand(managedDescribeCmd)
	managedCmd.AddCommand(managedResizeCmd)
	instanceGroupsCmd.AddCommand(managedCmd)
}

func runManagedListInstances(cmd *cobra.Command, args []string) error {
	group := args[0]
	// Resolve only the project here. The zone is resolved later, on the zonal
	// path, so that supplying --region does not trigger a zone prompt/requirement.
	props, err := loadProps()
	if err != nil {
		return err
	}
	project := resolveProjectOnly(props)
	if project == "" {
		return fmt.Errorf("project is required; set via --project flag, CLOUDSDK_CORE_PROJECT env, or config")
	}
	region := flagManagedRegion

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
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

		return formatManagedInstances(allInstances, flagManagedListFormat, true)
	}

	// Zonal MIG fallback.
	_, zone, err := resolveProjectZone()
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

	return formatManagedInstances(allInstances, flagManagedListFormat, false)
}

// formatManagedInstances renders managed instances per --format. It supports
// json, get(FIELD) and csv(FIELDS); otherwise it prints the default table.
// showZone includes the ZONE column in the default table (regional listings
// span zones; zonal listings do not).
func formatManagedInstances(instances []*compute.ManagedInstance, format string, showZone bool) error {
	switch {
	case format == "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(instances)
	case isGetFormat(format):
		field := extractGetField(format)
		for _, mi := range instances {
			fmt.Println(managedInstanceField(mi, field))
		}
		return nil
	case isCsvFormat(format):
		fields := extractCsvFields(format)
		// Heading row, using the canonical lowercase column names (matching gcloud).
		headings := make([]string, len(fields))
		for i, f := range fields {
			headings[i] = strings.ToLower(f)
		}
		fmt.Println(strings.Join(headings, ","))
		for _, mi := range instances {
			vals := make([]string, 0, len(fields))
			for _, f := range fields {
				vals = append(vals, managedInstanceField(mi, f))
			}
			fmt.Println(strings.Join(vals, ","))
		}
		return nil
	}

	if showZone {
		fmt.Printf("%-40s %-10s %-15s %-20s %-15s\n", "NAME", "ZONE", "STATUS", "ACTION", "HEALTH")
		for _, mi := range instances {
			fmt.Printf("%-40s %-10s %-15s %-20s %-15s\n",
				managedInstanceField(mi, "NAME"), managedInstanceField(mi, "ZONE"),
				managedInstanceField(mi, "STATUS"), managedInstanceField(mi, "ACTION"),
				managedInstanceField(mi, "HEALTH"))
		}
		return nil
	}
	fmt.Printf("%-40s %-15s %-20s %-15s\n", "NAME", "STATUS", "ACTION", "HEALTH")
	for _, mi := range instances {
		fmt.Printf("%-40s %-15s %-20s %-15s\n",
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

func runManagedDescribe(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, _, err := resolveProjectZone()
	if err != nil {
		return err
	}
	region := flagManagedDescribeRegion

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	if region != "" {
		mig, err := svc.RegionInstanceGroupManagers.Get(project, region, group).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("describing managed instance group: %w", err)
		}
		return formatOutput(mig, "")
	}

	_, zone, err := resolveProjectZone()
	if err != nil {
		return fmt.Errorf("either --region or --zone is required")
	}

	mig, err := svc.InstanceGroupManagers.Get(project, zone, group).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing managed instance group: %w", err)
	}
	return formatOutput(mig, "")
}

func runManagedResize(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, _, err := resolveProjectZone()
	if err != nil {
		return err
	}
	region := flagManagedResizeRegion

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	if region != "" {
		op, err := svc.RegionInstanceGroupManagers.Resize(project, region, group, flagManagedResizeSize).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("resizing managed instance group: %w", err)
		}
		fmt.Printf("Resize operation started: %s\n", op.Name)
		return nil
	}

	_, zone, err := resolveProjectZone()
	if err != nil {
		return fmt.Errorf("either --region or --zone is required")
	}

	op, err := svc.InstanceGroupManagers.Resize(project, zone, group, flagManagedResizeSize).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resizing managed instance group: %w", err)
	}
	fmt.Printf("Resize operation started: %s\n", op.Name)
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
