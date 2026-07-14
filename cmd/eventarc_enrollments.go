package cmd

import (
	"context"
	"fmt"
	"path"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	eventarc "google.golang.org/api/eventarc/v1"
)

var eventarcEnrollmentsCmd = &cobra.Command{
	Use:   "enrollments",
	Short: "Manage Eventarc enrollments",
}

var (
	evEnrCreateCmd = &cobra.Command{
		Use: "create ENROLLMENT", Short: "Create an enrollment from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runEvEnrCreate,
	}
	evEnrDeleteCmd = &cobra.Command{
		Use: "delete ENROLLMENT", Short: "Delete an enrollment",
		Args: cobra.ExactArgs(1), RunE: runEvEnrDelete,
	}
	evEnrDescribeCmd = &cobra.Command{
		Use: "describe ENROLLMENT", Short: "Describe an enrollment",
		Args: cobra.ExactArgs(1), RunE: runEvEnrDescribe,
	}
	evEnrListCmd = &cobra.Command{
		Use: "list", Short: "List enrollments in a location",
		Args: cobra.NoArgs, RunE: runEvEnrList,
	}
	evEnrUpdateCmd = &cobra.Command{
		Use: "update ENROLLMENT", Short: "Update an enrollment from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runEvEnrUpdate,
	}
)

var (
	flagEvEnrLocation   string
	flagEvEnrConfigFile string
	flagEvEnrUpdateMask string
	flagEvEnrFormat     string
	flagEvEnrAsync      bool
	flagEvEnrListLimit  int64
	flagEvEnrListPage   int64
	flagEvEnrListFilter string
	flagEvEnrListURI    bool
)

func init() {
	for _, c := range []*cobra.Command{evEnrCreateCmd, evEnrDeleteCmd, evEnrDescribeCmd, evEnrListCmd, evEnrUpdateCmd} {
		eventarcAddRegionFlag(c, &flagEvEnrLocation, true)
	}
	for _, c := range []*cobra.Command{evEnrCreateCmd, evEnrUpdateCmd} {
		c.Flags().StringVar(&flagEvEnrConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Enrollment message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	evEnrUpdateCmd.Flags().StringVar(&flagEvEnrUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{evEnrCreateCmd, evEnrDeleteCmd, evEnrUpdateCmd} {
		c.Flags().BoolVar(&flagEvEnrAsync, "async", false, "Return the long-running operation without waiting")
	}
	evEnrDescribeCmd.Flags().StringVar(&flagEvEnrFormat, "format", "", "Output format")
	evEnrListCmd.Flags().StringVar(&flagEvEnrFormat, "format", "", "Output format")
	evEnrListCmd.Flags().Int64Var(&flagEvEnrListPage, "page-size", 0, "Page size")
	evEnrListCmd.Flags().Int64Var(&flagEvEnrListLimit, "limit", 0, "Cap total results (0 = no cap)")
	evEnrListCmd.Flags().StringVar(&flagEvEnrListFilter, "filter", "", "Server-side filter expression")
	evEnrListCmd.Flags().BoolVar(&flagEvEnrListURI, "uri", false, "Print resource names only")

	eventarcEnrollmentsCmd.AddCommand(evEnrCreateCmd, evEnrDeleteCmd, evEnrDescribeCmd, evEnrListCmd, evEnrUpdateCmd)
	eventarcCmd.AddCommand(eventarcEnrollmentsCmd)
}

func runEvEnrCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	e := &eventarc.Enrollment{}
	if err := loadYAMLOrJSONInto(flagEvEnrConfigFile, e); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Enrollments.Create(eventarcLocationParent(project, flagEvEnrLocation), e).
		EnrollmentId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating enrollment: %w", err)
	}
	return eventarcFinishOp(ctx, svc, op, "Create enrollment", args[0], flagEvEnrAsync)
}

func runEvEnrDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Enrollments.Delete(eventarcResourceName("enrollments", args[0], project, flagEvEnrLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting enrollment: %w", err)
	}
	return eventarcFinishOp(ctx, svc, op, "Delete enrollment", args[0], flagEvEnrAsync)
}

func runEvEnrDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	e, err := svc.Projects.Locations.Enrollments.Get(eventarcResourceName("enrollments", args[0], project, flagEvEnrLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing enrollment: %w", err)
	}
	return emitFormatted(e, flagEvEnrFormat)
}

func runEvEnrList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := eventarcLocationParent(project, flagEvEnrLocation)
	var all []*eventarc.Enrollment
	pageToken := ""
	for {
		call := svc.Projects.Locations.Enrollments.List(parent).Context(ctx)
		if flagEvEnrListFilter != "" {
			call = call.Filter(flagEvEnrListFilter)
		}
		if flagEvEnrListPage > 0 {
			call = call.PageSize(flagEvEnrListPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing enrollments: %w", err)
		}
		all = append(all, resp.Enrollments...)
		if flagEvEnrListLimit > 0 && int64(len(all)) >= flagEvEnrListLimit {
			all = all[:flagEvEnrListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagEvEnrListURI {
		for _, e := range all {
			fmt.Println(e.Name)
		}
		return nil
	}
	if flagEvEnrFormat != "" {
		return emitFormatted(all, flagEvEnrFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "MESSAGE_BUS")
	for _, e := range all {
		fmt.Printf("%-40s %s\n", path.Base(e.Name), e.MessageBus)
	}
	return nil
}

func runEvEnrUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	e := &eventarc.Enrollment{}
	if err := loadYAMLOrJSONInto(flagEvEnrConfigFile, e); err != nil {
		return err
	}
	mask := flagEvEnrUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(e))
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Enrollments.Patch(eventarcResourceName("enrollments", args[0], project, flagEvEnrLocation), e).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating enrollment: %w", err)
	}
	return eventarcFinishOp(ctx, svc, op, "Update enrollment", args[0], flagEvEnrAsync)
}
