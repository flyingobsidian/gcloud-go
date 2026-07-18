package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	dataproc "google.golang.org/api/dataproc/v1"
)

// --- gcloud dataproc autoscaling-policies (#1510) ---

var dpASPCmd = &cobra.Command{Use: "autoscaling-policies", Short: "Manage Dataproc autoscaling policies"}

var (
	flagDPASPRegion     string
	flagDPASPFormat     string
	flagDPASPConfigFile string
	flagDPASPPageSize   int64
)

var (
	dpASPDeleteCmd = &cobra.Command{
		Use: "delete POLICY", Short: "Delete an autoscaling policy",
		Args: cobra.ExactArgs(1), RunE: runDPASPDelete,
	}
	dpASPDescribeCmd = &cobra.Command{
		Use: "describe POLICY", Short: "Describe an autoscaling policy",
		Args: cobra.ExactArgs(1), RunE: runDPASPDescribe,
	}
	dpASPExportCmd = &cobra.Command{
		Use: "export POLICY", Short: "Export an autoscaling policy (returns JSON)",
		Args: cobra.ExactArgs(1), RunE: runDPASPExport,
	}
	dpASPImportCmd = &cobra.Command{
		Use: "import POLICY", Short: "Import (create or update) an autoscaling policy from a YAML/JSON file",
		Args: cobra.ExactArgs(1), RunE: runDPASPImport,
	}
	dpASPListCmd = &cobra.Command{
		Use: "list", Short: "List autoscaling policies",
		Args: cobra.NoArgs, RunE: runDPASPList,
	}
	dpASPGetIamCmd = &cobra.Command{
		Use: "get-iam-policy POLICY", Short: "Get the IAM policy for an autoscaling policy",
		Args: cobra.ExactArgs(1), RunE: runDPASPGetIam,
	}
	dpASPSetIamCmd = &cobra.Command{
		Use: "set-iam-policy POLICY POLICY_FILE", Short: "Set the IAM policy for an autoscaling policy",
		Args: cobra.ExactArgs(2), RunE: runDPASPSetIam,
	}
)

func init() {
	all := []*cobra.Command{
		dpASPDeleteCmd, dpASPDescribeCmd, dpASPExportCmd, dpASPImportCmd, dpASPListCmd,
		dpASPGetIamCmd, dpASPSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDPASPRegion, "region", "", "Dataproc region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagDPASPFormat, "format", "", "Output format")
	}
	dpASPImportCmd.Flags().StringVar(&flagDPASPConfigFile, "source", "",
		"Path to a YAML/JSON file with the AutoscalingPolicy body (required)")
	_ = dpASPImportCmd.MarkFlagRequired("source")
	dpASPListCmd.Flags().Int64Var(&flagDPASPPageSize, "page-size", 0, "Maximum results per page")

	dpASPCmd.AddCommand(all...)
	dataprocCmd.AddCommand(dpASPCmd)
}

func dpASPParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return dpRegionParent(project, flagDPASPRegion), nil
}

func dpASPName(id string) (string, error) {
	parent, err := dpASPParent()
	if err != nil {
		return "", err
	}
	return dpChild("autoscalingPolicies", id, parent), nil
}

func runDPASPDelete(cmd *cobra.Command, args []string) error {
	name, err := dpASPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPASPRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Regions.AutoscalingPolicies.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting autoscaling policy: %w", err)
	}
	fmt.Printf("Deleted autoscaling policy [%s].\n", args[0])
	return nil
}

func runDPASPDescribe(cmd *cobra.Command, args []string) error {
	name, err := dpASPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPASPRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.AutoscalingPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing autoscaling policy: %w", err)
	}
	return emitFormatted(got, flagDPASPFormat)
}

func runDPASPExport(cmd *cobra.Command, args []string) error {
	name, err := dpASPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPASPRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Regions.AutoscalingPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting autoscaling policy: %w", err)
	}
	format := flagDPASPFormat
	if format == "" {
		format = "yaml"
	}
	return emitFormatted(got, format)
}

func runDPASPImport(cmd *cobra.Command, args []string) error {
	parent, err := dpASPParent()
	if err != nil {
		return err
	}
	body := &dataproc.AutoscalingPolicy{}
	if err := loadYAMLOrJSONInto(flagDPASPConfigFile, body); err != nil {
		return err
	}
	body.Id = args[0]
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPASPRegion)
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
		fmt.Printf("Updated autoscaling policy [%s].\n", args[0])
		return emitFormatted(got, flagDPASPFormat)
	}
	got, err := svc.Projects.Regions.AutoscalingPolicies.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating autoscaling policy: %w", err)
	}
	fmt.Printf("Created autoscaling policy [%s].\n", args[0])
	return emitFormatted(got, flagDPASPFormat)
}

func runDPASPList(cmd *cobra.Command, args []string) error {
	parent, err := dpASPParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPASPRegion)
	if err != nil {
		return err
	}
	var all []*dataproc.AutoscalingPolicy
	pageToken := ""
	for {
		call := svc.Projects.Regions.AutoscalingPolicies.List(parent).Context(ctx)
		if flagDPASPPageSize > 0 {
			call = call.PageSize(flagDPASPPageSize)
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
	return emitFormatted(all, flagDPASPFormat)
}

func runDPASPGetIam(cmd *cobra.Command, args []string) error {
	name, err := dpASPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPASPRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Regions.AutoscalingPolicies.GetIamPolicy(name, &dataproc.GetIamPolicyRequest{
		Options: &dataproc.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagDPASPFormat)
}

func runDPASPSetIam(cmd *cobra.Command, args []string) error {
	name, err := dpASPName(args[0])
	if err != nil {
		return err
	}
	policy := &dataproc.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPASPRegion)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Regions.AutoscalingPolicies.SetIamPolicy(name, &dataproc.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	dpUpdatedIam(fmt.Sprintf("autoscaling policy [%s]", args[0]))
	return emitFormatted(updated, flagDPASPFormat)
}
