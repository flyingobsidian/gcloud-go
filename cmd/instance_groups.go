package cmd

import (
	"context"
	"fmt"
	"path"

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

func init() {
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

	if len(allItems) == 0 {
		fmt.Println("No instances in group.")
		return nil
	}

	fmt.Printf("%-40s %s\n", "NAME", "STATUS")
	for _, item := range allItems {
		name := path.Base(item.Instance)
		fmt.Printf("%-40s %s\n", name, item.Status)
	}
	return nil
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
