package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	healthcare "google.golang.org/api/healthcare/v1"
)

// --- gcloud healthcare hl7v2-stores (#1224) ---

var healthcareHl7v2StoresCmd = &cobra.Command{Use: "hl7v2-stores", Short: "Manage Healthcare HL7v2 stores"}

var (
	flagHcH7Location   string
	flagHcH7Dataset    string
	flagHcH7Format     string
	flagHcH7ConfigFile string
	flagHcH7UpdateMask string
	flagHcH7PageSize   int64
)

var (
	hcH7CreateCmd = &cobra.Command{
		Use: "create HL7V2_STORE", Short: "Create a Healthcare HL7v2 store",
		Args: cobra.ExactArgs(1), RunE: runHcH7Create,
	}
	hcH7DeleteCmd = &cobra.Command{
		Use: "delete HL7V2_STORE", Short: "Delete a Healthcare HL7v2 store",
		Args: cobra.ExactArgs(1), RunE: runHcH7Delete,
	}
	hcH7DescribeCmd = &cobra.Command{
		Use: "describe HL7V2_STORE", Short: "Describe a Healthcare HL7v2 store",
		Args: cobra.ExactArgs(1), RunE: runHcH7Describe,
	}
	hcH7ListCmd = &cobra.Command{
		Use: "list", Short: "List Healthcare HL7v2 stores in a dataset",
		Args: cobra.NoArgs, RunE: runHcH7List,
	}
	hcH7UpdateCmd = &cobra.Command{
		Use: "update HL7V2_STORE", Short: "Update a Healthcare HL7v2 store",
		Args: cobra.ExactArgs(1), RunE: runHcH7Update,
	}
	hcH7GetIamCmd = &cobra.Command{
		Use: "get-iam-policy HL7V2_STORE", Short: "Get the IAM policy for a Healthcare HL7v2 store",
		Args: cobra.ExactArgs(1), RunE: runHcH7GetIam,
	}
	hcH7SetIamCmd = &cobra.Command{
		Use: "set-iam-policy HL7V2_STORE POLICY_FILE", Short: "Set the IAM policy for a Healthcare HL7v2 store",
		Args: cobra.ExactArgs(2), RunE: runHcH7SetIam,
	}
)

func init() {
	all := []*cobra.Command{
		hcH7CreateCmd, hcH7DeleteCmd, hcH7DescribeCmd,
		hcH7ListCmd, hcH7UpdateCmd, hcH7GetIamCmd, hcH7SetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagHcH7Location, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagHcH7Dataset, "dataset", "", "Parent dataset (required)")
		_ = c.MarkFlagRequired("dataset")
		c.Flags().StringVar(&flagHcH7Format, "format", "", "Output format")
	}
	hcH7CreateCmd.Flags().StringVar(&flagHcH7ConfigFile, "config-file", "", "YAML/JSON file with the Hl7V2Store body (required)")
	_ = hcH7CreateCmd.MarkFlagRequired("config-file")
	hcH7ListCmd.Flags().Int64Var(&flagHcH7PageSize, "page-size", 0, "Maximum results per page")
	hcH7UpdateCmd.Flags().StringVar(&flagHcH7ConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = hcH7UpdateCmd.MarkFlagRequired("config-file")
	hcH7UpdateCmd.Flags().StringVar(&flagHcH7UpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	healthcareHl7v2StoresCmd.AddCommand(all...)
	healthcareCmd.AddCommand(healthcareHl7v2StoresCmd)
}

// hcHl7V2StoreName returns projects/.../datasets/DATASET/hl7V2Stores/ID.
func hcHl7V2StoreName(location, dataset, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	ds, err := hcDatasetName(location, dataset)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/hl7V2Stores/%s", ds, id), nil
}

func runHcH7Create(cmd *cobra.Command, args []string) error {
	parent, err := hcDatasetName(flagHcH7Location, flagHcH7Dataset)
	if err != nil {
		return err
	}
	body := &healthcare.Hl7V2Store{}
	if err := loadYAMLOrJSONInto(flagHcH7ConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.Hl7V2Stores.Create(parent, body).Hl7V2StoreId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating HL7v2 store: %w", err)
	}
	fmt.Printf("Created HL7v2 store [%s].\n", args[0])
	return emitFormatted(got, flagHcH7Format)
}

func runHcH7Delete(cmd *cobra.Command, args []string) error {
	name, err := hcHl7V2StoreName(flagHcH7Location, flagHcH7Dataset, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Datasets.Hl7V2Stores.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting HL7v2 store: %w", err)
	}
	fmt.Printf("Deleted HL7v2 store [%s].\n", args[0])
	return nil
}

func runHcH7Describe(cmd *cobra.Command, args []string) error {
	name, err := hcHl7V2StoreName(flagHcH7Location, flagHcH7Dataset, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.Hl7V2Stores.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing HL7v2 store: %w", err)
	}
	return emitFormatted(got, flagHcH7Format)
}

func runHcH7List(cmd *cobra.Command, args []string) error {
	parent, err := hcDatasetName(flagHcH7Location, flagHcH7Dataset)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*healthcare.Hl7V2Store
	pageToken := ""
	for {
		call := svc.Projects.Locations.Datasets.Hl7V2Stores.List(parent).Context(ctx)
		if flagHcH7PageSize > 0 {
			call = call.PageSize(flagHcH7PageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing HL7v2 stores: %w", err)
		}
		all = append(all, resp.Hl7V2Stores...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagHcH7Format)
}

func runHcH7Update(cmd *cobra.Command, args []string) error {
	name, err := hcHl7V2StoreName(flagHcH7Location, flagHcH7Dataset, args[0])
	if err != nil {
		return err
	}
	body := &healthcare.Hl7V2Store{}
	if err := loadYAMLOrJSONInto(flagHcH7ConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagHcH7UpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Datasets.Hl7V2Stores.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating HL7v2 store: %w", err)
	}
	fmt.Printf("Updated HL7v2 store [%s].\n", args[0])
	return emitFormatted(got, flagHcH7Format)
}

func runHcH7GetIam(cmd *cobra.Command, args []string) error {
	name, err := hcHl7V2StoreName(flagHcH7Location, flagHcH7Dataset, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Datasets.Hl7V2Stores.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagHcH7Format)
}

func runHcH7SetIam(cmd *cobra.Command, args []string) error {
	name, err := hcHl7V2StoreName(flagHcH7Location, flagHcH7Dataset, args[0])
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
	updated, err := svc.Projects.Locations.Datasets.Hl7V2Stores.SetIamPolicy(name, &healthcare.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for HL7v2 store [%s].\n", args[0])
	return emitFormatted(updated, flagHcH7Format)
}
