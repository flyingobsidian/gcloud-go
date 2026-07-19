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

// --- gcloud healthcare consent-stores (#1220) ---

var healthcareConsentStoresCmd = &cobra.Command{Use: "consent-stores", Short: "Manage Healthcare consent stores"}

var (
	flagHcCsLocation   string
	flagHcCsDataset    string
	flagHcCsFormat     string
	flagHcCsConfigFile string
	flagHcCsUpdateMask string
	flagHcCsPageSize   int64
)

var (
	hcCsCreateCmd = &cobra.Command{
		Use: "create CONSENT_STORE", Short: "Create a Healthcare consent store",
		Args: cobra.ExactArgs(1), RunE: runHcCsCreate,
	}
	hcCsDeleteCmd = &cobra.Command{
		Use: "delete CONSENT_STORE", Short: "Delete a Healthcare consent store",
		Args: cobra.ExactArgs(1), RunE: runHcCsDelete,
	}
	hcCsDescribeCmd = &cobra.Command{
		Use: "describe CONSENT_STORE", Short: "Describe a Healthcare consent store",
		Args: cobra.ExactArgs(1), RunE: runHcCsDescribe,
	}
	hcCsListCmd = &cobra.Command{
		Use: "list", Short: "List Healthcare consent stores in a dataset",
		Args: cobra.NoArgs, RunE: runHcCsList,
	}
	hcCsUpdateCmd = &cobra.Command{
		Use: "update CONSENT_STORE", Short: "Update a Healthcare consent store",
		Args: cobra.ExactArgs(1), RunE: runHcCsUpdate,
	}
	hcCsGetIamCmd = &cobra.Command{
		Use: "get-iam-policy CONSENT_STORE", Short: "Get the IAM policy for a Healthcare consent store",
		Args: cobra.ExactArgs(1), RunE: runHcCsGetIam,
	}
	hcCsSetIamCmd = &cobra.Command{
		Use: "set-iam-policy CONSENT_STORE POLICY_FILE", Short: "Set the IAM policy for a Healthcare consent store",
		Args: cobra.ExactArgs(2), RunE: runHcCsSetIam,
	}
)

func init() {
	all := []*cobra.Command{
		hcCsCreateCmd, hcCsDeleteCmd, hcCsDescribeCmd,
		hcCsListCmd, hcCsUpdateCmd, hcCsGetIamCmd, hcCsSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagHcCsLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagHcCsDataset, "dataset", "", "Parent dataset (required)")
		_ = c.MarkFlagRequired("dataset")
		c.Flags().StringVar(&flagHcCsFormat, "format", "", "Output format")
	}
	hcCsCreateCmd.Flags().StringVar(&flagHcCsConfigFile, "config-file", "", "YAML/JSON file with the ConsentStore body (required)")
	_ = hcCsCreateCmd.MarkFlagRequired("config-file")
	hcCsListCmd.Flags().Int64Var(&flagHcCsPageSize, "page-size", 0, "Maximum results per page")
	hcCsUpdateCmd.Flags().StringVar(&flagHcCsConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = hcCsUpdateCmd.MarkFlagRequired("config-file")
	hcCsUpdateCmd.Flags().StringVar(&flagHcCsUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	healthcareConsentStoresCmd.AddCommand(all...)
	healthcareCmd.AddCommand(healthcareConsentStoresCmd)
}

// hcConsentStoreName returns projects/.../datasets/DATASET/consentStores/ID.
func hcConsentStoreName(location, dataset, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	ds, err := hcDatasetName(location, dataset)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/consentStores/%s", ds, id), nil
}

func runHcCsCreate(cmd *cobra.Command, args []string) error {
	parent, err := hcDatasetName(flagHcCsLocation, flagHcCsDataset)
	if err != nil {
		return err
	}
	body := &healthcare.ConsentStore{}
	if err := loadYAMLOrJSONInto(flagHcCsConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.ConsentStores.Create(parent, body).ConsentStoreId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating consent store: %w", err)
	}
	fmt.Printf("Created consent store [%s].\n", args[0])
	return emitFormatted(got, flagHcCsFormat)
}

func runHcCsDelete(cmd *cobra.Command, args []string) error {
	name, err := hcConsentStoreName(flagHcCsLocation, flagHcCsDataset, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Datasets.ConsentStores.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting consent store: %w", err)
	}
	fmt.Printf("Deleted consent store [%s].\n", args[0])
	return nil
}

func runHcCsDescribe(cmd *cobra.Command, args []string) error {
	name, err := hcConsentStoreName(flagHcCsLocation, flagHcCsDataset, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.ConsentStores.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing consent store: %w", err)
	}
	return emitFormatted(got, flagHcCsFormat)
}

func runHcCsList(cmd *cobra.Command, args []string) error {
	parent, err := hcDatasetName(flagHcCsLocation, flagHcCsDataset)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*healthcare.ConsentStore
	pageToken := ""
	for {
		call := svc.Projects.Locations.Datasets.ConsentStores.List(parent).Context(ctx)
		if flagHcCsPageSize > 0 {
			call = call.PageSize(flagHcCsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing consent stores: %w", err)
		}
		all = append(all, resp.ConsentStores...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagHcCsFormat)
}

func runHcCsUpdate(cmd *cobra.Command, args []string) error {
	name, err := hcConsentStoreName(flagHcCsLocation, flagHcCsDataset, args[0])
	if err != nil {
		return err
	}
	body := &healthcare.ConsentStore{}
	if err := loadYAMLOrJSONInto(flagHcCsConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagHcCsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Datasets.ConsentStores.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating consent store: %w", err)
	}
	fmt.Printf("Updated consent store [%s].\n", args[0])
	return emitFormatted(got, flagHcCsFormat)
}

func runHcCsGetIam(cmd *cobra.Command, args []string) error {
	name, err := hcConsentStoreName(flagHcCsLocation, flagHcCsDataset, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.HealthcareService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Datasets.ConsentStores.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagHcCsFormat)
}

func runHcCsSetIam(cmd *cobra.Command, args []string) error {
	name, err := hcConsentStoreName(flagHcCsLocation, flagHcCsDataset, args[0])
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
	updated, err := svc.Projects.Locations.Datasets.ConsentStores.SetIamPolicy(name, &healthcare.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for consent store [%s].\n", args[0])
	return emitFormatted(updated, flagHcCsFormat)
}
