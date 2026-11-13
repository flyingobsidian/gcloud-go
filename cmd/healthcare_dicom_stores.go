package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	healthcare "google.golang.org/api/healthcare/v1"
)

// --- gcloud healthcare dicom-stores (#1222) ---

func (s *healthcareCmds) initHealthcareDicomStores() {
	all := []*cobra.Command{
		s.hcDsxCreateCmd, s.hcDsxDeleteCmd, s.hcDsxDescribeCmd,
		s.hcDsxListCmd, s.hcDsxUpdateCmd, s.hcDsxGetIamCmd, s.hcDsxSetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagHcDsxLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagHcDsxDataset, "dataset", "", "Parent dataset (required)")
		_ = c.MarkFlagRequired("dataset")
		c.Flags().StringVar(&s.flagHcDsxFormat, "format", "", "Output format")
	}
	s.hcDsxCreateCmd.Flags().StringVar(&s.flagHcDsxConfigFile, "config-file", "", "YAML/JSON file with the DicomStore body (required)")
	_ = s.hcDsxCreateCmd.MarkFlagRequired("config-file")
	s.hcDsxListCmd.Flags().Int64Var(&s.flagHcDsxPageSize, "page-size", 0, "Maximum results per page")
	s.hcDsxUpdateCmd.Flags().StringVar(&s.flagHcDsxConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = s.hcDsxUpdateCmd.MarkFlagRequired("config-file")
	s.hcDsxUpdateCmd.Flags().StringVar(&s.flagHcDsxUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	s.healthcareDicomStoresCmd.AddCommand(all...)
	s.healthcareCmd.AddCommand(s.healthcareDicomStoresCmd)
}

// hcDicomStoreName returns projects/.../datasets/DATASET/dicomStores/ID.
func (s *healthcareCmds) hcDicomStoreName(location, dataset, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	ds, err := s.hcDatasetName(location, dataset)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/dicomStores/%s", ds, id), nil
}

func (s *healthcareCmds) runHcDsxCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.hcDatasetName(s.flagHcDsxLocation, s.flagHcDsxDataset)
	if err != nil {
		return err
	}
	body := &healthcare.DicomStore{}
	if err := loadYAMLOrJSONInto(s.flagHcDsxConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.DicomStores.Create(parent, body).DicomStoreId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating DICOM store: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created DICOM store [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagHcDsxFormat)
}

func (s *healthcareCmds) runHcDsxDelete(cmd *cobra.Command, args []string) error {
	name, err := s.hcDicomStoreName(s.flagHcDsxLocation, s.flagHcDsxDataset, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Datasets.DicomStores.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting DICOM store: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted DICOM store [%s].\n", args[0])
	return nil
}

func (s *healthcareCmds) runHcDsxDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.hcDicomStoreName(s.flagHcDsxLocation, s.flagHcDsxDataset, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.DicomStores.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing DICOM store: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagHcDsxFormat)
}

func (s *healthcareCmds) runHcDsxList(cmd *cobra.Command, args []string) error {
	parent, err := s.hcDatasetName(s.flagHcDsxLocation, s.flagHcDsxDataset)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	var all []*healthcare.DicomStore
	pageToken := ""
	for {
		call := svc.Projects.Locations.Datasets.DicomStores.List(parent).Context(ctx)
		if s.flagHcDsxPageSize > 0 {
			call = call.PageSize(s.flagHcDsxPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagHcDsxFormat)
}

func (s *healthcareCmds) runHcDsxUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.hcDicomStoreName(s.flagHcDsxLocation, s.flagHcDsxDataset, args[0])
	if err != nil {
		return err
	}
	body := &healthcare.DicomStore{}
	if err := loadYAMLOrJSONInto(s.flagHcDsxConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagHcDsxUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
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
	fmt.Fprintf(s.inv.stdout, "Updated DICOM store [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagHcDsxFormat)
}

func (s *healthcareCmds) runHcDsxGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.hcDicomStoreName(s.flagHcDsxLocation, s.flagHcDsxDataset, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Datasets.DicomStores.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagHcDsxFormat)
}

func (s *healthcareCmds) runHcDsxSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.hcDicomStoreName(s.flagHcDsxLocation, s.flagHcDsxDataset, args[0])
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
	updated, err := svc.Projects.Locations.Datasets.DicomStores.SetIamPolicy(name, &healthcare.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for DICOM store [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, updated, s.flagHcDsxFormat)
}
