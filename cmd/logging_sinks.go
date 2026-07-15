package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud logging sinks (#923) ---

var loggingSinksCmd = &cobra.Command{Use: "sinks", Short: "Manage log sinks"}

var (
	flagLogSinkDestination     string
	flagLogSinkFilter          string
	flagLogSinkDescription     string
	flagLogSinkDisabled        bool
	flagLogSinkIncludeChildren bool
	flagLogSinkUniqueWriter    bool
	flagLogSinkCustomWriter    string
)

var (
	loggingSinksCreateCmd = &cobra.Command{
		Use: "create SINK", Short: "Create a log sink",
		Args: cobra.ExactArgs(1), RunE: runLogSinkCreate,
	}
	loggingSinksDeleteCmd = &cobra.Command{
		Use: "delete SINK", Short: "Delete a log sink",
		Args: cobra.ExactArgs(1), RunE: runLogSinkDelete,
	}
	loggingSinksDescribeCmd = &cobra.Command{
		Use: "describe SINK", Short: "Describe a log sink",
		Args: cobra.ExactArgs(1), RunE: runLogSinkDescribe,
	}
	loggingSinksListCmd = &cobra.Command{
		Use: "list", Short: "List log sinks",
		Args: cobra.NoArgs, RunE: runLogSinkList,
	}
	loggingSinksUpdateCmd = &cobra.Command{
		Use: "update SINK", Short: "Update a log sink",
		Args: cobra.ExactArgs(1), RunE: runLogSinkUpdate,
	}
)

func sinkFromFlags(body *logging.LogSink) {
	if flagLogSinkDestination != "" {
		body.Destination = flagLogSinkDestination
	}
	if flagLogSinkFilter != "" {
		body.Filter = flagLogSinkFilter
	}
	if flagLogSinkDescription != "" {
		body.Description = flagLogSinkDescription
	}
	if flagLogSinkDisabled {
		body.Disabled = true
		body.ForceSendFields = append(body.ForceSendFields, "Disabled")
	}
	if flagLogSinkIncludeChildren {
		body.IncludeChildren = true
		body.ForceSendFields = append(body.ForceSendFields, "IncludeChildren")
	}
}

func runLogSinkCreate(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	body := &logging.LogSink{Name: args[0]}
	if flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagLogConfigFile, body); err != nil {
			return err
		}
	}
	sinkFromFlags(body)
	if body.Destination == "" {
		return fmt.Errorf("--destination is required (or set `destination` via --config-file)")
	}
	body.Name = args[0]
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.LogSink
	switch loggingScope(parent) {
	case "projects":
		call := svc.Projects.Sinks.Create(parent, body).Context(ctx)
		if flagLogSinkUniqueWriter {
			call = call.UniqueWriterIdentity(true)
		}
		if flagLogSinkCustomWriter != "" {
			call = call.CustomWriterIdentity(flagLogSinkCustomWriter)
		}
		got, err = call.Do()
	case "folders":
		call := svc.Folders.Sinks.Create(parent, body).Context(ctx)
		if flagLogSinkUniqueWriter {
			call = call.UniqueWriterIdentity(true)
		}
		if flagLogSinkCustomWriter != "" {
			call = call.CustomWriterIdentity(flagLogSinkCustomWriter)
		}
		got, err = call.Do()
	case "organizations":
		call := svc.Organizations.Sinks.Create(parent, body).Context(ctx)
		if flagLogSinkUniqueWriter {
			call = call.UniqueWriterIdentity(true)
		}
		if flagLogSinkCustomWriter != "" {
			call = call.CustomWriterIdentity(flagLogSinkCustomWriter)
		}
		got, err = call.Do()
	case "billingAccounts":
		call := svc.BillingAccounts.Sinks.Create(parent, body).Context(ctx)
		if flagLogSinkUniqueWriter {
			call = call.UniqueWriterIdentity(true)
		}
		if flagLogSinkCustomWriter != "" {
			call = call.CustomWriterIdentity(flagLogSinkCustomWriter)
		}
		got, err = call.Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("creating log sink: %w", err)
	}
	return emitFormatted(got, flagLogFormat)
}

func runLogSinkDelete(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	name := loggingChildName(parent, "sinks", args[0])
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	switch loggingScope(parent) {
	case "projects":
		_, err = svc.Projects.Sinks.Delete(name).Context(ctx).Do()
	case "folders":
		_, err = svc.Folders.Sinks.Delete(name).Context(ctx).Do()
	case "organizations":
		_, err = svc.Organizations.Sinks.Delete(name).Context(ctx).Do()
	case "billingAccounts":
		_, err = svc.BillingAccounts.Sinks.Delete(name).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("deleting log sink: %w", err)
	}
	fmt.Printf("Deleted sink [%s].\n", args[0])
	return nil
}

func runLogSinkDescribe(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	name := loggingChildName(parent, "sinks", args[0])
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.LogSink
	switch loggingScope(parent) {
	case "projects":
		got, err = svc.Projects.Sinks.Get(name).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Sinks.Get(name).Context(ctx).Do()
	case "organizations":
		got, err = svc.Organizations.Sinks.Get(name).Context(ctx).Do()
	case "billingAccounts":
		got, err = svc.BillingAccounts.Sinks.Get(name).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("describing log sink: %w", err)
	}
	return emitFormatted(got, flagLogFormat)
}

