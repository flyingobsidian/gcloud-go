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
	"google.golang.org/api/compute/v1"
)

var unmanagedListInstancesCmd = &cobra.Command{
	Use:   "list-instances INSTANCE_GROUP",
	Short: "List instances in an unmanaged instance group",
	Args:  cobra.ExactArgs(1),
	RunE:  runUnmanagedListInstances,
}

var unmanagedAddInstancesCmd = &cobra.Command{
	Use:   "add-instances INSTANCE_GROUP",
	Short: "Add instances to an unmanaged instance group",
	Args:  cobra.ExactArgs(1),
	RunE:  runUnmanagedAddInstances,
}

var flagInstances []string

var (
	flagUnmanagedLIFilter string
	flagUnmanagedLIFormat string
)

func init() {
	unmanagedListInstancesCmd.Flags().StringVar(&flagUnmanagedLIFilter, "filter", "", "Client-side filter by instance name (substring match)")
	unmanagedListInstancesCmd.Flags().StringVar(&flagUnmanagedLIFormat, "format", "", "Output format (e.g. json, 'csv(NAME,STATUS)')")

	unmanagedAddInstancesCmd.Flags().StringSliceVar(&flagInstances, "instances", nil, "Instance names to add (comma-separated)")
	unmanagedAddInstancesCmd.MarkFlagRequired("instances")

	unmanagedCmd.AddCommand(unmanagedListInstancesCmd)
	unmanagedCmd.AddCommand(unmanagedAddInstancesCmd)
}

func runUnmanagedListInstances(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, zone, err := resolveProjectZone()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var allItems []*compute.InstanceWithNamedPorts
	pageToken := ""
	for {
		call := svc.InstanceGroups.ListInstances(project, zone, group,
			&compute.InstanceGroupsListInstancesRequest{}).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing instances: %w", err)
		}
		allItems = append(allItems, resp.Items...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	allItems = filterUnmanagedInstances(allItems, flagUnmanagedLIFilter)

	return formatUnmanagedInstances(allItems, flagUnmanagedLIFormat)
}

// filterUnmanagedInstances keeps only the instances whose name contains filter
// (case-insensitive). An empty filter returns all instances.
func filterUnmanagedInstances(items []*compute.InstanceWithNamedPorts, filter string) []*compute.InstanceWithNamedPorts {
	if filter == "" {
		return items
	}
	needle := strings.ToLower(filter)
	out := make([]*compute.InstanceWithNamedPorts, 0, len(items))
	for _, it := range items {
		if strings.Contains(strings.ToLower(path.Base(it.Instance)), needle) {
			out = append(out, it)
		}
	}
	return out
}

// formatUnmanagedInstances renders the instances per --format. It supports json,
// get(FIELD) and csv(FIELDS); otherwise it prints the default NAME/STATUS table.
func formatUnmanagedInstances(items []*compute.InstanceWithNamedPorts, format string) error {
	switch {
	case format == "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(items)
	case isGetFormat(format):
		field := extractGetField(format)
		for _, it := range items {
			fmt.Println(unmanagedInstanceField(it, field))
		}
		return nil
	case isCsvFormat(format):
		fields := extractCsvFields(format)
		headings := make([]string, len(fields))
		for i, f := range fields {
			headings[i] = strings.ToLower(f)
		}
		fmt.Println(strings.Join(headings, ","))
		for _, it := range items {
			vals := make([]string, len(fields))
			for i, f := range fields {
				vals[i] = unmanagedInstanceField(it, f)
			}
			fmt.Println(strings.Join(vals, ","))
		}
		return nil
	}

	if len(items) == 0 {
		fmt.Println("No instances in group.")
		return nil
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATUS")
	for _, it := range items {
		fmt.Printf("%-40s %s\n", unmanagedInstanceField(it, "NAME"), unmanagedInstanceField(it, "STATUS"))
	}
	return nil
}

// unmanagedInstanceField extracts a display column from an instance. Field names
// are case-insensitive.
func unmanagedInstanceField(it *compute.InstanceWithNamedPorts, field string) string {
	switch strings.ToUpper(field) {
	case "NAME":
		return path.Base(it.Instance)
	case "STATUS":
		return it.Status
	case "ZONE":
		return extractZoneFromURL(it.Instance)
	}
	return ""
}

func runUnmanagedAddInstances(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, zone, err := resolveProjectZone()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var refs []*compute.InstanceReference
	for _, inst := range flagInstances {
		selfLink := fmt.Sprintf("projects/%s/zones/%s/instances/%s", project, zone, inst)
		refs = append(refs, &compute.InstanceReference{Instance: selfLink})
	}

	req := &compute.InstanceGroupsAddInstancesRequest{Instances: refs}
	op, err := svc.InstanceGroups.AddInstances(project, zone, group, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding instances: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Added %d instance(s) to [%s].\n", len(flagInstances), group)
	return nil
}
