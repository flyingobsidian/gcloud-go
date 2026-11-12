package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud logging operations (#917) ---

func (g *loggingCmds) opLocationParent() (string, error) {
	parent, err := g.loggingParent()
	if err != nil {
		return "", err
	}
	return loggingLocationParent(parent, g.loggingLocation()), nil
}

func (g *loggingCmds) opName(id string) (string, string, error) {
	parent, err := g.loggingParent()
	if err != nil {
		return "", "", err
	}
	scope := loggingScope(parent)
	name := loggingLocationChildName(parent, g.loggingLocation(), "operations", id)
	return scope, name, nil
}

func (g *loggingCmds) runLogOperationCancel(cmd *cobra.Command, args []string) error {
	scope, name, err := g.opName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
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
	fmt.Fprintf(g.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (g *loggingCmds) runLogOperationDescribe(cmd *cobra.Command, args []string) error {
	scope, name, err := g.opName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagLogFormat)
}

func (g *loggingCmds) runLogOperationList(cmd *cobra.Command, args []string) error {
	parent, err := g.opLocationParent()
	if err != nil {
		return err
	}
	scope := loggingScope(parent)
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
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
			if g.flagLogFilter != "" {
				call = call.Filter(g.flagLogFilter)
			}
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
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
			if g.flagLogFilter != "" {
				call = call.Filter(g.flagLogFilter)
			}
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
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
			if g.flagLogFilter != "" {
				call = call.Filter(g.flagLogFilter)
			}
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
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
			if g.flagLogFilter != "" {
				call = call.Filter(g.flagLogFilter)
			}
			if g.flagLogPageSize > 0 {
				call = call.PageSize(g.flagLogPageSize)
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
	if g.flagLogFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagLogFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-60s %s\n", "NAME", "DONE")
	for _, o := range all {
		fmt.Fprintf(g.inv.stdout, "%-60s %v\n", loggingBasename(o.Name), o.Done)
	}
	return nil
}

func (g *loggingCmds) initLoggingOperations() {
	all := []*cobra.Command{g.loggingOperationsCancelCmd, g.loggingOperationsDescribeCmd, g.loggingOperationsListCmd}
	g.addLogScopeFlags(all...)
	g.addLogLocationFlag(all...)
	g.addLogFormatFlag(g.loggingOperationsDescribeCmd, g.loggingOperationsListCmd)
	g.addLogFilterFlag(g.loggingOperationsListCmd)
	g.addLogPageSizeFlag(g.loggingOperationsListCmd)
	g.loggingOperationsCmd.AddCommand(all...)
	g.loggingCmd.AddCommand(g.loggingOperationsCmd)
}
