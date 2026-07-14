package cmd

import (
	"context"
	"fmt"
	"path"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	eventarc "google.golang.org/api/eventarc/v1"
)

var eventarcLocationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "Explore Eventarc locations",
}

var eventarcLocationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Eventarc-enabled locations for the project",
	Args:  cobra.NoArgs,
	RunE:  runEventarcLocationsList,
}

var eventarcProvidersCmd = &cobra.Command{
	Use:   "providers",
	Short: "Explore Eventarc event providers",
}

var evProvDescribeCmd = &cobra.Command{
	Use:   "describe PROVIDER",
	Short: "Describe an event provider",
	Args:  cobra.ExactArgs(1),
	RunE:  runEvProvDescribe,
}

var evProvListCmd = &cobra.Command{
	Use:   "list",
	Short: "List event providers in a location",
	Args:  cobra.NoArgs,
	RunE:  runEvProvList,
}

var (
	flagEvLocFormat     string
	flagEvLocListLimit  int64
	flagEvLocListPage   int64

	flagEvProvLocation   string
	flagEvProvFormat     string
	flagEvProvListLimit  int64
	flagEvProvListPage   int64
	flagEvProvListFilter string
	flagEvProvListURI    bool
)

func init() {
	eventarcLocationsListCmd.Flags().StringVar(&flagEvLocFormat, "format", "", "Output format")
	eventarcLocationsListCmd.Flags().Int64Var(&flagEvLocListPage, "page-size", 0, "Page size")
	eventarcLocationsListCmd.Flags().Int64Var(&flagEvLocListLimit, "limit", 0, "Cap total results (0 = no cap)")
	eventarcLocationsCmd.AddCommand(eventarcLocationsListCmd)
	eventarcCmd.AddCommand(eventarcLocationsCmd)

	for _, c := range []*cobra.Command{evProvDescribeCmd, evProvListCmd} {
		eventarcAddRegionFlag(c, &flagEvProvLocation, true)
	}
	evProvDescribeCmd.Flags().StringVar(&flagEvProvFormat, "format", "", "Output format")
	evProvListCmd.Flags().StringVar(&flagEvProvFormat, "format", "", "Output format")
	evProvListCmd.Flags().Int64Var(&flagEvProvListPage, "page-size", 0, "Page size")
	evProvListCmd.Flags().Int64Var(&flagEvProvListLimit, "limit", 0, "Cap total results (0 = no cap)")
	evProvListCmd.Flags().StringVar(&flagEvProvListFilter, "filter", "", "Server-side filter expression")
	evProvListCmd.Flags().BoolVar(&flagEvProvListURI, "uri", false, "Print resource names only")
	eventarcProvidersCmd.AddCommand(evProvDescribeCmd, evProvListCmd)
	eventarcCmd.AddCommand(eventarcProvidersCmd)
}

func runEventarcLocationsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*eventarc.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if flagEvLocListPage > 0 {
			call = call.PageSize(flagEvLocListPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing locations: %w", err)
		}
		all = append(all, resp.Locations...)
		if flagEvLocListLimit > 0 && int64(len(all)) >= flagEvLocListLimit {
			all = all[:flagEvLocListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagEvLocFormat != "" {
		return emitFormatted(all, flagEvLocFormat)
	}
	fmt.Printf("%-20s %s\n", "LOCATION", "DISPLAY_NAME")
	for _, l := range all {
		fmt.Printf("%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

func runEvProvDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	p, err := svc.Projects.Locations.Providers.Get(eventarcResourceName("providers", args[0], project, flagEvProvLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing provider: %w", err)
	}
	return emitFormatted(p, flagEvProvFormat)
}

func runEvProvList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := eventarcLocationParent(project, flagEvProvLocation)
	var all []*eventarc.Provider
	pageToken := ""
	for {
		call := svc.Projects.Locations.Providers.List(parent).Context(ctx)
		if flagEvProvListFilter != "" {
			call = call.Filter(flagEvProvListFilter)
		}
		if flagEvProvListPage > 0 {
			call = call.PageSize(flagEvProvListPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing providers: %w", err)
		}
		all = append(all, resp.Providers...)
		if flagEvProvListLimit > 0 && int64(len(all)) >= flagEvProvListLimit {
			all = all[:flagEvProvListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagEvProvListURI {
		for _, p := range all {
			fmt.Println(p.Name)
		}
		return nil
	}
	if flagEvProvFormat != "" {
		return emitFormatted(all, flagEvProvFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, p := range all {
		fmt.Printf("%-40s %s\n", path.Base(p.Name), p.DisplayName)
	}
	return nil
}
