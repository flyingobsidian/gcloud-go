package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	netapp "google.golang.org/api/netapp/v1"
)

// --- gcloud netapp active-directories (#1197) ---

var netappADCmd = &cobra.Command{
	Use:   "active-directories",
	Short: "Manage NetApp Active Directories",
}

var (
	flagNetAppADLocation   string
	flagNetAppADConfigFile string
	flagNetAppADUpdateMask string
	flagNetAppADFormat     string
	flagNetAppADFilter     string
	flagNetAppADPageSize   int64
)

var (
	netappADCreateCmd = &cobra.Command{
		Use: "create ACTIVE_DIRECTORY", Short: "Create an Active Directory",
		Args: cobra.ExactArgs(1), RunE: runNetAppADCreate,
	}
	netappADDeleteCmd = &cobra.Command{
		Use: "delete ACTIVE_DIRECTORY", Short: "Delete an Active Directory",
		Args: cobra.ExactArgs(1), RunE: runNetAppADDelete,
	}
	netappADDescribeCmd = &cobra.Command{
		Use: "describe ACTIVE_DIRECTORY", Short: "Describe an Active Directory",
		Args: cobra.ExactArgs(1), RunE: runNetAppADDescribe,
	}
	netappADListCmd = &cobra.Command{
		Use: "list", Short: "List Active Directories",
		Args: cobra.NoArgs, RunE: runNetAppADList,
	}
	netappADUpdateCmd = &cobra.Command{
		Use: "update ACTIVE_DIRECTORY", Short: "Update an Active Directory",
		Args: cobra.ExactArgs(1), RunE: runNetAppADUpdate,
	}
)

func init() {
	all := []*cobra.Command{netappADCreateCmd, netappADDeleteCmd, netappADDescribeCmd, netappADListCmd, netappADUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagNetAppADLocation, "location", "", "Location for the Active Directory (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNetAppADFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{netappADCreateCmd, netappADUpdateCmd} {
		c.Flags().StringVar(&flagNetAppADConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the ActiveDirectory body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	netappADUpdateCmd.Flags().StringVar(&flagNetAppADUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	netappADListCmd.Flags().StringVar(&flagNetAppADFilter, "filter", "", "Server-side filter expression")
	netappADListCmd.Flags().Int64Var(&flagNetAppADPageSize, "page-size", 0, "Maximum number of results per page")

	netappADCmd.AddCommand(all...)
	netappCmd.AddCommand(netappADCmd)
}

func netappADParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return netappLocationParent(project, flagNetAppADLocation), nil
}

func netappADName(id string) (string, error) {
	parent, err := netappADParent()
	if err != nil {
		return "", err
	}
	return netappChild("activeDirectories", id, parent), nil
}

func runNetAppADCreate(cmd *cobra.Command, args []string) error {
	parent, err := netappADParent()
	if err != nil {
		return err
	}
	body := &netapp.ActiveDirectory{}
	if err := loadYAMLOrJSONInto(flagNetAppADConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ActiveDirectories.Create(parent, body).ActiveDirectoryId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating active directory: %w", err)
	}
	fmt.Printf("Create request issued for active directory [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppADFormat)
}

func runNetAppADDelete(cmd *cobra.Command, args []string) error {
	name, err := netappADName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.ActiveDirectories.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting active directory: %w", err)
	}
	fmt.Printf("Delete request issued for active directory [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppADFormat)
}

func runNetAppADDescribe(cmd *cobra.Command, args []string) error {
	name, err := netappADName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ActiveDirectories.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing active directory: %w", err)
	}
	return emitFormatted(got, flagNetAppADFormat)
}

func runNetAppADList(cmd *cobra.Command, args []string) error {
	parent, err := netappADParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*netapp.ActiveDirectory
	pageToken := ""
	for {
		call := svc.Projects.Locations.ActiveDirectories.List(parent).Context(ctx)
		if flagNetAppADFilter != "" {
			call = call.Filter(flagNetAppADFilter)
		}
		if flagNetAppADPageSize > 0 {
			call = call.PageSize(flagNetAppADPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing active directories: %w", err)
		}
		all = append(all, resp.ActiveDirectories...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNetAppADFormat)
}

func runNetAppADUpdate(cmd *cobra.Command, args []string) error {
	name, err := netappADName(args[0])
	if err != nil {
		return err
	}
	body := &netapp.ActiveDirectory{}
	if err := loadYAMLOrJSONInto(flagNetAppADConfigFile, body); err != nil {
		return err
	}
	mask := flagNetAppADUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.ActiveDirectories.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating active directory: %w", err)
	}
	fmt.Printf("Update request issued for active directory [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppADFormat)
}
