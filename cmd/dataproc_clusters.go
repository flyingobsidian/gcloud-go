package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	dataproc "google.golang.org/api/dataproc/v1"
)

// --- gcloud dataproc clusters (#1512) ---

var dpClustersCmd = &cobra.Command{Use: "clusters", Short: "Manage Dataproc clusters"}

var dpClustersGKECmd = &cobra.Command{Use: "gke", Short: "Manage GKE-based Dataproc clusters"}

var (
	flagDPClusterRegion       string
	flagDPClusterFormat       string
	flagDPClusterConfigFile   string
	flagDPClusterUpdateMask   string
	flagDPClusterFilter       string
	flagDPClusterPageSize     int64
	flagDPClusterRequestID    string
	flagDPClusterGracefulTO   string
	flagDPClusterDiagnoseJobs string
	flagDPClusterDiagnoseYARN string
	flagDPClusterDiagnoseTS   string
)

var (
	dpClusterCreateCmd = &cobra.Command{
		Use: "create CLUSTER", Short: "Create a Dataproc cluster",
		Args: cobra.ExactArgs(1), RunE: runDPClusterCreate,
	}
	dpClusterDeleteCmd = &cobra.Command{
		Use: "delete CLUSTER", Short: "Delete a Dataproc cluster",
		Args: cobra.ExactArgs(1), RunE: runDPClusterDelete,
	}
	dpClusterDescribeCmd = &cobra.Command{
		Use: "describe CLUSTER", Short: "Describe a Dataproc cluster",
		Args: cobra.ExactArgs(1), RunE: runDPClusterDescribe,
	}
	dpClusterDiagnoseCmd = &cobra.Command{
		Use: "diagnose CLUSTER", Short: "Diagnose a Dataproc cluster",
		Args: cobra.ExactArgs(1), RunE: runDPClusterDiagnose,
	}
	dpClusterExportCmd = &cobra.Command{
		Use: "export CLUSTER", Short: "Export a Dataproc cluster (returns YAML/JSON)",
		Args: cobra.ExactArgs(1), RunE: runDPClusterExport,
	}
	dpClusterImportCmd = &cobra.Command{
		Use: "import CLUSTER", Short: "Import (create or update) a cluster from a YAML/JSON file",
		Args: cobra.ExactArgs(1), RunE: runDPClusterImport,
	}
	dpClusterListCmd = &cobra.Command{
		Use: "list", Short: "List Dataproc clusters",
		Args: cobra.NoArgs, RunE: runDPClusterList,
	}
	dpClusterStartCmd = &cobra.Command{
		Use: "start CLUSTER", Short: "Start a stopped Dataproc cluster",
		Args: cobra.ExactArgs(1), RunE: runDPClusterStart,
	}
	dpClusterStopCmd = &cobra.Command{
		Use: "stop CLUSTER", Short: "Stop a Dataproc cluster",
		Args: cobra.ExactArgs(1), RunE: runDPClusterStop,
	}
	dpClusterUpdateCmd = &cobra.Command{
		Use: "update CLUSTER", Short: "Update a Dataproc cluster",
		Args: cobra.ExactArgs(1), RunE: runDPClusterUpdate,
	}
	dpClusterGetIamCmd = &cobra.Command{
		Use: "get-iam-policy CLUSTER", Short: "Get the IAM policy for a Dataproc cluster",
		Args: cobra.ExactArgs(1), RunE: runDPClusterGetIam,
	}
	dpClusterSetIamCmd = &cobra.Command{
		Use: "set-iam-policy CLUSTER POLICY_FILE", Short: "Set the IAM policy for a Dataproc cluster",
		Args: cobra.ExactArgs(2), RunE: runDPClusterSetIam,
	}
	dpClusterGKECreateCmd = &cobra.Command{
		Use: "create CLUSTER",
		Short: "Create a GKE-based Dataproc cluster (--config-file must set virtualClusterConfig)",
		Args: cobra.ExactArgs(1), RunE: runDPClusterCreate,
	}
)

