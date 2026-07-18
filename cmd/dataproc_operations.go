package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	dataproc "google.golang.org/api/dataproc/v1"
)

// --- gcloud dataproc operations (#1515) ---

var dpOpCmd = &cobra.Command{Use: "operations", Short: "Manage Dataproc operations"}

var (
	flagDPOpRegion   string
	flagDPOpFormat   string
	flagDPOpFilter   string
	flagDPOpPageSize int64
)

var (
	dpOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Dataproc operation",
		Args: cobra.ExactArgs(1), RunE: runDPOpCancel,
	}
	dpOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a Dataproc operation record",
		Args: cobra.ExactArgs(1), RunE: runDPOpDelete,
	}
	dpOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Dataproc operation",
		Args: cobra.ExactArgs(1), RunE: runDPOpDescribe,
	}
	dpOpListCmd = &cobra.Command{
		Use: "list", Short: "List Dataproc operations",
		Args: cobra.NoArgs, RunE: runDPOpList,
	}
	dpOpGetIamCmd = &cobra.Command{
		Use: "get-iam-policy OPERATION", Short: "Get the IAM policy for a Dataproc operation",
		Args: cobra.ExactArgs(1), RunE: runDPOpGetIam,
	}
	dpOpSetIamCmd = &cobra.Command{
		Use: "set-iam-policy OPERATION POLICY_FILE", Short: "Set the IAM policy for a Dataproc operation",
		Args: cobra.ExactArgs(2), RunE: runDPOpSetIam,
	}
)

func init() {
	all := []*cobra.Command{
		dpOpCancelCmd, dpOpDeleteCmd, dpOpDescribeCmd, dpOpListCmd, dpOpGetIamCmd, dpOpSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDPOpRegion, "region", "", "Dataproc region (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagDPOpFormat, "format", "", "Output format")
	}
	dpOpListCmd.Flags().StringVar(&flagDPOpFilter, "filter", "", "Server-side filter expression")
	dpOpListCmd.Flags().Int64Var(&flagDPOpPageSize, "page-size", 0, "Maximum results per page")

	dpOpCmd.AddCommand(all...)
	dataprocCmd.AddCommand(dpOpCmd)
}

func dpOpParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/regions/%s/operations", project, flagDPOpRegion), nil
}

func dpOpName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := dpOpParent()
	if err != nil {
		return "", err
	}
	return parent + "/" + id, nil
}

func runDPOpCancel(cmd *cobra.Command, args []string) error {
	name, err := dpOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPOpRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Regions.Operations.Cancel(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancel request issued for operation %s.\n", args[0])
	return nil
}

func runDPOpDelete(cmd *cobra.Command, args []string) error {
	name, err := dpOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPOpRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Regions.Operations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation %s.\n", args[0])
	return nil
}

func runDPOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := dpOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPOpRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Regions.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagDPOpFormat)
}

func runDPOpList(cmd *cobra.Command, args []string) error {
	parent, err := dpOpParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPOpRegion)
	if err != nil {
		return err
	}
	var all []*dataproc.Operation
	pageToken := ""
	for {
		call := svc.Projects.Regions.Operations.List(parent).Context(ctx)
		if flagDPOpFilter != "" {
			call = call.Filter(flagDPOpFilter)
		}
		if flagDPOpPageSize > 0 {
			call = call.PageSize(flagDPOpPageSize)
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
	if flagDPOpFormat != "" {
		return emitFormatted(all, flagDPOpFormat)
	}
	fmt.Printf("%-50s %s\n", "NAME", "DONE")
	for _, o := range all {
		fmt.Printf("%-50s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

func runDPOpGetIam(cmd *cobra.Command, args []string) error {
	name, err := dpOpName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPOpRegion)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Regions.Operations.GetIamPolicy(name, &dataproc.GetIamPolicyRequest{
		Options: &dataproc.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagDPOpFormat)
}

func runDPOpSetIam(cmd *cobra.Command, args []string) error {
	name, err := dpOpName(args[0])
	if err != nil {
		return err
	}
	policy := &dataproc.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.DataprocService(ctx, flagAccount, flagDPOpRegion)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Regions.Operations.SetIamPolicy(name, &dataproc.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	dpUpdatedIam(fmt.Sprintf("operation [%s]", args[0]))
	return emitFormatted(updated, flagDPOpFormat)
}
