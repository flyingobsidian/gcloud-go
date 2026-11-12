package cmd

import (
	"fmt"
	"path"
	"strings"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

func (s *computeCmds) initInstanceGroups() {
	s.unmanagedListInstancesCmd.Flags().StringVar(&s.flagUnmanagedLIFilter, "filter", "", "Client-side filter by instance name (substring match)")
	s.unmanagedListInstancesCmd.Flags().StringVar(&s.flagUnmanagedLIFormat, "format", "", "Output format (e.g. json, 'csv(NAME,STATUS)')")

	s.unmanagedAddInstancesCmd.Flags().StringSliceVar(&s.flagInstances, "instances", nil, "Instance names to add (comma-separated)")
	s.unmanagedAddInstancesCmd.MarkFlagRequired("instances")

	s.unmanagedCmd.AddCommand(s.unmanagedListInstancesCmd)
	s.unmanagedCmd.AddCommand(s.unmanagedAddInstancesCmd)
}

func (s *computeCmds) runUnmanagedListInstances(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, zone, err := s.inv.projectZone()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
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

	allItems = filterUnmanagedInstances(allItems, s.flagUnmanagedLIFilter)

	return s.formatUnmanagedInstances(allItems, s.flagUnmanagedLIFormat)
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
func (s *computeCmds) formatUnmanagedInstances(items []*compute.InstanceWithNamedPorts, format string) error {
	switch {
	case format == "json":
		return s.inv.printJSON(items)
	case isGetFormat(format):
		field := extractGetField(format)
		for _, it := range items {
			fmt.Fprintln(s.inv.stdout, unmanagedInstanceField(it, field))
		}
		return nil
	case isCsvFormat(format):
		fields := extractCsvFields(format)
		headings := make([]string, len(fields))
		for i, f := range fields {
			headings[i] = strings.ToLower(f)
		}
		fmt.Fprintln(s.inv.stdout, strings.Join(headings, ","))
		for _, it := range items {
			vals := make([]string, len(fields))
			for i, f := range fields {
				vals[i] = unmanagedInstanceField(it, f)
			}
			fmt.Fprintln(s.inv.stdout, strings.Join(vals, ","))
		}
		return nil
	}

	if len(items) == 0 {
		fmt.Fprintln(s.inv.stdout, "No instances in group.")
		return nil
	}
	printTableHeadingTo(s.inv.stdout, len(items), "%-40s %s\n", "NAME", "STATUS")
	for _, it := range items {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", unmanagedInstanceField(it, "NAME"), unmanagedInstanceField(it, "STATUS"))
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

func (s *computeCmds) runUnmanagedAddInstances(cmd *cobra.Command, args []string) error {
	group := args[0]
	project, zone, err := s.inv.projectZone()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	var refs []*compute.InstanceReference
	for _, inst := range s.flagInstances {
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
	fmt.Fprintf(s.inv.stdout, "Added %d instance(s) to [%s].\n", len(s.flagInstances), group)
	return nil
}
