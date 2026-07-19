package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	baremetalsolution "google.golang.org/api/baremetalsolution/v2"
)

// --- gcloud bms nfs-shares (#1228) ---

var bmsNfsSharesCmd = &cobra.Command{Use: "nfs-shares", Short: "Manage bare metal NFS shares"}

var (
	flagBmsNfsLocation   string
	flagBmsNfsFormat     string
	flagBmsNfsConfigFile string
	flagBmsNfsUpdateMask string
	flagBmsNfsNewName    string
	flagBmsNfsPageSize   int64
)

var (
	bmsNfsDescribeCmd = &cobra.Command{
		Use: "describe NFS_SHARE", Short: "Describe an NFS share",
		Args: cobra.ExactArgs(1), RunE: runBmsNfsDescribe,
	}
	bmsNfsListCmd = &cobra.Command{
		Use: "list", Short: "List NFS shares in a location",
		Args: cobra.NoArgs, RunE: runBmsNfsList,
	}
	bmsNfsUpdateCmd = &cobra.Command{
		Use: "update NFS_SHARE", Short: "Update an NFS share",
		Args: cobra.ExactArgs(1), RunE: runBmsNfsUpdate,
	}
	bmsNfsRenameCmd = &cobra.Command{
		Use: "rename NFS_SHARE", Short: "Rename an NFS share",
		Args: cobra.ExactArgs(1), RunE: runBmsNfsRename,
	}
)

func init() {
	all := []*cobra.Command{bmsNfsDescribeCmd, bmsNfsListCmd, bmsNfsUpdateCmd, bmsNfsRenameCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagBmsNfsLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagBmsNfsFormat, "format", "", "Output format")
	}
	bmsNfsListCmd.Flags().Int64Var(&flagBmsNfsPageSize, "page-size", 0, "Maximum results per page")
	bmsNfsUpdateCmd.Flags().StringVar(&flagBmsNfsConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = bmsNfsUpdateCmd.MarkFlagRequired("config-file")
	bmsNfsUpdateCmd.Flags().StringVar(&flagBmsNfsUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	bmsNfsRenameCmd.Flags().StringVar(&flagBmsNfsNewName, "new-name", "", "New NFS share id (required)")
	_ = bmsNfsRenameCmd.MarkFlagRequired("new-name")

	bmsNfsSharesCmd.AddCommand(all...)
	bmsCmd.AddCommand(bmsNfsSharesCmd)
}

func bmsNfsName(id string) (string, error) {
	return bmsResource(flagBmsNfsLocation, "nfsShares", id)
}

func runBmsNfsDescribe(cmd *cobra.Command, args []string) error {
	name, err := bmsNfsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.NfsShares.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing nfs share: %w", err)
	}
	return emitFormatted(got, flagBmsNfsFormat)
}

func runBmsNfsList(cmd *cobra.Command, args []string) error {
	parent, err := bmsLocationParent(flagBmsNfsLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*baremetalsolution.NfsShare
	pageToken := ""
	for {
		call := svc.Projects.Locations.NfsShares.List(parent).Context(ctx)
		if flagBmsNfsPageSize > 0 {
			call = call.PageSize(flagBmsNfsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing nfs shares: %w", err)
		}
		all = append(all, resp.NfsShares...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagBmsNfsFormat)
}

func runBmsNfsUpdate(cmd *cobra.Command, args []string) error {
	name, err := bmsNfsName(args[0])
	if err != nil {
		return err
	}
	body := &baremetalsolution.NfsShare{}
	if err := loadYAMLOrJSONInto(flagBmsNfsConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagBmsNfsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.NfsShares.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating nfs share: %w", err)
	}
	fmt.Printf("Update nfs share [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBmsNfsFormat)
}

func runBmsNfsRename(cmd *cobra.Command, args []string) error {
	name, err := bmsNfsName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.NfsShares.Rename(name, &baremetalsolution.RenameNfsShareRequest{NewNfsshareId: flagBmsNfsNewName}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("renaming nfs share: %w", err)
	}
	fmt.Printf("Renamed nfs share [%s] to [%s].\n", args[0], flagBmsNfsNewName)
	return emitFormatted(got, flagBmsNfsFormat)
}
