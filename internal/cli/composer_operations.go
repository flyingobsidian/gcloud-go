package cli

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
	composer "google.golang.org/api/composer/v1"
)

// --- gcloud composer operations (#1503) ---

func (s *composerCmds) initComposerOperations() {
	all := []*cobra.Command{s.composerOpDeleteCmd, s.composerOpDescribeCmd, s.composerOpListCmd, s.composerOpWaitCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagComposerOpLocation, "location", "", "Composer location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagComposerOpFormat, "format", "", "Output format")
	}
	s.composerOpListCmd.Flags().StringVar(&s.flagComposerOpFilter, "filter", "", "Server-side filter expression")
	s.composerOpListCmd.Flags().Int64Var(&s.flagComposerOpPageSize, "page-size", 0, "Maximum results per page")
	s.composerOpWaitCmd.Flags().DurationVar(&s.flagComposerOpTimeout, "timeout", 30*time.Minute,
		"Maximum time to wait before returning (e.g. 15m, 1h)")

	s.composerOpCmd.AddCommand(all...)
	s.composerCmd.AddCommand(s.composerOpCmd)
}

func (s *composerCmds) composerOpParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return composerLocationParent(project, s.flagComposerOpLocation), nil
}

func (s *composerCmds) composerOpName(id string) (string, error) {
	parent, err := s.composerOpParent()
	if err != nil {
		return "", err
	}
	if len(id) > 0 && id[0] == '/' {
		return id[1:], nil
	}
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	return fmt.Sprintf("%s/operations/%s", parent, id), nil
}

func (s *composerCmds) runComposerOpDelete(cmd *cobra.Command, args []string) error {
	name, err := s.composerOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted operation %s.\n", args[0])
	return nil
}

func (s *composerCmds) runComposerOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.composerOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagComposerOpFormat)
}

func (s *composerCmds) runComposerOpList(cmd *cobra.Command, args []string) error {
	parent, err := s.composerOpParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	var all []*composer.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
		if s.flagComposerOpFilter != "" {
			call = call.Filter(s.flagComposerOpFilter)
		}
		if s.flagComposerOpPageSize > 0 {
			call = call.PageSize(s.flagComposerOpPageSize)
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
	if s.flagComposerOpFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagComposerOpFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-50s %s\n", "NAME", "DONE")
	for _, o := range all {
		fmt.Fprintf(s.inv.stdout, "%-50s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

func (s *composerCmds) runComposerOpWait(cmd *cobra.Command, args []string) error {
	name, err := s.composerOpName(args[0])
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), s.flagComposerOpTimeout)
	defer cancel()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return err
	}
	backoff := 2 * time.Second
	for {
		op, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling operation: %w", err)
		}
		if op.Done {
			if op.Error != nil {
				return fmt.Errorf("operation failed: %s", op.Error.Message)
			}
			fmt.Fprintf(s.inv.stdout, "Operation %s completed.\n", args[0])
			return emitFormattedTo(s.inv.stdout, op, s.flagComposerOpFormat)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for operation %s: %w", args[0], ctx.Err())
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff = time.Duration(float64(backoff) * 1.5)
		}
	}
}
