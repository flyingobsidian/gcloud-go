package cmd

import (
	"context"
	"fmt"
	"path"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	eventarc "google.golang.org/api/eventarc/v1"
)

var eventarcPipelinesCmd = &cobra.Command{
	Use:   "pipelines",
	Short: "Manage Eventarc pipelines",
}

var (
	evPipCreateCmd = &cobra.Command{
		Use: "create PIPELINE", Short: "Create a pipeline from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runEvPipCreate,
	}
	evPipDeleteCmd = &cobra.Command{
		Use: "delete PIPELINE", Short: "Delete a pipeline",
		Args: cobra.ExactArgs(1), RunE: runEvPipDelete,
	}
	evPipDescribeCmd = &cobra.Command{
		Use: "describe PIPELINE", Short: "Describe a pipeline",
		Args: cobra.ExactArgs(1), RunE: runEvPipDescribe,
	}
	evPipListCmd = &cobra.Command{
		Use: "list", Short: "List pipelines in a location",
		Args: cobra.NoArgs, RunE: runEvPipList,
	}
	evPipUpdateCmd = &cobra.Command{
		Use: "update PIPELINE", Short: "Update a pipeline from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runEvPipUpdate,
	}
)

var (
	flagEvPipLocation   string
	flagEvPipConfigFile string
	flagEvPipUpdateMask string
	flagEvPipFormat     string
	flagEvPipAsync      bool
	flagEvPipListLimit  int64
	flagEvPipListPage   int64
	flagEvPipListFilter string
	flagEvPipListURI    bool
)

func init() {
	for _, c := range []*cobra.Command{evPipCreateCmd, evPipDeleteCmd, evPipDescribeCmd, evPipListCmd, evPipUpdateCmd} {
		eventarcAddRegionFlag(c, &flagEvPipLocation, true)
	}
	for _, c := range []*cobra.Command{evPipCreateCmd, evPipUpdateCmd} {
		c.Flags().StringVar(&flagEvPipConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Pipeline message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	evPipUpdateCmd.Flags().StringVar(&flagEvPipUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{evPipCreateCmd, evPipDeleteCmd, evPipUpdateCmd} {
		c.Flags().BoolVar(&flagEvPipAsync, "async", false, "Return the long-running operation without waiting")
	}
	evPipDescribeCmd.Flags().StringVar(&flagEvPipFormat, "format", "", "Output format")
	evPipListCmd.Flags().StringVar(&flagEvPipFormat, "format", "", "Output format")
	evPipListCmd.Flags().Int64Var(&flagEvPipListPage, "page-size", 0, "Page size")
	evPipListCmd.Flags().Int64Var(&flagEvPipListLimit, "limit", 0, "Cap total results (0 = no cap)")
	evPipListCmd.Flags().StringVar(&flagEvPipListFilter, "filter", "", "Server-side filter expression")
	evPipListCmd.Flags().BoolVar(&flagEvPipListURI, "uri", false, "Print resource names only")

	eventarcPipelinesCmd.AddCommand(evPipCreateCmd, evPipDeleteCmd, evPipDescribeCmd, evPipListCmd, evPipUpdateCmd)
	eventarcCmd.AddCommand(eventarcPipelinesCmd)
}

func runEvPipCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	p := &eventarc.Pipeline{}
	if err := loadYAMLOrJSONInto(flagEvPipConfigFile, p); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Pipelines.Create(eventarcLocationParent(project, flagEvPipLocation), p).
		PipelineId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating pipeline: %w", err)
	}
	return eventarcFinishOp(ctx, svc, op, "Create pipeline", args[0], flagEvPipAsync)
}

func runEvPipDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Pipelines.Delete(eventarcResourceName("pipelines", args[0], project, flagEvPipLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting pipeline: %w", err)
	}
	return eventarcFinishOp(ctx, svc, op, "Delete pipeline", args[0], flagEvPipAsync)
}

func runEvPipDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	p, err := svc.Projects.Locations.Pipelines.Get(eventarcResourceName("pipelines", args[0], project, flagEvPipLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing pipeline: %w", err)
	}
	return emitFormatted(p, flagEvPipFormat)
}

func runEvPipList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := eventarcLocationParent(project, flagEvPipLocation)
	var all []*eventarc.Pipeline
	pageToken := ""
	for {
		call := svc.Projects.Locations.Pipelines.List(parent).Context(ctx)
		if flagEvPipListFilter != "" {
			call = call.Filter(flagEvPipListFilter)
		}
		if flagEvPipListPage > 0 {
			call = call.PageSize(flagEvPipListPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing pipelines: %w", err)
		}
		all = append(all, resp.Pipelines...)
		if flagEvPipListLimit > 0 && int64(len(all)) >= flagEvPipListLimit {
			all = all[:flagEvPipListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagEvPipListURI {
		for _, p := range all {
			fmt.Println(p.Name)
		}
		return nil
	}
	if flagEvPipFormat != "" {
		return emitFormatted(all, flagEvPipFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "UPDATE_TIME")
	for _, p := range all {
		fmt.Printf("%-40s %s\n", path.Base(p.Name), p.UpdateTime)
	}
	return nil
}

func runEvPipUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	p := &eventarc.Pipeline{}
	if err := loadYAMLOrJSONInto(flagEvPipConfigFile, p); err != nil {
		return err
	}
	mask := flagEvPipUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(p))
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Pipelines.Patch(eventarcResourceName("pipelines", args[0], project, flagEvPipLocation), p).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating pipeline: %w", err)
	}
	return eventarcFinishOp(ctx, svc, op, "Update pipeline", args[0], flagEvPipAsync)
}
