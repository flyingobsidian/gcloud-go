package cmd

import (
	"context"
	"fmt"
	"path"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	eventarc "google.golang.org/api/eventarc/v1"
)

var eventarcTriggersCmd = &cobra.Command{
	Use:   "triggers",
	Short: "Manage Eventarc triggers",
}

var (
	evTrgCreateCmd = &cobra.Command{
		Use: "create TRIGGER", Short: "Create a trigger from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runEvTrgCreate,
	}
	evTrgDeleteCmd = &cobra.Command{
		Use: "delete TRIGGER", Short: "Delete a trigger",
		Args: cobra.ExactArgs(1), RunE: runEvTrgDelete,
	}
	evTrgDescribeCmd = &cobra.Command{
		Use: "describe TRIGGER", Short: "Describe a trigger",
		Args: cobra.ExactArgs(1), RunE: runEvTrgDescribe,
	}
	evTrgListCmd = &cobra.Command{
		Use: "list", Short: "List triggers in a location",
		Args: cobra.NoArgs, RunE: runEvTrgList,
	}
	evTrgUpdateCmd = &cobra.Command{
		Use: "update TRIGGER", Short: "Update a trigger from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runEvTrgUpdate,
	}
)

var (
	flagEvTrgLocation   string
	flagEvTrgConfigFile string
	flagEvTrgUpdateMask string
	flagEvTrgFormat     string
	flagEvTrgAsync      bool
	flagEvTrgListLimit  int64
	flagEvTrgListPage   int64
	flagEvTrgListFilter string
	flagEvTrgListURI    bool
)

func init() {
	for _, c := range []*cobra.Command{evTrgCreateCmd, evTrgDeleteCmd, evTrgDescribeCmd, evTrgListCmd, evTrgUpdateCmd} {
		eventarcAddRegionFlag(c, &flagEvTrgLocation, true)
	}
	for _, c := range []*cobra.Command{evTrgCreateCmd, evTrgUpdateCmd} {
		c.Flags().StringVar(&flagEvTrgConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Trigger message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	evTrgUpdateCmd.Flags().StringVar(&flagEvTrgUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{evTrgCreateCmd, evTrgDeleteCmd, evTrgUpdateCmd} {
		c.Flags().BoolVar(&flagEvTrgAsync, "async", false, "Return the long-running operation without waiting")
	}
	evTrgDescribeCmd.Flags().StringVar(&flagEvTrgFormat, "format", "", "Output format")
	evTrgListCmd.Flags().StringVar(&flagEvTrgFormat, "format", "", "Output format")
	evTrgListCmd.Flags().Int64Var(&flagEvTrgListPage, "page-size", 0, "Page size")
	evTrgListCmd.Flags().Int64Var(&flagEvTrgListLimit, "limit", 0, "Cap total results (0 = no cap)")
	evTrgListCmd.Flags().StringVar(&flagEvTrgListFilter, "filter", "", "Server-side filter expression")
	evTrgListCmd.Flags().BoolVar(&flagEvTrgListURI, "uri", false, "Print resource names only")

	eventarcTriggersCmd.AddCommand(evTrgCreateCmd, evTrgDeleteCmd, evTrgDescribeCmd, evTrgListCmd, evTrgUpdateCmd)
	eventarcCmd.AddCommand(eventarcTriggersCmd)
}

func runEvTrgCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	t := &eventarc.Trigger{}
	if err := loadYAMLOrJSONInto(flagEvTrgConfigFile, t); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Triggers.Create(eventarcLocationParent(project, flagEvTrgLocation), t).
		TriggerId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating trigger: %w", err)
	}
	return eventarcFinishOp(ctx, svc, op, "Create trigger", args[0], flagEvTrgAsync)
}

func runEvTrgDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Triggers.Delete(eventarcResourceName("triggers", args[0], project, flagEvTrgLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting trigger: %w", err)
	}
	return eventarcFinishOp(ctx, svc, op, "Delete trigger", args[0], flagEvTrgAsync)
}

func runEvTrgDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	t, err := svc.Projects.Locations.Triggers.Get(eventarcResourceName("triggers", args[0], project, flagEvTrgLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing trigger: %w", err)
	}
	return emitFormatted(t, flagEvTrgFormat)
}

func runEvTrgList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := eventarcLocationParent(project, flagEvTrgLocation)
	var all []*eventarc.Trigger
	pageToken := ""
	for {
		call := svc.Projects.Locations.Triggers.List(parent).Context(ctx)
		if flagEvTrgListFilter != "" {
			call = call.Filter(flagEvTrgListFilter)
		}
		if flagEvTrgListPage > 0 {
			call = call.PageSize(flagEvTrgListPage)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing triggers: %w", err)
		}
		all = append(all, resp.Triggers...)
		if flagEvTrgListLimit > 0 && int64(len(all)) >= flagEvTrgListLimit {
			all = all[:flagEvTrgListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagEvTrgListURI {
		for _, t := range all {
			fmt.Println(t.Name)
		}
		return nil
	}
	if flagEvTrgFormat != "" {
		return emitFormatted(all, flagEvTrgFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "TYPE")
	for _, t := range all {
		fmt.Printf("%-40s %s\n", path.Base(t.Name), triggerEventType(t))
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

func runEvTrgUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	t := &eventarc.Trigger{}
	if err := loadYAMLOrJSONInto(flagEvTrgConfigFile, t); err != nil {
		return err
	}
	mask := flagEvTrgUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(t))
	}
	ctx := context.Background()
	svc, err := gcp.EventarcService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Triggers.Patch(eventarcResourceName("triggers", args[0], project, flagEvTrgLocation), t).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating trigger: %w", err)
	}
	return eventarcFinishOp(ctx, svc, op, "Update trigger", args[0], flagEvTrgAsync)
}
