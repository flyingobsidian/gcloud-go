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

// --- gcloud healthcare fhir-stores (#1223) ---

var healthcareFhirStoresCmd = &cobra.Command{Use: "fhir-stores", Short: "Manage Healthcare FHIR stores"}

var (
	flagHcFsLocation   string
	flagHcFsDataset    string
	flagHcFsFormat     string
	flagHcFsConfigFile string
	flagHcFsUpdateMask string
	flagHcFsPageSize   int64
)

var (
	hcFsCreateCmd = &cobra.Command{
		Use: "create FHIR_STORE", Short: "Create a Healthcare FHIR store",
		Args: cobra.ExactArgs(1), RunE: runHcFsCreate,
	}
	hcFsDeleteCmd = &cobra.Command{
		Use: "delete FHIR_STORE", Short: "Delete a Healthcare FHIR store",
		Args: cobra.ExactArgs(1), RunE: runHcFsDelete,
	}
	hcFsDescribeCmd = &cobra.Command{
		Use: "describe FHIR_STORE", Short: "Describe a Healthcare FHIR store",
		Args: cobra.ExactArgs(1), RunE: runHcFsDescribe,
	}
	hcFsListCmd = &cobra.Command{
		Use: "list", Short: "List Healthcare FHIR stores in a dataset",
		Args: cobra.NoArgs, RunE: runHcFsList,
	}
	hcFsUpdateCmd = &cobra.Command{
		Use: "update FHIR_STORE", Short: "Update a Healthcare FHIR store",
		Args: cobra.ExactArgs(1), RunE: runHcFsUpdate,
	}
	hcFsGetIamCmd = &cobra.Command{
		Use: "get-iam-policy FHIR_STORE", Short: "Get the IAM policy for a Healthcare FHIR store",
		Args: cobra.ExactArgs(1), RunE: runHcFsGetIam,
	}
	hcFsSetIamCmd = &cobra.Command{
		Use: "set-iam-policy FHIR_STORE POLICY_FILE", Short: "Set the IAM policy for a Healthcare FHIR store",
		Args: cobra.ExactArgs(2), RunE: runHcFsSetIam,
	}
)

func init() {
	all := []*cobra.Command{
		hcFsCreateCmd, hcFsDeleteCmd, hcFsDescribeCmd,
		hcFsListCmd, hcFsUpdateCmd, hcFsGetIamCmd, hcFsSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagHcFsLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagHcFsDataset, "dataset", "", "Parent dataset (required)")
		_ = c.MarkFlagRequired("dataset")
		c.Flags().StringVar(&flagHcFsFormat, "format", "", "Output format")
	}
	hcFsCreateCmd.Flags().StringVar(&flagHcFsConfigFile, "config-file", "", "YAML/JSON file with the FhirStore body (required)")
	_ = hcFsCreateCmd.MarkFlagRequired("config-file")
	hcFsListCmd.Flags().Int64Var(&flagHcFsPageSize, "page-size", 0, "Maximum results per page")
	hcFsUpdateCmd.Flags().StringVar(&flagHcFsConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = hcFsUpdateCmd.MarkFlagRequired("config-file")
	hcFsUpdateCmd.Flags().StringVar(&flagHcFsUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	healthcareFhirStoresCmd.AddCommand(all...)
	healthcareCmd.AddCommand(healthcareFhirStoresCmd)
}

// hcFhirStoreName returns projects/.../datasets/DATASET/fhirStores/ID.
func hcFhirStoreName(location, dataset, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	ds, err := hcDatasetName(location, dataset)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/fhirStores/%s", ds, id), nil
}

func runHcFsCreate(cmd *cobra.Command, args []string) error {
	parent, err := hcDatasetName(flagHcFsLocation, flagHcFsDataset)
	if err != nil {
		return err
	}
	body := &healthcare.FhirStore{}
	if err := loadYAMLOrJSONInto(flagHcFsConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.FhirStores.Create(parent, body).FhirStoreId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating FHIR store: %w", err)
	}
	fmt.Printf("Created FHIR store [%s].\n", args[0])
	return emitFormatted(got, flagHcFsFormat)
}

func runHcFsDelete(cmd *cobra.Command, args []string) error {
	name, err := hcFhirStoreName(flagHcFsLocation, flagHcFsDataset, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Datasets.FhirStores.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting FHIR store: %w", err)
	}
	fmt.Printf("Deleted FHIR store [%s].\n", args[0])
	return nil
}

func runHcFsDescribe(cmd *cobra.Command, args []string) error {
	name, err := hcFhirStoreName(flagHcFsLocation, flagHcFsDataset, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.FhirStores.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing FHIR store: %w", err)
	}
	return emitFormatted(got, flagHcFsFormat)
}

func runHcFsList(cmd *cobra.Command, args []string) error {
	parent, err := hcDatasetName(flagHcFsLocation, flagHcFsDataset)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*healthcare.FhirStore
	pageToken := ""
	for {
		call := svc.Projects.Locations.Datasets.FhirStores.List(parent).Context(ctx)
		if flagHcFsPageSize > 0 {
			call = call.PageSize(flagHcFsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing FHIR stores: %w", err)
		}
		all = append(all, resp.FhirStores...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagHcFsFormat)
}

func runHcFsUpdate(cmd *cobra.Command, args []string) error {
	name, err := hcFhirStoreName(flagHcFsLocation, flagHcFsDataset, args[0])
	if err != nil {
		return err
	}
	body := &healthcare.FhirStore{}
	if err := loadYAMLOrJSONInto(flagHcFsConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagHcFsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Datasets.FhirStores.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating FHIR store: %w", err)
	}
	fmt.Printf("Updated FHIR store [%s].\n", args[0])
	return emitFormatted(got, flagHcFsFormat)
}

func runHcFsGetIam(cmd *cobra.Command, args []string) error {
	name, err := hcFhirStoreName(flagHcFsLocation, flagHcFsDataset, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Datasets.FhirStores.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagHcFsFormat)
}

func runHcFsSetIam(cmd *cobra.Command, args []string) error {
	name, err := hcFhirStoreName(flagHcFsLocation, flagHcFsDataset, args[0])
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
	updated, err := svc.Projects.Locations.Datasets.FhirStores.SetIamPolicy(name, &healthcare.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for FHIR store [%s].\n", args[0])
	return emitFormatted(updated, flagHcFsFormat)
}
