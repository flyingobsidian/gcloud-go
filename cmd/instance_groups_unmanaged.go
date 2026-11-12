package cmd

import (
	"fmt"
	"strconv"
	"strings"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

// --- unmanaged create ---

// --- unmanaged delete ---

// --- unmanaged remove-instances ---

// --- unmanaged list ---

// --- unmanaged describe ---

// --- unmanaged set-named-ports ---

// --- unmanaged get-named-ports ---

func (s *computeCmds) initInstanceGroupsUnmanaged() {
	s.unmanagedCreateCmd.Flags().StringVar(&s.flagUnmanagedDescription, "description", "", "Group description")

	s.unmanagedRemoveInstancesCmd.Flags().StringSliceVar(&s.flagRemoveInstances, "instances", nil, "Instance names to remove")
	s.unmanagedRemoveInstancesCmd.MarkFlagRequired("instances")

	s.unmanagedListCmd.Flags().StringVar(&s.flagUnmanagedListFormat, "format", "", "Output format (e.g. json)")
	s.unmanagedListCmd.Flags().BoolVar(&s.flagUnmanagedListURI, "uri", false, "Print resource URIs")

	s.unmanagedSetNamedPortsCmd.Flags().StringSliceVar(&s.flagNamedPorts, "named-ports", nil, "Named ports (NAME:PORT,...)")
	s.unmanagedSetNamedPortsCmd.MarkFlagRequired("named-ports")

	s.unmanagedCmd.AddCommand(s.unmanagedCreateCmd)
	s.unmanagedCmd.AddCommand(s.unmanagedDeleteCmd)
	s.unmanagedCmd.AddCommand(s.unmanagedRemoveInstancesCmd)
	s.unmanagedCmd.AddCommand(s.unmanagedListCmd)
	s.unmanagedCmd.AddCommand(s.unmanagedDescribeCmd)
	s.unmanagedCmd.AddCommand(s.unmanagedSetNamedPortsCmd)
	s.unmanagedCmd.AddCommand(s.unmanagedGetNamedPortsCmd)
}

func (s *computeCmds) runUnmanagedCreate(cmd *cobra.Command, args []string) error {
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

	ig := &compute.InstanceGroup{
		Name:        group,
		Description: s.flagUnmanagedDescription,
	}

	op, err := svc.InstanceGroups.Insert(project, zone, ig).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance group: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Created instance group [%s].\n", group)
	return nil
}

func (s *computeCmds) runUnmanagedDelete(cmd *cobra.Command, args []string) error {
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

	op, err := svc.InstanceGroups.Delete(project, zone, group).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting instance group: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Deleted instance group [%s].\n", group)
	return nil
}

func (s *computeCmds) runUnmanagedRemoveInstances(cmd *cobra.Command, args []string) error {
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
	for _, inst := range s.flagRemoveInstances {
		selfLink := fmt.Sprintf("projects/%s/zones/%s/instances/%s", project, zone, inst)
		refs = append(refs, &compute.InstanceReference{Instance: selfLink})
	}

	req := &compute.InstanceGroupsRemoveInstancesRequest{Instances: refs}
	op, err := svc.InstanceGroups.RemoveInstances(project, zone, group, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing instances: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Removed %d instance(s) from [%s].\n", len(s.flagRemoveInstances), group)
	return nil
}

func (s *computeCmds) runUnmanagedList(cmd *cobra.Command, args []string) error {
	project, zone, err := s.inv.projectZone()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	var groups []*compute.InstanceGroup
	pageToken := ""
	for {
		call := svc.InstanceGroups.List(project, zone).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing instance groups: %w", err)
		}
		groups = append(groups, resp.Items...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if s.flagUnmanagedListURI {
		for _, g := range groups {
			fmt.Fprintln(s.inv.stdout, g.SelfLink)
		}
		return nil
	}

	if s.flagUnmanagedListFormat == "json" {
		return s.inv.printJSON(groups)
	}

	printTableHeadingTo(s.inv.stdout, len(groups), "%-30s %-15s %-10s %-30s\n", "NAME", "ZONE", "SIZE", "DESCRIPTION")
	for _, g := range groups {
		fmt.Fprintf(s.inv.stdout, "%-30s %-15s %-10d %-30s\n", g.Name, zone, g.Size, g.Description)
	}
	return nil
}

func (s *computeCmds) runUnmanagedDescribe(cmd *cobra.Command, args []string) error {
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

	ig, err := svc.InstanceGroups.Get(project, zone, group).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance group: %w", err)
	}

	return s.inv.formatOutput(ig, "")
}

func (s *computeCmds) runUnmanagedSetNamedPorts(cmd *cobra.Command, args []string) error {
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

	var ports []*compute.NamedPort
	for _, np := range s.flagNamedPorts {
		name, portStr, ok := strings.Cut(np, ":")
		if !ok {
			return fmt.Errorf("invalid named port format %q, expected NAME:PORT", np)
		}
		port, err := strconv.ParseInt(portStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid port number %q: %w", portStr, err)
		}
		ports = append(ports, &compute.NamedPort{Name: name, Port: port})
	}

	req := &compute.InstanceGroupsSetNamedPortsRequest{NamedPorts: ports}
	op, err := svc.InstanceGroups.SetNamedPorts(project, zone, group, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting named ports: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Set named ports on [%s].\n", group)
	return nil
}

func (s *computeCmds) runUnmanagedGetNamedPorts(cmd *cobra.Command, args []string) error {
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

	ig, err := svc.InstanceGroups.Get(project, zone, group).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting instance group: %w", err)
	}

	if len(ig.NamedPorts) == 0 {
		fmt.Fprintln(s.inv.stdout, "No named ports.")
		return nil
	}

	printTableHeadingTo(s.inv.stdout, len(ig.NamedPorts), "%-20s %s\n", "NAME", "PORT")
	for _, np := range ig.NamedPorts {
		fmt.Fprintf(s.inv.stdout, "%-20s %d\n", np.Name, np.Port)
	}
	return nil
}
