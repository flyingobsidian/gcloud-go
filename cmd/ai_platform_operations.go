package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	ml "google.golang.org/api/ml/v1"
)

// --- gcloud ai-platform operations (#985) ---

func (s *aiPlatformCmds) initAiPlatformOperations() {
	all := []*cobra.Command{
		s.aiPlatformOpsCancelCmd, s.aiPlatformOpsDescribeCmd, s.aiPlatformOpsListCmd, s.aiPlatformOpsWaitCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagAIPlatformOpsFormat, "format", "", "Output format")
	}
	s.aiPlatformOpsListCmd.Flags().Int64Var(&s.flagAIPlatformOpsPageSize, "page-size", 0, "Maximum results per page")
	s.aiPlatformOpsListCmd.Flags().StringVar(&s.flagAIPlatformOpsFilter, "filter", "", "List filter expression")

	s.aiPlatformOpsWaitCmd.Flags().DurationVar(&s.flagAIPlatformOpsPollEvery, "poll-interval", 5*time.Second,
		"Interval between operation polls")
	s.aiPlatformOpsWaitCmd.Flags().DurationVar(&s.flagAIPlatformOpsTimeout, "timeout", 30*time.Minute,
		"Maximum time to wait for the operation to complete")

	s.aiPlatformOpsCmd.AddCommand(all...)
	s.aiPlatformCmd.AddCommand(s.aiPlatformOpsCmd)
}

func (s *aiPlatformCmds) runAIPlatformOpsCancel(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Operations.Cancel(mlOperationName(project, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (s *aiPlatformCmds) runAIPlatformOpsDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Operations.Get(mlOperationName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAIPlatformOpsFormat)
}

func (s *aiPlatformCmds) runAIPlatformOpsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	var all []*ml.GoogleLongrunning__Operation
	pageToken := ""
	for {
		call := svc.Projects.Operations.List(mlProjectPath(project)).Context(ctx)
		if s.flagAIPlatformOpsPageSize > 0 {
			call = call.PageSize(s.flagAIPlatformOpsPageSize)
		}
		if s.flagAIPlatformOpsFilter != "" {
			call = call.Filter(s.flagAIPlatformOpsFilter)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagAIPlatformOpsFormat)
}

func (s *aiPlatformCmds) runAIPlatformOpsWait(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ML(ctx)
	if err != nil {
		return err
	}
	name := mlOperationName(project, args[0])
	deadline := time.Now().Add(s.flagAIPlatformOpsTimeout)
	for {
		op, err := svc.Projects.Operations.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling operation: %w", err)
		}
		if op.Done {
			if op.Error != nil {
				return fmt.Errorf("operation %s failed: %s (code %d)", args[0], op.Error.Message, op.Error.Code)
			}
			return emitFormattedTo(s.inv.stdout, op, s.flagAIPlatformOpsFormat)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out after %s waiting for operation [%s]", s.flagAIPlatformOpsTimeout, args[0])
		}
		time.Sleep(s.flagAIPlatformOpsPollEvery)
	}
}
