package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	baremetalsolution "google.golang.org/api/baremetalsolution/v2"
)

// --- gcloud bms networks (#1227) ---

var bmsNetworksCmd = &cobra.Command{Use: "networks", Short: "Manage bare metal networks"}

var (
	flagBmsNetLocation   string
	flagBmsNetFormat     string
	flagBmsNetConfigFile string
	flagBmsNetUpdateMask string
	flagBmsNetNewName    string
	flagBmsNetPageSize   int64
)

var (
	bmsNetDescribeCmd = &cobra.Command{
		Use: "describe NETWORK", Short: "Describe a bare metal network",
		Args: cobra.ExactArgs(1), RunE: runBmsNetDescribe,
	}
	bmsNetListCmd = &cobra.Command{
		Use: "list", Short: "List bare metal networks in a location",
		Args: cobra.NoArgs, RunE: runBmsNetList,
	}
	bmsNetUpdateCmd = &cobra.Command{
		Use: "update NETWORK", Short: "Update a bare metal network",
		Args: cobra.ExactArgs(1), RunE: runBmsNetUpdate,
	}
	bmsNetRenameCmd = &cobra.Command{
		Use: "rename NETWORK", Short: "Rename a bare metal network",
		Args: cobra.ExactArgs(1), RunE: runBmsNetRename,
	}
)

func init() {
	all := []*cobra.Command{bmsNetDescribeCmd, bmsNetListCmd, bmsNetUpdateCmd, bmsNetRenameCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagBmsNetLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagBmsNetFormat, "format", "", "Output format")
	}
	bmsNetListCmd.Flags().Int64Var(&flagBmsNetPageSize, "page-size", 0, "Maximum results per page")
	bmsNetUpdateCmd.Flags().StringVar(&flagBmsNetConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = bmsNetUpdateCmd.MarkFlagRequired("config-file")
	bmsNetUpdateCmd.Flags().StringVar(&flagBmsNetUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	bmsNetRenameCmd.Flags().StringVar(&flagBmsNetNewName, "new-name", "", "New network id (required)")
	_ = bmsNetRenameCmd.MarkFlagRequired("new-name")

	bmsNetworksCmd.AddCommand(all...)
	bmsCmd.AddCommand(bmsNetworksCmd)
}

func bmsNetName(id string) (string, error) {
	return bmsResource(flagBmsNetLocation, "networks", id)
}

func runBmsNetDescribe(cmd *cobra.Command, args []string) error {
	name, err := bmsNetName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Networks.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing network: %w", err)
	}
	return emitFormatted(got, flagBmsNetFormat)
}

func runBmsNetList(cmd *cobra.Command, args []string) error {
	parent, err := bmsLocationParent(flagBmsNetLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*baremetalsolution.Network
	pageToken := ""
	for {
		call := svc.Projects.Locations.Networks.List(parent).Context(ctx)
		if flagBmsNetPageSize > 0 {
			call = call.PageSize(flagBmsNetPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing networks: %w", err)
		}
		all = append(all, resp.Networks...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagBmsNetFormat)
}

func runBmsNetUpdate(cmd *cobra.Command, args []string) error {
	name, err := bmsNetName(args[0])
	if err != nil {
		return err
	}
	body := &baremetalsolution.Network{}
	if err := loadYAMLOrJSONInto(flagBmsNetConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagBmsNetUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Networks.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating network: %w", err)
	}
	fmt.Printf("Update network [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBmsNetFormat)
}

func runBmsNetRename(cmd *cobra.Command, args []string) error {
	name, err := bmsNetName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Networks.Rename(name, &baremetalsolution.RenameNetworkRequest{NewNetworkId: flagBmsNetNewName}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("renaming network: %w", err)
	}
	fmt.Printf("Renamed network [%s] to [%s].\n", args[0], flagBmsNetNewName)
	return emitFormatted(got, flagBmsNetFormat)
}