func runLogSinkList(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	var all []*logging.LogSink
	pageToken := ""
	for {
		var (
			page []*logging.LogSink
			next string
		)
		switch loggingScope(parent) {
		case "projects":
			call := svc.Projects.Sinks.List(parent).Context(ctx)
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing log sinks: %w", err)
			}
			page, next = resp.Sinks, resp.NextPageToken
		case "folders":
			call := svc.Folders.Sinks.List(parent).Context(ctx)
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing log sinks: %w", err)
			}
			page, next = resp.Sinks, resp.NextPageToken
		case "organizations":
			call := svc.Organizations.Sinks.List(parent).Context(ctx)
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing log sinks: %w", err)
			}
			page, next = resp.Sinks, resp.NextPageToken
		case "billingAccounts":
			call := svc.BillingAccounts.Sinks.List(parent).Context(ctx)
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing log sinks: %w", err)
			}
			page, next = resp.Sinks, resp.NextPageToken
		default:
			return fmt.Errorf("invalid parent %q", parent)
		}
		all = append(all, page...)
		if next == "" {
			break
		}
		pageToken = next
	}
	if flagLogFormat != "" {
		return emitFormatted(all, flagLogFormat)
	}
	fmt.Printf("%-30s %-60s %s\n", "NAME", "DESTINATION", "FILTER")
	for _, s := range all {
		fmt.Printf("%-30s %-60s %s\n", s.Name, s.Destination, s.Filter)
	}
	return nil
}

func runLogSinkUpdate(cmd *cobra.Command, args []string) error {
	parent, err := loggingParent()
	if err != nil {
		return err
	}
	name := loggingChildName(parent, "sinks", args[0])
	body := &logging.LogSink{}
	if flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagLogConfigFile, body); err != nil {
			return err
		}
	}
	sinkFromFlags(body)
	mask := loggingResolveMask(body)
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.LogSink
	switch loggingScope(parent) {
	case "projects":
		call := svc.Projects.Sinks.Update(name, body).UpdateMask(mask).Context(ctx)
		if flagLogSinkUniqueWriter {
			call = call.UniqueWriterIdentity(true)
		}
		if flagLogSinkCustomWriter != "" {
			call = call.CustomWriterIdentity(flagLogSinkCustomWriter)
		}
		got, err = call.Do()
	case "folders":
		call := svc.Folders.Sinks.Update(name, body).UpdateMask(mask).Context(ctx)
		if flagLogSinkUniqueWriter {
			call = call.UniqueWriterIdentity(true)
		}
		if flagLogSinkCustomWriter != "" {
			call = call.CustomWriterIdentity(flagLogSinkCustomWriter)
		}
		got, err = call.Do()
	case "organizations":
		call := svc.Organizations.Sinks.Update(name, body).UpdateMask(mask).Context(ctx)
		if flagLogSinkUniqueWriter {
			call = call.UniqueWriterIdentity(true)
		}
		if flagLogSinkCustomWriter != "" {
			call = call.CustomWriterIdentity(flagLogSinkCustomWriter)
		}
		got, err = call.Do()
	case "billingAccounts":
		call := svc.BillingAccounts.Sinks.Update(name, body).UpdateMask(mask).Context(ctx)
		if flagLogSinkUniqueWriter {
			call = call.UniqueWriterIdentity(true)
		}
		if flagLogSinkCustomWriter != "" {
			call = call.CustomWriterIdentity(flagLogSinkCustomWriter)
		}
		got, err = call.Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("updating log sink: %w", err)
	}
	return emitFormatted(got, flagLogFormat)
}

func init() {
	all := []*cobra.Command{loggingSinksCreateCmd, loggingSinksDeleteCmd, loggingSinksDescribeCmd,
		loggingSinksListCmd, loggingSinksUpdateCmd}
	addLogScopeFlags(all...)
	addLogFormatFlag(loggingSinksCreateCmd, loggingSinksDescribeCmd, loggingSinksListCmd, loggingSinksUpdateCmd)
	addLogPageSizeFlag(loggingSinksListCmd)
	for _, c := range []*cobra.Command{loggingSinksCreateCmd, loggingSinksUpdateCmd} {
		c.Flags().StringVar(&flagLogSinkDestination, "destination", "", "Sink destination URI (e.g. bigquery.googleapis.com/projects/P/datasets/D)")
		c.Flags().StringVar(&flagLogSinkFilter, "log-filter", "", "Advanced logs filter")
		c.Flags().StringVar(&flagLogSinkDescription, "description", "", "A textual description for the sink")
		c.Flags().BoolVar(&flagLogSinkDisabled, "disabled", false, "Disable the sink")
		c.Flags().BoolVar(&flagLogSinkIncludeChildren, "include-children", false, "Include logs from children (folder/org sinks)")
		c.Flags().BoolVar(&flagLogSinkUniqueWriter, "unique-writer-identity", false, "Request a unique writer identity")
		c.Flags().StringVar(&flagLogSinkCustomWriter, "custom-writer-identity", "", "Custom writer identity service account")
		c.Flags().StringVar(&flagLogConfigFile, "config-file", "", "Path to a JSON/YAML file with the LogSink body")
	}
	loggingSinksUpdateCmd.Flags().StringVar(&flagLogUpdateMask, "update-mask", "", "Comma-separated list of fields to update")
	loggingSinksCmd.AddCommand(all...)
	loggingCmd.AddCommand(loggingSinksCmd)
}
