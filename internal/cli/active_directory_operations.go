package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	managedidentities "google.golang.org/api/managedidentities/v1"
)

// --- gcloud active-directory operations (#1449) ---

func (s *activeDirectoryCmds) initActiveDirectoryOperations() {
	all := []*cobra.Command{s.adOpCancelCmd, s.adOpDescribeCmd, s.adOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagADOpFormat, "format", "", "Output format")
	}
	s.adOpListCmd.Flags().Int64Var(&s.flagADOpPageSize, "page-size", 0, "Maximum results per page")
	s.adOpListCmd.Flags().StringVar(&s.flagADOpFilter, "filter", "", "Server-side list filter")

	s.adOperationsCmd.AddCommand(all...)
	s.activeDirectoryCmd.AddCommand(s.adOperationsCmd)
}

func adOpResource(project, op string) string {
	return fmt.Sprintf("projects/%s/locations/global/operations/%s", project, op)
}

func adOpParent(project string) string {
	return fmt.Sprintf("projects/%s/locations/global/operations", project)
}

func (s *activeDirectoryCmds) runADOpCancel(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Global.Operations.
		Cancel(adOpResource(project, args[0]),
			&managedidentities.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancel request issued for operation [%s].\n", args[0])
	return nil
}

func (s *activeDirectoryCmds) runADOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Global.Operations.
		Get(adOpResource(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagADOpFormat)
}

func (s *activeDirectoryCmds) runADOpList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	var all []*managedidentities.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Global.Operations.List(adOpParent(project)).Context(ctx)
		if s.flagADOpPageSize > 0 {
			call = call.PageSize(s.flagADOpPageSize)
		}
		if s.flagADOpFilter != "" {
			call = call.Filter(s.flagADOpFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagADOpFormat)
}
