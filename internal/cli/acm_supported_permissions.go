package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// --- gcloud access-context-manager supported-permissions (#1446) ---

func (s *accessContextManagerCmds) initAcmSupportedPermissions() {
	for _, c := range []*cobra.Command{s.acmSPListCmd, s.acmSPDescribeCmd} {
		c.Flags().StringVar(&s.flagACMSPFormat, "format", "", "Output format")
	}
	s.acmSPListCmd.Flags().Int64Var(&s.flagACMSPPageSize, "page-size", 0, "Maximum results per page")

	s.acmSPCmd.AddCommand(s.acmSPListCmd, s.acmSPDescribeCmd)
	s.accessContextManagerCmd.AddCommand(s.acmSPCmd)
}

func (s *accessContextManagerCmds) acmSPListAll(ctx context.Context) ([]string, error) {
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return nil, err
	}
	var all []string
	pageToken := ""
	for {
		call := svc.Permissions.List().Context(ctx)
		if s.flagACMSPPageSize > 0 {
			call = call.PageSize(s.flagACMSPPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("listing supported permissions: %w", err)
		}
		all = append(all, resp.SupportedPermissions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return all, nil
}

func (s *accessContextManagerCmds) runACMSPList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	all, err := s.acmSPListAll(ctx)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagACMSPFormat)
}

func (s *accessContextManagerCmds) runACMSPDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	all, err := s.acmSPListAll(ctx)
	if err != nil {
		return err
	}
	for _, p := range all {
		if p == args[0] {
			return emitFormattedTo(s.inv.stdout, p, s.flagACMSPFormat)
		}
	}
	return fmt.Errorf("supported permission %q not found", args[0])
}
