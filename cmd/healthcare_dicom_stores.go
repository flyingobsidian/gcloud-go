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

// --- gcloud healthcare dicom-stores (#1222) ---

var healthcareDicomStoresCmd = &cobra.Command{Use: "dicom-stores", Short: "Manage Healthcare DICOM stores"}

var (
	flagHcDsxLocation   string
	flagHcDsxDataset    string
	flagHcDsxFormat     string
	flagHcDsxConfigFile string
	flagHcDsxUpdateMask string
	flagHcDsxPageSize   int64
)

var (
	hcDsxCreateCmd = &cobra.Command{
		Use: "create DICOM_STORE", Short: "Create a Healthcare DICOM store",
		Args: cobra.ExactArgs(1), RunE: runHcDsxCreate,
	}
	hcDsxDeleteCmd = &cobra.Command{
		Use: "delete DICOM_STORE", Short: "Delete a Healthcare DICOM store",
		Args: cobra.ExactArgs(1), RunE: runHcDsxDelete,
	}
	hcDsxDescribeCmd = &cobra.Command{
		Use: "describe DICOM_STORE", Short: "Describe a Healthcare DICOM store",
		Args: cobra.ExactArgs(1), RunE: runHcDsxDescribe,
	}
	hcDsxListCmd = &cobra.Command{
		Use: "list", Short: "List Healthcare DICOM stores in a dataset",
		Args: cobra.NoArgs, RunE: runHcDsxList,
	}
	hcDsxUpdateCmd = &cobra.Command{
		Use: "update DICOM_STORE", Short: "Update a Healthcare DICOM store",
		Args: cobra.ExactArgs(1), RunE: runHcDsxUpdate,
	}
	hcDsxGetIamCmd = &cobra.Command{
		Use: "get-iam-policy DICOM_STORE", Short: "Get the IAM policy for a Healthcare DICOM store",
		Args: cobra.ExactArgs(1), RunE: runHcDsxGetIam,
	}
	hcDsxSetIamCmd = &cobra.Command{
		Use: "set-iam-policy DICOM_STORE POLICY_FILE", Short: "Set the IAM policy for a Healthcare DICOM store",
		Args: cobra.ExactArgs(2), RunE: runHcDsxSetIam,
	}
)

func init() {
	all := []*cobra.Command{
		hcDsxCreateCmd, hcDsxDeleteCmd, hcDsxDescribeCmd,
		hcDsxListCmd, hcDsxUpdateCmd, hcDsxGetIamCmd, hcDsxSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagHcDsxLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagHcDsxDataset, "dataset", "", "Parent dataset (required)")
		_ = c.MarkFlagRequired("dataset")
		c.Flags().StringVar(&flagHcDsxFormat, "format", "", "Output format")
	}
	hcDsxCreateCmd.Flags().StringVar(&flagHcDsxConfigFile, "config-file", "", "YAML/JSON file with the DicomStore body (required)")
	_ = hcDsxCreateCmd.MarkFlagRequired("config-file")
	hcDsxListCmd.Flags().Int64Var(&flagHcDsxPageSize, "page-size", 0, "Maximum results per page")
	hcDsxUpdateCmd.Flags().StringVar(&flagHcDsxConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = hcDsxUpdateCmd.MarkFlagRequired("config-file")
	hcDsxUpdateCmd.Flags().StringVar(&flagHcDsxUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	healthcareDicomStoresCmd.AddCommand(all...)
	healthcareCmd.AddCommand(healthcareDicomStoresCmd)
}

// hcDicomStoreName returns projects/.../datasets/DATASET/dicomStores/ID.
func hcDicomStoreName(location, dataset, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	ds, err := hcDatasetName(location, dataset)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/dicomStores/%s", ds, id), nil
}

func runHcDsxCreate(cmd *cobra.Command, args []string) error {
	parent, err := hcDatasetName(flagHcDsxLocation, flagHcDsxDataset)
	if err != nil {
		return err
	}
	body := &healthcare.DicomStore{}
	if err := loadYAMLOrJSONInto(flagHcDsxConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.DicomStores.Create(parent, body).DicomStoreId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating DICOM store: %w", err)
	}
	fmt.Printf("Created DICOM store [%s].\n", args[0])
	return emitFormatted(got, flagHcDsxFormat)
}

func runHcDsxDelete(cmd *cobra.Command, args []string) error {
	name, err := hcDicomStoreName(flagHcDsxLocation, flagHcDsxDataset, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Datasets.DicomStores.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting DICOM store: %w", err)
	}
	fmt.Printf("Deleted DICOM store [%s].\n", args[0])
	return nil
}

func runHcDsxDescribe(cmd *cobra.Command, args []string) error {
	name, err := hcDicomStoreName(flagHcDsxLocation, flagHcDsxDataset, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.DicomStores.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing DICOM store: %w", err)
	}
	return emitFormatted(got, flagHcDsxFormat)
}

func runHcDsxList(cmd *cobra.Command, args []string) error {
	parent, err := hcDatasetName(flagHcDsxLocation, flagHcDsxDataset)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*healthcare.DicomStore
	pageToken := ""
	for {
		call := svc.Projects.Locations.Datasets.DicomStores.List(parent).Context(ctx)
		if flagHcDsxPageSize > 0 {
			call = call.PageSize(flagHcDsxPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing DICOM stores: %w", err)
		}
		all = append(all, resp.DicomStores...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagHcDsxFormat)
}

func runHcDsxUpdate(cmd *cobra.Command, args []string) error {
	name, err := hcDicomStoreName(flagHcDsxLocation, flagHcDsxDataset, args[0])
	if err != nil {
		return err
	}
	body := &healthcare.DicomStore{}
	if err := loadYAMLOrJSONInto(flagHcDsxConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagHcDsxUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Datasets.DicomStores.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating DICOM store: %w", err)
	}
	fmt.Printf("Updated DICOM store [%s].\n", args[0])
	return emitFormatted(got, flagHcDsxFormat)
}

func runHcDsxGetIam(cmd *cobra.Command, args []string) error {
	name, err := hcDicomStoreName(flagHcDsxLocation, flagHcDsxDataset, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Datasets.DicomStores.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagHcDsxFormat)
}

func runHcDsxSetIam(cmd *cobra.Command, args []string) error {
	name, err := hcDicomStoreName(flagHcDsxLocation, flagHcDsxDataset, args[0])
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
	updated, err := svc.Projects.Locations.Datasets.DicomStores.SetIamPolicy(name, &healthcare.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for DICOM store [%s].\n", args[0])
	return emitFormatted(updated, flagHcDsxFormat)
}
