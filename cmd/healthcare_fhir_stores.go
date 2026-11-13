package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	healthcare "google.golang.org/api/healthcare/v1"
)

// --- gcloud healthcare fhir-stores (#1223) ---

func (s *healthcareCmds) initHealthcareFhirStores() {
	all := []*cobra.Command{
		s.hcFsCreateCmd, s.hcFsDeleteCmd, s.hcFsDescribeCmd,
		s.hcFsListCmd, s.hcFsUpdateCmd, s.hcFsGetIamCmd, s.hcFsSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagHcFsLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagHcFsDataset, "dataset", "", "Parent dataset (required)")
		_ = c.MarkFlagRequired("dataset")
		c.Flags().StringVar(&s.flagHcFsFormat, "format", "", "Output format")
	}
	s.hcFsCreateCmd.Flags().StringVar(&s.flagHcFsConfigFile, "config-file", "", "YAML/JSON file with the FhirStore body (required)")
	_ = s.hcFsCreateCmd.MarkFlagRequired("config-file")
	s.hcFsListCmd.Flags().Int64Var(&s.flagHcFsPageSize, "page-size", 0, "Maximum results per page")
	s.hcFsUpdateCmd.Flags().StringVar(&s.flagHcFsConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = s.hcFsUpdateCmd.MarkFlagRequired("config-file")
	s.hcFsUpdateCmd.Flags().StringVar(&s.flagHcFsUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	s.healthcareFhirStoresCmd.AddCommand(all...)
	s.healthcareCmd.AddCommand(s.healthcareFhirStoresCmd)
}

// hcFhirStoreName returns projects/.../datasets/DATASET/fhirStores/ID.
func (s *healthcareCmds) hcFhirStoreName(location, dataset, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	ds, err := s.hcDatasetName(location, dataset)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/fhirStores/%s", ds, id), nil
}

func (s *healthcareCmds) runHcFsCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.hcDatasetName(s.flagHcFsLocation, s.flagHcFsDataset)
	if err != nil {
		return err
	}
	body := &healthcare.FhirStore{}
	if err := loadYAMLOrJSONInto(s.flagHcFsConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.FhirStores.Create(parent, body).FhirStoreId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating FHIR store: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created FHIR store [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagHcFsFormat)
}

func (s *healthcareCmds) runHcFsDelete(cmd *cobra.Command, args []string) error {
	name, err := s.hcFhirStoreName(s.flagHcFsLocation, s.flagHcFsDataset, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Datasets.FhirStores.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting FHIR store: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted FHIR store [%s].\n", args[0])
	return nil
}

func (s *healthcareCmds) runHcFsDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.hcFhirStoreName(s.flagHcFsLocation, s.flagHcFsDataset, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.FhirStores.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing FHIR store: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagHcFsFormat)
}

func (s *healthcareCmds) runHcFsList(cmd *cobra.Command, args []string) error {
	parent, err := s.hcDatasetName(s.flagHcFsLocation, s.flagHcFsDataset)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	var all []*healthcare.FhirStore
	pageToken := ""
	for {
		call := svc.Projects.Locations.Datasets.FhirStores.List(parent).Context(ctx)
		if s.flagHcFsPageSize > 0 {
			call = call.PageSize(s.flagHcFsPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagHcFsFormat)
}

func (s *healthcareCmds) runHcFsUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.hcFhirStoreName(s.flagHcFsLocation, s.flagHcFsDataset, args[0])
	if err != nil {
		return err
	}
	body := &healthcare.FhirStore{}
	if err := loadYAMLOrJSONInto(s.flagHcFsConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagHcFsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
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
	fmt.Fprintf(s.inv.stdout, "Updated FHIR store [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagHcFsFormat)
}

func (s *healthcareCmds) runHcFsGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.hcFhirStoreName(s.flagHcFsLocation, s.flagHcFsDataset, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Datasets.FhirStores.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagHcFsFormat)
}

func (s *healthcareCmds) runHcFsSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.hcFhirStoreName(s.flagHcFsLocation, s.flagHcFsDataset, args[0])
	if err != nil {
		return err
	}
	policy := &healthcare.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.Datasets.FhirStores.SetIamPolicy(name, &healthcare.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for FHIR store [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, updated, s.flagHcFsFormat)
}
