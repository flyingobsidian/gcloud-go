package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	baremetalsolution "google.golang.org/api/baremetalsolution/v2"
)

// --- gcloud bms volumes (#1232) ---

var bmsVolumesCmd = &cobra.Command{Use: "volumes", Short: "Manage bare metal volumes"}

var (
	flagBmsVolLocation   string
	flagBmsVolFormat     string
	flagBmsVolConfigFile string
	flagBmsVolUpdateMask string
	flagBmsVolNewName    string
	flagBmsVolSizeGib    int64
	flagBmsVolPageSize   int64
)

var (
	bmsVolDescribeCmd = &cobra.Command{
		Use: "describe VOLUME", Short: "Describe a bare metal volume",
		Args: cobra.ExactArgs(1), RunE: runBmsVolDescribe,
	}
	bmsVolListCmd = &cobra.Command{
		Use: "list", Short: "List bare metal volumes in a location",
		Args: cobra.NoArgs, RunE: runBmsVolList,
	}
	bmsVolUpdateCmd = &cobra.Command{
		Use: "update VOLUME", Short: "Update a bare metal volume",
		Args: cobra.ExactArgs(1), RunE: runBmsVolUpdate,
	}
	bmsVolRenameCmd = &cobra.Command{
		Use: "rename VOLUME", Short: "Rename a bare metal volume",
		Args: cobra.ExactArgs(1), RunE: runBmsVolRename,
	}
	bmsVolResizeCmd = &cobra.Command{
		Use: "resize VOLUME", Short: "Resize a bare metal volume",
		Args: cobra.ExactArgs(1), RunE: runBmsVolResize,
	}
)

func init() {
	all := []*cobra.Command{bmsVolDescribeCmd, bmsVolListCmd, bmsVolUpdateCmd, bmsVolRenameCmd, bmsVolResizeCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagBmsVolLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagBmsVolFormat, "format", "", "Output format")
	}
	bmsVolListCmd.Flags().Int64Var(&flagBmsVolPageSize, "page-size", 0, "Maximum results per page")
	bmsVolUpdateCmd.Flags().StringVar(&flagBmsVolConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = bmsVolUpdateCmd.MarkFlagRequired("config-file")
	bmsVolUpdateCmd.Flags().StringVar(&flagBmsVolUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	bmsVolRenameCmd.Flags().StringVar(&flagBmsVolNewName, "new-name", "", "New volume id (required)")
	_ = bmsVolRenameCmd.MarkFlagRequired("new-name")
	bmsVolResizeCmd.Flags().Int64Var(&flagBmsVolSizeGib, "size-gib", 0, "New size in GiB (required)")
	_ = bmsVolResizeCmd.MarkFlagRequired("size-gib")

	bmsVolumesCmd.AddCommand(all...)
	bmsCmd.AddCommand(bmsVolumesCmd)
}

func bmsVolName(id string) (string, error) {
	return bmsResource(flagBmsVolLocation, "volumes", id)
}

func runBmsVolDescribe(cmd *cobra.Command, args []string) error {
	name, err := bmsVolName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Volumes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing volume: %w", err)
	}
	return emitFormatted(got, flagBmsVolFormat)
}

func runBmsVolList(cmd *cobra.Command, args []string) error {
	parent, err := bmsLocationParent(flagBmsVolLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*baremetalsolution.Volume
	pageToken := ""
	for {
		call := svc.Projects.Locations.Volumes.List(parent).Context(ctx)
		if flagBmsVolPageSize > 0 {
			call = call.PageSize(flagBmsVolPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing volumes: %w", err)
		}
		all = append(all, resp.Volumes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagBmsVolFormat)
}

func runBmsVolUpdate(cmd *cobra.Command, args []string) error {
	name, err := bmsVolName(args[0])
	if err != nil {
		return err
	}
	body := &baremetalsolution.Volume{}
	if err := loadYAMLOrJSONInto(flagBmsVolConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagBmsVolUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Volumes.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating volume: %w", err)
	}
	fmt.Printf("Update volume [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBmsVolFormat)
}

func runBmsVolRename(cmd *cobra.Command, args []string) error {
	name, err := bmsVolName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Volumes.Rename(name, &baremetalsolution.RenameVolumeRequest{NewVolumeId: flagBmsVolNewName}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("renaming volume: %w", err)
	}
	fmt.Printf("Renamed volume [%s] to [%s].\n", args[0], flagBmsVolNewName)
	return emitFormatted(got, flagBmsVolFormat)
}

func runBmsVolResize(cmd *cobra.Command, args []string) error {
	name, err := bmsVolName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Volumes.Resize(name, &baremetalsolution.ResizeVolumeRequest{SizeGib: flagBmsVolSizeGib}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resizing volume: %w", err)
	}
	fmt.Printf("Resize volume [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBmsVolFormat)
}
