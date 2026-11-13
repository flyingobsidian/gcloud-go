package cli

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	dataproc "google.golang.org/api/dataproc/v1"
)

// --- gcloud dataproc operations (#1515) ---

func (s *dataprocCmds) initDataprocOperations() {
	all := []*cobra.Command{
		s.dpOpCancelCmd, s.dpOpDeleteCmd, s.dpOpDescribeCmd, s.dpOpListCmd, s.dpOpGetIamCmd, s.dpOpSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDPOpRegion, "region", "", "Dataproc region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagDPOpFormat, "format", "", "Output format")
	}
	s.dpOpListCmd.Flags().StringVar(&s.flagDPOpFilter, "filter", "", "Server-side filter expression")
	s.dpOpListCmd.Flags().Int64Var(&s.flagDPOpPageSize, "page-size", 0, "Maximum results per page")

	s.dpOpCmd.AddCommand(all...)
	s.dataprocCmd.AddCommand(s.dpOpCmd)
}

func (s *dataprocCmds) dpOpParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/regions/%s/operations", project, s.flagDPOpRegion), nil
}

func (s *dataprocCmds) dpOpName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.dpOpParent()
	if err != nil {
		return "", err
	}
	return parent + "/" + id, nil
}

func (s *dataprocCmds) runDPOpCancel(cmd *cobra.Command, args []string) error {
	name, err := s.dpOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPOpRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Regions.Operations.Cancel(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancel request issued for operation %s.\n", args[0])
	return nil
}

func (s *dataprocCmds) runDPOpDelete(cmd *cobra.Command, args []string) error {
	name, err := s.dpOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPOpRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Regions.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted operation %s.\n", args[0])
	return nil
}

func (s *dataprocCmds) runDPOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.dpOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPOpRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Regions.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagDPOpFormat)
}

func (s *dataprocCmds) runDPOpList(cmd *cobra.Command, args []string) error {
	parent, err := s.dpOpParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPOpRegion)
	if err != nil {
		return err
	}
	var all []*dataproc.Operation
	pageToken := ""
	for {
		call := svc.Projects.Regions.Operations.List(parent).Context(ctx)
		if s.flagDPOpFilter != "" {
			call = call.Filter(s.flagDPOpFilter)
		}
		if s.flagDPOpPageSize > 0 {
			call = call.PageSize(s.flagDPOpPageSize)
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
	if s.flagDPOpFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagDPOpFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-50s %s\n", "NAME", "DONE")
	for _, o := range all {
		fmt.Fprintf(s.inv.stdout, "%-50s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

func (s *dataprocCmds) runDPOpGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.dpOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPOpRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Regions.Operations.GetIamPolicy(name, &dataproc.GetIamPolicyRequest{
		Options: &dataproc.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagDPOpFormat)
}

func (s *dataprocCmds) runDPOpSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.dpOpName(args[0])
	if err != nil {
		return err
	}
	policy := &dataproc.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPOpRegion)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Regions.Operations.SetIamPolicy(name, &dataproc.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.dpUpdatedIam(fmt.Sprintf("operation [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDPOpFormat)
}
