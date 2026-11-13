package cmd

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	eventarc "google.golang.org/api/eventarc/v1"
)

func (grp *eventarcCmds) initEventarcEnrollments() {
	for _, c := range []*cobra.Command{grp.evEnrCreateCmd, grp.evEnrDeleteCmd, grp.evEnrDescribeCmd, grp.evEnrListCmd, grp.evEnrUpdateCmd} {
		eventarcAddRegionFlag(c, &grp.flagEvEnrLocation, true)
	}
	for _, c := range []*cobra.Command{grp.evEnrCreateCmd, grp.evEnrUpdateCmd} {
		c.Flags().StringVar(&grp.flagEvEnrConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Enrollment message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	grp.evEnrUpdateCmd.Flags().StringVar(&grp.flagEvEnrUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{grp.evEnrCreateCmd, grp.evEnrDeleteCmd, grp.evEnrUpdateCmd} {
		c.Flags().BoolVar(&grp.flagEvEnrAsync, "async", false, "Return the long-running operation without waiting")
	}
	grp.evEnrDescribeCmd.Flags().StringVar(&grp.flagEvEnrFormat, "format", "", "Output format")
	grp.evEnrListCmd.Flags().StringVar(&grp.flagEvEnrFormat, "format", "", "Output format")
	grp.evEnrListCmd.Flags().Int64Var(&grp.flagEvEnrListPage, "page-size", 0, "Page size")
	grp.evEnrListCmd.Flags().Int64Var(&grp.flagEvEnrListLimit, "limit", 0, "Cap total results (0 = no cap)")
	grp.evEnrListCmd.Flags().StringVar(&grp.flagEvEnrListFilter, "filter", "", "Server-side filter expression")
	grp.evEnrListCmd.Flags().BoolVar(&grp.flagEvEnrListURI, "uri", false, "Print resource names only")

	grp.eventarcEnrollmentsCmd.AddCommand(grp.evEnrCreateCmd, grp.evEnrDeleteCmd, grp.evEnrDescribeCmd, grp.evEnrListCmd, grp.evEnrUpdateCmd)
	grp.eventarcCmd.AddCommand(grp.eventarcEnrollmentsCmd)
}

func (grp *eventarcCmds) runEvEnrCreate(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	e := &eventarc.Enrollment{}
	if err := loadYAMLOrJSONInto(grp.flagEvEnrConfigFile, e); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Enrollments.Create(eventarcLocationParent(project, grp.flagEvEnrLocation), e).
		EnrollmentId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating enrollment: %w", err)
	}
	return grp.eventarcFinishOp(ctx, svc, op, "Create enrollment", args[0], grp.flagEvEnrAsync)
}

func (grp *eventarcCmds) runEvEnrDelete(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Enrollments.Delete(eventarcResourceName("enrollments", args[0], project, grp.flagEvEnrLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting enrollment: %w", err)
	}
	return grp.eventarcFinishOp(ctx, svc, op, "Delete enrollment", args[0], grp.flagEvEnrAsync)
}

func (grp *eventarcCmds) runEvEnrDescribe(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	e, err := svc.Projects.Locations.Enrollments.Get(eventarcResourceName("enrollments", args[0], project, grp.flagEvEnrLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing enrollment: %w", err)
	}
	return emitFormattedTo(grp.inv.stdout, e, grp.flagEvEnrFormat)
}

func (grp *eventarcCmds) runEvEnrList(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	parent := eventarcLocationParent(project, grp.flagEvEnrLocation)
	var all []*eventarc.Enrollment
	pageToken := ""
	for {
		call := svc.Projects.Locations.Enrollments.List(parent).Context(ctx)
		if grp.flagEvEnrListFilter != "" {
			call = call.Filter(grp.flagEvEnrListFilter)
		}
		if grp.flagEvEnrListPage > 0 {
			call = call.PageSize(grp.flagEvEnrListPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing enrollments: %w", err)
		}
		all = append(all, resp.Enrollments...)
		if grp.flagEvEnrListLimit > 0 && int64(len(all)) >= grp.flagEvEnrListLimit {
			all = all[:grp.flagEvEnrListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if grp.flagEvEnrListURI {
		for _, e := range all {
			fmt.Fprintln(grp.inv.stdout, e.Name)
		}
		return nil
	}
	if grp.flagEvEnrFormat != "" {
		return emitFormattedTo(grp.inv.stdout, all, grp.flagEvEnrFormat)
	}
	printTableHeadingTo(grp.inv.stdout, len(all), "%-40s %s\n", "NAME", "MESSAGE_BUS")
	for _, e := range all {
		fmt.Fprintf(grp.inv.stdout, "%-40s %s\n", path.Base(e.Name), e.MessageBus)
	}
	return nil
}

func (grp *eventarcCmds) runEvEnrUpdate(cmd *cobra.Command, args []string) error {
	project, err := grp.inv.project()
	if err != nil {
		return err
	}
	e := &eventarc.Enrollment{}
	if err := loadYAMLOrJSONInto(grp.flagEvEnrConfigFile, e); err != nil {
		return err
	}
	mask := grp.flagEvEnrUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(e))
	}
	ctx := cmd.Context()
	svc, err := grp.inv.clients().Eventarc(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Enrollments.Patch(eventarcResourceName("enrollments", args[0], project, grp.flagEvEnrLocation), e).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating enrollment: %w", err)
	}
	return grp.eventarcFinishOp(ctx, svc, op, "Update enrollment", args[0], grp.flagEvEnrAsync)
}
