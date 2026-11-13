package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	dataproc "google.golang.org/api/dataproc/v1"
)

// --- gcloud dataproc autoscaling-policies (#1510) ---

func (s *dataprocCmds) initDataprocAutoscalingPolicies() {
	all := []*cobra.Command{
		s.dpASPDeleteCmd, s.dpASPDescribeCmd, s.dpASPExportCmd, s.dpASPImportCmd, s.dpASPListCmd,
		s.dpASPGetIamCmd, s.dpASPSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDPASPRegion, "region", "", "Dataproc region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagDPASPFormat, "format", "", "Output format")
	}
	s.dpASPImportCmd.Flags().StringVar(&s.flagDPASPConfigFile, "source", "",
		"Path to a YAML/JSON file with the AutoscalingPolicy body (required)")
	_ = s.dpASPImportCmd.MarkFlagRequired("source")
	s.dpASPListCmd.Flags().Int64Var(&s.flagDPASPPageSize, "page-size", 0, "Maximum results per page")

	s.dpASPCmd.AddCommand(all...)
	s.dataprocCmd.AddCommand(s.dpASPCmd)
}

func (s *dataprocCmds) dpASPParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return dpRegionParent(project, s.flagDPASPRegion), nil
}

func (s *dataprocCmds) dpASPName(id string) (string, error) {
	parent, err := s.dpASPParent()
	if err != nil {
		return "", err
	}
	return dpChild("autoscalingPolicies", id, parent), nil
}

func (s *dataprocCmds) runDPASPDelete(cmd *cobra.Command, args []string) error {
	name, err := s.dpASPName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPASPRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Regions.AutoscalingPolicies.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting autoscaling policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted autoscaling policy [%s].\n", args[0])
	return nil
}

func (s *dataprocCmds) runDPASPDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.dpASPName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPASPRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.AutoscalingPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing autoscaling policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDPASPFormat)
}

func (s *dataprocCmds) runDPASPExport(cmd *cobra.Command, args []string) error {
	name, err := s.dpASPName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPASPRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.AutoscalingPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting autoscaling policy: %w", err)
	}
	format := s.flagDPASPFormat
	if format == "" {
		format = "yaml"
	}
	return emitFormattedTo(s.inv.stdout, got, format)
}

func (s *dataprocCmds) runDPASPImport(cmd *cobra.Command, args []string) error {
	parent, err := s.dpASPParent()
	if err != nil {
		return err
	}
	body := &dataproc.AutoscalingPolicy{}
	if err := loadYAMLOrJSONInto(s.flagDPASPConfigFile, body); err != nil {
		return err
	}
	body.Id = args[0]
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPASPRegion)
	if err != nil {
		return err
	}
	name := dpChild("autoscalingPolicies", args[0], parent)
	if existing, err := svc.Projects.Regions.AutoscalingPolicies.Get(name).Context(ctx).Do(); err == nil {
		body.Name = existing.Name
		got, err := svc.Projects.Regions.AutoscalingPolicies.Update(name, body).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("updating autoscaling policy: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Updated autoscaling policy [%s].\n", args[0])
		return emitFormattedTo(s.inv.stdout, got, s.flagDPASPFormat)
	}
	got, err := svc.Projects.Regions.AutoscalingPolicies.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating autoscaling policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created autoscaling policy [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagDPASPFormat)
}

func (s *dataprocCmds) runDPASPList(cmd *cobra.Command, args []string) error {
	parent, err := s.dpASPParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPASPRegion)
	if err != nil {
		return err
	}
	var all []*dataproc.AutoscalingPolicy
	pageToken := ""
	for {
		call := svc.Projects.Regions.AutoscalingPolicies.List(parent).Context(ctx)
		if s.flagDPASPPageSize > 0 {
			call = call.PageSize(s.flagDPASPPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing autoscaling policies: %w", err)
		}
		all = append(all, resp.Policies...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagDPASPFormat)
}

func (s *dataprocCmds) runDPASPGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.dpASPName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPASPRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Regions.AutoscalingPolicies.GetIamPolicy(name, &dataproc.GetIamPolicyRequest{
		Options: &dataproc.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagDPASPFormat)
}

func (s *dataprocCmds) runDPASPSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.dpASPName(args[0])
	if err != nil {
		return err
	}
	policy := &dataproc.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPASPRegion)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Regions.AutoscalingPolicies.SetIamPolicy(name, &dataproc.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.dpUpdatedIam(fmt.Sprintf("autoscaling policy [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDPASPFormat)
}
