package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	dataproc "google.golang.org/api/dataproc/v1"
)

// --- gcloud dataproc clusters (#1512) ---

func (s *dataprocCmds) initDataprocClusters() {
	all := []*cobra.Command{
		s.dpClusterCreateCmd, s.dpClusterDeleteCmd, s.dpClusterDescribeCmd, s.dpClusterDiagnoseCmd,
		s.dpClusterExportCmd, s.dpClusterImportCmd, s.dpClusterListCmd, s.dpClusterStartCmd,
		s.dpClusterStopCmd, s.dpClusterUpdateCmd, s.dpClusterGetIamCmd, s.dpClusterSetIamCmd,
		s.dpClusterGKECreateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDPClusterRegion, "region", "", "Dataproc region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagDPClusterFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.dpClusterCreateCmd, s.dpClusterGKECreateCmd, s.dpClusterUpdateCmd, s.dpClusterImportCmd} {
		c.Flags().StringVar(&s.flagDPClusterConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Cluster body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	for _, c := range []*cobra.Command{s.dpClusterCreateCmd, s.dpClusterGKECreateCmd, s.dpClusterUpdateCmd, s.dpClusterDeleteCmd, s.dpClusterStartCmd, s.dpClusterStopCmd} {
		c.Flags().StringVar(&s.flagDPClusterRequestID, "request-id", "",
			"Optional client-supplied ID for idempotent submission")
	}
	s.dpClusterUpdateCmd.Flags().StringVar(&s.flagDPClusterUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.dpClusterUpdateCmd.Flags().StringVar(&s.flagDPClusterGracefulTO, "graceful-decommission-timeout", "",
		"Timeout for graceful decommission of workers (e.g. 10m)")
	s.dpClusterDeleteCmd.Flags().StringVar(&s.flagDPClusterGracefulTO, "graceful-termination-timeout", "",
		"Timeout for graceful termination (e.g. 10m)")
	s.dpClusterListCmd.Flags().StringVar(&s.flagDPClusterFilter, "filter", "", "Server-side filter expression")
	s.dpClusterListCmd.Flags().Int64Var(&s.flagDPClusterPageSize, "page-size", 0, "Maximum results per page")
	s.dpClusterDiagnoseCmd.Flags().StringVar(&s.flagDPClusterDiagnoseJobs, "job-ids", "",
		"Comma-separated job IDs to include in the diagnostic tarball")
	s.dpClusterDiagnoseCmd.Flags().StringVar(&s.flagDPClusterDiagnoseYARN, "yarn-application-ids", "",
		"Comma-separated YARN application IDs to include")
	s.dpClusterDiagnoseCmd.Flags().StringVar(&s.flagDPClusterDiagnoseTS, "tarball-gcs-dir", "",
		"Optional GCS dir where the tarball should be written")

	s.dpClustersCmd.AddCommand(s.dpClusterCreateCmd, s.dpClusterDeleteCmd, s.dpClusterDescribeCmd,
		s.dpClusterDiagnoseCmd, s.dpClusterExportCmd, s.dpClusterImportCmd, s.dpClusterListCmd,
		s.dpClusterStartCmd, s.dpClusterStopCmd, s.dpClusterUpdateCmd, s.dpClusterGetIamCmd, s.dpClusterSetIamCmd)
	s.dpClustersGKECmd.AddCommand(s.dpClusterGKECreateCmd)
	s.dpClustersCmd.AddCommand(s.dpClustersGKECmd)
	s.dataprocCmd.AddCommand(s.dpClustersCmd)
}

func dpClusterResourceName(project, region, cluster string) string {
	return fmt.Sprintf("projects/%s/regions/%s/clusters/%s", project, region, cluster)
}

func (s *dataprocCmds) runDPClusterCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &dataproc.Cluster{}
	if err := loadYAMLOrJSONInto(s.flagDPClusterConfigFile, body); err != nil {
		return err
	}
	body.ClusterName = args[0]
	body.ProjectId = project
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPClusterRegion)
	if err != nil {
		return err
	}
	call := svc.Projects.Regions.Clusters.Create(project, s.flagDPClusterRegion, body).Context(ctx)
	if s.flagDPClusterRequestID != "" {
		call = call.RequestId(s.flagDPClusterRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for cluster [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDPClusterFormat)
}

func (s *dataprocCmds) runDPClusterDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPClusterRegion)
	if err != nil {
		return err
	}
	call := svc.Projects.Regions.Clusters.Delete(project, s.flagDPClusterRegion, args[0]).Context(ctx)
	if s.flagDPClusterRequestID != "" {
		call = call.RequestId(s.flagDPClusterRequestID)
	}
	if s.flagDPClusterGracefulTO != "" {
		call = call.GracefulTerminationTimeout(s.flagDPClusterGracefulTO)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for cluster [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDPClusterFormat)
}

func (s *dataprocCmds) runDPClusterDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPClusterRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.Clusters.Get(project, s.flagDPClusterRegion, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cluster: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDPClusterFormat)
}

func (s *dataprocCmds) runDPClusterDiagnose(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	req := &dataproc.DiagnoseClusterRequest{
		TarballGcsDir: s.flagDPClusterDiagnoseTS,
	}
	if s.flagDPClusterDiagnoseJobs != "" {
		req.Jobs = splitCSV(s.flagDPClusterDiagnoseJobs)
	}
	if s.flagDPClusterDiagnoseYARN != "" {
		req.YarnApplicationIds = splitCSV(s.flagDPClusterDiagnoseYARN)
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPClusterRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Regions.Clusters.Diagnose(project, s.flagDPClusterRegion, args[0], req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("diagnosing cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Diagnose request issued for cluster [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDPClusterFormat)
}

func (s *dataprocCmds) runDPClusterExport(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPClusterRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.Clusters.Get(project, s.flagDPClusterRegion, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting cluster: %w", err)
	}
	format := s.flagDPClusterFormat
	if format == "" {
		format = "yaml"
	}
	return emitFormattedTo(s.inv.stdout, got, format)
}

func (s *dataprocCmds) runDPClusterImport(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &dataproc.Cluster{}
	if err := loadYAMLOrJSONInto(s.flagDPClusterConfigFile, body); err != nil {
		return err
	}
	body.ClusterName = args[0]
	body.ProjectId = project
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPClusterRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Regions.Clusters.Get(project, s.flagDPClusterRegion, args[0]).Context(ctx).Do(); err == nil {
		mask := s.flagDPClusterUpdateMask
		if mask == "" {
			mask = joinMask(nonEmptyJSONFields(body))
		}
		call := svc.Projects.Regions.Clusters.Patch(project, s.flagDPClusterRegion, args[0], body).Context(ctx)
		if mask != "" {
			call = call.UpdateMask(mask)
		}
		op, err := call.Do()
		if err != nil {
			return fmt.Errorf("updating cluster: %w", err)
		}
		fmt.Fprintf(s.inv.stdout, "Update request issued for cluster [%s] (operation: %s).\n", args[0], op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagDPClusterFormat)
	}
	op, err := svc.Projects.Regions.Clusters.Create(project, s.flagDPClusterRegion, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for cluster [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDPClusterFormat)
}

func (s *dataprocCmds) runDPClusterList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPClusterRegion)
	if err != nil {
		return err
	}
	var all []*dataproc.Cluster
	pageToken := ""
	for {
		call := svc.Projects.Regions.Clusters.List(project, s.flagDPClusterRegion).Context(ctx)
		if s.flagDPClusterFilter != "" {
			call = call.Filter(s.flagDPClusterFilter)
		}
		if s.flagDPClusterPageSize > 0 {
			call = call.PageSize(s.flagDPClusterPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing clusters: %w", err)
		}
		all = append(all, resp.Clusters...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagDPClusterFormat)
}

func (s *dataprocCmds) runDPClusterStart(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPClusterRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Regions.Clusters.Start(project, s.flagDPClusterRegion, args[0], &dataproc.StartClusterRequest{
		RequestId: s.flagDPClusterRequestID,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Start request issued for cluster [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDPClusterFormat)
}

func (s *dataprocCmds) runDPClusterStop(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPClusterRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Regions.Clusters.Stop(project, s.flagDPClusterRegion, args[0], &dataproc.StopClusterRequest{
		RequestId: s.flagDPClusterRequestID,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Stop request issued for cluster [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDPClusterFormat)
}

func (s *dataprocCmds) runDPClusterUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &dataproc.Cluster{}
	if err := loadYAMLOrJSONInto(s.flagDPClusterConfigFile, body); err != nil {
		return err
	}
	body.ClusterName = args[0]
	body.ProjectId = project
	mask := s.flagDPClusterUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPClusterRegion)
	if err != nil {
		return err
	}
	call := svc.Projects.Regions.Clusters.Patch(project, s.flagDPClusterRegion, args[0], body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	if s.flagDPClusterRequestID != "" {
		call = call.RequestId(s.flagDPClusterRequestID)
	}
	if s.flagDPClusterGracefulTO != "" {
		call = call.GracefulDecommissionTimeout(s.flagDPClusterGracefulTO)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating cluster: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for cluster [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagDPClusterFormat)
}

func (s *dataprocCmds) runDPClusterGetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	name := dpClusterResourceName(project, s.flagDPClusterRegion, args[0])
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPClusterRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Regions.Clusters.GetIamPolicy(name, &dataproc.GetIamPolicyRequest{
		Options: &dataproc.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagDPClusterFormat)
}

func (s *dataprocCmds) runDPClusterSetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	name := dpClusterResourceName(project, s.flagDPClusterRegion, args[0])
	policy := &dataproc.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().Dataproc(ctx, s.flagDPClusterRegion)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Regions.Clusters.SetIamPolicy(name, &dataproc.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.dpUpdatedIam(fmt.Sprintf("cluster [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagDPClusterFormat)
}
