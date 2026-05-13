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
	managedListInstancesCmd.Flags().StringVar(&flagManagedListFormat, "format", "", "Output format (e.g. json)")

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
	project, _, err := resolveProjectZone()
	if err != nil {
		// Try with just project if zone not set.
		props, loadErr := loadProps()
		if loadErr != nil {
			return err
		}
		project = resolveProjectOnly(props)
		if project == "" {
			return err
		}
	}
	region := flagManagedRegion

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	if region != "" {
		// Regional MIG.
		resp, err := svc.RegionInstanceGroupManagers.ListManagedInstances(project, region, group).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing managed instances: %w", err)
		}

		if flagManagedListFormat == "json" {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(resp.ManagedInstances)
		}

		fmt.Printf("%-40s %-10s %-15s\n", "NAME", "ZONE", "STATUS")
		for _, mi := range resp.ManagedInstances {
			name := path.Base(mi.Instance)
			zone := extractZoneFromURL(mi.Instance)
			fmt.Printf("%-40s %-10s %-15s\n", name, zone, mi.InstanceStatus)
		}
		return nil
	}

	// Zonal MIG fallback.
	_, zone, err := resolveProjectZone()
	if err != nil {
		return fmt.Errorf("either --region or --zone is required")
	}

	resp, err := svc.InstanceGroupManagers.ListManagedInstances(project, zone, group).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing managed instances: %w", err)
	}

	if flagManagedListFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(resp.ManagedInstances)
	}

	fmt.Printf("%-40s %-15s\n", "NAME", "STATUS")
	for _, mi := range resp.ManagedInstances {
		name := path.Base(mi.Instance)
		fmt.Printf("%-40s %-15s\n", name, mi.InstanceStatus)
	}
	return nil
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
