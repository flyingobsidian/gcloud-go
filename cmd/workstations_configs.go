package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	workstations "google.golang.org/api/workstations/v1"
)

// --- gcloud workstations configs (#1260) ---

var workstationsConfigsCmd = &cobra.Command{Use: "configs", Short: "Manage workstation configurations"}

var (
	flagWsCfgLocation   string
	flagWsCfgCluster    string
	flagWsCfgFormat     string
	flagWsCfgConfigFile string
	flagWsCfgUpdateMask string
	flagWsCfgPageSize   int64
)

var (
	workstationsCfgCreateCmd = &cobra.Command{
		Use: "create CONFIG", Short: "Create a workstation configuration",
		Args: cobra.ExactArgs(1), RunE: runWsCfgCreate,
	}
	workstationsCfgDeleteCmd = &cobra.Command{
		Use: "delete CONFIG", Short: "Delete a workstation configuration",
		Args: cobra.ExactArgs(1), RunE: runWsCfgDelete,
	}
	workstationsCfgDescribeCmd = &cobra.Command{
		Use: "describe CONFIG", Short: "Describe a workstation configuration",
		Args: cobra.ExactArgs(1), RunE: runWsCfgDescribe,
	}
	workstationsCfgListCmd = &cobra.Command{
		Use: "list", Short: "List workstation configurations in a cluster",
		Args: cobra.NoArgs, RunE: runWsCfgList,
	}
	workstationsCfgUpdateCmd = &cobra.Command{
		Use: "update CONFIG", Short: "Update a workstation configuration",
		Args: cobra.ExactArgs(1), RunE: runWsCfgUpdate,
	}
	workstationsCfgGetIamCmd = &cobra.Command{
		Use: "get-iam-policy CONFIG", Short: "Get the IAM policy for a workstation configuration",
		Args: cobra.ExactArgs(1), RunE: runWsCfgGetIam,
	}
	workstationsCfgSetIamCmd = &cobra.Command{
		Use: "set-iam-policy CONFIG POLICY_FILE", Short: "Set the IAM policy for a workstation configuration",
		Args: cobra.ExactArgs(2), RunE: runWsCfgSetIam,
	}
)

func init() {
	all := []*cobra.Command{
		workstationsCfgCreateCmd, workstationsCfgDeleteCmd,
		workstationsCfgDescribeCmd, workstationsCfgListCmd,
		workstationsCfgUpdateCmd, workstationsCfgGetIamCmd,
		workstationsCfgSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagWsCfgLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagWsCfgCluster, "cluster", "", "Workstation cluster (required)")
		_ = c.MarkFlagRequired("cluster")
		c.Flags().StringVar(&flagWsCfgFormat, "format", "", "Output format")
	}
	workstationsCfgCreateCmd.Flags().StringVar(&flagWsCfgConfigFile, "config-file", "", "YAML/JSON file with the WorkstationConfig body (required)")
	_ = workstationsCfgCreateCmd.MarkFlagRequired("config-file")
	workstationsCfgListCmd.Flags().Int64Var(&flagWsCfgPageSize, "page-size", 0, "Maximum results per page")
	workstationsCfgUpdateCmd.Flags().StringVar(&flagWsCfgConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = workstationsCfgUpdateCmd.MarkFlagRequired("config-file")
	workstationsCfgUpdateCmd.Flags().StringVar(&flagWsCfgUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	workstationsConfigsCmd.AddCommand(all...)
	workstationsCmd.AddCommand(workstationsConfigsCmd)
}

func runWsCfgCreate(cmd *cobra.Command, args []string) error {
	parent, err := wsClusterName(flagWsCfgLocation, flagWsCfgCluster)
	if err != nil {
		return err
	}
	body := &workstations.WorkstationConfig{}
	if err := loadYAMLOrJSONInto(flagWsCfgConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.WorkstationsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WorkstationClusters.WorkstationConfigs.Create(parent, body).WorkstationConfigId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating workstation configuration: %w", err)
	}
	fmt.Printf("Create workstation configuration [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagWsCfgFormat)
}

func runWsCfgDelete(cmd *cobra.Command, args []string) error {
	name, err := wsConfigName(flagWsCfgLocation, flagWsCfgCluster, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.WorkstationsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.WorkstationClusters.WorkstationConfigs.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting workstation configuration: %w", err)
	}
	fmt.Printf("Delete workstation configuration [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagWsCfgFormat)
}

func runWsCfgDescribe(cmd *cobra.Command, args []string) error {
	name, err := wsConfigName(flagWsCfgLocation, flagWsCfgCluster, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.WorkstationsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.WorkstationClusters.WorkstationConfigs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing workstation configuration: %w", err)
	}
	return emitFormatted(got, flagWsCfgFormat)
}

func runWsCfgList(cmd *cobra.Command, args []string) error {
	parent, err := wsClusterName(flagWsCfgLocation, flagWsCfgCluster)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.WorkstationsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*workstations.WorkstationConfig
	pageToken := ""
	for {
		call := svc.Projects.Locations.WorkstationClusters.WorkstationConfigs.List(parent).Context(ctx)
		if flagWsCfgPageSize > 0 {
			call = call.PageSize(flagWsCfgPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing workstation configurations: %w", err)
		}
		all = append(all, resp.WorkstationConfigs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagWsCfgFormat)
}

func runWsCfgUpdate(cmd *cobra.Command, args []string) error {
	name, err := wsConfigName(flagWsCfgLocation, flagWsCfgCluster, args[0])
	if err != nil {
		return err
	}
	body := &workstations.WorkstationConfig{}
	if err := loadYAMLOrJSONInto(flagWsCfgConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagWsCfgUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.WorkstationsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.WorkstationClusters.WorkstationConfigs.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating workstation configuration: %w", err)
	}
	fmt.Printf("Update workstation configuration [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagWsCfgFormat)
}

func runWsCfgGetIam(cmd *cobra.Command, args []string) error {
	name, err := wsConfigName(flagWsCfgLocation, flagWsCfgCluster, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.WorkstationsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.WorkstationClusters.WorkstationConfigs.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagWsCfgFormat)
}

func runWsCfgSetIam(cmd *cobra.Command, args []string) error {
	name, err := wsConfigName(flagWsCfgLocation, flagWsCfgCluster, args[0])
	if err != nil {
		return err
	}
	policy := &workstations.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.WorkstationsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.WorkstationClusters.WorkstationConfigs.SetIamPolicy(name, &workstations.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for workstation configuration [%s].\n", args[0])
	return emitFormatted(updated, flagWsCfgFormat)
}
