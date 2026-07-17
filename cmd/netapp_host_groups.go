package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	netapp "google.golang.org/api/netapp/v1"
)

// --- gcloud netapp host-groups (#1200) ---

var netappHGCmd = &cobra.Command{Use: "host-groups", Short: "Manage NetApp host groups"}

var (
	flagNetAppHGLocation   string
	flagNetAppHGConfigFile string
	flagNetAppHGUpdateMask string
	flagNetAppHGFormat     string
	flagNetAppHGFilter     string
	flagNetAppHGPageSize   int64
)

var (
	netappHGCreateCmd = &cobra.Command{
		Use: "create HOST_GROUP", Short: "Create a host group",
		Args: cobra.ExactArgs(1), RunE: runNetAppHGCreate,
	}
	netappHGDeleteCmd = &cobra.Command{
		Use: "delete HOST_GROUP", Short: "Delete a host group",
		Args: cobra.ExactArgs(1), RunE: runNetAppHGDelete,
	}
	netappHGDescribeCmd = &cobra.Command{
		Use: "describe HOST_GROUP", Short: "Describe a host group",
		Args: cobra.ExactArgs(1), RunE: runNetAppHGDescribe,
	}
	netappHGListCmd = &cobra.Command{
		Use: "list", Short: "List host groups",
		Args: cobra.NoArgs, RunE: runNetAppHGList,
	}
	netappHGUpdateCmd = &cobra.Command{
		Use: "update HOST_GROUP", Short: "Update a host group",
		Args: cobra.ExactArgs(1), RunE: runNetAppHGUpdate,
	}
)

func init() {
	all := []*cobra.Command{netappHGCreateCmd, netappHGDeleteCmd, netappHGDescribeCmd, netappHGListCmd, netappHGUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagNetAppHGLocation, "location", "", "Location for the host group (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNetAppHGFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{netappHGCreateCmd, netappHGUpdateCmd} {
		c.Flags().StringVar(&flagNetAppHGConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the HostGroup body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	netappHGUpdateCmd.Flags().StringVar(&flagNetAppHGUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	netappHGListCmd.Flags().StringVar(&flagNetAppHGFilter, "filter", "", "Server-side filter expression")
	netappHGListCmd.Flags().Int64Var(&flagNetAppHGPageSize, "page-size", 0, "Maximum number of results per page")

	netappHGCmd.AddCommand(all...)
	netappCmd.AddCommand(netappHGCmd)
}

func netappHGParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return netappLocationParent(project, flagNetAppHGLocation), nil
}

func netappHGName(id string) (string, error) {
	parent, err := netappHGParent()
	if err != nil {
		return "", err
	}
	return netappChild("hostGroups", id, parent), nil
}

func runNetAppHGCreate(cmd *cobra.Command, args []string) error {
	parent, err := netappHGParent()
	if err != nil {
		return err
	}
	body := &netapp.HostGroup{}
	if err := loadYAMLOrJSONInto(flagNetAppHGConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.HostGroups.Create(parent, body).HostGroupId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating host group: %w", err)
	}
	fmt.Printf("Create request issued for host group [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppHGFormat)
}

func runNetAppHGDelete(cmd *cobra.Command, args []string) error {
	name, err := netappHGName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.HostGroups.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting host group: %w", err)
	}
	fmt.Printf("Delete request issued for host group [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppHGFormat)
}

func runNetAppHGDescribe(cmd *cobra.Command, args []string) error {
	name, err := netappHGName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.HostGroups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing host group: %w", err)
	}
	return emitFormatted(got, flagNetAppHGFormat)
}

func runNetAppHGList(cmd *cobra.Command, args []string) error {
	parent, err := netappHGParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*netapp.HostGroup
	pageToken := ""
	for {
		call := svc.Projects.Locations.HostGroups.List(parent).Context(ctx)
		if flagNetAppHGFilter != "" {
			call = call.Filter(flagNetAppHGFilter)
		}
		if flagNetAppHGPageSize > 0 {
			call = call.PageSize(flagNetAppHGPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing host groups: %w", err)
		}
		all = append(all, resp.HostGroups...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNetAppHGFormat)
}

func runNetAppHGUpdate(cmd *cobra.Command, args []string) error {
	name, err := netappHGName(args[0])
	if err != nil {
		return err
	}
	body := &netapp.HostGroup{}
	if err := loadYAMLOrJSONInto(flagNetAppHGConfigFile, body); err != nil {
		return err
	}
	mask := flagNetAppHGUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.HostGroups.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating host group: %w", err)
	}
	fmt.Printf("Update request issued for host group [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppHGFormat)
}
