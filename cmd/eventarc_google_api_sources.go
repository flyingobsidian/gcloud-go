package cmd

import (
	"context"
	"fmt"
	"path"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	eventarc "google.golang.org/api/eventarc/v1"
)

var eventarcGASCmd = &cobra.Command{
	Use:   "google-api-sources",
	Short: "Manage Eventarc Google API sources",
}

var (
	evGASCreateCmd = &cobra.Command{
		Use: "create GOOGLE_API_SOURCE", Short: "Create a Google API source from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runEvGASCreate,
	}
	evGASDeleteCmd = &cobra.Command{
		Use: "delete GOOGLE_API_SOURCE", Short: "Delete a Google API source",
		Args: cobra.ExactArgs(1), RunE: runEvGASDelete,
	}
	evGASDescribeCmd = &cobra.Command{
		Use: "describe GOOGLE_API_SOURCE", Short: "Describe a Google API source",
		Args: cobra.ExactArgs(1), RunE: runEvGASDescribe,
	}
	evGASListCmd = &cobra.Command{
		Use: "list", Short: "List Google API sources in a location",
		Args: cobra.NoArgs, RunE: runEvGASList,
	}
	evGASUpdateCmd = &cobra.Command{
		Use: "update GOOGLE_API_SOURCE", Short: "Update a Google API source from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runEvGASUpdate,
	}
)

var (
	flagEvGASLocation   string
	flagEvGASConfigFile string
	flagEvGASUpdateMask string
	flagEvGASFormat     string
	flagEvGASAsync      bool
	flagEvGASListLimit  int64
	flagEvGASListPage   int64
	flagEvGASListFilter string
	flagEvGASListURI    bool
)

func init() {
	for _, c := range []*cobra.Command{evGASCreateCmd, evGASDeleteCmd, evGASDescribeCmd, evGASListCmd, evGASUpdateCmd} {
		eventarcAddRegionFlag(c, &flagEvGASLocation, true)
	}
	for _, c := range []*cobra.Command{evGASCreateCmd, evGASUpdateCmd} {
		c.Flags().StringVar(&flagEvGASConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the GoogleApiSource message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	evGASUpdateCmd.Flags().StringVar(&flagEvGASUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{evGASCreateCmd, evGASDeleteCmd, evGASUpdateCmd} {
		c.Flags().BoolVar(&flagEvGASAsync, "async", false, "Return the long-running operation without waiting")
	}
	evGASDescribeCmd.Flags().StringVar(&flagEvGASFormat, "format", "", "Output format")
	evGASListCmd.Flags().StringVar(&flagEvGASFormat, "format", "", "Output format")
	evGASListCmd.Flags().Int64Var(&flagEvGASListPage, "page-size", 0, "Page size")
	evGASListCmd.Flags().Int64Var(&flagEvGASListLimit, "limit", 0, "Cap total results (0 = no cap)")
	evGASListCmd.Flags().StringVar(&flagEvGASListFilter, "filter", "", "Server-side filter expression")
	evGASListCmd.Flags().BoolVar(&flagEvGASListURI, "uri", false, "Print resource names only")

	eventarcGASCmd.AddCommand(evGASCreateCmd, evGASDeleteCmd, evGASDescribeCmd, evGASListCmd, evGASUpdateCmd)
	eventarcCmd.AddCommand(eventarcGASCmd)
}

func runEvGASCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	g := &eventarc.GoogleApiSource{}
	if err := loadYAMLOrJSONInto(flagEvGASConfigFile, g); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoogleApiSources.Create(eventarcLocationParent(project, flagEvGASLocation), g).
		GoogleApiSourceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating Google API source: %w", err)
	}
	return eventarcFinishOp(ctx, svc, op, "Create Google API source", args[0], flagEvGASAsync)
}

func runEvGASDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoogleApiSources.Delete(eventarcResourceName("googleApiSources", args[0], project, flagEvGASLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting Google API source: %w", err)
	}
	return eventarcFinishOp(ctx, svc, op, "Delete Google API source", args[0], flagEvGASAsync)
}

func runEvGASDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	g, err := svc.Projects.Locations.GoogleApiSources.Get(eventarcResourceName("googleApiSources", args[0], project, flagEvGASLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing Google API source: %w", err)
	}
	return emitFormatted(g, flagEvGASFormat)
}

func runEvGASList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := eventarcLocationParent(project, flagEvGASLocation)
	var all []*eventarc.GoogleApiSource
	pageToken := ""
	for {
		call := svc.Projects.Locations.GoogleApiSources.List(parent).Context(ctx)
		if flagEvGASListFilter != "" {
			call = call.Filter(flagEvGASListFilter)
		}
		if flagEvGASListPage > 0 {
			call = call.PageSize(flagEvGASListPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing Google API sources: %w", err)
		}
		all = append(all, resp.GoogleApiSources...)
		if flagEvGASListLimit > 0 && int64(len(all)) >= flagEvGASListLimit {
			all = all[:flagEvGASListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagEvGASListURI {
		for _, g := range all {
			fmt.Println(g.Name)
		}
		return nil
	}
	if flagEvGASFormat != "" {
		return emitFormatted(all, flagEvGASFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DESTINATION")
	for _, g := range all {
		fmt.Printf("%-40s %s\n", path.Base(g.Name), g.Destination)
	}
	return nil
}

func runEvGASUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	g := &eventarc.GoogleApiSource{}
	if err := loadYAMLOrJSONInto(flagEvGASConfigFile, g); err != nil {
		return err
	}
	mask := flagEvGASUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(g))
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.GoogleApiSources.Patch(eventarcResourceName("googleApiSources", args[0], project, flagEvGASLocation), g).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating Google API source: %w", err)
	}
	return eventarcFinishOp(ctx, svc, op, "Update Google API source", args[0], flagEvGASAsync)
}
