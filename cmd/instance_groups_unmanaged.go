package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

// --- unmanaged create ---

var unmanagedCreateCmd = &cobra.Command{
	Use:   "create INSTANCE_GROUP",
	Short: "Create an unmanaged instance group",
	Args:  cobra.ExactArgs(1),
	RunE:  runUnmanagedCreate,
}

var flagUnmanagedDescription string

// --- unmanaged delete ---

var unmanagedDeleteCmd = &cobra.Command{
	Use:   "delete INSTANCE_GROUP",
	Short: "Delete an unmanaged instance group",
	Args:  cobra.ExactArgs(1),
	RunE:  runUnmanagedDelete,
}

// --- unmanaged remove-instances ---

var unmanagedRemoveInstancesCmd = &cobra.Command{
	Use:   "remove-instances INSTANCE_GROUP",
	Short: "Remove instances from an unmanaged instance group",
	Args:  cobra.ExactArgs(1),
	RunE:  runUnmanagedRemoveInstances,
}

var flagRemoveInstances []string

// --- unmanaged list ---

var unmanagedListCmd = &cobra.Command{
	Use:   "list",
	Short: "List unmanaged instance groups",
	Args:  cobra.NoArgs,
	RunE:  runUnmanagedList,
}

var flagUnmanagedListFormat string

// --- unmanaged describe ---

var unmanagedDescribeCmd = &cobra.Command{
	Use:   "describe INSTANCE_GROUP",
	Short: "Describe an unmanaged instance group",
	Args:  cobra.ExactArgs(1),
	RunE:  runUnmanagedDescribe,
}

// --- unmanaged set-named-ports ---

var unmanagedSetNamedPortsCmd = &cobra.Command{
	Use:   "set-named-ports INSTANCE_GROUP",
	Short: "Set named ports for an instance group",
	Args:  cobra.ExactArgs(1),
	RunE:  runUnmanagedSetNamedPorts,
}

var flagNamedPorts []string

// --- unmanaged get-named-ports ---

var unmanagedGetNamedPortsCmd = &cobra.Command{
	Use:   "get-named-ports INSTANCE_GROUP",
	Short: "Get named ports for an instance group",
	Args:  cobra.ExactArgs(1),
	RunE:  runUnmanagedGetNamedPorts,
}

func init() {
	unmanagedCreateCmd.Flags().StringVar(&flagUnmanagedDescription, "description", "", "Group description")

	unmanagedRemoveInstancesCmd.Flags().StringSliceVar(&flagRemoveInstances, "instances", nil, "Instance names to remove")
	unmanagedRemoveInstancesCmd.MarkFlagRequired("instances")

	unmanagedListCmd.Flags().StringVar(&flagUnmanagedListFormat, "format", "", "Output format (e.g. json)")

	unmanagedSetNamedPortsCmd.Flags().StringSliceVar(&flagNamedPorts, "named-ports", nil, "Named ports (NAME:PORT,...)")
	unmanagedSetNamedPortsCmd.MarkFlagRequired("named-ports")

	unmanagedCmd.AddCommand(unmanagedCreateCmd)
	unmanagedCmd.AddCommand(unmanagedDeleteCmd)
	unmanagedCmd.AddCommand(unmanagedRemoveInstancesCmd)
	unmanagedCmd.AddCommand(unmanagedListCmd)
	unmanagedCmd.AddCommand(unmanagedDescribeCmd)
	unmanagedCmd.AddCommand(unmanagedSetNamedPortsCmd)
	unmanagedCmd.AddCommand(unmanagedGetNamedPortsCmd)
}

func runUnmanagedCreate(cmd *cobra.Command, args []string) error {
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

	ig := &compute.InstanceGroup{
		Name:        group,
		Description: flagUnmanagedDescription,
	}

	op, err := svc.InstanceGroups.Insert(project, zone, ig).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance group: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Created instance group [%s].\n", group)
	return nil
}

func runUnmanagedDelete(cmd *cobra.Command, args []string) error {
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

	op, err := svc.InstanceGroups.Delete(project, zone, group).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting instance group: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Printf("Deleted instance group [%s].\n", group)
	return nil
}

func runUnmanagedRemoveInstances(cmd *cobra.Command, args []string) error {
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
	for _, inst := range flagRemoveInstances {
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
	fmt.Printf("Removed %d instance(s) from [%s].\n", len(flagRemoveInstances), group)
	return nil
}

func runUnmanagedList(cmd *cobra.Command, args []string) error {
	project, zone, err := resolveProjectZone()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
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

	if flagUnmanagedListFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(groups)
	}

	fmt.Printf("%-30s %-15s %-10s %-30s\n", "NAME", "ZONE", "SIZE", "DESCRIPTION")
	for _, g := range groups {
		fmt.Printf("%-30s %-15s %-10d %-30s\n", g.Name, zone, g.Size, g.Description)
	}
	return nil
}

func runUnmanagedDescribe(cmd *cobra.Command, args []string) error {
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

	ig, err := svc.InstanceGroups.Get(project, zone, group).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance group: %w", err)
	}

	return formatOutput(ig, "")
}

func runUnmanagedSetNamedPorts(cmd *cobra.Command, args []string) error {
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

	var ports []*compute.NamedPort
	for _, np := range flagNamedPorts {
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
	fmt.Printf("Set named ports on [%s].\n", group)
	return nil
}

func runUnmanagedGetNamedPorts(cmd *cobra.Command, args []string) error {
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

	ig, err := svc.InstanceGroups.Get(project, zone, group).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting instance group: %w", err)
	}

	if len(ig.NamedPorts) == 0 {
		fmt.Println("No named ports.")
		return nil
	}

	fmt.Printf("%-20s %s\n", "NAME", "PORT")
	for _, np := range ig.NamedPorts {
		fmt.Printf("%-20s %d\n", np.Name, np.Port)
	}
	return nil
}
