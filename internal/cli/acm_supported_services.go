package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	accesscontextmanager "google.golang.org/api/accesscontextmanager/v1"
)

// --- gcloud access-context-manager supported-services (#1447) ---

func (s *accessContextManagerCmds) initAcmSupportedServices() {
	for _, c := range []*cobra.Command{s.acmSSListCmd, s.acmSSDescribeCmd} {
		c.Flags().StringVar(&s.flagACMSSFormat, "format", "", "Output format")
	}
	s.acmSSListCmd.Flags().Int64Var(&s.flagACMSSPageSize, "page-size", 0, "Maximum results per page")

	s.acmSSCmd.AddCommand(s.acmSSListCmd, s.acmSSDescribeCmd)
	s.accessContextManagerCmd.AddCommand(s.acmSSCmd)
}

func (s *accessContextManagerCmds) runACMSSList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	var all []*accesscontextmanager.SupportedService
	pageToken := ""
	for {
		call := svc.Services.List().Context(ctx)
		if s.flagACMSSPageSize > 0 {
			call = call.PageSize(s.flagACMSSPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing supported services: %w", err)
		}
		all = append(all, resp.SupportedServices...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagACMSSFormat)
}

func (s *accessContextManagerCmds) runACMSSDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Services.Get(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing supported service: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagACMSSFormat)
}