func init() {
	all := []*cobra.Command{
		dpClusterCreateCmd, dpClusterDeleteCmd, dpClusterDescribeCmd, dpClusterDiagnoseCmd,
		dpClusterExportCmd, dpClusterImportCmd, dpClusterListCmd, dpClusterStartCmd,
		dpClusterStopCmd, dpClusterUpdateCmd, dpClusterGetIamCmd, dpClusterSetIamCmd,
		dpClusterGKECreateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDPClusterRegion, "region", "", "Dataproc region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagDPClusterFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{dpClusterCreateCmd, dpClusterGKECreateCmd, dpClusterUpdateCmd, dpClusterImportCmd} {
		c.Flags().StringVar(&flagDPClusterConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Cluster body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	for _, c := range []*cobra.Command{dpClusterCreateCmd, dpClusterGKECreateCmd, dpClusterUpdateCmd, dpClusterDeleteCmd, dpClusterStartCmd, dpClusterStopCmd} {
		c.Flags().StringVar(&flagDPClusterRequestID, "request-id", "",
			"Optional client-supplied ID for idempotent submission")
	}
	dpClusterUpdateCmd.Flags().StringVar(&flagDPClusterUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	dpClusterUpdateCmd.Flags().StringVar(&flagDPClusterGracefulTO, "graceful-decommission-timeout", "",
		"Timeout for graceful decommission of workers (e.g. 10m)")
	dpClusterDeleteCmd.Flags().StringVar(&flagDPClusterGracefulTO, "graceful-termination-timeout", "",
		"Timeout for graceful termination (e.g. 10m)")
	dpClusterListCmd.Flags().StringVar(&flagDPClusterFilter, "filter", "", "Server-side filter expression")
	dpClusterListCmd.Flags().Int64Var(&flagDPClusterPageSize, "page-size", 0, "Maximum results per page")
	dpClusterDiagnoseCmd.Flags().StringVar(&flagDPClusterDiagnoseJobs, "job-ids", "",
		"Comma-separated job IDs to include in the diagnostic tarball")
	dpClusterDiagnoseCmd.Flags().StringVar(&flagDPClusterDiagnoseYARN, "yarn-application-ids", "",
		"Comma-separated YARN application IDs to include")
	dpClusterDiagnoseCmd.Flags().StringVar(&flagDPClusterDiagnoseTS, "tarball-gcs-dir", "",
		"Optional GCS dir where the tarball should be written")

	dpClustersCmd.AddCommand(dpClusterCreateCmd, dpClusterDeleteCmd, dpClusterDescribeCmd,
		dpClusterDiagnoseCmd, dpClusterExportCmd, dpClusterImportCmd, dpClusterListCmd,
		dpClusterStartCmd, dpClusterStopCmd, dpClusterUpdateCmd, dpClusterGetIamCmd, dpClusterSetIamCmd)
	dpClustersGKECmd.AddCommand(dpClusterGKECreateCmd)
	dpClustersCmd.AddCommand(dpClustersGKECmd)
	dataprocCmd.AddCommand(dpClustersCmd)
}

func dpClusterResourceName(project, region, cluster string) string {
	return fmt.Sprintf("projects/%s/regions/%s/clusters/%s", project, region, cluster)
}

func runDPClusterCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &dataproc.Cluster{}
	if err := loadYAMLOrJSONInto(flagDPClusterConfigFile, body); err != nil {
		return err
	}
	body.ClusterName = args[0]
	body.ProjectId = project
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPClusterRegion)
	if err != nil {
		return err
	}
	call := svc.Projects.Regions.Clusters.Create(project, flagDPClusterRegion, body).Context(ctx)
	if flagDPClusterRequestID != "" {
		call = call.RequestId(flagDPClusterRequestID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	fmt.Printf("Create request issued for cluster [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDPClusterFormat)
}

func runDPClusterDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPClusterRegion)
	if err != nil {
		return err
	}
	call := svc.Projects.Regions.Clusters.Delete(project, flagDPClusterRegion, args[0]).Context(ctx)
	if flagDPClusterRequestID != "" {
		call = call.RequestId(flagDPClusterRequestID)
	}
	if flagDPClusterGracefulTO != "" {
		call = call.GracefulTerminationTimeout(flagDPClusterGracefulTO)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting cluster: %w", err)
	}
	fmt.Printf("Delete request issued for cluster [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDPClusterFormat)
}

func runDPClusterDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPClusterRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.Clusters.Get(project, flagDPClusterRegion, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cluster: %w", err)
	}
	return emitFormatted(got, flagDPClusterFormat)
}

func runDPClusterDiagnose(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &dataproc.DiagnoseClusterRequest{
		TarballGcsDir: flagDPClusterDiagnoseTS,
	}
	if flagDPClusterDiagnoseJobs != "" {
		req.Jobs = splitCSV(flagDPClusterDiagnoseJobs)
	}
	if flagDPClusterDiagnoseYARN != "" {
		req.YarnApplicationIds = splitCSV(flagDPClusterDiagnoseYARN)
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPClusterRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Regions.Clusters.Diagnose(project, flagDPClusterRegion, args[0], req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("diagnosing cluster: %w", err)
	}
	fmt.Printf("Diagnose request issued for cluster [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDPClusterFormat)
}

func runDPClusterExport(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPClusterRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.Clusters.Get(project, flagDPClusterRegion, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting cluster: %w", err)
	}
	format := flagDPClusterFormat
	if format == "" {
		format = "yaml"
	}
	return emitFormatted(got, format)
}

func runDPClusterImport(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &dataproc.Cluster{}
	if err := loadYAMLOrJSONInto(flagDPClusterConfigFile, body); err != nil {
		return err
	}
	body.ClusterName = args[0]
	body.ProjectId = project
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPClusterRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Regions.Clusters.Get(project, flagDPClusterRegion, args[0]).Context(ctx).Do(); err == nil {
		mask := flagDPClusterUpdateMask
		if mask == "" {
			mask = joinMask(nonEmptyJSONFields(body))
		}
		call := svc.Projects.Regions.Clusters.Patch(project, flagDPClusterRegion, args[0], body).Context(ctx)
		if mask != "" {
			call = call.UpdateMask(mask)
		}
		op, err := call.Do()
		if err != nil {
			return fmt.Errorf("updating cluster: %w", err)
		}
		fmt.Printf("Update request issued for cluster [%s] (operation: %s).\n", args[0], op.Name)
		return emitFormatted(op, flagDPClusterFormat)
	}
	op, err := svc.Projects.Regions.Clusters.Create(project, flagDPClusterRegion, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}
	fmt.Printf("Create request issued for cluster [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDPClusterFormat)
}

func runDPClusterList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPClusterRegion)
	if err != nil {
		return err
	}
	var all []*dataproc.Cluster
	pageToken := ""
	for {
		call := svc.Projects.Regions.Clusters.List(project, flagDPClusterRegion).Context(ctx)
		if flagDPClusterFilter != "" {
			call = call.Filter(flagDPClusterFilter)
		}
		if flagDPClusterPageSize > 0 {
			call = call.PageSize(flagDPClusterPageSize)
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
	return emitFormatted(all, flagDPClusterFormat)
}

func runDPClusterStart(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPClusterRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Regions.Clusters.Start(project, flagDPClusterRegion, args[0], &dataproc.StartClusterRequest{
		RequestId: flagDPClusterRequestID,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting cluster: %w", err)
	}
	fmt.Printf("Start request issued for cluster [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDPClusterFormat)
}

func runDPClusterStop(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPClusterRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Regions.Clusters.Stop(project, flagDPClusterRegion, args[0], &dataproc.StopClusterRequest{
		RequestId: flagDPClusterRequestID,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping cluster: %w", err)
	}
	fmt.Printf("Stop request issued for cluster [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDPClusterFormat)
}

func runDPClusterUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &dataproc.Cluster{}
	if err := loadYAMLOrJSONInto(flagDPClusterConfigFile, body); err != nil {
		return err
	}
	body.ClusterName = args[0]
	body.ProjectId = project
	mask := flagDPClusterUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPClusterRegion)
	if err != nil {
		return err
	}
	call := svc.Projects.Regions.Clusters.Patch(project, flagDPClusterRegion, args[0], body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	if flagDPClusterRequestID != "" {
		call = call.RequestId(flagDPClusterRequestID)
	}
	if flagDPClusterGracefulTO != "" {
		call = call.GracefulDecommissionTimeout(flagDPClusterGracefulTO)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating cluster: %w", err)
	}
	fmt.Printf("Update request issued for cluster [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDPClusterFormat)
}

func runDPClusterGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	name := dpClusterResourceName(project, flagDPClusterRegion, args[0])
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPClusterRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Regions.Clusters.GetIamPolicy(name, &dataproc.GetIamPolicyRequest{
		Options: &dataproc.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagDPClusterFormat)
}

func runDPClusterSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	name := dpClusterResourceName(project, flagDPClusterRegion, args[0])
	policy := &dataproc.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPClusterRegion)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Regions.Clusters.SetIamPolicy(name, &dataproc.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	dpUpdatedIam(fmt.Sprintf("cluster [%s]", args[0]))
	return emitFormatted(updated, flagDPClusterFormat)
}
