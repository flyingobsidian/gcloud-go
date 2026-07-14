package cmd

import (
	"context"
	"fmt"
	"path"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	firestore "google.golang.org/api/firestore/v1"
)

var firestoreLocationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "Explore Cloud Firestore locations",
}

var fsLocListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Firestore-enabled locations for the project",
	Args:  cobra.NoArgs,
	RunE:  runFSLocList,
}

var fsLocDescribeCmd = &cobra.Command{
	Use:   "describe LOCATION",
	Short: "Describe a Firestore location",
	Args:  cobra.ExactArgs(1),
	RunE:  runFSLocDescribe,
}

var (
	flagFSLocFormat   string
	flagFSLocPageSize int64
	flagFSLocLimit    int64
)

func init() {
	fsLocListCmd.Flags().StringVar(&flagFSLocFormat, "format", "", "Output format")
	fsLocListCmd.Flags().Int64Var(&flagFSLocPageSize, "page-size", 0, "Page size")
	fsLocListCmd.Flags().Int64Var(&flagFSLocLimit, "limit", 0, "Cap total results (0 = no cap)")
	fsLocDescribeCmd.Flags().StringVar(&flagFSLocFormat, "format", "", "Output format")

	firestoreLocationsCmd.AddCommand(fsLocDescribeCmd, fsLocListCmd)
	firestoreCmd.AddCommand(firestoreLocationsCmd)
}

func runFSLocList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*firestore.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if flagFSLocPageSize > 0 {
			call = call.PageSize(flagFSLocPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing locations: %w", err)
		}
		all = append(all, resp.Locations...)
		if flagFSLocLimit > 0 && int64(len(all)) >= flagFSLocLimit {
			all = all[:flagFSLocLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagFSLocFormat != "" {
		return emitFormatted(all, flagFSLocFormat)
	}
	fmt.Printf("%-20s %s\n", "LOCATION", "DISPLAY_NAME")
	for _, l := range all {
		fmt.Printf("%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

func runFSLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	loc, err := svc.Projects.Locations.Get(firestoreLocationName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	if flagFSLocFormat != "" {
		return emitFormatted(loc, flagFSLocFormat)
	}
	fmt.Printf("%s (%s)\n", path.Base(loc.Name), loc.DisplayName)
	return nil
}
