package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	healthcare "google.golang.org/api/healthcare/v1"
)

// --- gcloud healthcare consent-stores (#1220) ---

func (s *healthcareCmds) initHealthcareConsentStores() {
	all := []*cobra.Command{
		s.hcCsCreateCmd, s.hcCsDeleteCmd, s.hcCsDescribeCmd,
		s.hcCsListCmd, s.hcCsUpdateCmd, s.hcCsGetIamCmd, s.hcCsSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagHcCsLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagHcCsDataset, "dataset", "", "Parent dataset (required)")
		_ = c.MarkFlagRequired("dataset")
		c.Flags().StringVar(&s.flagHcCsFormat, "format", "", "Output format")
	}
	s.hcCsCreateCmd.Flags().StringVar(&s.flagHcCsConfigFile, "config-file", "", "YAML/JSON file with the ConsentStore body (required)")
	_ = s.hcCsCreateCmd.MarkFlagRequired("config-file")
	s.hcCsListCmd.Flags().Int64Var(&s.flagHcCsPageSize, "page-size", 0, "Maximum results per page")
	s.hcCsUpdateCmd.Flags().StringVar(&s.flagHcCsConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = s.hcCsUpdateCmd.MarkFlagRequired("config-file")
	s.hcCsUpdateCmd.Flags().StringVar(&s.flagHcCsUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	s.healthcareConsentStoresCmd.AddCommand(all...)
	s.healthcareCmd.AddCommand(s.healthcareConsentStoresCmd)
}

// hcConsentStoreName returns projects/.../datasets/DATASET/consentStores/ID.
func (s *healthcareCmds) hcConsentStoreName(location, dataset, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	ds, err := s.hcDatasetName(location, dataset)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/consentStores/%s", ds, id), nil
}

func (s *healthcareCmds) runHcCsCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.hcDatasetName(s.flagHcCsLocation, s.flagHcCsDataset)
	if err != nil {
		return err
	}
	body := &healthcare.ConsentStore{}
	if err := loadYAMLOrJSONInto(s.flagHcCsConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.ConsentStores.Create(parent, body).ConsentStoreId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating consent store: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created consent store [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagHcCsFormat)
}

func (s *healthcareCmds) runHcCsDelete(cmd *cobra.Command, args []string) error {
	name, err := s.hcConsentStoreName(s.flagHcCsLocation, s.flagHcCsDataset, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Datasets.ConsentStores.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting consent store: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted consent store [%s].\n", args[0])
	return nil
}

func (s *healthcareCmds) runHcCsDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.hcConsentStoreName(s.flagHcCsLocation, s.flagHcCsDataset, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.ConsentStores.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing consent store: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagHcCsFormat)
}

func (s *healthcareCmds) runHcCsList(cmd *cobra.Command, args []string) error {
	parent, err := s.hcDatasetName(s.flagHcCsLocation, s.flagHcCsDataset)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	var all []*healthcare.ConsentStore
	pageToken := ""
	for {
		call := svc.Projects.Locations.Datasets.ConsentStores.List(parent).Context(ctx)
		if s.flagHcCsPageSize > 0 {
			call = call.PageSize(s.flagHcCsPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagHcCsFormat)
}

func (s *healthcareCmds) runHcCsUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.hcConsentStoreName(s.flagHcCsLocation, s.flagHcCsDataset, args[0])
	if err != nil {
		return err
	}
	body := &healthcare.ConsentStore{}
	if err := loadYAMLOrJSONInto(s.flagHcCsConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagHcCsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
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
	fmt.Fprintf(s.inv.stdout, "Updated consent store [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagHcCsFormat)
}

func (s *healthcareCmds) runHcCsGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.hcConsentStoreName(s.flagHcCsLocation, s.flagHcCsDataset, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Datasets.ConsentStores.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagHcCsFormat)
}

func (s *healthcareCmds) runHcCsSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.hcConsentStoreName(s.flagHcCsLocation, s.flagHcCsDataset, args[0])
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
	updated, err := svc.Projects.Locations.Datasets.ConsentStores.SetIamPolicy(name, &healthcare.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for consent store [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, updated, s.flagHcCsFormat)
}
