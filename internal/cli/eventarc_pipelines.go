package cli

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	eventarc "google.golang.org/api/eventarc/v1"
)

func (grp *eventarcCmds) initEventarcPipelines() {
	for _, c := range []*cobra.Command{grp.evPipCreateCmd, grp.evPipDeleteCmd, grp.evPipDescribeCmd, grp.evPipListCmd, grp.evPipUpdateCmd} {
		eventarcAddRegionFlag(c, &grp.flagEvPipLocation, true)
	}
	for _, c := range []*cobra.Command{grp.evPipCreateCmd, grp.evPipUpdateCmd} {
		c.Flags().StringVar(&grp.flagEvPipConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Pipeline message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	grp.evPipUpdateCmd.Flags().StringVar(&grp.flagEvPipUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{grp.evPipCreateCmd, grp.evPipDeleteCmd, grp.evPipUpdateCmd} {
		c.Flags().BoolVar(&grp.flagEvPipAsync, "async", false, "Return the long-running operation without waiting")
	}
	grp.evPipDescribeCmd.Flags().StringVar(&grp.flagEvPipFormat, "format", "", "Output format")
	grp.evPipListCmd.Flags().StringVar(&grp.flagEvPipFormat, "format", "", "Output format")
	grp.evPipListCmd.Flags().Int64Var(&grp.flagEvPipListPage, "page-size", 0, "Page size")
	grp.evPipListCmd.Flags().Int64Var(&grp.flagEvPipListLimit, "limit", 0, "Cap total results (0 = no cap)")
	grp.evPipListCmd.Flags().StringVar(&grp.flagEvPipListFilter, "filter", "", "Server-side filter expression")
	grp.evPipListCmd.Flags().BoolVar(&grp.flagEvPipListURI, "uri", false, "Print resource names only")

	grp.eventarcPipelinesCmd.AddCommand(grp.evPipCreateCmd, grp.evPipDeleteCmd, grp.evPipDescribeCmd, grp.evPipListCmd, grp.evPipUpdateCmd)
	grp.eventarcCmd.AddCommand(grp.eventarcPipelinesCmd)
}

func (grp *eventarcCmds) runEvPipCreate(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	p := &eventarc.Pipeline{}
	if err := loadYAMLOrJSONInto(grp.flagEvPipConfigFile, p); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Pipelines.Create(eventarcLocationParent(project, grp.flagEvPipLocation), p).
		PipelineId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating pipeline: %w", err)
	}
	return grp.eventarcFinishOp(ctx, svc, op, "Create pipeline", args[0], grp.flagEvPipAsync)
}

func (grp *eventarcCmds) runEvPipDelete(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Pipelines.Delete(eventarcResourceName("pipelines", args[0], project, grp.flagEvPipLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting pipeline: %w", err)
	}
	return grp.eventarcFinishOp(ctx, svc, op, "Delete pipeline", args[0], grp.flagEvPipAsync)
}

func (grp *eventarcCmds) runEvPipDescribe(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	p, err := svc.Projects.Locations.Pipelines.Get(eventarcResourceName("pipelines", args[0], project, grp.flagEvPipLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing pipeline: %w", err)
	}
	return emitFormattedTo(grp.inv.stdout, p, grp.flagEvPipFormat)
}

func (grp *eventarcCmds) runEvPipList(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	parent := eventarcLocationParent(project, grp.flagEvPipLocation)
	var all []*eventarc.Pipeline
	pageToken := ""
	for {
		call := svc.Projects.Locations.Pipelines.List(parent).Context(ctx)
		if grp.flagEvPipListFilter != "" {
			call = call.Filter(grp.flagEvPipListFilter)
		}
		if grp.flagEvPipListPage > 0 {
			call = call.PageSize(grp.flagEvPipListPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing pipelines: %w", err)
		}
		all = append(all, resp.Pipelines...)
		if grp.flagEvPipListLimit > 0 && int64(len(all)) >= grp.flagEvPipListLimit {
			all = all[:grp.flagEvPipListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if grp.flagEvPipListURI {
		for _, p := range all {
			fmt.Fprintln(grp.inv.stdout, p.Name)
		}
		return nil
	}
	if grp.flagEvPipFormat != "" {
		return emitFormattedTo(grp.inv.stdout, all, grp.flagEvPipFormat)
	}
	printTableHeadingTo(grp.inv.stdout, len(all), "%-40s %s\n", "NAME", "UPDATE_TIME")
	for _, p := range all {
		fmt.Fprintf(grp.inv.stdout, "%-40s %s\n", path.Base(p.Name), p.UpdateTime)
	}
	return nil
}

func (grp *eventarcCmds) runEvPipUpdate(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	p := &eventarc.Pipeline{}
	if err := loadYAMLOrJSONInto(grp.flagEvPipConfigFile, p); err != nil {
		return err
	}
	mask := grp.flagEvPipUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(p))
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Pipelines.Patch(eventarcResourceName("pipelines", args[0], project, grp.flagEvPipLocation), p).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating pipeline: %w", err)
	}
	return grp.eventarcFinishOp(ctx, svc, op, "Update pipeline", args[0], grp.flagEvPipAsync)
}
