package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	healthcare "google.golang.org/api/healthcare/v1"
)

// --- gcloud healthcare datasets (#1221) ---

var healthcareDatasetsCmd = &cobra.Command{Use: "datasets", Short: "Manage Cloud Healthcare datasets"}

var (
	flagHcDsLocation    string
	flagHcDsFormat      string
	flagHcDsConfigFile  string
	flagHcDsUpdateMask  string
	flagHcDsPageSize    int64
	flagHcDsDestDataset string
)

var (
	healthcareDsCreateCmd = &cobra.Command{
		Use: "create DATASET", Short: "Create a healthcare dataset",
		Args: cobra.ExactArgs(1), RunE: runHcDsCreate,
	}
	healthcareDsDeleteCmd = &cobra.Command{
		Use: "delete DATASET", Short: "Delete a healthcare dataset",
		Args: cobra.ExactArgs(1), RunE: runHcDsDelete,
	}
	healthcareDsDescribeCmd = &cobra.Command{
		Use: "describe DATASET", Short: "Describe a healthcare dataset",
		Args: cobra.ExactArgs(1), RunE: runHcDsDescribe,
	}
	healthcareDsListCmd = &cobra.Command{
		Use: "list", Short: "List healthcare datasets in a location",
		Args: cobra.NoArgs, RunE: runHcDsList,
	}
	healthcareDsUpdateCmd = &cobra.Command{
		Use: "update DATASET", Short: "Update a healthcare dataset",
		Args: cobra.ExactArgs(1), RunE: runHcDsUpdate,
	}
	healthcareDsGetIamCmd = &cobra.Command{
		Use: "get-iam-policy DATASET", Short: "Get the IAM policy for a healthcare dataset",
		Args: cobra.ExactArgs(1), RunE: runHcDsGetIam,
	}
	healthcareDsSetIamCmd = &cobra.Command{
		Use: "set-iam-policy DATASET POLICY_FILE", Short: "Set the IAM policy for a healthcare dataset",
		Args: cobra.ExactArgs(2), RunE: runHcDsSetIam,
	}
	healthcareDsDeidentifyCmd = &cobra.Command{
		Use: "deidentify DATASET", Short: "De-identify a healthcare dataset into another dataset",
		Args: cobra.ExactArgs(1), RunE: runHcDsDeidentify,
	}
)

func init() {
	all := []*cobra.Command{
		healthcareDsCreateCmd, healthcareDsDeleteCmd, healthcareDsDescribeCmd,
		healthcareDsListCmd, healthcareDsUpdateCmd, healthcareDsGetIamCmd,
		healthcareDsSetIamCmd, healthcareDsDeidentifyCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagHcDsLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagHcDsFormat, "format", "", "Output format")
	}
	healthcareDsCreateCmd.Flags().StringVar(&flagHcDsConfigFile, "config-file", "", "Optional YAML/JSON file with the Dataset body")
	healthcareDsListCmd.Flags().Int64Var(&flagHcDsPageSize, "page-size", 0, "Maximum results per page")
	healthcareDsUpdateCmd.Flags().StringVar(&flagHcDsConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = healthcareDsUpdateCmd.MarkFlagRequired("config-file")
	healthcareDsUpdateCmd.Flags().StringVar(&flagHcDsUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	healthcareDsDeidentifyCmd.Flags().StringVar(&flagHcDsConfigFile, "config-file", "", "YAML/JSON file with the DeidentifyDatasetRequest body (required)")
	_ = healthcareDsDeidentifyCmd.MarkFlagRequired("config-file")
	healthcareDsDeidentifyCmd.Flags().StringVar(&flagHcDsDestDataset, "destination-dataset", "", "Fully qualified destination dataset (overrides destinationDataset in the config file)")

	healthcareDatasetsCmd.AddCommand(all...)
	healthcareCmd.AddCommand(healthcareDatasetsCmd)
}

func runHcDsCreate(cmd *cobra.Command, args []string) error {
	parent, err := hcLocationParent(flagHcDsLocation)
	if err != nil {
		return err
	}
	body := &healthcare.Dataset{}
	if flagHcDsConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagHcDsConfigFile, body); err != nil {
			return err
		}
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.Create(parent, body).DatasetId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating dataset: %w", err)
	}
	fmt.Printf("Created dataset [%s].\n", args[0])
	return emitFormatted(got, flagHcDsFormat)
}

func runHcDsDelete(cmd *cobra.Command, args []string) error {
	name, err := hcDatasetName(flagHcDsLocation, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Datasets.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting dataset: %w", err)
	}
	fmt.Printf("Deleted dataset [%s].\n", args[0])
	return nil
}

func runHcDsDescribe(cmd *cobra.Command, args []string) error {
	name, err := hcDatasetName(flagHcDsLocation, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing dataset: %w", err)
	}
	return emitFormatted(got, flagHcDsFormat)
}

func runHcDsList(cmd *cobra.Command, args []string) error {
	parent, err := hcLocationParent(flagHcDsLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*healthcare.Dataset
	pageToken := ""
	for {
		call := svc.Projects.Locations.Datasets.List(parent).Context(ctx)
		if flagHcDsPageSize > 0 {
			call = call.PageSize(flagHcDsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing datasets: %w", err)
		}
		all = append(all, resp.Datasets...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagHcDsFormat)
}

func runHcDsUpdate(cmd *cobra.Command, args []string) error {
	name, err := hcDatasetName(flagHcDsLocation, args[0])
	if err != nil {
		return err
	}
	body := &healthcare.Dataset{}
	if err := loadYAMLOrJSONInto(flagHcDsConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagHcDsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Datasets.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating dataset: %w", err)
	}
	fmt.Printf("Updated dataset [%s].\n", args[0])
	return emitFormatted(got, flagHcDsFormat)
}

func runHcDsGetIam(cmd *cobra.Command, args []string) error {
	name, err := hcDatasetName(flagHcDsLocation, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Datasets.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagHcDsFormat)
}

func runHcDsSetIam(cmd *cobra.Command, args []string) error {
	name, err := hcDatasetName(flagHcDsLocation, args[0])
	if err != nil {
		return err
	}
	policy := &healthcare.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.Datasets.SetIamPolicy(name, &healthcare.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for dataset [%s].\n", args[0])
	return emitFormatted(updated, flagHcDsFormat)
}

func runHcDsDeidentify(cmd *cobra.Command, args []string) error {
	src, err := hcDatasetName(flagHcDsLocation, args[0])
	if err != nil {
		return err
	}
	body := &healthcare.DeidentifyDatasetRequest{}
	if err := loadYAMLOrJSONInto(flagHcDsConfigFile, body); err != nil {
		return err
	}
	if flagHcDsDestDataset != "" {
		body.DestinationDataset = flagHcDsDestDataset
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Datasets.Deidentify(src, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deidentifying dataset: %w", err)
	}
	fmt.Printf("Deidentify dataset [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagHcDsFormat)
}
