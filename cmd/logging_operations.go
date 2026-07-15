package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud logging operations (#917) ---

var loggingOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Cloud Logging long-running operations"}

var (
	loggingOperationsCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a long-running operation",
		Args: cobra.ExactArgs(1), RunE: runLogOperationCancel,
	}
	loggingOperationsDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a long-running operation",
		Args: cobra.ExactArgs(1), RunE: runLogOperationDescribe,
	}
	loggingOperationsListCmd = &cobra.Command{
		Use: "list", Short: "List long-running operations",
		Args: cobra.NoArgs, RunE: runLogOperationList,
	}
)

func opLocationParent() (string, error) {
	parent, err := loggingParent()
	if err != nil {
		return "", err
	}
	return loggingLocationParent(parent, loggingLocation()), nil
}

func opName(id string) (string, string, error) {
	parent, err := loggingParent()
	if err != nil {
		return "", "", err
	}
	scope := loggingScope(parent)
	name := loggingLocationChildName(parent, loggingLocation(), "operations", id)
	return scope, name, nil
}

func runLogOperationCancel(cmd *cobra.Command, args []string) error {
	scope, name, err := opName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	req := &logging.CancelOperationRequest{}
	switch scope {
	case "projects":
		_, err = svc.Projects.Locations.Operations.Cancel(name, req).Context(ctx).Do()
	case "folders":
		_, err = svc.Folders.Locations.Operations.Cancel(name, req).Context(ctx).Do()
	case "organizations":
		_, err = svc.Organizations.Locations.Operations.Cancel(name, req).Context(ctx).Do()
	case "billingAccounts":
		_, err = svc.BillingAccounts.Locations.Operations.Cancel(name, req).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid scope %q", scope)
	}
	if err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runLogOperationDescribe(cmd *cobra.Command, args []string) error {
	scope, name, err := opName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.Operation
	switch scope {
	case "projects":
		got, err = svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.Locations.Operations.Get(name).Context(ctx).Do()
	case "organizations":
		got, err = svc.Organizations.Locations.Operations.Get(name).Context(ctx).Do()
	case "billingAccounts":
		got, err = svc.BillingAccounts.Locations.Operations.Get(name).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid scope %q", scope)
	}
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(got, flagLogFormat)
}

func runLogOperationList(cmd *cobra.Command, args []string) error {
	parent, err := opLocationParent()
	if err != nil {
		return err
	}
	scope := loggingScope(parent)
	ctx := context.Background()
	svc, err := loggingClient(ctx)
	if err != nil {
		return err
	}
	var all []*logging.Operation
	pageToken := ""
	for {
		var (
			page []*logging.Operation
			next string
		)
		switch scope {
		case "projects":
			call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
			if flagLogFilter != "" {
				call = call.Filter(flagLogFilter)
			}
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing operations: %w", err)
			}
			page, next = resp.Operations, resp.NextPageToken
		case "folders":
			call := svc.Folders.Locations.Operations.List(parent).Context(ctx)
			if flagLogFilter != "" {
				call = call.Filter(flagLogFilter)
			}
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing operations: %w", err)
			}
			page, next = resp.Operations, resp.NextPageToken
		case "organizations":
			call := svc.Organizations.Locations.Operations.List(parent).Context(ctx)
			if flagLogFilter != "" {
				call = call.Filter(flagLogFilter)
			}
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing operations: %w", err)
			}
			page, next = resp.Operations, resp.NextPageToken
		case "billingAccounts":
			call := svc.BillingAccounts.Locations.Operations.List(parent).Context(ctx)
			if flagLogFilter != "" {
				call = call.Filter(flagLogFilter)
			}
			if flagLogPageSize > 0 {
				call = call.PageSize(flagLogPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing operations: %w", err)
			}
			page, next = resp.Operations, resp.NextPageToken
		default:
			return fmt.Errorf("invalid scope %q", scope)
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
	fmt.Printf("%-60s %s\n", "NAME", "DONE")
	for _, o := range all {
		fmt.Printf("%-60s %v\n", loggingBasename(o.Name), o.Done)
	}
	return nil
}

func init() {
	all := []*cobra.Command{loggingOperationsCancelCmd, loggingOperationsDescribeCmd, loggingOperationsListCmd}
	addLogScopeFlags(all...)
	addLogLocationFlag(all...)
	addLogFormatFlag(loggingOperationsDescribeCmd, loggingOperationsListCmd)
	addLogFilterFlag(loggingOperationsListCmd)
	addLogPageSizeFlag(loggingOperationsListCmd)
	loggingOperationsCmd.AddCommand(all...)
	loggingCmd.AddCommand(loggingOperationsCmd)
}
