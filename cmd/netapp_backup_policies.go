package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	netapp "google.golang.org/api/netapp/v1"
)

// --- gcloud netapp backup-policies (#1198) ---

var netappBPCmd = &cobra.Command{Use: "backup-policies", Short: "Manage NetApp backup policies"}

var (
	flagNetAppBPLocation   string
	flagNetAppBPConfigFile string
	flagNetAppBPUpdateMask string
	flagNetAppBPFormat     string
	flagNetAppBPFilter     string
	flagNetAppBPPageSize   int64
)

var (
	netappBPCreateCmd = &cobra.Command{
		Use: "create BACKUP_POLICY", Short: "Create a backup policy",
		Args: cobra.ExactArgs(1), RunE: runNetAppBPCreate,
	}
	netappBPDeleteCmd = &cobra.Command{
		Use: "delete BACKUP_POLICY", Short: "Delete a backup policy",
		Args: cobra.ExactArgs(1), RunE: runNetAppBPDelete,
	}
	netappBPDescribeCmd = &cobra.Command{
		Use: "describe BACKUP_POLICY", Short: "Describe a backup policy",
		Args: cobra.ExactArgs(1), RunE: runNetAppBPDescribe,
	}
	netappBPListCmd = &cobra.Command{
		Use: "list", Short: "List backup policies",
		Args: cobra.NoArgs, RunE: runNetAppBPList,
	}
	netappBPUpdateCmd = &cobra.Command{
		Use: "update BACKUP_POLICY", Short: "Update a backup policy",
		Args: cobra.ExactArgs(1), RunE: runNetAppBPUpdate,
	}
)

func init() {
	all := []*cobra.Command{netappBPCreateCmd, netappBPDeleteCmd, netappBPDescribeCmd, netappBPListCmd, netappBPUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagNetAppBPLocation, "location", "", "Location for the backup policy (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagNetAppBPFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{netappBPCreateCmd, netappBPUpdateCmd} {
		c.Flags().StringVar(&flagNetAppBPConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the BackupPolicy body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	netappBPUpdateCmd.Flags().StringVar(&flagNetAppBPUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	netappBPListCmd.Flags().StringVar(&flagNetAppBPFilter, "filter", "", "Server-side filter expression")
	netappBPListCmd.Flags().Int64Var(&flagNetAppBPPageSize, "page-size", 0, "Maximum number of results per page")

	netappBPCmd.AddCommand(all...)
	netappCmd.AddCommand(netappBPCmd)
}

func netappBPParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return netappLocationParent(project, flagNetAppBPLocation), nil
}

func netappBPName(id string) (string, error) {
	parent, err := netappBPParent()
	if err != nil {
		return "", err
	}
	return netappChild("backupPolicies", id, parent), nil
}

func runNetAppBPCreate(cmd *cobra.Command, args []string) error {
	parent, err := netappBPParent()
	if err != nil {
		return err
	}
	body := &netapp.BackupPolicy{}
	if err := loadYAMLOrJSONInto(flagNetAppBPConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.BackupPolicies.Create(parent, body).BackupPolicyId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup policy: %w", err)
	}
	fmt.Printf("Create request issued for backup policy [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppBPFormat)
}

func runNetAppBPDelete(cmd *cobra.Command, args []string) error {
	name, err := netappBPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.BackupPolicies.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting backup policy: %w", err)
	}
	fmt.Printf("Delete request issued for backup policy [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppBPFormat)
}

func runNetAppBPDescribe(cmd *cobra.Command, args []string) error {
	name, err := netappBPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.BackupPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup policy: %w", err)
	}
	return emitFormatted(got, flagNetAppBPFormat)
}

func runNetAppBPList(cmd *cobra.Command, args []string) error {
	parent, err := netappBPParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*netapp.BackupPolicy
	pageToken := ""
	for {
		call := svc.Projects.Locations.BackupPolicies.List(parent).Context(ctx)
		if flagNetAppBPFilter != "" {
			call = call.Filter(flagNetAppBPFilter)
		}
		if flagNetAppBPPageSize > 0 {
			call = call.PageSize(flagNetAppBPPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing backup policies: %w", err)
		}
		all = append(all, resp.BackupPolicies...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagNetAppBPFormat)
}

func runNetAppBPUpdate(cmd *cobra.Command, args []string) error {
	name, err := netappBPName(args[0])
	if err != nil {
		return err
	}
	body := &netapp.BackupPolicy{}
	if err := loadYAMLOrJSONInto(flagNetAppBPConfigFile, body); err != nil {
		return err
	}
	mask := flagNetAppBPUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.NetAppService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.BackupPolicies.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating backup policy: %w", err)
	}
	fmt.Printf("Update request issued for backup policy [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagNetAppBPFormat)
}
