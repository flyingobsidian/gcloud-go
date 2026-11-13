package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	container "google.golang.org/api/container/v1"
)

// --- gcloud container operations (#1141) ---

func (s *containerCmds) initContainerOperations() {
	all := []*cobra.Command{s.containerOpCancelCmd, s.containerOpDescribeCmd, s.containerOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagCtnOpLocation, "location", "", "Location (region or zone) (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagCtnOpFormat, "format", "", "Output format")
	}

	s.containerOperationsCmd.AddCommand(all...)
	s.containerCmd.AddCommand(s.containerOperationsCmd)
}

func (s *containerCmds) ctnOpParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	if s.flagCtnOpLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagCtnOpLocation), nil
}

func (s *containerCmds) ctnOpName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.ctnOpParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/operations/%s", parent, id), nil
}

func (s *containerCmds) runCtnOpCancel(cmd *cobra.Command, args []string) error {
	name, err := s.ctnOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Container(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(name, &container.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (s *containerCmds) runCtnOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.ctnOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Container(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagCtnOpFormat)
}

func (s *containerCmds) runCtnOpList(cmd *cobra.Command, args []string) error {
	parent, err := s.ctnOpParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Container(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp.Operations, s.flagCtnOpFormat)
}
