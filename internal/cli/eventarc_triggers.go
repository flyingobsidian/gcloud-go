package cli

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	eventarc "google.golang.org/api/eventarc/v1"
)

func (grp *eventarcCmds) initEventarcTriggers() {
	for _, c := range []*cobra.Command{grp.evTrgCreateCmd, grp.evTrgDeleteCmd, grp.evTrgDescribeCmd, grp.evTrgListCmd, grp.evTrgUpdateCmd} {
		eventarcAddRegionFlag(c, &grp.flagEvTrgLocation, true)
	}
	for _, c := range []*cobra.Command{grp.evTrgCreateCmd, grp.evTrgUpdateCmd} {
		c.Flags().StringVar(&grp.flagEvTrgConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Trigger message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	grp.evTrgUpdateCmd.Flags().StringVar(&grp.flagEvTrgUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{grp.evTrgCreateCmd, grp.evTrgDeleteCmd, grp.evTrgUpdateCmd} {
		c.Flags().BoolVar(&grp.flagEvTrgAsync, "async", false, "Return the long-running operation without waiting")
	}
	grp.evTrgDescribeCmd.Flags().StringVar(&grp.flagEvTrgFormat, "format", "", "Output format")
	grp.evTrgListCmd.Flags().StringVar(&grp.flagEvTrgFormat, "format", "", "Output format")
	grp.evTrgListCmd.Flags().Int64Var(&grp.flagEvTrgListPage, "page-size", 0, "Page size")
	grp.evTrgListCmd.Flags().Int64Var(&grp.flagEvTrgListLimit, "limit", 0, "Cap total results (0 = no cap)")
	grp.evTrgListCmd.Flags().StringVar(&grp.flagEvTrgListFilter, "filter", "", "Server-side filter expression")
	grp.evTrgListCmd.Flags().BoolVar(&grp.flagEvTrgListURI, "uri", false, "Print resource names only")

	grp.eventarcTriggersCmd.AddCommand(grp.evTrgCreateCmd, grp.evTrgDeleteCmd, grp.evTrgDescribeCmd, grp.evTrgListCmd, grp.evTrgUpdateCmd)
	grp.eventarcCmd.AddCommand(grp.eventarcTriggersCmd)
}

func (grp *eventarcCmds) runEvTrgCreate(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	t := &eventarc.Trigger{}
	if err := loadYAMLOrJSONInto(grp.flagEvTrgConfigFile, t); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Triggers.Create(eventarcLocationParent(project, grp.flagEvTrgLocation), t).
		TriggerId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating trigger: %w", err)
	}
	return grp.eventarcFinishOp(ctx, svc, op, "Create trigger", args[0], grp.flagEvTrgAsync)
}

func (grp *eventarcCmds) runEvTrgDelete(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Triggers.Delete(eventarcResourceName("triggers", args[0], project, grp.flagEvTrgLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting trigger: %w", err)
	}
	return grp.eventarcFinishOp(ctx, svc, op, "Delete trigger", args[0], grp.flagEvTrgAsync)
}

func (grp *eventarcCmds) runEvTrgDescribe(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	t, err := svc.Projects.Locations.Triggers.Get(eventarcResourceName("triggers", args[0], project, grp.flagEvTrgLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing trigger: %w", err)
	}
	return emitFormattedTo(grp.inv.stdout, t, grp.flagEvTrgFormat)
}

func (grp *eventarcCmds) runEvTrgList(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	parent := eventarcLocationParent(project, grp.flagEvTrgLocation)
	var all []*eventarc.Trigger
	pageToken := ""
	for {
		call := svc.Projects.Locations.Triggers.List(parent).Context(ctx)
		if grp.flagEvTrgListFilter != "" {
			call = call.Filter(grp.flagEvTrgListFilter)
		}
		if grp.flagEvTrgListPage > 0 {
			call = call.PageSize(grp.flagEvTrgListPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing triggers: %w", err)
		}
		all = append(all, resp.Triggers...)
		if grp.flagEvTrgListLimit > 0 && int64(len(all)) >= grp.flagEvTrgListLimit {
			all = all[:grp.flagEvTrgListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if grp.flagEvTrgListURI {
		for _, t := range all {
			fmt.Fprintln(grp.inv.stdout, t.Name)
		}
		return nil
	}
	if grp.flagEvTrgFormat != "" {
		return emitFormattedTo(grp.inv.stdout, all, grp.flagEvTrgFormat)
	}
	printTableHeadingTo(grp.inv.stdout, len(all), "%-40s %s\n", "NAME", "TYPE")
	for _, t := range all {
		fmt.Fprintf(grp.inv.stdout, "%-40s %s\n", path.Base(t.Name), triggerEventType(t))
	}
	return nil
}

func triggerEventType(t *eventarc.Trigger) string {
	for _, f := range t.EventFilters {
		if f.Attribute == "type" {
			return f.Value
		}
	}
	return ""
}

func (grp *eventarcCmds) runEvTrgUpdate(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	t := &eventarc.Trigger{}
	if err := loadYAMLOrJSONInto(grp.flagEvTrgConfigFile, t); err != nil {
		return err
	}
	mask := grp.flagEvTrgUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(t))
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Triggers.Patch(eventarcResourceName("triggers", args[0], project, grp.flagEvTrgLocation), t).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating trigger: %w", err)
	}
	return grp.eventarcFinishOp(ctx, svc, op, "Update trigger", args[0], grp.flagEvTrgAsync)
}
